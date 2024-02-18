package handler

import (
	"log"
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
		h.repo.CreateRoom(roomID)
	}

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to upgrade to websocket")
	}
	defer ws.Close()

	// store client
	user := domain.NewUser(username, ws)
	defer close(user.Cancel)
	if err := h.repo.AddUser(roomID, user); err != nil {
		return err
	}
	user.Run()

	for {
		var msg domain.Comment
		if err := ws.ReadJSON(&msg); err != nil {
			break
		}
		msg.Timestamp = time.Now().Unix()
		msg.UserName = username

		err := h.repo.PushMsg(roomID, &msg)
		if err != nil {
			break
		}
		// for _, u := range users {
		// 	u.Ch <- msg
		// }
	}

	return nil
}
