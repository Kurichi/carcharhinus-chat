package handler

import (
	"net/http"

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
	userID, ok := c.Get("userID").(string)
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
	user := domain.NewUser(userID, ws)
	defer close(user.Cancel)
	if err := h.repo.AddUser(roomID, user); err != nil {
		return err
	}

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			break
		}
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
