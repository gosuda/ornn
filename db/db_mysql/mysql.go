package db_mysql

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gosuda/ornn/db"
)

func Dsn(id, pw, addr, port, dbName string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?multiStatements=true", id, pw, addr, port, dbName)
}

func New(dsn, dbName string) (*db.Conn, error) {
	db := &db.Conn{}
	err := db.Connect("mysql", dsn, dbName)
	if err != nil {
		return nil, err
	}
	return db, nil
}
