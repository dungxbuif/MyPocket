package repository

type MonthNoteRepository interface {
	Find(ownerID, month string) (string, error)
	Save(ownerID, month, note string) error
}
