package usecase

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	categoryrepo "github.com/mypocket/backend/internal/repository"
)

const (
	CategoryKindExpense = "expense"
	CategoryKindIncome  = "income"
	CategoryKindDebt    = "debt"
)

var (
	ErrCategoryNameRequired        = errors.New("tên nhóm không được để trống")
	ErrCategoryKindInvalid         = errors.New("loại nhóm không hợp lệ")
	ErrCategoryParentInvalid       = errors.New("nhóm cha không hợp lệ")
	ErrCategoryWalletInvalid       = errors.New("ví áp dụng không hợp lệ")
	ErrCategoryIconInvalid         = errors.New("icon nhóm không hợp lệ")
	ErrCategoryMutationUnavailable = errors.New("category mutation ledger unavailable")
)

type CategoryInput struct {
	Name            string
	Kind            string
	ParentID        *string
	WalletIDs       []string
	IconKey         string
	ExpectedVersion *int64
}
type CategoryInteractor struct {
	repository categoryrepo.CategoryRepository
}

func NewCategoryInteractor(repository categoryrepo.CategoryRepository) *CategoryInteractor {
	return &CategoryInteractor{repository: repository}
}
func (u *CategoryInteractor) List(ownerID string) ([]entity.Category, error) {
	if err := u.repository.EnsurePersonalDefaults(ownerID); err != nil {
		return nil, err
	}
	return u.repository.ListVisible(ownerID)
}

func (u *CategoryInteractor) Create(ownerID string, input CategoryInput) (*entity.Category, error) {
	input, err := u.validateInput(ownerID, "", input)
	if err != nil {
		return nil, err
	}
	if err := u.repository.ValidateWallets(ownerID, input.WalletIDs); err != nil {
		return nil, err
	}
	item := newCategory(input)
	if err := u.repository.Create(ownerID, item); err != nil {
		return nil, err
	}
	return item, nil
}

// CreateIdempotent validates and prepares the same category payload as Create,
// then delegates the actual write and durable replay receipt to the repository.
// The ledger capability is deliberately optional so legacy in-memory/test
// repositories cannot accidentally claim replay safety.
func (u *CategoryInteractor) CreateIdempotent(ownerID string, input CategoryInput, idempotencyKey, requestHash string) (*entity.Category, bool, error) {
	input, err := u.validateInput(ownerID, "", input)
	if err != nil {
		return nil, false, err
	}
	if err := u.repository.ValidateWallets(ownerID, input.WalletIDs); err != nil {
		return nil, false, err
	}
	mutator, ok := u.repository.(categoryrepo.CategoryCreateMutationRepository)
	if !ok {
		return nil, false, ErrCategoryMutationUnavailable
	}
	return mutator.CreateIdempotent(ownerID, newCategory(input), idempotencyKey, requestHash)
}

func newCategory(input CategoryInput) *entity.Category {
	return &entity.Category{ID: uuid.NewString(), Name: input.Name, Kind: input.Kind, ParentID: input.ParentID, IconKey: input.IconKey, WalletIDs: input.WalletIDs, IsActive: true}
}

func (u *CategoryInteractor) SetActive(ownerID, id string, active bool) (*entity.Category, error) {
	return u.repository.SetActive(ownerID, id, active)
}

func (u *CategoryInteractor) Update(ownerID, id string, input CategoryInput) (*entity.Category, error) {
	item, err := u.repository.FindPersonal(ownerID, id)
	if err != nil {
		return nil, err
	}
	input, err = u.validateInput(ownerID, id, input)
	if err != nil {
		return nil, err
	}
	if err := u.repository.ValidateWallets(ownerID, input.WalletIDs); err != nil {
		return nil, err
	}
	return u.updateValidated(ownerID, id, item, input, maxCategoryVersion(item.Version))
}

// UpdateVersioned applies an atomic compare-and-swap when the backing
// repository supports it. It is used by confirmed AI actions and keeps the
// legacy Update path compatible with older repository implementations.
func (u *CategoryInteractor) UpdateVersioned(ownerID, id string, input CategoryInput, expectedVersion int64) (*entity.Category, error) {
	item, err := u.repository.FindPersonal(ownerID, id)
	if err != nil {
		return nil, err
	}
	input, err = u.validateInput(ownerID, id, input)
	if err != nil {
		return nil, err
	}
	if err := u.repository.ValidateWallets(ownerID, input.WalletIDs); err != nil {
		return nil, err
	}
	return u.updateValidated(ownerID, id, item, input, expectedVersion)
}

func maxCategoryVersion(version int64) int64 {
	if version < 1 {
		return 1
	}
	return version
}

func (u *CategoryInteractor) updateValidated(ownerID, id string, item *entity.Category, input CategoryInput, expectedVersion int64) (*entity.Category, error) {
	item.Name, item.Kind, item.ParentID, item.IconKey = input.Name, input.Kind, input.ParentID, input.IconKey
	if versioned, ok := u.repository.(categoryrepo.VersionedCategoryRepository); ok {
		if err := versioned.UpdateWithVersion(ownerID, id, expectedVersion, item, input.WalletIDs); err != nil {
			return nil, err
		}
		return item, nil
	}
	if err := u.repository.Update(ownerID, item); err != nil {
		return nil, err
	}
	if err := u.repository.ReplaceWallets(ownerID, item, input.WalletIDs); err != nil {
		return nil, err
	}
	return item, nil
}

func (u *CategoryInteractor) UpdateWallets(ownerID, id string, walletIDs []string) (*entity.Category, error) {
	item, err := u.repository.FindVisible(ownerID, id)
	if err != nil {
		return nil, err
	}
	walletIDs = uniqueWalletIDs(walletIDs)
	if err := u.repository.ValidateWallets(ownerID, walletIDs); err != nil {
		return nil, err
	}
	if err := u.repository.ReplaceWallets(ownerID, item, walletIDs); err != nil {
		return nil, err
	}
	return item, nil
}

func (u *CategoryInteractor) Delete(ownerID, id string) error {
	if _, err := u.repository.FindPersonal(ownerID, id); err != nil {
		return err
	}
	return u.repository.Delete(ownerID, id)
}

func (u *CategoryInteractor) validateInput(ownerID, selfID string, input CategoryInput) (CategoryInput, error) {
	input.Name, input.Kind = strings.TrimSpace(input.Name), strings.TrimSpace(input.Kind)
	input.IconKey = strings.TrimSpace(input.IconKey)
	if input.IconKey == "" {
		input.IconKey = "tag"
	}
	input.WalletIDs = uniqueWalletIDs(input.WalletIDs)
	if input.Name == "" {
		return input, ErrCategoryNameRequired
	}
	if input.Kind == "" {
		input.Kind = CategoryKindExpense
	}
	if input.Kind != CategoryKindExpense && input.Kind != CategoryKindIncome && input.Kind != CategoryKindDebt {
		return input, ErrCategoryKindInvalid
	}
	if input.ParentID == nil || strings.TrimSpace(*input.ParentID) == "" {
		input.ParentID = nil
		return input, nil
	}
	parentID := strings.TrimSpace(*input.ParentID)
	if parentID == selfID {
		return input, ErrCategoryParentInvalid
	}
	items, err := u.repository.ListVisible(ownerID)
	if err != nil {
		return input, err
	}
	for _, item := range items {
		if item.ID != parentID {
			continue
		}
		if item.ParentID != nil || item.Kind != input.Kind {
			return input, ErrCategoryParentInvalid
		}
		input.ParentID = &parentID
		return input, nil
	}
	return input, ErrCategoryParentInvalid
}

func uniqueWalletIDs(walletIDs []string) []string {
	seen := make(map[string]bool, len(walletIDs))
	result := make([]string, 0, len(walletIDs))
	for _, walletID := range walletIDs {
		walletID = strings.TrimSpace(walletID)
		if walletID != "" && !seen[walletID] {
			seen[walletID] = true
			result = append(result, walletID)
		}
	}
	return result
}
