package usecase

import (
	"errors"
	"testing"

	"github.com/mypocket/backend/internal/entity"
)

type categoryRepositoryStub struct {
	items   []entity.Category
	created *entity.Category
	wallets map[string]bool
}

func (s *categoryRepositoryStub) EnsurePersonalDefaults(string) error { return nil }

func (s *categoryRepositoryStub) ListVisible(string) ([]entity.Category, error) { return s.items, nil }
func (s *categoryRepositoryStub) FindVisible(_ string, id string) (*entity.Category, error) {
	for index := range s.items {
		if s.items[index].ID == id {
			return &s.items[index], nil
		}
	}
	return nil, errors.New("not found")
}
func (s *categoryRepositoryStub) FindPersonal(_ string, id string) (*entity.Category, error) {
	for index := range s.items {
		if s.items[index].ID == id && !s.items[index].IsSystem {
			return &s.items[index], nil
		}
	}
	return nil, errors.New("not found")
}
func (s *categoryRepositoryStub) Create(_ string, item *entity.Category) error {
	s.created = item
	s.items = append(s.items, *item)
	return nil
}
func (s *categoryRepositoryStub) Update(_ string, item *entity.Category) error {
	for index := range s.items {
		if s.items[index].ID == item.ID {
			s.items[index] = *item
			return nil
		}
	}
	return errors.New("not found")
}
func (s *categoryRepositoryStub) HasChildren(_ string, id string) (bool, error) {
	for _, item := range s.items {
		if item.ParentID != nil && *item.ParentID == id {
			return true, nil
		}
	}
	return false, nil
}
func (s *categoryRepositoryStub) Delete(_ string, id string) error {
	if err := s.DetachChildren("", id); err != nil {
		return err
	}
	for index := range s.items {
		if s.items[index].ID == id {
			s.items = append(s.items[:index], s.items[index+1:]...)
			return nil
		}
	}
	return errors.New("not found")
}
func (s *categoryRepositoryStub) DetachChildren(_ string, id string) error {
	for index := range s.items {
		if s.items[index].ParentID != nil && *s.items[index].ParentID == id {
			s.items[index].ParentID = nil
		}
	}
	return nil
}
func (s *categoryRepositoryStub) ReplaceWallets(_ string, item *entity.Category, walletIDs []string) error {
	for _, walletID := range walletIDs {
		if !s.wallets[walletID] {
			return ErrCategoryWalletInvalid
		}
	}
	item.WalletIDs = walletIDs
	return nil
}
func (s *categoryRepositoryStub) ValidateWallets(_ string, walletIDs []string) error {
	for _, walletID := range walletIDs {
		if !s.wallets[walletID] {
			return ErrCategoryWalletInvalid
		}
	}
	return nil
}

func TestCategoryCreateRejectsThirdLevelParent(t *testing.T) {
	rootID, childID := "root", "child"
	stub := &categoryRepositoryStub{items: []entity.Category{{ID: rootID, Kind: CategoryKindExpense}, {ID: childID, Kind: CategoryKindExpense, ParentID: &rootID}}}
	_, err := NewCategoryInteractor(stub).Create("owner", CategoryInput{Name: "grandchild", Kind: CategoryKindExpense, ParentID: &childID})
	if !errors.Is(err, ErrCategoryParentInvalid) {
		t.Fatalf("error = %v", err)
	}
}
func TestCategoryUpdateRejectsSelfParent(t *testing.T) {
	itemID := "personal"
	stub := &categoryRepositoryStub{items: []entity.Category{{ID: itemID, Kind: CategoryKindExpense}}}
	_, err := NewCategoryInteractor(stub).Update("owner", itemID, CategoryInput{Name: "self", Kind: CategoryKindExpense, ParentID: &itemID})
	if !errors.Is(err, ErrCategoryParentInvalid) {
		t.Fatalf("error = %v", err)
	}
}
func TestCategoryDeleteDetachesChildrenOfPersonalParent(t *testing.T) {
	parentID := "personal"
	stub := &categoryRepositoryStub{items: []entity.Category{{ID: parentID, Kind: CategoryKindExpense}, {ID: "child", Kind: CategoryKindExpense, ParentID: &parentID}}}
	err := NewCategoryInteractor(stub).Delete("owner", parentID)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if stub.items[0].ID != "child" || stub.items[0].ParentID != nil {
		t.Fatalf("child parent = %v, want nil", stub.items[0].ParentID)
	}
}

func TestCategoryCreateRejectsWalletOutsideOwnerScope(t *testing.T) {
	stub := &categoryRepositoryStub{wallets: map[string]bool{"owner-wallet": true}}
	_, err := NewCategoryInteractor(stub).Create("owner", CategoryInput{Name: "Cá nhân", Kind: CategoryKindExpense, WalletIDs: []string{"other-user-wallet"}})
	if !errors.Is(err, ErrCategoryWalletInvalid) {
		t.Fatalf("error = %v", err)
	}
}

func TestCategoryUpdateWalletsAllowsSystemCategory(t *testing.T) {
	stub := &categoryRepositoryStub{items: []entity.Category{{ID: "system", IsSystem: true}}, wallets: map[string]bool{"wallet": true}}
	item, err := NewCategoryInteractor(stub).UpdateWallets("owner", "system", []string{"wallet"})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if len(item.WalletIDs) != 1 || item.WalletIDs[0] != "wallet" {
		t.Fatalf("wallet ids = %v, want [wallet]", item.WalletIDs)
	}
}
