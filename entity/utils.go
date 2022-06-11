package entity

import (
	"math/rand"
	"regexp"
	"time"
)

var (
	ValidUsernameRegex = regexp.MustCompilePOSIX("^[a-zA-Z0-9_]*$")
	InvalidCharsRegex  = regexp.MustCompilePOSIX("([[:cntrl:]]|[[:space:]])+")
	EmailRegex         = regexp.MustCompile("^.+@.+\\..+$")
)

func GenerateUsername() string {
	const usernameAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 10)
	for i := range b {
		b[i] = usernameAlphabet[rand.Intn(len(usernameAlphabet))]
	}
	return string(b)
}

func String2Bool(str string) bool {
	return str == "true" || str == "1"
}

func MaxIn64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func RandomInt(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min+1) + min
}
