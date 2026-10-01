package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	travelrepo "github.com/mypocket/backend/internal/repository"
)

type TravelHandler struct {
	Events travelrepo.TravelRepository
}

type travelEventInput struct {
	Name     string  `json:"name"`
	Context  *string `json:"context"`
	StartsOn *string `json:"starts_on"`
	EndsOn   *string `json:"ends_on"`
	Active   *bool   `json:"active"`
}

func (r *Router) RegisterTravelRoutes(h *TravelHandler) {
	g := r.Engine.Group("/api/v1/travel/events")
	g.Use(r.AuthMiddleware.RequireAuth)
	g.GET("", h.List)
	g.POST("", h.Create)
	g.PATCH("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
	g.POST("/:id/activate", h.Activate)
	g.POST("/:id/deactivate", h.Deactivate)
}

// ListTravelEvents godoc
// @Summary List travel events
// @Tags Travel
// @Produce json
// @Security BearerAuth
// @Success 200 {array} entity.TravelEvent
// @Router /api/v1/travel/events [get]
func (h *TravelHandler) List(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h.Events == nil {
		travelUnavailable(c)
		return
	}
	rows, err := h.Events.List(owner)
	if err != nil {
		travelFailure(c)
		return
	}
	OK(c, rows)
}

// CreateTravelEvent godoc
// @Summary Create a travel event
// @Tags Travel
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param event body travelEventInput true "Travel event"
// @Success 201 {object} entity.TravelEvent
// @Failure 400 {object} Problem
// @Router /api/v1/travel/events [post]
func (h *TravelHandler) Create(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h.Events == nil {
		travelUnavailable(c)
		return
	}
	input, startsOn, endsOn, valid := bindTravelEvent(c, false, nil)
	if !valid {
		return
	}
	active := false
	if input.Active != nil {
		active = *input.Active
	}
	event := &entity.TravelEvent{ID: uuid.NewString(), OwnerID: owner, Name: input.Name, Context: normalizeOptional(input.Context), StartsOn: startsOn, EndsOn: endsOn, Active: active}
	if err := h.Events.Create(owner, event); err != nil {
		travelMapError(c, err)
		return
	}
	Created(c, event)
}

// UpdateTravelEvent godoc
// @Summary Update a travel event
// @Tags Travel
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Travel event ID"
// @Param event body travelEventInput true "Travel event"
// @Success 200 {object} entity.TravelEvent
// @Failure 400 {object} Problem
// @Failure 404 {object} Problem
// @Router /api/v1/travel/events/{id} [patch]
func (h *TravelHandler) Update(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h.Events == nil {
		travelUnavailable(c)
		return
	}
	current, err := h.Events.Find(owner, strings.TrimSpace(c.Param("id")))
	if err != nil {
		travelNotFound(c)
		return
	}
	input, startsOn, endsOn, valid := bindTravelEvent(c, true, current)
	if !valid {
		return
	}
	updated, err := h.Events.Update(owner, current.ID, travelrepo.TravelUpdate{Name: input.Name, Context: normalizeOptional(input.Context), StartsOn: startsOn, EndsOn: endsOn})
	if err != nil {
		travelMapError(c, err)
		return
	}
	OK(c, updated)
}

// DeleteTravelEvent godoc
// @Summary Delete a travel event and clear its transaction links
// @Tags Travel
// @Security BearerAuth
// @Param id path string true "Travel event ID"
// @Success 204
// @Failure 404 {object} Problem
// @Router /api/v1/travel/events/{id} [delete]
func (h *TravelHandler) Delete(c *gin.Context) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h.Events == nil {
		travelUnavailable(c)
		return
	}
	if err := h.Events.Delete(owner, strings.TrimSpace(c.Param("id"))); err != nil {
		travelMapError(c, err)
		return
	}
	NoContent(c)
}

// ActivateTravelEvent godoc
// @Summary Activate one travel event
// @Tags Travel
// @Produce json
// @Security BearerAuth
// @Param id path string true "Travel event ID"
// @Success 200 {object} entity.TravelEvent
// @Failure 404 {object} Problem
// @Router /api/v1/travel/events/{id}/activate [post]
func (h *TravelHandler) Activate(c *gin.Context) { h.setActive(c, true) }

// DeactivateTravelEvent godoc
// @Summary Deactivate a travel event
// @Tags Travel
// @Produce json
// @Security BearerAuth
// @Param id path string true "Travel event ID"
// @Success 200 {object} entity.TravelEvent
// @Failure 404 {object} Problem
// @Router /api/v1/travel/events/{id}/deactivate [post]
func (h *TravelHandler) Deactivate(c *gin.Context) { h.setActive(c, false) }

func (h *TravelHandler) setActive(c *gin.Context, active bool) {
	owner, ok := transactionOwner(c)
	if !ok {
		transactionUnauthorized(c)
		return
	}
	if h.Events == nil {
		travelUnavailable(c)
		return
	}
	row, err := h.Events.SetActive(owner, strings.TrimSpace(c.Param("id")), active)
	if err != nil {
		travelMapError(c, err)
		return
	}
	OK(c, row)
}

func bindTravelEvent(c *gin.Context, update bool, current *entity.TravelEvent) (travelEventInput, *entity.CalendarDate, *entity.CalendarDate, bool) {
	var input travelEventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		transactionBadRequest(c, problemDetailInvalidJSON)
		return travelEventInput{}, nil, nil, false
	}
	if update && current != nil {
		if strings.TrimSpace(input.Name) == "" {
			input.Name = current.Name
		}
		if input.Context == nil {
			input.Context = current.Context
		}
		if input.StartsOn == nil && current.StartsOn != nil {
			value := current.StartsOn.String()
			input.StartsOn = &value
		}
		if input.EndsOn == nil && current.EndsOn != nil {
			value := current.EndsOn.String()
			input.EndsOn = &value
		}
	}
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		transactionBadRequest(c, "Tên chuyến là bắt buộc.")
		return travelEventInput{}, nil, nil, false
	}
	starts, ends, valid := parseTravelDates(input.StartsOn, input.EndsOn)
	if !valid {
		transactionBadRequest(c, "Ngày chuyến phải có định dạng YYYY-MM-DD và ngày kết thúc không trước ngày bắt đầu.")
		return travelEventInput{}, nil, nil, false
	}
	return input, starts, ends, true
}

func parseTravelDates(startsRaw, endsRaw *string) (*entity.CalendarDate, *entity.CalendarDate, bool) {
	var starts, ends *entity.CalendarDate
	if startsRaw != nil && strings.TrimSpace(*startsRaw) != "" {
		parsed, err := entity.ParseCalendarDate(strings.TrimSpace(*startsRaw))
		if err != nil {
			return nil, nil, false
		}
		starts = &parsed
	}
	if endsRaw != nil && strings.TrimSpace(*endsRaw) != "" {
		parsed, err := entity.ParseCalendarDate(strings.TrimSpace(*endsRaw))
		if err != nil {
			return nil, nil, false
		}
		ends = &parsed
	}
	if starts != nil && ends != nil && ends.Time.Before(starts.Time) {
		return nil, nil, false
	}
	return starts, ends, true
}

func travelMapError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, travelrepo.ErrTravelNotFound):
		travelNotFound(c)
	case errors.Is(err, travelrepo.ErrTravelInvalid):
		transactionBadRequest(c, "Thông tin chuyến chưa hợp lệ.")
	case errors.Is(err, travelrepo.ErrTravelTransaction):
		transactionBadRequest(c, "Giao dịch không hỗ trợ liên kết chuyến.")
	default:
		travelFailure(c)
	}
}

func travelNotFound(c *gin.Context) {
	Fail(c, http.StatusNotFound, Problem{Code: problemCodeTransactionNotFound, Title: problemTitleNotFound, Detail: "Không tìm thấy chuyến."})
}

func travelFailure(c *gin.Context) {
	Fail(c, http.StatusInternalServerError, Problem{Code: problemCodeTransactionSaveFailed, Title: problemTitleInternalServer, Detail: "Không xử lý được Travel Mode."})
}

func travelUnavailable(c *gin.Context) {
	Fail(c, http.StatusNotImplemented, Problem{Code: problemCodeTransactionSaveFailed, Title: problemTitleInternalServer, Detail: "Travel Mode chưa được cấu hình."})
}
