package models

import (
	"fmt"
	"strconv"

	"github.com/gocs/davy/generator"
)

// Lobby is a manager for accessing users in the database
type Lobby struct {
	id int64
}

// NewLobby creates new lobby, saves it to the database, and returns the newly created lobby
func NewLobby(hostID int64, length int) (*Lobby, error) {
	lid, err := client.HGet(fmt.Sprintf("user:%d", hostID), "lobby").Int64()
	if err != nil {
		return nil, err
	}
	if lid != -1 {
		return nil, ErrUserInLobby
	}

	id, err := client.Incr("lobby:next-id").Result()
	if err != nil {
		return nil, err
	}

	code := generator.Code(length)

	key := fmt.Sprintf("lobby:%d", id)
	pipe := client.Pipeline()
	pipe.HSet(key, "id", id)
	pipe.HSet(key, "code", code)
	pipe.HSet(key, "host_id", hostID)
	pipe.HSet(key, "status", StatusWaiting)
	pipe.HSet("lobby:by-code", code, id)
	pipe.HSet(fmt.Sprintf("user:%d", hostID), "lobby", id)
	pipe.SAdd(fmt.Sprintf("lobby:%d:members", id), hostID)
	_, err = pipe.Exec()
	if err != nil {
		return nil, err
	}

	return &Lobby{id: id}, nil
}

const (
	// StatusWaiting means the status is "waiting"
	StatusWaiting  = 0
	// StatusStarting means the status is "starting"
	StatusStarting = 1
	// StatusOngoing means the status is "on-going"
	StatusOngoing  = 2
	// StatusEnded means the status is "ended"
	StatusEnded    = 3
)

// GetLobbyID LobbyID getter
func (l *Lobby) GetLobbyID() (int64, error) {
	key := fmt.Sprintf("lobby:%d", l.id)
	return client.HGet(key, "id").Int64()
}

// GetCode Lobby Code getter
func (l *Lobby) GetCode() (string, error) {
	key := fmt.Sprintf("lobby:%d", l.id)
	code, err := client.HGet(key, "code").Result()
	if err != nil {
		return "", err
	}
	return code, err
}

// GetHostID Lobby HostID getter
func (l *Lobby) GetHostID() (int64, error) {
	key := fmt.Sprintf("lobby:%d", l.id)
	return client.HGet(key, "host_id").Int64()
}

// SetHostID Lobby HostID setter
func (l *Lobby) SetHostID(hostID int64) error {
	key := fmt.Sprintf("lobby:%d", l.id)
	pipe := client.Pipeline()
	pipe.HSet(key, "host_id", hostID)
	_, err := pipe.Exec()
	return err
}

// IsMember checks if the user is a member
func (l *Lobby) IsMember(userID int64) (bool, error) {
	return client.SIsMember(fmt.Sprintf("lobby:%d:members", l.id), userID).Result()
}

// GetTopMember gets the member supposedly inherits the host role
func (l *Lobby) GetTopMember() (*User, error) {
	ids, err := client.SMembers(fmt.Sprintf("lobby:%d:members", l.id)).Result()
	if err != nil {
		return nil, err
	}

	id, err := strconv.ParseInt(ids[0], 10, 64)
	if err != nil {
		return nil, err
	}

	return &User{id: id}, nil
}

// GetMembers gets all the members
func (l *Lobby) GetMembers() ([]*User, error) {
	ids, err := client.SMembers(fmt.Sprintf("lobby:%d:members", l.id)).Result()
	if err != nil {
		return nil, err
	}

	var users []*User
	for _, v := range ids {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, err
		}
		users = append(users, &User{id: id})
	}

	return users, nil
}

// GetPlayers get all the players' usernames
func (l *Lobby) GetPlayers() ([]string, error) {
	users, err := l.GetMembers()
	if err != nil {
		return nil, err
	}
	us := []string{}
	for _, u := range users {
		un, err := u.GetUsername()
		if err != nil {
			return nil, err
		}
		us = append(us, un)
	}
	return us, nil
}

// LeaveLobby makes the user leave from the lobby
func (l *Lobby) LeaveLobby(userID int64) error {
	hostID, err := l.GetHostID()
	if err != nil {
		return err
	}

	pipe := client.Pipeline()
	pipe.HSet(fmt.Sprintf("user:%d", hostID), "lobby", -1)
	pipe.SRem(fmt.Sprintf("lobby:%d:members", l.id), userID)
	_, err = pipe.Exec()
	if err != nil {
		return err
	}

	if hostID != userID {
		return nil
	}

	u, err := l.GetTopMember()
	if err != nil {
		return err
	}

	key := fmt.Sprintf("lobby:%d", l.id)
	pipe.HSet(key, "host_id", u.id)
	_, err = pipe.Exec()
	return err
}

// AddMember adds a member to the lobby
func (l *Lobby) AddMember(userID int64) error {
	userIsMember, err := l.IsMember(userID)
	if err != nil {
		return err
	}
	if userIsMember {
		return ErrUserInLobby
	}

	id, err := client.HGet(fmt.Sprintf("user:%d", userID), "lobby").Int64()
	if err != nil {
		return err
	}
	if id == l.id {
		return ErrUserInLobby
	}

	pipe := client.Pipeline()
	pipe.HSet(fmt.Sprintf("user:%d", userID), "lobby", l.id)
	pipe.SAdd(fmt.Sprintf("lobby:%d:members", l.id), userID)
	_, err = pipe.Exec()
	return err
}

// GetLobbyByCode get the lobby based on the given code
func GetLobbyByCode(code string) (*Lobby, error) {
	id, err := client.HGet("lobby:by-code", code).Int64()
	if err != nil {
		return nil, err
	}
	return &Lobby{id: id}, nil
}

// JoinLobby joins the user to the lobby from the given code
func JoinLobby(code string, userID int64) error {
	l, err := GetLobbyByCode(code)
	if err != nil {
		return err
	}
	return l.AddMember(userID)
}

// JoinOrCreateLobby process whether the lobby is joined or created by the current user
func JoinOrCreateLobby(choice, code string, userID int64) error {
	if choice == "join" {
		return JoinLobby(code, userID)
	}
	_, err := NewLobby(userID, 5)
	return err
}
