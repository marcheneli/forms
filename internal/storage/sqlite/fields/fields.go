package fields

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/mattn/go-sqlite3"

	"github.com/marcheneli/forms/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.fields.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS fields(
		id INTEGER PRIMARY KEY,
		schema_id INTEGER REFERENCES schemas(id),
		name TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS idx_schema_name ON fields(name);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Create(name string, schemaId int) (int64, error) {
	const op = "storage.sqlite.CreateField"

	stmt, err := s.db.Prepare("INSERT INTO fields(name, schema_id) VALUES(?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.Exec(name, schemaId)
	if err != nil {

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", op, err)
	}

	return id, nil
}

func (s *Storage) Update(newName string, id int) error {
	const op = "storage.sqlite.fields.Update"

	stmt, err := s.db.Prepare("UPDATE fields SET name = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, updateErr := stmt.Exec(newName, id)
	if err != nil {

		return fmt.Errorf("%s: %w", op, updateErr)
	}

	return nil
}

func (s *Storage) Delete(id int) error {
	const op = "storage.sqlite.DeleteField"

	stmt, err := s.db.Prepare("DELETE FROM fields WHERE id = ?")
	if err != nil {
		return fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	_, err = stmt.Exec(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrFieldNotFound
		}

		return fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return nil
}
