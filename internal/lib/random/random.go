package random

import (
	"math/rand"
	"time"
)

func GenerateRandomString(length int) string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	var runes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

	b := make([]rune, length)
	for i := range b {
		b[i] = runes[rnd.Intn(len(runes))]
	}
	return string(b)
}
