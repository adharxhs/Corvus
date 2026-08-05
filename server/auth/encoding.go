package auth

import (
	"encoding/base64"
)

func encodeB64(data []byte) string {
	return base64.RawStdEncoding.EncodeToString(data)
}

func decodeB64(s string) ([]byte, error) {
	return base64.RawStdEncoding.DecodeString(s)
}
