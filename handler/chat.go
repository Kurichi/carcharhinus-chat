package handler

import (
	"net/http"
	"time"

	"github.com/Kurichi/carcharhinus-chat/domain"
	"github.com/Kurichi/carcharhinus-chat/infra"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type ChatHandler struct {
	repo *infra.RoomRepository
}

func NewChatHandler(repo *infra.RoomRepository) *ChatHandler {
	return &ChatHandler{
		repo: repo,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *ChatHandler) Join(c echo.Context) error {
	username, ok := c.Get("username").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}
	roomID := c.Param("roomID")
	if !h.repo.IsRoomExist(roomID) {
		return echo.NewHTTPError(http.StatusNotFound, "room not found")
	}

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	// store client
	user := domain.NewUser(username, ws)
	defer close(user.Cancel)
	if err := h.repo.AddUser(roomID, user); err != nil {
		return err
	}

	for {
		var msg domain.Comment
		if err := ws.ReadJSON(&msg); err != nil {
			break
		}
		msg.Timestamp = time.Now().Unix()
		msg.UserName = username

		users, err := h.repo.GetUsers(roomID)
		if err != nil {
			break
		}
		for _, u := range users {
			u.Ch <- msg
		}
	}

	return nil
}
