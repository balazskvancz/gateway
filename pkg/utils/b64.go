package utils

import (
	"encoding/base64"
	"fmt"
)

func EncodeB64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func DecodeB64(b64 string) string {
	decoded, err := base64.StdEncoding.DecodeString(b64)

	if err != nil {
		fmt.Printf("[UTILS]: b64 decode error: %v\nInput string: %s\n", err, b64)

		return ""
	}

	return string(decoded)
}
