package handler

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

type OTTHandler struct {
	store *sync.Map
}

func NewOTTHandler() *OTTHandler {
	return &OTTHandler{
		store: &sync.Map{},
	}
}

func genToken(n int) string {
	const rs2Letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	b := make([]byte, n)
	for i := range b {
		b[i] = rs2Letters[rand.Intn(len(rs2Letters))]
	}
	return string(b)
}

func (h *OTTHandler) GenToken(c echo.Context) error {
	tokenString := c.Request().Header.Get("Authorization")

	jwtToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("hfodsahfoasjfdas"), nil
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok || !jwtToken.Valid {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	userID := claims["id"].(string)
	token := genToken(32)
	h.store.Store(token, userID)

	return c.JSON(http.StatusOK, map[string]string{
		"token": token,
	})
}

// Join middleware
// Join middleware is used to check if the one-time-token is valid
func (h *OTTHandler) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.QueryParam("token")

		userID, ok := h.store.LoadAndDelete(token)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		}

		c.Set("userID", userID)

		return next(c)
	}
}
