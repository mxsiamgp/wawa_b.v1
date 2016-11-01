package sha1

import (
	"crypto/sha1"
	"encoding/hex"
)

func StringDigest(orig []byte) string {
	hash := sha1.New()
	hash.Write(orig)
	return hex.EncodeToString(hash.Sum(nil))
}
