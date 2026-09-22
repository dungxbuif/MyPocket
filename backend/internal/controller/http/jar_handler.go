package httpapi

import (
	"errors"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
)

type JarHandler struct {
	Jars  repository.JarRepository
	Users repository.UserRepository
}

type jarInput struct {
	Month             string   `json:"month"`
	Name              string   `json:"name"`
	AllocationMode    string   `json:"allocation_mode"`
	AllocationAmount  *int64   `json:"allocation_amount"`
	AllocationPercent *float64 `json:"allocation_percent"`
}

func (r *Router) RegisterJarRoutes(h *JarHandler) {
	g := r.Engine.Group("/api/v1/jars")
	g.Use(r.AuthMiddleware.RequireAuth)
	g.GET("", h.List)
	g.POST("", h.Create)
	g.PUT("/:jar_id/months/:month", h.UpdateMonth)
	g.DELETE("/:jar_id/months/:month", h.RemoveMonth)
	g.GET("/:jar_id/report", h.Cumulative)
}

// ListJars godoc
// @Summary Get account jars and derived spending for a month
// @Tags Jars
// @Produce json
// @Security BearerAuth
// @Param month query string false "Calendar month YYYY-MM; defaults to account-local current month"
// @Success 200 {object} entity.JarMonthSummary
// @Router /api/v1/jars [get]
func (h *JarHandler) List(c *gin.Context) {
	owner, ok := walletOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	month, timezone, err := h.period(owner, c.Query("month"))
	if err != nil {
		jarFail(c, err)
		return
	}
	result, err := h.Jars.ListMonth(owner, month, timezone)
	if err != nil {
		jarFail(c, err)
		return
	}
	OK(c, result)
}

// CreateJar godoc
// @Summary Create a stable jar and its configuration for one month
// @Tags Jars
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param jar body jarInput true "Jar configuration"
// @Success 201 {object} entity.JarMonthConfig
// @Router /api/v1/jars [post]
func (h *JarHandler) Create(c *gin.Context) {
	owner, ok := walletOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	var input jarInput
	if c.ShouldBindJSON(&input) != nil {
		jarFail(c, repository.ErrJarInvalid)
		return
	}
	month, timezone, err := h.period(owner, input.Month)
	if err != nil {
		jarFail(c, err)
		return
	}
	name, mode, amount, bps, err := validateJarInput(input)
	if err != nil {
		jarFail(c, err)
		return
	}
	config, err := h.Jars.Create(owner, month, timezone, name, mode, amount, bps)
	if err != nil {
		jarFail(c, err)
		return
	}
	Created(c, config)
}

// UpdateJarMonth godoc
// @Summary Update one jar's month-specific configuration
// @Tags Jars
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param jar_id path string true "Stable jar ID"
// @Param month path string true "Calendar month YYYY-MM"
// @Param jar body jarInput true "Jar month configuration"
// @Success 200 {object} entity.JarMonthConfig
// @Router /api/v1/jars/{jar_id}/months/{month} [put]
func (h *JarHandler) UpdateMonth(c *gin.Context) {
	owner, ok := walletOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	jarID, month := c.Param("jar_id"), c.Param("month")
	if _, err := uuid.Parse(jarID); err != nil {
		jarFail(c, repository.ErrJarInvalid)
		return
	}
	var input jarInput
	if c.ShouldBindJSON(&input) != nil {
		jarFail(c, repository.ErrJarInvalid)
		return
	}
	name, mode, amount, bps, err := validateJarInput(input)
	if err != nil {
		jarFail(c, err)
		return
	}
	config, err := h.Jars.UpdateMonthConfig(owner, jarID, month, name, mode, amount, bps)
	if err != nil {
		jarFail(c, err)
		return
	}
	OK(c, config)
}

// RemoveJarMonth godoc
// @Summary Remove a jar from one month's active configuration and preserve its history
// @Tags Jars
// @Security BearerAuth
// @Param jar_id path string true "Stable jar ID"
// @Param month path string true "Calendar month YYYY-MM"
// @Success 204
// @Router /api/v1/jars/{jar_id}/months/{month} [delete]
func (h *JarHandler) RemoveMonth(c *gin.Context) {
	owner, ok := walletOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if _, err := uuid.Parse(c.Param("jar_id")); err != nil {
		jarFail(c, repository.ErrJarInvalid)
		return
	}
	if err := h.Jars.RemoveMonthConfig(owner, c.Param("jar_id"), c.Param("month")); err != nil {
		jarFail(c, err)
		return
	}
	NoContent(c)
}

// CumulativeJar godoc
// @Summary Get a stable jar's monthly history through an inclusive end month
// @Tags Jars
// @Produce json
// @Security BearerAuth
// @Param jar_id path string true "Stable jar ID"
// @Param from query string false "First month YYYY-MM; omitted means full history"
// @Param to query string false "Last month YYYY-MM; omitted means account-local current month"
// @Success 200 {object} entity.JarCumulativeSummary
// @Router /api/v1/jars/{jar_id}/report [get]
func (h *JarHandler) Cumulative(c *gin.Context) {
	owner, ok := walletOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	jarID := c.Param("jar_id")
	if _, err := uuid.Parse(jarID); err != nil {
		jarFail(c, repository.ErrJarInvalid)
		return
	}
	user, err := h.Users.FindByID(owner)
	if err != nil {
		jarFail(c, err)
		return
	}
	result, err := h.Jars.Cumulative(owner, jarID, c.Query("from"), c.Query("to"), user.Timezone)
	if err != nil {
		jarFail(c, err)
		return
	}
	OK(c, result)
}

func (h *JarHandler) period(owner, requested string) (string, string, error) {
	user, err := h.Users.FindByID(owner)
	if err != nil {
		return "", "", err
	}
	if requested == "" {
		location, err := time.LoadLocation(user.Timezone)
		if err != nil {
			return "", "", err
		}
		requested = time.Now().In(location).Format("2006-01")
	}
	if _, err := entity.ParseMonth(requested); err != nil {
		return "", "", repository.ErrJarInvalid
	}
	return requested, user.Timezone, nil
}

func validateJarInput(input jarInput) (string, string, *int64, *int, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" || len(name) > 100 {
		return "", "", nil, nil, repository.ErrJarInvalid
	}
	mode := input.AllocationMode
	switch mode {
	case entity.JarAllocationNone:
		if input.AllocationAmount != nil || input.AllocationPercent != nil {
			return "", "", nil, nil, repository.ErrJarInvalid
		}
	case entity.JarAllocationFixed:
		if input.AllocationAmount == nil || *input.AllocationAmount <= 0 || *input.AllocationAmount > 9007199254740991 || input.AllocationPercent != nil {
			return "", "", nil, nil, repository.ErrJarInvalid
		}
	case entity.JarAllocationPercent:
		if input.AllocationPercent == nil || math.IsNaN(*input.AllocationPercent) || math.IsInf(*input.AllocationPercent, 0) || *input.AllocationPercent < 0 || *input.AllocationPercent > 100 || math.Round(*input.AllocationPercent*100) != *input.AllocationPercent*100 || input.AllocationAmount != nil {
			return "", "", nil, nil, repository.ErrJarInvalid
		}
		bps := int(math.Round(*input.AllocationPercent * 100))
		return name, mode, nil, &bps, nil
	default:
		return "", "", nil, nil, repository.ErrJarInvalid
	}
	return name, mode, input.AllocationAmount, nil, nil
}

func jarFail(c *gin.Context, err error) {
	status, code, title, message := http.StatusInternalServerError, "JAR_FAILED", problemTitleInternalServer, "Không thể xử lý hũ. Vui lòng thử lại."
	switch {
	case errors.Is(err, repository.ErrJarInvalid):
		status, code, title, message = http.StatusBadRequest, problemCodeBadRequest, problemTitleBadRequest, "Kiểm tra tháng, tên hũ và mức phân bổ."
	case errors.Is(err, gorm.ErrRecordNotFound), errors.Is(err, repository.ErrNotFound):
		status, code, title, message = http.StatusNotFound, "JAR_NOT_FOUND", problemTitleNotFound, "Không tìm thấy hũ trong tháng này."
	}
	Fail(c, status, Problem{Code: code, Title: title, Detail: message})
}
