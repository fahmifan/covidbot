package covidbot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrNotFound = errors.New("not found")
)

type User struct {
	ID               string
	TelegramUpdateID string
	ChatID           int64
	Username         string
}

type UserService interface {
	Create(ctx context.Context, user *User) error
	ForEach(ctx context.Context, cb func(user User) error) (err error)
}

func Int64ToString(i int64) string {
	return fmt.Sprint(i)
}

// Dumps Dump in string
func Dumps(i interface{}) string {
	return string(Dump(i))
}

// Dump ..
func Dump(i interface{}) []byte {
	bt, _ := json.Marshal(i)
	return bt
}
