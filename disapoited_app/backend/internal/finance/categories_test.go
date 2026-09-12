package finance_test

import (
	"errors"
	"testing"

	"mypocket/internal/finance"
)

func TestSystemCategoryCannotBeRenamed(t *testing.T) {
	err := finance.ValidateCategoryUpdate(finance.Category{IsSystem: true, Name: "Ăn uống"}, finance.UpdateCategoryInput{
		Name: "Tên mới",
	})

	if !errors.Is(err, finance.ErrSystemCategoryLocked) {
		t.Fatalf("expected system lock error, got %v", err)
	}
}

func TestCreateCategoryRejectsDepthAboveTwoLevels(t *testing.T) {
	_, err := finance.ValidateCreateCategory(finance.CreateCategoryInput{
		Kind:        finance.CategoryExpense,
		Name:        "Quán quen",
		ParentDepth: ptrInt(2),
	})

	if !errors.Is(err, finance.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}
