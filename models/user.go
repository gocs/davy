package models

import (
	"fmt"

	"github.com/gocs/davy/validator"
	"golang.org/x/crypto/bcrypt"
)

// User is a manager for accessing users in the database
type User struct {
	id int64
}

// NewUser create a new user, saves it to the database, and returns the newly created user
func NewUser(username string, hash []byte) (*User, error) {
	exists, err := client.HExists("user:by-username", username).Result()
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUsernameTaken
	}

	err = validator.Username(username)
	if err != nil {
		return nil, err
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
	pipe.HSet(key, "lobby", -1)
	pipe.HSet("user:by-username", username, id)
	_, err = pipe.Exec()
	if err != nil {
		return nil, err
	}

	return &User{id: id}, nil
}

// GetUserID UserID getter
func (u *User) GetUserID() int64 { return u.id }

// GetUsername Username getter
func (u *User) GetUsername() (string, error) {
	key := fmt.Sprintf("user:%d", u.id)
	return client.HGet(key, "username").Result()
}

// GetHash Hash getter
func (u *User) GetHash() ([]byte, error) {
	key := fmt.Sprintf("user:%d", u.id)
	return client.HGet(key, "hash").Bytes()
}

// GetLobby Lobby getter
func (u *User) GetLobby() (*Lobby, error) {
	key := fmt.Sprintf("user:%d", u.id)

	id, err := client.HGet(key, "lobby").Int64()
	if err != nil {
		return nil, err
	}
	if id == -1 {
		return nil, ErrUserNotInLobby
	}
	return &Lobby{id: id}, nil
}

// Authenticate will validates the login attempt
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

// RegisterUser register a valid user
func RegisterUser(username, password string) error {
	cost := bcrypt.DefaultCost
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return err
	}

	u, err := NewUser(username, hash)
	if err != nil {
		return err
	}

	_, err = newUserQuestion(u.id)
	return err
}

// IsUser checks if username has registered in this site
func IsUser(userID int64) (bool, error) {
	u := &User{id: userID}
	un, err := u.GetUsername()
	if err != nil || un == "" {
		return false, err
	}
	return true, nil
}

// GetUserByUserID gets user using a user id
func GetUserByUserID(userID int64) (*User, error) {
	return &User{id: userID}, nil
}

// GetUserIDByUser gets the user id using the user
func GetUserIDByUser(user *User) int64 {
	return user.GetUserID()
}

// GetUserByUsername gets the user using the username
func GetUserByUsername(username string) (*User, error) {
	id, err := client.HGet("user:by-username", username).Int64()
	if err == redisNil {
		return nil, ErrUserNotFound
	} else if err != nil {
		return nil, err
	}

	return &User{id: id}, nil
}

// AuthenticateUser authenticates the user by its username and password
func AuthenticateUser(username, password string) (*User, error) {
	user, err := GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	if err := user.Authenticate(password); err != nil {
		return nil, err
	}
	return user, nil
}
