package db_access

import (
	"database/sql"
	"context"
	"container/list"
	"errors"

	"github.com/tori209/data-executor/log/format"
	"github.com/tori209/data-executor/log/dao"
	"github.com/tori209/data-executor/log/util"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

var ErrInvalidElement = errors.New("Invalid Argument Given")

type PostgresDbOpt struct {
	DSN		string
	Ctx		context.Context
}

type PostgresQueryRunner struct {
	sqldb	*sql.DB
	bundb	*bun.DB
	ctx		context.Context
}

func NewPostgresQueryRunner(opt *PostgresDbOpt) (*PostgresQueryRunner) {
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(opt.DSN)))
	db := bun.NewDB(sqldb, pgdialect.New())
	return &PostgresQueryRunner{
		sqldb:	sqldb,
		bundb:	db,
		ctx:	opt.Ctx,
	}
}

// DB 초기화
func (pqr *PostgresQueryRunner) Init() error {
	if _, err := 
		pqr.bundb.NewCreateTable().
		Model((*dao.TaskDao)(nil)).
		IfNotExists().
		Exec(pqr.ctx); err != nil {
		return err
	}
	if _, err := pqr.bundb.NewCreateTable().
		Model((*dao.JobDao)(nil)).
		IfNotExists().
		Exec(pqr.ctx); err != nil {
		return err
	}
	if _, err := pqr.bundb.NewCreateTable().
		Model((*dao.LogDao)(nil)).
		IfNotExists().
		Exec(pqr.ctx); err != nil {
		return err
	}
	return nil
}

func (pqr *PostgresQueryRunner) SaveJob(tr *format.TaskRequestMessage) error {
	jd := util.TaskRequestToJobDao(tr)
	_, err := pqr.bundb.NewInsert().Model(&jd).Exec(pqr.ctx)
	return err
}

func (pqr *PostgresQueryRunner) SaveJobs(trs *[]format.TaskRequestMessage) error {
	jdl := make([]dao.JobDao, 0)
	for _, tr := range *trs {
		jdl = append(jdl, util.TaskRequestToJobDao(&tr))
	}
	_, err := pqr.bundb.NewInsert().Model(&jdl).Exec(pqr.ctx)
	return err
}

func (pqr *PostgresQueryRunner) SaveTask(tr *format.TaskRequestMessage) error {
	td := util.TaskRequestToTaskDao(tr)
	_, err := pqr.bundb.NewInsert().Model(&td).Exec(pqr.ctx)
	return err
}

func (pqr *PostgresQueryRunner) SaveTasks(trs *[]format.TaskRequestMessage) error {
	tdl := make([]dao.TaskDao, 0)
	for _, tr := range *trs {
		tdl = append(tdl, util.TaskRequestToTaskDao(&tr))
	}

	_, err := pqr.bundb.NewInsert().Model(&tdl).Exec(pqr.ctx)
	return err
}

func (pqr *PostgresQueryRunner) SaveTasksFromList(trs *list.List) error {
	tdl := make([]dao.TaskDao, 0)
	for wrapped_tr := trs.Front(); wrapped_tr != nil; wrapped_tr = wrapped_tr.Next() {
		tr, success := wrapped_tr.Value.(*format.TaskRequestMessage)
		if !success {
			return ErrInvalidElement
		}
		tdl = append(tdl, util.TaskRequestToTaskDao(tr))
	}

	_, err := pqr.bundb.NewInsert().Model(&tdl).Exec(pqr.ctx)
	return err
}

func (pqr *PostgresQueryRunner) SaveLogs(msgl *[]format.L4Message) error {
	ldl := util.MultipleL4MessageToLogDao(msgl)
	if len(ldl) > 0 {
		_, err := pqr.bundb.NewInsert().Model(&ldl).Exec(pqr.ctx)
		return err
	}
	return nil
}
