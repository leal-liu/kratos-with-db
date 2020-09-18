package utils

import (
	"crypto/sha256"
	"hash"
	"sync"
)

var (
	sha256Pool = sync.Pool{
		New: func() interface{} {
			return sha256.New()
		},
	}
)

// SHA256 sha256
func SHA256(input []byte) (output []byte) {
	h := sha256Pool.Get().(hash.Hash)
	h.Write(input)
	output = h.Sum(nil)
	h.Reset()
	sha256Pool.Put(h)
	return
}
