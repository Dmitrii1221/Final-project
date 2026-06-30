package http

import (
	"context"
	"net/http"
	"sort"
	"strconv"
	"time"

	"final-project/internal/domain"

	"github.com/labstack/echo/v4"
)

type BudgetHandler struct {
	budgetRepo interface {
		GetByID(ctx context.Context, id int64) (domain.Budget, error)
	}
	periodRepo interface {
		GetByID(ctx context.Context, id int64) (domain.BudgetPeriod, error)
	}
	spendingRepo interface {
		ListByBudgetID(ctx context.Context, budgetID int64, currencyID *int64, periodID *int64, from, to *time.Time) ([]domain.Spending, error)
	}
	currencyRepo interface {
		GetByCode(ctx context.Context, code string) (domain.Currency, error)
	}
}

func NewBudgetHandler(budgetRepo interface {
	GetByID(ctx context.Context, id int64) (domain.Budget, error)
}, periodRepo interface {
	GetByID(ctx context.Context, id int64) (domain.BudgetPeriod, error)
}, spendingRepo interface {
	ListByBudgetID(ctx context.Context, budgetID int64, currencyID *int64, periodID *int64, from, to *time.Time) ([]domain.Spending, error)
},
	currencyRepo interface {
		GetByCode(ctx context.Context, code string) (domain.Currency, error)
	}) *BudgetHandler {
	return &BudgetHandler{budgetRepo: budgetRepo,
		periodRepo:   periodRepo,
		spendingRepo: spendingRepo,
		currencyRepo: currencyRepo,
	}

}

func (h *BudgetHandler) GetBudget(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid budget id")
	}

	budget, err := h.budgetRepo.GetByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "budget not found")
	}

	return c.JSON(http.StatusOK, budget)
}
func (h *BudgetHandler) GetStats(c echo.Context) error {
	budgetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid budget id")
	}

	var currencyID *int64
	if code := c.QueryParam("currency"); code != "" {
		currency, err := h.currencyRepo.GetByCode(c.Request().Context(), code)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "unknown currency")
		}
		currencyID = &currency.ID
	}

	var periodID *int64
	if s := c.QueryParam("period_id"); s != "" {
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid period_id")
		}
		periodID = &id
	}

	var from *time.Time
	if s := c.QueryParam("from"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid from")
		}
		from = &t
	}

	var to *time.Time
	if s := c.QueryParam("to"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid to")
		}
		to = &t
	}

	spendings, err := h.spendingRepo.ListByBudgetID(
		c.Request().Context(), budgetID, currencyID, periodID, from, to,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list spendings failed")
	}

	if c.QueryParam("group_by") == "day" {
		byDay := make(map[string][]domain.Spending)
		for _, s := range spendings {
			day := s.SpentAt.Format("2006-01-02")
			byDay[day] = append(byDay[day], s)
		}
		type dayGroup struct {
			Date      string            `json:"date"`
			Spendings []domain.Spending `json:"spendings"`
		}
		out := make([]dayGroup, 0, len(byDay))
		for date, ss := range byDay {
			out = append(out, dayGroup{Date: date, Spendings: ss})
		}
		sort.Slice(out, func(i, j int) bool { return out[i].Date < out[j].Date })
		return c.JSON(http.StatusOK, out)
	}

	return c.JSON(http.StatusOK, spendings)
}

func (h *BudgetHandler) GetPeriod(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("period_id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid period id")
	}

	period, err := h.periodRepo.GetByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "period not found")
	}

	return c.JSON(http.StatusOK, period)
}
