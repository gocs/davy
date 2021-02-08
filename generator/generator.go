package generator

import (
	"fmt"
	"math/rand"
	"time"
)


func init() {
	rand.Seed(time.Now().UnixNano())
}

// Code this will generate a alphanumeric with dash and underscore code
func Code(length int) string {
	result := ""
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	charsLen := len(chars)
	for i := 0; i < length; i++ {
		result = fmt.Sprintf("%s%c", result, chars[rand.Intn(charsLen)])
	}
	return result
}