package filehost

import (
	"math/rand"
	"strings"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

// RandString generates a random string with length n
func RandString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func whiteListed(allowed []string, input string) bool {
	spl := strings.Split(input, "/")
	if len(spl) < 2 {
		return false
	}
	s1, s2 := spl[0], spl[1]
	for _, a := range allowed {
		if input == a {
			return true
		}
		spl := strings.Split(a, "/")
		if len(spl) < 2 {
			log.Error("invalid mime type ", a)
			continue
		}
		passed := 0
		if spl[0] == "*" || spl[0] == s1 {
			passed++
		}
		if spl[1] == "*" || spl[1] == s2 {
			passed++
		}
		if passed > 1 {
			return true
		}
	}
	return false
}
