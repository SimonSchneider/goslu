package sid

import (
	"crypto/rand"
	"math/big"
	"slices"
	"sync"
)

var bufPool = &sync.Pool{New: func() any {
	return make([]byte, 0)
}}

const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
const lettersLen = int64(len(letters))

var lettersBigLen = big.NewInt(lettersLen)

func bufWithLen(n int) []byte {
	buf := bufPool.Get().([]byte)
	if cap(buf) < n {
		buf = slices.Grow(buf, n-len(buf))
	}
	return buf[:n]
}

func NewString(n int) (string, error) {
	buf := bufWithLen(n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, lettersBigLen)
		if err != nil {
			return "", err
		}
		buf[i] = letters[num.Int64()]
	}
	ret := string(buf)
	bufPool.Put(buf)
	return ret, nil
}

func MustNewString(n int) string {
	str, err := NewString(n)
	if err != nil {
		panic(err)
	}
	return str
}
