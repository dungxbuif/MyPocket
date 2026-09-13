package httpapi

import (
	"net/http"
	"strings"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"

	"github.com/gin-gonic/gin"
	categoryrepo "github.com/mypocket/backend/internal/repository"
)

const (
	categoryUnauthorizedMessage = "chưa đăng nhập"
	categoryLoadErrorMessage    = "không đọc được danh sách nhóm"
)

type CategoryHandler struct {
	Categories categoryrepo.CategoryRepository
}

type categoryInput struct { Name string `json:"name"`; Kind string `json:"kind"`; ParentID *string `json:"parent_id"` }
func categoryOwner(c *gin.Context) (string, bool) { value, ok := c.Get(contextUserIDKey); id, valid := value.(string); return id, ok && valid && strings.TrimSpace(id) != "" }

func (h *CategoryHandler) CreateCategory(c *gin.Context) { owner, ok := categoryOwner(c); if !ok { Fail(c, http.StatusUnauthorized, Problem{Code: problemCodeAuthRequired, Title: problemTitleUnauthorized, Detail: categoryUnauthorizedMessage}); return }; var input categoryInput; if c.ShouldBindJSON(&input) != nil || strings.TrimSpace(input.Name) == "" { Fail(c, http.StatusBadRequest, Problem{Code: problemCodeBadRequest, Title: "Bad Request", Detail: "tên nhóm không được để trống"}); return }; item := &entity.Category{ID: uuid.NewString(), Name: strings.TrimSpace(input.Name), Kind: input.Kind, ParentID: input.ParentID}; if item.Kind == "" { item.Kind = "expense" }; if err := h.Categories.Create(owner, item); err != nil { FailError(c, http.StatusInternalServerError, problemCodeCategoryLoadFailed, problemTitleInternalServer, err); return }; Created(c, item) }
func (h *CategoryHandler) UpdateCategory(c *gin.Context) { owner, ok := categoryOwner(c); if !ok { Fail(c, http.StatusUnauthorized, Problem{Code: problemCodeAuthRequired, Title: problemTitleUnauthorized, Detail: categoryUnauthorizedMessage}); return }; var input categoryInput; if c.ShouldBindJSON(&input) != nil || strings.TrimSpace(input.Name) == "" { Fail(c, http.StatusBadRequest, Problem{Code: problemCodeBadRequest, Title: "Bad Request", Detail: "tên nhóm không được để trống"}); return }; item, err := h.Categories.Update(owner, c.Param("id"), map[string]any{"name": strings.TrimSpace(input.Name), "parent_id": input.ParentID}); if err != nil { Fail(c, http.StatusNotFound, Problem{Code: "CATEGORY_NOT_FOUND", Title: "Not Found", Detail: "không tìm thấy nhóm hoặc nhóm hệ thống không thể sửa"}); return }; OK(c, item) }
func (h *CategoryHandler) DeleteCategory(c *gin.Context) { owner, ok := categoryOwner(c); if !ok { Fail(c, http.StatusUnauthorized, Problem{Code: problemCodeAuthRequired, Title: problemTitleUnauthorized, Detail: categoryUnauthorizedMessage}); return }; if err := h.Categories.Delete(owner, c.Param("id")); err != nil { Fail(c, http.StatusNotFound, Problem{Code: "CATEGORY_NOT_FOUND", Title: "Not Found", Detail: "không tìm thấy nhóm hoặc nhóm hệ thống không thể xóa"}); return }; NoContent(c) }

func NewCategoryHandler(categories categoryrepo.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{Categories: categories}
}

// ListCategories godoc
// @Summary List visible categories
// @Tags Categories
// @Produce json
// @Security BearerAuth
// @Success 200 {array} entity.Category
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/categories [get]
func (h *CategoryHandler) ListCategories(c *gin.Context) {
	value, ok := c.Get(contextUserIDKey)
	if !ok {
		Fail(c, http.StatusUnauthorized, Problem{Code: problemCodeAuthRequired, Title: problemTitleUnauthorized, Detail: categoryUnauthorizedMessage})
		return
	}
	userID, ok := value.(string)
	if !ok || userID == "" {
		Fail(c, http.StatusUnauthorized, Problem{Code: problemCodeAuthRequired, Title: problemTitleUnauthorized, Detail: categoryUnauthorizedMessage})
		return
	}
	categories, err := h.Categories.ListVisible(userID)
	if err != nil {
		Fail(c, http.StatusInternalServerError, Problem{Code: problemCodeCategoryLoadFailed, Title: problemTitleInternalServer, Detail: categoryLoadErrorMessage})
		return
	}
	OK(c, categories)
}
