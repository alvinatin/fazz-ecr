package randstring

import (
	"crypto/rand"
	"encoding/base64"
)

func Get(size int) string {
	data := make([]byte, size)
	rand.Read(data)
	return base64.RawURLEncoding.EncodeToString(data)[:size]
}
