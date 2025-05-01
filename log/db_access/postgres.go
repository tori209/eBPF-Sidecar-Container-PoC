package db_access

import (
	"database/sql"
	"context"

	"github.com/tori209/data-executor/log/format"
	"github.com/tori209/data-executor/log/dao"
	"github.com/tori209/data-executor/log/util"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

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
	return nil
}

func (pqr *PostgresQueryRunner) SaveTasks(trs *[]format.TaskRequestMessage) error {
	tdl := make([]dao.TaskDao, 0)
	for _, tr := range *trs {
		tdl = append(tdl, util.TaskRequestToTaskDao(&tr))
	}

	_, err := pqr.bundb.NewInsert().Model(&tdl).Exec(pqr.ctx)
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

