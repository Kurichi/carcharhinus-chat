package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func main() {
	// Setup
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())
	e.Use(middleware.Recover())

	m := newManager()
	e.GET("/:roomID/ws", m.join)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	// Start server
	go func() {
		if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}

type client struct {
	token  string
	ws     *websocket.Conn
	ch     chan []byte
	cancel chan struct{}
}

type RoomType struct {
	clients sync.Map
}

type manager struct {
	rooms sync.Map
}

func newManager() *manager {
	return &manager{
		rooms: sync.Map{},
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (m *manager) join(c echo.Context) error {
	token := c.QueryParam("token")
	roomID := c.Param("roomID")

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	// get room
	r, _ := m.rooms.LoadOrStore(roomID, &RoomType{clients: sync.Map{}})
	room, ok := r.(*RoomType)
	if !ok {
		log.Error("room type error")
		return nil
	}

	// store client
	cl := &client{
		token:  token,
		ws:     ws,
		ch:     make(chan []byte, 10),
		cancel: make(chan struct{}),
	}
	defer close(cl.ch)
	go cl.run()
	room.clients.Store(token, cl)

	// read
	// go func() {
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			room.clients.Delete(token)
			break
		}
		room.clients.Range(func(key, value interface{}) bool {
			value.(*client).ch <- msg
			return true
		})
	}
	// }()

	return nil
}

func (c *client) run() {
	for {
		select {
		case msg := <-c.ch:
			err := c.ws.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				return
			}
		case <-c.cancel:
			return
		}
	}
}
