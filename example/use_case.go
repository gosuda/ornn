package gen

import (
	"database/sql"

	"github.com/gosuda/ornn/db"
)

func NewLogic(db *db.Conn) *Logic {
	return &Logic{
		db: db,
	}
}

type Logic struct {
	db *db.Conn
}

func (t *Logic) ExampleNoTx() error {
	job := t.db.Job()

	gen := &Gen{}
	gen.Init(job)

	//---------------------- start ----------------------//
	// Execute generated function down here!!
	// _, err = gen.TestGroup.Select()

	//----------------------- end -----------------------//
	return nil
}

func (t *Logic) ExampleTx() error {
	job, err := t.db.TxJob(sql.LevelSerializable, false)
	if err != nil {
		return err
	}

	gen := &Gen{}
	gen.Init(job)

	//---------------------- start ----------------------//
	// Execute generated function down here!!
	// _, err = schema.TestGroup.Select()

	//----------------------- end -----------------------//
	if err != nil {
		job.Rollback()
		return err
	}
	err = job.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (t *Logic) ExampleTxFunc() error {
	return t.db.TxJobFunc(sql.LevelSerializable, false, func(job *db.Job) error {
		gen := &Gen{}
		gen.Init(job)

		//---------------------- start ----------------------//
		// Execute generated function down here!!
		// _, err = schema.TestGroup.Select()

		//----------------------- end -----------------------//

		return nil
	})
}
