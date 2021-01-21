package models

import (
	"errors"

	"github.com/go-redis/redis"
)

// NewRedisDB instantiates a package-level redis client access
func NewRedisDB() {
	client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	}) // fuuuuck
}

var (
	redisNil = redis.Nil

	// ErrUserNotFound common error on login form when the user is not found
	ErrUserNotFound = errors.New("user not found")

	// ErrInvalidLogin common error on login form when the user does not match its login credentials
	ErrInvalidLogin = errors.New("invalid login")

	// ErrUsernameTaken common error on registration form when the username already existed
	ErrUsernameTaken = errors.New("username taken")

	client *redis.Client // fuuuuck// fuuuuck// fuuuuck// fuuuuck
)
