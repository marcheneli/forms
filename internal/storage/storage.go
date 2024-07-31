package storage

import "errors"

var (
	ErrSchemaNotFound = errors.New("schema not found")
	ErrFieldNotFound  = errors.New("field not found")
)
