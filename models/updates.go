package models

import (
	"fmt"
	"strconv"
)

type Update struct {
	id int64
}

func NewUpdate(userID int64, body string) (*Update, error) {
	id, err := client.Incr("update:next-id").Result()
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("update:%d", id)
	pipe := client.Pipeline()
	pipe.HSet(key, "id", id)
	pipe.HSet(key, "user_id", userID)
	pipe.HSet(key, "body", body)
	pipe.LPush("updates", id)
	pipe.LPush(fmt.Sprintf("user:%d:updates", userID), id)
	_, err = pipe.Exec()
	if err != nil {
		return nil, err
	}

	return &Update{id: id}, nil
}

func (u *Update) GetBody() (string, error) {
	key := fmt.Sprintf("update:%d", u.id)
	return client.HGet(key, "body").Result()
}

func (u *Update) GetUser() (*User, error) {
	key := fmt.Sprintf("update:%d", u.id)
	userID, err := client.HGet(key, "user_id").Int64()
	if err != nil {
		return nil, err
	}

	return GetUserByUserID(userID)
}

func queryUpdates(key string) ([]*Update, error) {
	updateIDs, err := client.LRange(key, 0, 10).Result()
	if err != nil {
		return nil, err
	}

	updates := make([]*Update, len(updateIDs))
	for i, val := range updateIDs {
		id, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, err
		}
		updates[i] = &Update{id: id}
	}
	return updates, nil
}

func GetAllUpdates() ([]*Update, error) {
	return queryUpdates("updates")
}

func GetUpdates(userID int64) ([]*Update, error) {
	key := fmt.Sprintf("user:%d:updates", userID)
	return queryUpdates(key)
}

func PostUpdate(userID int64, body string) error {
	_, err := NewUpdate(userID, body)
	return err
}
