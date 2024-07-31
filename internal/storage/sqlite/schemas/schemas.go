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

type Schema struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func (s *Storage) GetList() ([]Schema, error) {
	const op = "storage.sqlite.schemas.Get"

	stmt, err := s.db.Prepare("SELECT id, name FROM schemas")
	if err != nil {
		return nil, fmt.Errorf("%s: failed to prepare statement: %w", op, err)
	}

	queryResp, queryErr := stmt.Query()

	if err != nil {
		return nil, fmt.Errorf("%s: failed to query schema rows: %w", op, queryErr)
	}

	rows := make([]Schema, 0)

	for queryResp.Next() {
		var id int
		var name string

		if err := queryResp.Scan(&id, &name); err != nil {
			return nil, fmt.Errorf("%s: failed to scan schema row: %w", op, err)
		}

		rows = append(rows, Schema{id, name})
	}

	return rows, nil
}

func (s *Storage) Delete(id int) error {
	const op = "storage.sqlite.schemas.Delete"

	deleteStmt, err := s.db.Prepare(`
		BEGIN TRANSACTION;
			DELETE FROM fields WHERE schema_id = ?;
			DELETE FROM schemas WHERE id = ?;
		COMMIT;
	`)
	if err != nil {
		return fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	_, err = deleteStmt.Exec(id, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrSchemaNotFound
		}

		return fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return nil
}
