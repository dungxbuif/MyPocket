package usecase

import (
	"errors"
	"github.com/mypocket/backend/internal/entity"
	"strings"
	"time"
)

func ValidateAIEntryDraft(d entity.AIEntryDraft, wallet entity.Wallet, category *entity.Category) error {
	if d.Type != entity.TransactionTypeIncome && d.Type != entity.TransactionTypeExpense {
		return errors.New("chỉ duyệt thu/chi; chuyển nội bộ cần luồng hai ví riêng")
	}
	if d.Amount <= 0 || d.Amount > 9007199254740991 {
		return errors.New("số tiền phải là số nguyên đồng dương trong giới hạn cho phép")
	}
	if d.WalletID == "" || wallet.ID != d.WalletID || (wallet.Type != entity.WalletTypeBasic && wallet.Type != entity.WalletTypeGoal) {
		return errors.New("chọn ví thường hoặc ví tiết kiệm hợp lệ")
	}
	if _, err := time.Parse(time.RFC3339, d.OccurredAt); err != nil {
		return errors.New("cần ngày giờ giao dịch có múi giờ rõ ràng")
	}
	if len([]rune(d.Note)) > 2000 {
		return errors.New("ghi chú tối đa 2000 ký tự")
	}
	if d.CategoryID == nil || strings.TrimSpace(*d.CategoryID) == "" {
		return nil
	}
	if category == nil || category.ID != *d.CategoryID || category.Kind != d.Type {
		return errors.New("nhóm không phù hợp giao dịch")
	}
	if wallet.Type == entity.WalletTypeGoal {
		if category.SystemKey == nil {
			return errors.New("chọn nhóm dành cho ví tiết kiệm")
		}
		key := *category.SystemKey
		if !(d.Type == entity.TransactionTypeIncome && (key == "income_transfer_in" || key == "income_interest") || d.Type == entity.TransactionTypeExpense && key == "expense_transfer_out") {
			return errors.New("chọn nhóm dành cho ví tiết kiệm")
		}
	}
	if len(category.WalletIDs) > 0 {
		for _, id := range category.WalletIDs {
			if id == wallet.ID {
				return nil
			}
		}
		return errors.New("nhóm không áp dụng cho ví đã chọn")
	}
	return nil
}
