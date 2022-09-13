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

func IntToBool(num int) bool {
	return num != 0
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

func RangeWeek(t time.Time) (time.Time, time.Time) {
	x := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	for x.Weekday() != time.Monday {
		x = x.AddDate(0, 0, -1)
	}
	return x, x.AddDate(0, 0, 7).Add(-1 * time.Second)
}

func RangeThisWeek() (time.Time, time.Time) {
	now := time.Now()
	return RangeWeek(now)
}

func RangeLastWeek() (time.Time, time.Time) {
	endLastWeek, _ := RangeWeek(time.Now())
	beginWeek := endLastWeek.AddDate(0, 0, -7)
	endLastWeek = endLastWeek.Add(-1 * time.Second)

	return beginWeek, endLastWeek
}

func RangeMonth(t time.Time) (time.Time, time.Time) {
	beginMonth := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	endMonth := beginMonth.AddDate(0, 1, 0).Add(-1 * time.Second)
	return beginMonth, endMonth
}

func RangeThisMonth() (time.Time, time.Time) {
	return RangeMonth(time.Now())
}

func RangeLastMonth() (time.Time, time.Time) {
	endLastMonth, _ := RangeMonth(time.Now())
	endLastMonth = endLastMonth.Add(-1 * time.Second)
	beginLastMonth := time.Date(endLastMonth.Year(), endLastMonth.Month(), 1,
		0, 0, 0, 0, endLastMonth.Location())
	return beginLastMonth, endLastMonth
}

func AbsInt64(num int64) int64 {
	if num > 0 {
		return num
	}
	return -num
}