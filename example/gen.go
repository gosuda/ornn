// Code generated - DO NOT EDIT.
// This file is a generated and any changes will be lost.

package gen

import (
	"fmt"

	. "github.com/gosuda/ornn/db"
)

type Gen struct {
	User User
}

func (t *Gen) Init(
	job *Job,
) {
	t.User.Init(job)
}

func (t *User) Init(
	job *Job,
) {
	t.job = job
}

type User struct {
	job *Job
}

func (t *User) Update(
	set_seq uint64,
	set_id string,
	set_ord *int64,
	set_name *string,
	set_pw []byte,
	where_seq uint64,
) (
	rowAffected int64,
	err error,
) {
	sql := fmt.Sprintf(
		"UPDATE user SET seq = ?, id = ?, ord = ?, name = ?, pw = ? WHERE seq = ?",
	)
	args := []interface{}{
		set_seq,
		set_id,
		set_ord,
		set_name,
		set_pw,
		where_seq,
	}

	exec, err := t.job.Exec(
		sql,
		args...,
	)
	if err != nil {
		return 0, err
	}

	return exec.RowsAffected()
}

func (t *User) Insert(
	val_seq uint64,
	val_id string,
	val_ord *int64,
	val_name *string,
	val_pw []byte,
) (
	lastInsertId int64,
	err error,
) {
	args := []interface{}{
		val_seq,
		val_id,
		val_ord,
		val_name,
		val_pw,
	}

	sql := fmt.Sprintf(
		"INSERT INTO user VALUES (?, ?, ?, ?, ?)",
	)

	exec, err := t.job.Exec(
		sql,
		args...,
	)
	if err != nil {
		return 0, err
	}

	return exec.LastInsertId()
}

type User_select struct {
	Seq  uint64
	Id   string
	Ord  *int64
	Name *string
	Pw   []byte
}

func (t *User) Select(
	where_seq uint64,
) (
	selects []*User_select,
	err error,
) {
	args := []interface{}{
		where_seq,
	}

	sql := fmt.Sprintf(
		"SELECT * FROM user WHERE seq = ?",
	)
	ret, err := t.job.Query(
		sql,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer ret.Close()

	selects = make([]*User_select, 0, 100)
	for ret.Next() {
		scan := &User_select{}
		err := ret.Scan(scan)
		if err != nil {
			return nil, err
		}
		selects = append(selects, scan)
	}

	return selects, nil
}

func (t *User) Delete(
	where_seq uint64,
) (
	rowAffected int64,
	err error,
) {
	args := []interface{}{
		where_seq,
	}

	sql := fmt.Sprintf(
		"DELETE FROM user WHERE seq = ?",
	)

	exec, err := t.job.Exec(
		sql,
		args...,
	)
	if err != nil {
		return 0, err
	}

	return exec.RowsAffected()
}
