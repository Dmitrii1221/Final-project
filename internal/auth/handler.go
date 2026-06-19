package auth

import (
	"context"
	"final-project/internal/domain"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	userRepo interface {
		Insert(ctx context.Context, u domain.User) (domain.User, error)
		GetByUsername(ctx context.Context, username string) (domain.User, error)
	}
	secret []byte
	ttl    time.Duration
}

func NewHandler(
	userRepo interface {
		Insert(ctx context.Context, u domain.User) (domain.User, error)
		GetByUsername(ctx context.Context, username string) (domain.User, error)
	},
	secret string,
	ttl time.Duration,
) *Handler {
	return &Handler{
		userRepo: userRepo,
		secret:   []byte(secret),
		ttl:      ttl,
	}
}

func (h *Handler) Register(c echo.Context) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "hash error")
	}

	user, err := h.userRepo.Insert(c.Request().Context(), domain.User{
		Username:     req.Username,
		PasswordHash: hash,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusConflict, "username taken")
	}

	token, err := GenerateToken(user.ID, user.Username, h.secret, h.ttl)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "token error")
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"token": token,
		"user":  user,
	})
}
func (h *Handler) Login(c echo.Context) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}

	user, err := h.userRepo.GetByUsername(c.Request().Context(), req.Username)
	if err != nil {
		// dummy hash чтобы не отличать "юзер не найден" от "неверный пароль". делает одинаковым время ответа
		CheckPassword(req.Password, "$2a$10$dummyhashdummyhashdummyhashdummyhashdummyhashdu")
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}

	if err := CheckPassword(req.Password, user.PasswordHash); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}

	token, err := GenerateToken(user.ID, user.Username, h.secret, h.ttl)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "token error")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"token": token,
		"user":  user,
	})
}

func (h *Handler) Me(c echo.Context) error {
	claims := ClaimsFromContext(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	return c.JSON(http.StatusOK, map[string]any{
		"id":       claims.UserID,
		"username": claims.Username,
	})
}
