package toolkit

import (
	"crypto/rand"
	"encoding/base32"
	"log"
)

const randomStringSource = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_+"

// Tools is the type that allows access to its various utility methods.
type Tools struct {
}

// RandomString generates a base32 random string of n size.
func RandomString(n int) string {
	bb := make([]byte, n)
	_, err := rand.Read(bb)
	if err != nil {
		log.Println(err)
		return ""
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bb)[:n]
}

// RandomString generates a random string from randomStringSource of n size.
func (t *Tools) RandomString(n int) string {
	dst, src := make([]rune, n), []rune(randomStringSource)
	for i := range dst {
		p, _ := rand.Prime(rand.Reader, len(src))
		x, y := p.Uint64(), uint64(len(src))
		//fmt.Println("x, y, x%y:", x, y, x%y)
		dst[i] = src[x%y]
	}
	return string(dst)
}
