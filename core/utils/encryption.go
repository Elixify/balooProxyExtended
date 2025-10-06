package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"math/rand"
	"strings"
	"sync"

	"github.com/zeebo/blake3"
)

var (
	sha256Pool = sync.Pool{
		New: func() interface{} {
			return sha256.New()
		},
	}
)

/*
func Encrypt(input string, key string) string {
	hasher := md5.New()
	hasher.Write([]byte(input + key))
	return hex.EncodeToString(hasher.Sum(nil))
}
*/

func Encrypt(input string, key string) string {
	// Optimize string concatenation
	var sb strings.Builder
	sb.Grow(len(input) + len(key))
	sb.WriteString(input)
	sb.WriteString(key)
	
	hash := blake3.Sum256([]byte(sb.String()))
	return hex.EncodeToString(hash[:])
}

func EncryptSha(input string, key string) string {
	// Use pool to reduce allocations
	hasher := sha256Pool.Get().(hash.Hash)
	defer func() {
		hasher.Reset()
		sha256Pool.Put(hasher)
	}()
	
	// Optimize string concatenation
	var sb strings.Builder
	sb.Grow(len(input) + len(key))
	sb.WriteString(input)
	sb.WriteString(key)
	
	hasher.Write([]byte(sb.String()))
	return hex.EncodeToString(hasher.Sum(nil))
}

func RandomString(length int) string {
	var rnd = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	res := make([]rune, length)
	for i := range res {
		res[i] = rnd[rand.Intn(len(rnd))]
	}
	return string(res)
}

func HashToInt(hash string) int {
	subset := (uint16(hash[0]) << 8) | uint16(hash[1])
	return int(subset)%15 + 1
}
