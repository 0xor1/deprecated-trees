package isql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"math/rand"
	"reflect"
	"time"
)

type opener struct {
}

func (o *opener) Open(driverName, dataSourceName string) (DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		if db != nil {
			db.Close()
		}
		return nil, err
	}
	return NewDB(db), nil
}

type dbWrapper struct {
	db *sql.DB
}

func (d *dbWrapper) Begin() (Tx, error) {
	tx, err := d.db.Begin()
	return NewTx(tx), err
}

func (d *dbWrapper) BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
	tx, err := d.db.BeginTx(ctx, opts)
	return NewTx(tx), err
}

func (d *dbWrapper) Close() error {
	return d.db.Close()
}

func (d *dbWrapper) Driver() driver.Driver {
	return d.db.Driver()
}

func (d *dbWrapper) Exec(query string, args ...interface{}) (sql.Result, error) {
	return d.db.Exec(query, args...)
}

func (d *dbWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return d.db.ExecContext(ctx, query, args...)
}

func (d *dbWrapper) Ping() error {
	return d.db.Ping()
}

func (d *dbWrapper) PingContext(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

func (d *dbWrapper) Prepare(query string) (Stmt, error) {
	stmt, err := d.db.Prepare(query)
	return NewStmt(stmt), err
}

func (d *dbWrapper) PrepareContext(ctx context.Context, query string) (Stmt, error) {
	stmt, err := d.db.PrepareContext(ctx, query)
	return NewStmt(stmt), err
}

func (d *dbWrapper) Query(query string, args ...interface{}) (Rows, error) {
	rows, err := d.db.Query(query, args...)
	return NewRows(rows), err
}

func (d *dbWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	rows, err := d.db.QueryContext(ctx, query, args...)
	return NewRows(rows), err
}

func (d *dbWrapper) QueryRow(query string, args ...interface{}) Row {
	return NewRow(d.db.QueryRow(query, args...))
}

func (d *dbWrapper) QueryRowContext(ctx context.Context, query string, args ...interface{}) Row {
	return NewRow(d.db.QueryRowContext(ctx, query, args...))
}

func (d *dbWrapper) SetConnMaxLifetime(dur time.Duration) {
	d.db.SetConnMaxLifetime(dur)
}

func (d *dbWrapper) SetMaxIdleConns(n int) {
	d.db.SetMaxIdleConns(n)
}

func (d *dbWrapper) SetMaxOpenConns(n int) {
	d.db.SetMaxOpenConns(n)
}

func (d *dbWrapper) Stats() sql.DBStats {
	return d.db.Stats()
}

type replicaSet struct {
	primary DBCore
	slaves  []DBCore
}

func (r *replicaSet) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return r.primary.ExecContext(ctx, query, args...)
}

func (r *replicaSet) QueryContext(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	if len(r.slaves) > 0 {
		return r.slaves[rand.Intn(len(r.slaves))].QueryContext(ctx, query, args...)
	} else {
		return r.primary.QueryContext(ctx, query, args...)
	}
}

func (r *replicaSet) QueryRowContext(ctx context.Context, query string, args ...interface{}) Row {
	if len(r.slaves) > 0 {
		return r.slaves[rand.Intn(len(r.slaves))].QueryRowContext(ctx, query, args...)
	} else {
		return r.primary.QueryRowContext(ctx, query, args...)
	}
}

func (r *replicaSet) Primary() DBCore {
	return r.primary
}

func (r *replicaSet) Slaves() []DBCore {
	return r.slaves
}

type rowWrapper struct {
	row *sql.Row
}

func (r *rowWrapper) Scan(dest ...interface{}) error {
	return r.row.Scan(dest...)
}

type rowsWrapper struct {
	rows *sql.Rows
}

func (r *rowsWrapper) Close() error {
	return r.rows.Close()
}

func (r *rowsWrapper) ColumnTypes() ([]ColumnType, error) {
	columnTypes, err := r.rows.ColumnTypes()
	res := make([]ColumnType, 0, len(columnTypes))
	for _, ct := range columnTypes {
		res = append(res, NewColumnType(ct))
	}
	return res, err
}

func (r *rowsWrapper) Columns() ([]string, error) {
	return r.rows.Columns()
}

func (r *rowsWrapper) Err() error {
	return r.rows.Err()
}

func (r *rowsWrapper) Next() bool {
	return r.rows.Next()
}

func (r *rowsWrapper) NextResultSet() bool {
	return r.rows.NextResultSet()
}

func (r *rowsWrapper) Scan(dest ...interface{}) error {
	return r.rows.Scan(dest...)
}

type stmtWrapper struct {
	stmt *sql.Stmt
}

func (s *stmtWrapper) Close() error {
	return s.stmt.Close()
}

func (s *stmtWrapper) Exec(args ...interface{}) (sql.Result, error) {
	return s.stmt.Exec(args...)
}

func (s *stmtWrapper) ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error) {
	return s.stmt.ExecContext(ctx, args...)
}

func (s *stmtWrapper) Query(args ...interface{}) (Rows, error) {
	rows, err := s.stmt.Query(args...)
	return NewRows(rows), err
}

func (s *stmtWrapper) QueryContext(ctx context.Context, args ...interface{}) (Rows, error) {
	rows, err := s.stmt.QueryContext(ctx, args...)
	return NewRows(rows), err
}

func (s *stmtWrapper) QueryRow(args ...interface{}) Row {
	return NewRow(s.stmt.QueryRow(args...))
}

func (s *stmtWrapper) QueryRowContext(ctx context.Context, args ...interface{}) Row {
	return NewRow(s.stmt.QueryRowContext(ctx, args...))
}

type txWrapper struct {
	tx *sql.Tx
}

func (t *txWrapper) Commit() error {
	return t.tx.Commit()
}

func (t *txWrapper) Exec(query string, args ...interface{}) (sql.Result, error) {
	return t.tx.Exec(query, args...)
}

func (t *txWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

func (t *txWrapper) Prepare(query string) (Stmt, error) {
	stmt, err := t.tx.Prepare(query)
	return NewStmt(stmt), err
}

func (t *txWrapper) PrepareContext(ctx context.Context, query string) (Stmt, error) {
	stmt, err := t.tx.PrepareContext(ctx, query)
	return NewStmt(stmt), err
}

func (t *txWrapper) Query(query string, args ...interface{}) (Rows, error) {
	rows, err := t.tx.Query(query, args...)
	return NewRows(rows), err
}

func (t *txWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	rows, err := t.tx.QueryContext(ctx, query, args...)
	return NewRows(rows), err
}

func (t *txWrapper) QueryRow(query string, args ...interface{}) Row {
	return NewRow(t.tx.QueryRow(query, args...))
}

func (t *txWrapper) QueryRowContext(ctx context.Context, query string, args ...interface{}) Row {
	return NewRow(t.tx.QueryRowContext(ctx, query, args...))
}

func (t *txWrapper) Rollback() error {
	return t.tx.Rollback()
}

func (t *txWrapper) Stmt(stmt *sql.Stmt) Stmt {
	return NewStmt(t.tx.Stmt(stmt))
}

func (t *txWrapper) StmtContext(ctx context.Context, stmt *sql.Stmt) Stmt {
	return NewStmt(t.tx.StmtContext(ctx, stmt))
}

type columnTypeWrapper struct {
	columnType *sql.ColumnType
}

func (c *columnTypeWrapper) DatabaseTypeName() string {
	return c.columnType.DatabaseTypeName()
}

func (c *columnTypeWrapper) DecimalSize() (precision, scale int64, ok bool) {
	return c.columnType.DecimalSize()
}

func (c *columnTypeWrapper) Length() (length int64, ok bool) {
	return c.columnType.Length()
}

func (c *columnTypeWrapper) Name() string {
	return c.columnType.Name()
}

func (c *columnTypeWrapper) Nullable() (nullable, ok bool) {
	return c.columnType.Nullable()
}

func (c *columnTypeWrapper) ScanType() reflect.Type {
	return c.columnType.ScanType()
}
