package database

type Database struct {
}

func New() (*Database, error) {
	db := &Database{}
	return db, nil
}
