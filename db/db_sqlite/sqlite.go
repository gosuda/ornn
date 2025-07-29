package db_sqlite

import (
	"errors"
	"os"

	"github.com/gosuda/ornn/db"
	_ "github.com/mattn/go-sqlite3"
)

func New(path string) (*db.Conn, error) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		_, err := os.Create(path)
		if err != nil {
			return nil, err
		}
	}

	conn := &db.Conn{}
	err := conn.Connect("sqlite3", path, "")
	if err != nil {
		return nil, err
	}

	return conn, nil
}
