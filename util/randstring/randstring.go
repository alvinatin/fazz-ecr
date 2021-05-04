package randstring

import (
	"crypto/rand"
	"encoding/base64"
)

func Get(size int) string {
	data := make([]byte, size)
	for {
		if n, _ := rand.Read(data); n == size {
			break
		}
	}
	return base64.RawURLEncoding.EncodeToString(data)[:size]
}
