package dao

import "errors"

var (
	ErrDuplicateRecord = errors.New("Duplicate ID Found")
	ErrInvalidObject   = errors.New("Invalid Object, ID Empty or NULL")
)

const (
	DATABASE_NAME = "TekionDB"
)
