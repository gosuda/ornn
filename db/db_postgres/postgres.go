package db_postgres

import (
	"fmt"

	"github.com/gosuda/ornn/db"
	_ "github.com/lib/pq"
)

func Dsn(host, port, id, pw, dbName string) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, id, pw, dbName)
}

func New(dsn, dbName string) (*db.Conn, error) {
	conn := &db.Conn{}
	err := conn.Connect("postgres", dsn, dbName)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
