package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

func CreateHash(plain string) string {
	ha := sha256.New()
	ha.Write([]byte(plain))
	return hex.EncodeToString(ha.Sum(nil))
}
