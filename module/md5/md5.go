package md5

import (
	"crypto/md5"
	"encoding/hex"
)

func StringDigest(orig []byte) string {
	hash := md5.New()
	hash.Write(orig)
	return hex.EncodeToString(hash.Sum(nil))
}
