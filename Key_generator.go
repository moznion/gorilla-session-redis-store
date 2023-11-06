package redistore

import (
	"crypto/rand"
	"encoding/base32"
	"io"
	"strings"
)

type KeyGenerator interface {
	GenerateKey() (string, error)
}

type DefaultRandomKeyGenerator struct {
}

func (g *DefaultRandomKeyGenerator) GenerateKey() (string, error) {
	return strings.TrimRight(base32.StdEncoding.EncodeToString(g.generateRandomKey(32)), "="), nil
}

func (g *DefaultRandomKeyGenerator) generateRandomKey(length uint64) []byte {
	// imported from gorilla/securecookie: https://github.com/gorilla/securecookie/blob/eae3c1840ec4adda88a4af683ad0f60bb690e7c2/securecookie.go#L515
	k := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		return nil
	}
	return k
}
