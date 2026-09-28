// Package migrations contains immutable, explicitly executed database upgrades.
package migrations

import (
	"crypto/sha256"
	_ "embed"
	"fmt"

	"gorm.io/gorm"
)

//go:embed 0001_legacy.go
var legacySource []byte

type migration struct {
	version  int
	name     string
	checksum string
	up       func(*gorm.DB) error
}

//go:embed 0002_article_codegen_options.go
var articleOptionsSource []byte

var registry = []migration{
	{1, "legacy_schema_and_seeds", fmt.Sprintf("%x", sha256.Sum256(legacySource)), upgradeLegacySchema},
	{2, "article_codegen_options", fmt.Sprintf("%x", sha256.Sum256(articleOptionsSource)), repairArticleCodegenOptions},
}

// Record is the persistent migration ledger. Dirty records block startup and upgrades.
type Record struct {
	Version  int `gorm:"primaryKey"`
	Name     string
	Checksum string
	Dirty    bool
}

// TableName isolates template migrations from business-owned migration tables.
func (Record) TableName() string { return "ai_schema_migrations" }

const ledgerDDL = `CREATE TABLE IF NOT EXISTS ai_schema_migrations (
 version int NOT NULL PRIMARY KEY,
 name varchar(200) NOT NULL,
 checksum char(64) NOT NULL,
 dirty boolean NOT NULL,
 applied_at datetime NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`

// Status reads the ledger without creating tables or modifying data.
func Status(db *gorm.DB) ([]Record, error) {
	if !db.Migrator().HasTable(&Record{}) {
		return nil, nil
	}
	var records []Record
	err := db.Order("version ASC").Find(&records).Error
	return records, err
}

func pending(records []Record) ([]migration, error) {
	if len(records) > len(registry) {
		return nil, fmt.Errorf("database version is newer than this binary")
	}
	for i, record := range records {
		m := registry[i]
		if record.Version != m.version || record.Name != m.name || record.Checksum != m.checksum {
			return nil, fmt.Errorf("migration %d history/checksum mismatch", record.Version)
		}
		if record.Dirty {
			return nil, fmt.Errorf("migration %d is dirty; inspect and restore the database before retrying", record.Version)
		}
	}
	return registry[len(records):], nil
}

// Check fails closed on pending, dirty, altered or newer migrations; it never writes.
func Check(db *gorm.DB) error {
	records, err := Status(db)
	if err != nil {
		return err
	}
	remaining, err := pending(records)
	if err != nil {
		return err
	}
	if len(remaining) != 0 {
		return fmt.Errorf("database migration required: run migrate status, then migrate up")
	}
	return nil
}

// Plan returns the reviewable ordered upgrade list without applying it.
func Plan(db *gorm.DB) ([]string, error) {
	records, err := Status(db)
	if err != nil {
		return nil, err
	}
	remaining, err := pending(records)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(remaining))
	for _, m := range remaining {
		result = append(result, fmt.Sprintf("%04d %s sha256=%s", m.version, m.name, m.checksum))
	}
	return result, nil
}

// Up applies pending upgrades on one pinned MySQL connection under an advisory lock.
// MySQL DDL implicitly commits: a failure deliberately leaves a dirty record.
func Up(db *gorm.DB) error {
	return db.Connection(func(conn *gorm.DB) error {
		// Connection pins the pool but returns a reusable statement. Reset it between queries.
		conn = conn.Session(&gorm.Session{NewDB: true})
		var lockName string
		if err := conn.Raw("SELECT CONCAT('gra:migrate:', LEFT(SHA2(DATABASE(), 256), 40))").Scan(&lockName).Error; err != nil {
			return err
		}
		var acquired int
		if err := conn.Raw("SELECT GET_LOCK(?, 0)", lockName).Scan(&acquired).Error; err != nil {
			return err
		}
		if acquired != 1 {
			return fmt.Errorf("another database migration is running")
		}
		defer conn.Exec("SELECT RELEASE_LOCK(?)", lockName)
		if err := conn.Exec(ledgerDDL).Error; err != nil {
			return err
		}
		records, err := Status(conn)
		if err != nil {
			return err
		}
		remaining, err := pending(records)
		if err != nil {
			return err
		}
		for _, m := range remaining {
			record := Record{Version: m.version, Name: m.name, Checksum: m.checksum, Dirty: true}
			if err := conn.Create(&record).Error; err != nil {
				return err
			}
			if err := m.up(conn); err != nil {
				return fmt.Errorf("migration %d failed (dirty): %w", m.version, err)
			}
			if err := conn.Model(&record).Updates(map[string]interface{}{"dirty": false, "applied_at": gorm.Expr("NOW()")}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
