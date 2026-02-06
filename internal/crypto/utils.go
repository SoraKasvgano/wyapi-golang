package crypto

import (
	"crypto/md5"
	"encoding/base64"
)

func MD5Bytes(data []byte) []byte {
	sum := md5.Sum(data)
	return sum[:]
}

func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
