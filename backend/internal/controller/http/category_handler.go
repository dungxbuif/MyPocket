package httpapi

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/usecase"
	"gorm.io/gorm"
)

const (
	categoryUnauthorizedMessage   = "chưa đăng nhập"
	categoryLoadErrorMessage      = "không đọc được danh sách nhóm"
	categoryNotFoundMessage       = "không tìm thấy nhóm hoặc nhóm hệ thống không thể sửa"
	categoryDeleteNotFoundMessage = "không tìm thấy nhóm hoặc nhóm hệ thống không thể xóa"
)

type CategoryHandler struct{ Categories *usecase.CategoryInteractor }
type categoryInput struct {
	Name      string   `json:"name"`
	Kind      string   `json:"kind"`
	ParentID  *string  `json:"parent_id"`
	WalletIDs []string `json:"wallet_ids"`
	IconKey   string   `json:"icon_key"`
}
type categoryWalletInput struct {
	WalletIDs []string `json:"wallet_ids"`
}

func NewCategoryHandler(categories *usecase.CategoryInteractor) *CategoryHandler {
	return &CategoryHandler{Categories: categories}
}

// ListCategories godoc
// @Summary List visible categories
// @Tags Categories
// @Produce json
// @Security BearerAuth
// @Success 200 {array} entity.Category
// @Failure 401 {object} Problem
// @Router /api/v1/categories [get]
func (h *CategoryHandler) ListCategories(c *gin.Context) {
	owner, ok := categoryOwner(c)
	if !ok {
		categoryUnauthorized(c)
		return
	}
	items, err := h.Categories.List(owner)
	if err != nil {
		Fail(c, http.StatusInternalServerError, Problem{Code: problemCodeCategoryLoadFailed, Title: problemTitleInternalServer, Detail: categoryLoadErrorMessage})
		return
	}
	OK(c, items)
}

// CreateCategory godoc
// @Summary Create a personal category
// @Tags Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param category body categoryInput true "Category input"
// @Success 201 {object} entity.Category
// @Failure 400 {object} Problem
// @Router /api/v1/categories [post]
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	owner, ok := categoryOwner(c)
	if !ok {
		categoryUnauthorized(c)
		return
	}
	input, ok := bindCategoryInput(c)
	if !ok {
		return
	}
	item, err := h.Categories.Create(owner, input)
	if err != nil {
		categoryError(c, err, categoryNotFoundMessage)
		return
	}
	Created(c, item)
}

// UpdateCategory godoc
// @Summary Update a personal category
// @Tags Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Param category body categoryInput true "Category input"
// @Success 200 {object} entity.Category
// @Failure 400 {object} Problem
// @Failure 404 {object} Problem
// @Router /api/v1/categories/{id} [patch]
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	owner, ok := categoryOwner(c)
	if !ok {
		categoryUnauthorized(c)
		return
	}
	input, ok := bindCategoryInput(c)
	if !ok {
		return
	}
	item, err := h.Categories.Update(owner, c.Param("id"), input)
	if err != nil {
		categoryError(c, err, categoryNotFoundMessage)
		return
	}
	OK(c, item)
}

// UpdateCategoryWallets godoc
// @Summary Update wallets applicable to a visible category
// @Tags Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Param wallets body categoryWalletInput true "Applicable wallets"
// @Success 200 {object} entity.Category
// @Failure 400 {object} Problem
// @Failure 404 {object} Problem
// @Router /api/v1/categories/{id}/wallets [patch]
func (h *CategoryHandler) UpdateCategoryWallets(c *gin.Context) {
	owner, ok := categoryOwner(c)
	if !ok {
		categoryUnauthorized(c)
		return
	}
	var input categoryWalletInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Fail(c, http.StatusBadRequest, Problem{Code: problemCodeBadRequest, Title: problemTitleBadRequest, Detail: problemDetailInvalidJSON})
		return
	}
	item, err := h.Categories.UpdateWallets(owner, c.Param("id"), input.WalletIDs)
	if err != nil {
		categoryError(c, err, categoryNotFoundMessage)
		return
	}
	OK(c, item)
}

// DeleteCategory godoc
// @Summary Delete a personal category without children
// @Tags Categories
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Success 204
// @Failure 400 {object} Problem
// @Failure 404 {object} Problem
// @Router /api/v1/categories/{id} [delete]
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	owner, ok := categoryOwner(c)
	if !ok {
		categoryUnauthorized(c)
		return
	}
	err := h.Categories.Delete(owner, c.Param("id"))
	if err != nil {
		categoryError(c, err, categoryDeleteNotFoundMessage)
		return
	}
	NoContent(c)
}

func categoryOwner(c *gin.Context) (string, bool) { return walletOwner(c) }
func categoryUnauthorized(c *gin.Context) {
	Fail(c, http.StatusUnauthorized, Problem{Code: problemCodeAuthRequired, Title: problemTitleUnauthorized, Detail: categoryUnauthorizedMessage})
}
func bindCategoryInput(c *gin.Context) (usecase.CategoryInput, bool) {
	var input categoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Fail(c, http.StatusBadRequest, Problem{Code: problemCodeBadRequest, Title: problemTitleBadRequest, Detail: problemDetailInvalidJSON})
		return usecase.CategoryInput{}, false
	}
	return usecase.CategoryInput{Name: input.Name, Kind: input.Kind, ParentID: input.ParentID, WalletIDs: input.WalletIDs, IconKey: input.IconKey}, true
}
func categoryError(c *gin.Context, err error, notFoundDetail string) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		Fail(c, http.StatusNotFound, Problem{Code: problemCodeCategoryNotFound, Title: problemTitleNotFound, Detail: notFoundDetail})
		return
	}
	if errors.Is(err, usecase.ErrCategoryNameRequired) || errors.Is(err, usecase.ErrCategoryKindInvalid) || errors.Is(err, usecase.ErrCategoryParentInvalid) || errors.Is(err, usecase.ErrCategoryWalletInvalid) {
		Fail(c, http.StatusBadRequest, Problem{Code: problemCodeBadRequest, Title: problemTitleBadRequest, Detail: err.Error()})
		return
	}
	Fail(c, http.StatusInternalServerError, Problem{Code: problemCodeCategorySaveFailed, Title: problemTitleInternalServer, Detail: categoryLoadErrorMessage})
}
