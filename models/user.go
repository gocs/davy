package models

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	id int64
}

func NewUser(username string, hash []byte) (*User, error) {
	exists, err := client.HExists("user:by-username", username).Result()
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUsernameTaken
	}

	id, err := client.Incr("user:next-id").Result()
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("user:%d", id)
	pipe := client.Pipeline()
	pipe.HSet(key, "id", id)
	pipe.HSet(key, "username", username)
	pipe.HSet(key, "hash", hash)
	pipe.HSet("user:by-username", username, id)
	_, err = pipe.Exec()
	if err != nil {
		return nil, err
	}

	return &User{id: id}, nil
}

func (u *User) GetUserID() (int64, error) {
	return u.id, nil
}

func (u *User) GetUsername() (string, error) {
	key := fmt.Sprintf("user:%d", u.id)
	return client.HGet(key, "username").Result()
}

func (u *User) GetHash() ([]byte, error) {
	key := fmt.Sprintf("user:%d", u.id)
	return client.HGet(key, "hash").Bytes()
}

func (u *User) Authenticate(password string) error {
	hash, err := u.GetHash()
	if err != nil {
		return err
	}
	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return ErrInvalidLogin
	}
	return err
}

func RegisterUser(username, password string) error {
	cost := bcrypt.DefaultCost
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return err
	}

	_, err = NewUser(username, hash)
	return err
}

func GetUserByUserID(userID int64) (*User, error) {
	return &User{id: userID}, nil
}

func GetUserIDByUser(user *User) (int64, error) {
	return user.GetUserID()
}

func GetUserByUsername(username string) (*User, error) {
	id, err := client.HGet("user:by-username", username).Int64()
	if err == redisNil {
		return nil, ErrUserNotFound
	} else if err != nil {
		return nil, err
	}

	u, err := GetUserByUserID(id)
	return u, err
}

func AuthenticateUser(username, password string) (*User, error) {
	user, err := GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	return user, user.Authenticate(password)
}
