package db_postgres

import (
	"fmt"

	"github.com/gosuda/ornn/db"
	_ "github.com/lib/pq"
)

func Dsn(host, port, id, pw, dbName string) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, id, pw, dbName)
}

func New(addr, port, id, pw, dbName string) (*db.Conn, error) {
	conn := &db.Conn{}
	err := conn.Connect("postgres", Dsn(addr, port, id, pw, dbName), dbName)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
