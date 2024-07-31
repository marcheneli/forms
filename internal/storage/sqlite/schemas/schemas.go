package schemas

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
	const op = "storage.sqlite.schemas.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS schemas(
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS idx_schema_name ON schemas(name);
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

func (s *Storage) Create(name string) (int64, error) {
	const op = "storage.sqlite.schemas.Create"

	stmt, err := s.db.Prepare("INSERT INTO schemas(name) VALUES(?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.Exec(name)
	if err != nil {

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", op, err)
	}

	return id, nil
}

func (s *Storage) Update(newName string, schemaId int) error {
	const op = "storage.sqlite.schemas.Update"

	stmt, err := s.db.Prepare("UPDATE schemas SET name = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, updateErr := stmt.Exec(newName, schemaId)
	if err != nil {

		return fmt.Errorf("%s: %w", op, updateErr)
	}

	return nil
}

func (s *Storage) GetSchemaName(id int) (string, error) {
	const op = "storage.sqlite.schemas.Get"

	stmt, err := s.db.Prepare("SELECT name FROM schemas WHERE id = ?")
	if err != nil {
		return "", fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	var schemaName string

	err = stmt.QueryRow(id).Scan(&schemaName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrSchemaNotFound
		}

		return "", fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return schemaName, nil
}

func (s *Storage) Delete(id int) error {
	const op = "storage.sqlite.schemas.Delete"

	deleteFieldsStmt, err := s.db.Prepare("DELETE FROM fields WHERE schema_id = ?")
	if err != nil {
		return fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	_, err = deleteFieldsStmt.Exec(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrSchemaNotFound
		}

		return fmt.Errorf("%s: execute statement: %w", op, err)
	}

	stmt, err := s.db.Prepare("DELETE FROM schemas WHERE id = ?")
	if err != nil {
		return fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	_, err = stmt.Exec(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrSchemaNotFound
		}

		return fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return nil
}
