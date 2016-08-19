package filehost

import (
	"math/rand"
	"mime"
	"strings"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func randString(n int) string {
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

func extFromMime(s string) string {
	spl := strings.Split(s, "/")
	if len(spl) > 1 {
		// mime lib doesnt support audio for some reason
		if spl[0] == "audio" {
			switch spl[1] {
			case "mp3", "ogg", "flac", "wav", "mpeg3":
				if cfg.isBlocked(spl[1]) {
					return ""
				}
				return spl[1]
			default:
				return ""
			}
		}
	}
	exts, err := mime.ExtensionsByType(s)
	log.Debug(exts)
	if err != nil || exts == nil {
		return ""
	}
	var m string
	var blocked bool
	for _, x := range exts {
		m = x[1:]
		if !cfg.isBlocked(m) {
			blocked = false
			break
		}
	}
	if !blocked {
		return m
	}
	return ""
}

func validMimeType(s string) bool {
	spl := strings.Split(s, "/")
	if len(spl) < 2 {
		return false
	}
	givenExt := spl[1]
	exts, err := mime.ExtensionsByType(s)
	if err != nil {
		log.Error(err)
		return false
	}
	if exts == nil {
		return false
	}
	for _, x := range exts {
		if x[1:] == givenExt {
			return true
		}
	}
	return false
}
