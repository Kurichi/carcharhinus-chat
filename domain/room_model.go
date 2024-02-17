package domain

import (
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

type Room struct {
	ID    string
	Users map[string]*User
}

type User struct {
	ID     string
	Conn   *websocket.Conn
	Ch     chan interface{}
	Cancel chan struct{}
}

func NewUser(id string, conn *websocket.Conn) *User {
	return &User{
		ID:     id,
		Conn:   conn,
		Ch:     make(chan interface{}, 10),
		Cancel: make(chan struct{}),
	}
}

var ErrRoomNotFound = errors.New("room not found")

func (c *User) Run() {
	for {
		select {
		case msg := <-c.Ch:
			err := c.Conn.WriteJSON(msg)
			if err != nil {
				return
			}
		case <-c.Cancel:
			return
		}
	}
}
