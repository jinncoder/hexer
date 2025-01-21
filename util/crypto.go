package util

import (
	"crypto/sha256"
	"encoding/hex"
)

func SHA256SUM(data string) string {
	hash := sha256.New()
	hash.Write([]byte(data))
	sum := hash.Sum(nil)

	return hex.EncodeToString(sum)
}
