package domain

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

type Room struct {
	ID    string
	Users map[string]*User
}

type User struct {
	Name   string
	Conn   *websocket.Conn
	Ch     chan Comment
	Cancel chan struct{}
}

type Comment struct {
	UserName  string `json:"userName"`
	Price     int    `json:"price"`
	Comment   string `json:"comment"`
	Timestamp int64  `json:"timestamp"`
}

func NewUser(id string, conn *websocket.Conn) *User {
	return &User{
		Name:   id,
		Conn:   conn,
		Ch:     make(chan Comment, 10),
		Cancel: make(chan struct{}),
	}
}

var ErrRoomNotFound = errors.New("room not found")

func (c *User) Run() {
	for {
		select {
		case msg := <-c.Ch:
			fmt.Println("Run", c.Name, msg)
			err := c.Conn.WriteJSON(msg)
			if err != nil {
				return
			}
		case <-c.Cancel:
			return
		}
	}
}
