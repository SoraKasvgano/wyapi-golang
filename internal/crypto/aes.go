package crypto

import (
	"crypto/aes"
	"errors"
)

func PKCS7Pad(data []byte, blockSize int) []byte {
	if blockSize <= 0 {
		return data
	}
	padding := blockSize - (len(data) % blockSize)
	if padding == 0 {
		padding = blockSize
	}
	padText := make([]byte, padding)
	for i := range padText {
		padText[i] = byte(padding)
	}
	return append(data, padText...)
}

func EncryptECB(key, data []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, errors.New("empty key")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	data = PKCS7Pad(data, blockSize)
	if len(data)%blockSize != 0 {
		return nil, errors.New("data is not a multiple of block size")
	}

	encrypted := make([]byte, len(data))
	for start := 0; start < len(data); start += blockSize {
		block.Encrypt(encrypted[start:start+blockSize], data[start:start+blockSize])
	}
	return encrypted, nil
}
