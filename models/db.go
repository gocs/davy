package models

import (
	"errors"

	"github.com/go-redis/redis"
)

func NewRedisDB() {
	client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	}) // fuuuuck
}

var (
	redisNil = redis.Nil

	ErrUserNotFound = errors.New("user not found")
	ErrInvalidLogin = errors.New("invalid login")
	ErrUsernameTaken = errors.New("username taken")

	client *redis.Client // fuuuuck// fuuuuck// fuuuuck// fuuuuck
)
