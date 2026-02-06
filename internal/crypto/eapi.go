package crypto

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"net/url"
	"strings"
)

const (
	EAPIKey = "e82ckenh8dichen8"
)

func EncryptEAPIParams(rawURL string, payloadJSON []byte) (string, error) {
	if rawURL == "" {
		return "", errors.New("empty url")
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	path := strings.Replace(parsed.Path, "/eapi/", "/api/", 1)
	payload := string(payloadJSON)
	digest := md5Hex("nobody" + path + "use" + payload + "md5forencrypt")
	params := path + "-36cd479b6b5-" + payload + "-36cd479b6b5-" + digest

	encrypted, err := EncryptECB([]byte(EAPIKey), []byte(params))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(encrypted), nil
}

func md5Hex(text string) string {
	sum := md5.Sum([]byte(text))
	return hex.EncodeToString(sum[:])
}
