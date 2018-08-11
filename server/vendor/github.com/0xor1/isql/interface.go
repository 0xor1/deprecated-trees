package isql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"reflect"
	"github.com/0xor1/panic"
	"time"
)

func NewOpener() Opener {
	return &opener{}
}

type Opener interface {
	Open(driverName, dataSourceName string) (DB, error)
}

func NewDB(db *sql.DB) DB {
	if db == nil {
		return nil
	}
	return &dbWrapper{
		db: db,
	}
}

type DB interface {
	DBCore
	Begin() (Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error)
	Close() error
	Driver() driver.Driver
	Exec(query string, args ...interface{}) (sql.Result, error)
	Ping() error
	PingContext(ctx context.Context) error
	Prepare(query string) (Stmt, error)
	PrepareContext(ctx context.Context, query string) (Stmt, error)
	Query(query string, args ...interface{}) (Rows, error)
	QueryRow(query string, args ...interface{}) Row
	SetConnMaxLifetime(d time.Duration)
	SetMaxIdleConns(n int)
	SetMaxOpenConns(n int)
	Stats() sql.DBStats
}

func NewReplicaSet(driverName, primaryDataSourceName string, slaveDataSourceNames ...string) (ReplicaSet, error) {
	op := &opener{}
	primary, err := op.Open(driverName, primaryDataSourceName)
	if err != nil {
		return nil, err
	}
	rs := &replicaSet{
		primary: primary,
		slaves:  make([]DBCore, 0, len(slaveDataSourceNames)),
	}
	for _, slaveDataSourceName := range slaveDataSourceNames {
		slave, err := op.Open(driverName, slaveDataSourceName)
		if err != nil {
			return nil, err
		}
		rs.slaves = append(rs.slaves, slave)
	}
	return rs, nil
}

func MustNewReplicaSet(driverName, primaryDataSourceName string, slaveDataSourceNames ...string) ReplicaSet {
	rs, err := NewReplicaSet(driverName, primaryDataSourceName, slaveDataSourceNames...)
	panic.IfNotNil(err)
	return rs
}

type DBCore interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) Row
}

type ReplicaSet interface {
	DBCore
	Primary() DBCore
	Slaves() []DBCore
}

func NewRow(row *sql.Row) Row {
	if row == nil {
		return nil
	}
	return &rowWrapper{
		row: row,
	}
}

type Row interface {
	Scan(dest ...interface{}) error
}

func NewRows(rows *sql.Rows) Rows {
	if rows == nil {
		return nil
	}
	return &rowsWrapper{
		rows: rows,
	}
}

type Rows interface {
	Close() error
	ColumnTypes() ([]ColumnType, error)
	Columns() ([]string, error)
	Err() error
	Next() bool
	NextResultSet() bool
	Scan(dest ...interface{}) error
}

func NewStmt(stmt *sql.Stmt) Stmt {
	if stmt == nil {
		return nil
	}
	return &stmtWrapper{
		stmt: stmt,
	}
}

type Stmt interface {
	Close() error
	Exec(args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error)
	Query(args ...interface{}) (Rows, error)
	QueryContext(ctx context.Context, args ...interface{}) (Rows, error)
	QueryRow(args ...interface{}) Row
	QueryRowContext(ctx context.Context, args ...interface{}) Row
}

func NewTx(tx *sql.Tx) Tx {
	if tx == nil {
		return nil
	}
	return &txWrapper{
		tx: tx,
	}
}

type Tx interface {
	DBCore
	Commit() error
	Exec(query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (Stmt, error)
	PrepareContext(ctx context.Context, query string) (Stmt, error)
	Query(query string, args ...interface{}) (Rows, error)
	QueryRow(query string, args ...interface{}) Row
	Rollback() error
	Stmt(stmt *sql.Stmt) Stmt
	StmtContext(ctx context.Context, stmt *sql.Stmt) Stmt
}

func NewColumnType(columnType *sql.ColumnType) ColumnType {
	if columnType == nil {
		return nil
	}
	return &columnTypeWrapper{
		columnType: columnType,
	}
}

type ColumnType interface {
	DatabaseTypeName() string
	DecimalSize() (precision, scale int64, ok bool)
	Length() (length int64, ok bool)
	Name() string
	Nullable() (nullable, ok bool)
	ScanType() reflect.Type
}
