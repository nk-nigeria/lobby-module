package entity

import (
	"math/rand"
	"regexp"
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
