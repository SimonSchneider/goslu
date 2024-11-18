package migrate_test

import (
	"context"
	"fmt"
	"github.com/SimonSchneider/goslu/migrate"
	"testing"
	"testing/fstest"
)

func TestMigrate(t *testing.T) {
	dir := fstest.MapFS{
		"3.sql": &fstest.MapFile{
			Data: []byte("-- migrate:up\nCREATE TABLE fo2 (id INTEGER PRIMARY KEY);"),
		},
		"2.sql/ignored.sql": &fstest.MapFile{
			Data: []byte("-- migrate:up\nCREATE TABLE bar (id INTEGER PRIMARY KEY);"),
		},
		"1.sql": &fstest.MapFile{
			Data: []byte("-- migrate:up\nCREATE TABLE foo (id INTEGER PRIMARY KEY);"),
		},
	}
	db, drvr, err := OpenFakeDB()
	if err != nil {
		t.Fatal(err)
	}
	if err := migrate.Migrate(context.Background(), dir, db); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%s\n", drvr)
	//for i, stmt := range db.stmts {
	//	t.Logf("stmt %d (tx %s): %s, %v", i, stmt.tx, stmt.query, stmt.args)
	//}
}
