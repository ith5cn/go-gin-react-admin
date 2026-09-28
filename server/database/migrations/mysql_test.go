package migrations

import (
	"os"
	"strings"
	"testing"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestMySQLUpgrade is destructive only to the explicitly named disposable codex_ database.
// It is opt-in locally and runs against a throwaway MySQL service in CI.
func TestMySQLUpgrade(t *testing.T) {
	dsn := os.Getenv("MIGRATION_TEST_DSN")
	if dsn == "" {
		t.Skip("set MIGRATION_TEST_DSN to a disposable codex_ database")
	}
	cfg, err := mysqlDriver.ParseDSN(dsn)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(cfg.DBName, "codex_") {
		t.Fatal("integration tests require a disposable codex_ database")
	}
	cfg.MultiStatements = true
	cfg.ParseTime = true
	db, err := gorm.Open(mysql.Open(cfg.FormatDSN()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	baseline, err := os.ReadFile("../ai_system.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, old := range []bool{false, true} {
		t.Run(map[bool]string{false: "fresh", true: "legacy"}[old], func(t *testing.T) {
			if err := db.Exec("DROP TABLE IF EXISTS ai_schema_migrations").Error; err != nil {
				t.Fatal(err)
			}
			if err := db.Exec(string(baseline)).Error; err != nil {
				t.Fatal(err)
			}
			if old {
				for _, sql := range []string{
					"UPDATE nest_tool_generate_columns SET view_type = CASE WHEN column_name = 'category_id' THEN 'select' ELSE 'saSelect' END, dict_type = NULL, option_source = NULL, option_config = NULL WHERE table_id = 1 AND column_name IN ('category_id', 'status', 'is_link', 'is_hot')",
					"ALTER TABLE ai_system_config_group DROP COLUMN sort",
					"DROP TABLE IF EXISTS ai_system_role_dept",
					"DROP TABLE IF EXISTS ai_system_notice",
					"ALTER TABLE nest_tool_generate_columns DROP COLUMN option_source, DROP COLUMN option_config",
					"ALTER TABLE nest_tool_generate_tables ADD COLUMN is_full smallint NULL",
				} {
					if err := db.Exec(sql).Error; err != nil {
						t.Fatal(err)
					}
				}
			}
			if err := Check(db); err == nil {
				t.Fatal("unversioned schema was accepted")
			}
			if err := Up(db); err != nil {
				t.Fatal(err)
			}
			if err := Check(db); err != nil {
				t.Fatal(err)
			}
			var configured int64
			if err := db.Table("nest_tool_generate_columns").Where("table_id = 1 AND column_name IN ('category_id', 'status', 'is_link', 'is_hot') AND view_type = 'select' AND option_source = 'static' AND JSON_LENGTH(option_config, '$.options') > 0").Count(&configured).Error; err != nil {
				t.Fatal(err)
			}
			if configured != 4 {
				t.Fatalf("configured article fields = %d, want 4", configured)
			}
			// Reapplying the repair must preserve a project's custom source and dictionary.
			if err := db.Exec(`UPDATE nest_tool_generate_columns SET option_source = 'route', option_config = '{"path":"/system/custom/options"}' WHERE table_id = 1 AND column_name = 'category_id'`).Error; err != nil {
				t.Fatal(err)
			}
			if err := db.Exec("UPDATE nest_tool_generate_columns SET view_type = 'saSelect', dict_type = 'custom_status', option_source = NULL, option_config = NULL WHERE table_id = 1 AND column_name = 'status'").Error; err != nil {
				t.Fatal(err)
			}
			if err := repairArticleCodegenOptions(db); err != nil {
				t.Fatal(err)
			}
			var preserved int64
			if err := db.Table("nest_tool_generate_columns").Where("table_id = 1 AND ((column_name = 'category_id' AND option_source = 'route') OR (column_name = 'status' AND dict_type = 'custom_status' AND view_type = 'saSelect'))").Count(&preserved).Error; err != nil {
				t.Fatal(err)
			}
			if preserved != 2 {
				t.Fatal("customized article configuration was overwritten")
			}
			before, err := Status(db)
			if err != nil {
				t.Fatal(err)
			}
			if len(before) != len(registry) {
				t.Fatalf("history=%v", before)
			}
			if err := Up(db); err != nil {
				t.Fatalf("repeat migration: %v", err)
			}
			after, err := Status(db)
			if err != nil {
				t.Fatal(err)
			}
			if len(after) != len(before) {
				t.Fatal("repeat up added a version")
			}
			if !db.Migrator().HasColumn("ai_system_config_group", "sort") || !db.Migrator().HasTable("ai_system_role_dept") {
				t.Fatal("legacy upgrade incomplete")
			}
			if err := db.Model(&Record{}).Where("version = ?", 1).Update("dirty", true).Error; err != nil {
				t.Fatal(err)
			}
			if err := Check(db); err == nil {
				t.Fatal("dirty database accepted")
			}
			if err := Up(db); err == nil {
				t.Fatal("dirty migration rerun")
			}
		})
	}
}
