package migrate_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"strings"
	"sync"
)

var _drvr *FakeDriver

func init() {
	_drvr = &FakeDriver{mux: &sync.Mutex{}}
	sql.Register("fake", _drvr)
}

func OpenFakeDB() (*sql.DB, *FakeDriver, error) {
	db, err := sql.Open("fake", "")
	return db, _drvr, err
}

type FakeDriver struct {
	mux        *sync.Mutex
	statements []FakeStmt
}

func (f *FakeDriver) String() string {
	b := &strings.Builder{}
	b.WriteString(fmt.Sprintf("Driver(%d)\n", len(f.statements)))
	for i, stmt := range f.statements {
		b.WriteString(fmt.Sprintf("\t%d: %s\n", i, stmt.String()))
	}
	return b.String()
}

func (f *FakeDriver) Commit() error {
	return nil
}

func (f *FakeDriver) Rollback() error {
	return nil
}

func (f *FakeDriver) Prepare(query string) (driver.Stmt, error) {
	f.mux.Lock()
	defer f.mux.Unlock()
	f.statements = append(f.statements, FakeStmt{query: query})
	return &f.statements[len(f.statements)-1], nil
}

func (f *FakeDriver) Close() error {
	return nil
}

func (f *FakeDriver) Begin() (driver.Tx, error) {
	return f, nil
}

func (f *FakeDriver) Open(name string) (driver.Conn, error) {
	return f.Connect(context.Background())
}

func (f *FakeDriver) OpenConnector(name string) (driver.Connector, error) {
	return f, nil
}

func (f *FakeDriver) Connect(ctx context.Context) (driver.Conn, error) {
	return f, nil
}

func (f *FakeDriver) Driver() driver.Driver {
	return f
}

type FakeStmt struct {
	executed bool
	query    string
	args     []driver.Value
}

func (f *FakeStmt) String() string {
	return fmt.Sprintf("Stmt(%t), query: %s, args: %v}", f.executed, f.query, f.args)
}

func (f *FakeStmt) Close() error {
	return nil
}

func (f *FakeStmt) NumInput() int {
	return strings.Count(f.query, "?")
}

func (f *FakeStmt) Exec(args []driver.Value) (driver.Result, error) {
	if f.executed {
		panic("executed twice")
	}
	f.executed = true
	f.args = args
	return &FakeResult{}, nil
}

func (f *FakeStmt) Query(args []driver.Value) (driver.Rows, error) {
	if f.executed {
		panic("executed twice")
	}
	f.executed = true
	f.args = args
	return FakeRows{}, nil
}

type FakeResult struct {
}

func (f *FakeResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (f *FakeResult) RowsAffected() (int64, error) {
	return 0, nil
}

type FakeRows struct {
	idx  int
	vals []string
}

func (f FakeRows) Columns() []string {
	return []string{"version"}
}

func (f FakeRows) Close() error {
	return nil
}

func (f FakeRows) Next(dest []driver.Value) error {
	if f.idx >= len(f.vals) {
		return io.EOF
	}
	dest[0] = f.vals[f.idx]
	f.idx++
	return nil
}
