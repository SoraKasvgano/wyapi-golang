package api

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"time"
)

type CryptoSessionStore struct {
	mu       sync.Mutex
	sessions map[string]cryptoSession
	ttl      time.Duration
}

type cryptoSession struct {
	keyToken string
	key      []byte
	expires  time.Time
}

type PublicCryptoSession struct {
	KeyID    string `json:"keyId"`
	KeyToken string `json:"keyToken"`
	Key      string `json:"key"`
}

func NewCryptoSessionStore(ttl time.Duration) *CryptoSessionStore {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return &CryptoSessionStore{
		sessions: map[string]cryptoSession{},
		ttl:      ttl,
	}
}

func (s *CryptoSessionStore) Create() (*PublicCryptoSession, error) {
	if s == nil {
		return nil, errors.New("crypto session store is nil")
	}

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	keyID, err := randomHex(16)
	if err != nil {
		return nil, err
	}
	keyToken, err := randomHex(16)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupLocked(now)
	s.sessions[keyID] = cryptoSession{
		keyToken: keyToken,
		key:      key,
		expires:  now.Add(s.ttl),
	}

	return &PublicCryptoSession{
		KeyID:    keyID,
		KeyToken: keyToken,
		Key:      base64.StdEncoding.EncodeToString(key),
	}, nil
}

func (s *CryptoSessionStore) Lookup(keyID, keyToken string) ([]byte, bool) {
	if s == nil || keyID == "" || keyToken == "" {
		return nil, false
	}

	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupLocked(now)

	session, ok := s.sessions[keyID]
	if !ok || session.keyToken != keyToken || now.After(session.expires) {
		return nil, false
	}

	key := make([]byte, len(session.key))
	copy(key, session.key)
	return key, true
}

func (s *CryptoSessionStore) cleanupLocked(now time.Time) {
	for keyID, session := range s.sessions {
		if now.After(session.expires) {
			delete(s.sessions, keyID)
		}
	}
}

func randomHex(byteLen int) (string, error) {
	buf := make([]byte, byteLen)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func decryptCompatPayload(ciphertext string, key []byte) ([]byte, error) {
	parts := strings.Split(ciphertext, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid encrypted payload")
	}

	iv, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}
	tag, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	data, err := base64.StdEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	combined := make([]byte, 0, len(data)+len(tag))
	combined = append(combined, data...)
	combined = append(combined, tag...)
	return gcm.Open(nil, iv, combined, nil)
}
