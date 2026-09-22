package repository

import (
	"time"

	"github.com/mypocket/backend/internal/entity"
	contract "github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
)

type MonthPostgresRepository struct{ db *gorm.DB }

func NewMonthPostgresRepository(db *gorm.DB) contract.MonthNoteRepository {
	return &MonthPostgresRepository{db: db}
}

func (r *MonthPostgresRepository) Find(owner, monthLabel string) (string, error) {
	month, err := entity.ParseMonth(monthLabel)
	if err != nil {
		return "", contract.ErrJarInvalid
	}
	var note entity.MonthNote
	err = r.db.Where("owner_id = ? AND month = ?", owner, entity.CalendarDate{Time: month}).First(&note).Error
	if err == gorm.ErrRecordNotFound {
		return "", nil
	}
	return note.Note, err
}

func (r *MonthPostgresRepository) Save(owner, monthLabel, noteText string) error {
	month, err := entity.ParseMonth(monthLabel)
	if err != nil {
		return contract.ErrJarInvalid
	}
	date := entity.CalendarDate{Time: month}
	if noteText == "" {
		return r.db.Where("owner_id = ? AND month = ?", owner, date).Delete(&entity.MonthNote{}).Error
	}
	return r.db.Exec(`INSERT INTO month_notes(owner_id, month, note, created_at, updated_at) VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(owner_id, month) DO UPDATE SET note = EXCLUDED.note, updated_at = EXCLUDED.updated_at`, owner, date, noteText, time.Now().UTC(), time.Now().UTC()).Error
}
