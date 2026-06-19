package auth

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type ctxKey string

const ClaimsKey ctxKey = "claims"

func Middleware(secret []byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Request().Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(header, "Bearer ") {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing token")
			}

			tokenStr := strings.TrimPrefix(header, "Bearer ")
			claims, err := ValidateToken(tokenStr, secret)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			c.Set(string(ClaimsKey), claims)
			return next(c)
		}
	}
}

func ClaimsFromContext(c echo.Context) *Claims {
	if claims, ok := c.Get(string(ClaimsKey)).(*Claims); ok {
		return claims
	}
	return nil
}
