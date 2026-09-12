package finance

import "fmt"

func ValidateCreateCategory(input CreateCategoryInput) (CreateCategoryInput, error) {
	input.Name = trimmed(input.Name)
	input.ParentID = trimmed(input.ParentID)
	if input.Name == "" {
		return CreateCategoryInput{}, fmt.Errorf("%w: category name is required", ErrValidation)
	}
	if !validCategoryKind(input.Kind) {
		return CreateCategoryInput{}, fmt.Errorf("%w: unsupported category kind", ErrValidation)
	}
	if input.ParentDepth != nil && *input.ParentDepth >= 2 {
		return CreateCategoryInput{}, fmt.Errorf("%w: category depth cannot exceed two levels", ErrValidation)
	}
	return input, nil
}

func ValidateCategoryUpdate(category Category, input UpdateCategoryInput) error {
	name := trimmed(input.Name)
	if category.IsSystem {
		return ErrSystemCategoryLocked
	}
	if name == "" {
		return fmt.Errorf("%w: category name is required", ErrValidation)
	}
	return nil
}

func validCategoryKind(value CategoryKind) bool {
	switch value {
	case CategoryExpense, CategoryIncome, CategoryDebt:
		return true
	default:
		return false
	}
}
