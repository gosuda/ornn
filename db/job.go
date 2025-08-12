package db

import (
	"context"
	"database/sql"
	"errors"
)

// db or tx
func NewJob(db *sql.DB) *Job {
	return &Job{db: db}
}

type Job struct {
	db *sql.DB
	tx *sql.Tx
}

func (t *Job) Exec(query string, args ...any) (res sql.Result, err error) {
	if t.tx == nil {
		res, err = t.db.Exec(query, args...)
	} else {
		res, err = t.tx.Exec(query, args...)
	}
	return res, err
}

func (t *Job) Query(query string, args ...any) (rows *sql.Rows, err error) {
	if t.tx == nil {
		rows, err = t.db.Query(query, args...)
	} else {
		rows, err = t.tx.Query(query, args...)
	}
	return rows, err
}

func (t *Job) BeginTx(isoLevel sql.IsolationLevel, readonly bool) error {
	var err error
	t.tx, err = t.db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: isoLevel,
		ReadOnly:  readonly,
	})
	if err != nil {
		return err
	}
	return nil
}

func (t *Job) Commit() error {
	if t.tx == nil {
		return errors.New("not transaction job")
	}
	return t.tx.Commit()
}

func (t *Job) Rollback() error {
	if t.tx == nil {
		return errors.New("not transaction job")
	}
	return t.tx.Rollback()
}
