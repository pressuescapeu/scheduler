package middleware

import (
	"net/http"
	"scheduler/internal/utils"
	"strings"

	"github.com/labstack/echo/v4"
)

func JWTAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, utils.ErrUnauthorized)
			}

			if !strings.HasPrefix(authHeader, "Bearer ") {
				return c.JSON(http.StatusUnauthorized, utils.ErrInvalidToken)
			}

			token := strings.Split(authHeader, " ")[1]

			claims, err := utils.ValidateToken(token)

			if err != nil {
				return c.JSON(http.StatusUnauthorized, utils.ErrInvalidToken)
			}

			c.Set("user_id", claims.UserID)
			c.Set("email", claims.Email)

			return next(c)
		}
	}
}
