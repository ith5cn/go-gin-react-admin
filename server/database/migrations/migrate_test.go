package migrations

import (
	"errors"
	"testing"

	"server/internal/testutil"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/gorm"
)

func TestHistoryValidation(t *testing.T) {
	m := registry[0]
	valid := Record{Version: m.version, Name: m.name, Checksum: m.checksum}
	dirty, changed, newer := valid, valid, valid
	dirty.Dirty = true
	changed.Checksum = "changed"
	newer.Version = len(registry) + 1
	for _, tc := range []struct {
		name    string
		records []Record
		count   int
		bad     bool
	}{
		{"empty", nil, len(registry), false}, {"upgrade", []Record{valid}, len(registry) - 1, false},
		{"dirty", []Record{dirty}, 0, true}, {"modified", []Record{changed}, 0, true},
		{"newer", []Record{newer}, 0, true}, {"extra", []Record{valid, newer}, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := pending(tc.records)
			if (err != nil) != tc.bad || (!tc.bad && len(result) != tc.count) {
				t.Fatalf("pending = %v, %v", result, err)
			}
		})
	}
}

func expectLedger(mock sqlmock.Sqlmock, rows *sqlmock.Rows) {
	// MySQL's HasTable probes the database name before information_schema.
	mock.ExpectQuery("SELECT DATABASE").WillReturnRows(sqlmock.NewRows([]string{"db"}).AddRow("test"))
	mock.ExpectQuery("SELECT SCHEMA_NAME").WillReturnRows(sqlmock.NewRows([]string{"SCHEMA_NAME"}).AddRow("test"))
	mock.ExpectQuery("SELECT count.*information_schema.tables").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT .*ai_schema_migrations.*ORDER BY version ASC").WillReturnRows(rows)
}

func TestCheckIsReadOnly(t *testing.T) {
	db, mock := testutil.DB(t)
	rows := sqlmock.NewRows([]string{"version", "name", "checksum", "dirty"})
	for _, m := range registry {
		rows.AddRow(m.version, m.name, m.checksum, false)
	}
	expectLedger(mock, rows)
	if err := Check(db); err != nil {
		t.Fatal(err)
	}
}

func TestFailedMigrationLeavesDirtyAndReleasesLock(t *testing.T) {
	db, mock := testutil.DB(t)
	original := registry
	t.Cleanup(func() { registry = original })
	failure := errors.New("simulated DDL failure")
	registry = []migration{{1, "test", "checksum", func(*gorm.DB) error { return failure }}}
	mock.ExpectQuery("SELECT CONCAT").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("lock"))
	mock.ExpectQuery("SELECT GET_LOCK").WithArgs("lock").WillReturnRows(sqlmock.NewRows([]string{"locked"}).AddRow(1))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS ai_schema_migrations").WillReturnResult(sqlmock.NewResult(0, 0))
	expectLedger(mock, sqlmock.NewRows([]string{"version", "name", "checksum", "dirty"}))
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `ai_schema_migrations`").WithArgs("test", "checksum", true, 1).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectExec("SELECT RELEASE_LOCK").WithArgs("lock").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := Up(db); !errors.Is(err, failure) {
		t.Fatalf("Up error = %v", err)
	}
}
