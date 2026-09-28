package gormInit

import (
	"fmt"
	"os"
	"strings"

	"server/config"
	"server/database/migrations"

	"gorm.io/gorm"
)

var Gorm = new(_gorm)

type _gorm struct {
	// Databases 保存启动时初始化完成的 MySQL 连接，业务层后续从这里取连接。
	Databases *Databases
}

// Databases 按业务库名称保存 *gorm.DB。
// AISystem 是后台系统库，Named 是通过 MYSQL_CONNECTIONS 注册的业务库。
type Databases struct {
	AISystem *gorm.DB
	Named    map[string]*gorm.DB
}

// InitializeAll 初始化项目启动所需的全部 MySQL 连接。
// 以后新增数据库连接时，只需要在这里扩展，main/initSetup 不需要跟着变复杂。
func (g *_gorm) InitializeAll() error {
	dbAISystem, err := g.Initialize()
	if err != nil {
		return fmt.Errorf("%s connect failed: %w", config.MysqlAISystem, err)
	}
	if err := migrations.Check(dbAISystem); err != nil {
		if sqlDB, e := dbAISystem.DB(); e == nil {
			_ = sqlDB.Close()
		}
		return fmt.Errorf("%s schema check failed: %w", config.MysqlAISystem, err)
	}

	connections := &Databases{AISystem: dbAISystem, Named: map[string]*gorm.DB{}}
	for _, name := range strings.Split(os.Getenv("MYSQL_CONNECTIONS"), ",") {
		name = strings.TrimSpace(name)
		if name == "" || name == config.MysqlAISystem {
			continue
		}
		if _, exists := connections.Named[name]; exists {
			continue
		}
		db, err := g.InitializeByName(name)
		if err != nil {
			for _, opened := range connections.Named {
				if sqlDB, e := opened.DB(); e == nil {
					_ = sqlDB.Close()
				}
			}
			if sqlDB, e := dbAISystem.DB(); e == nil {
				_ = sqlDB.Close()
			}
			return fmt.Errorf("%s connect failed: %w", name, err)
		}
		connections.Named[name] = db
	}
	g.Databases = connections

	return nil
}

// Initialize 打开默认 MySQL 连接。
// 当前默认库是 ai_system。
func (g *_gorm) Initialize() (*gorm.DB, error) {
	return g.InitializeByName(config.MysqlAISystem)
}

// InitializeByName 按连接名打开指定 MySQL。
// 支持的连接名定义在 config 中，例如 ai_system。
func (g *_gorm) InitializeByName(name string) (*gorm.DB, error) {
	return g.initializeMysqlByName(name)
}

// Get returns an initialized named connection without silently falling back to another database.
func (g *_gorm) Get(name string) (*gorm.DB, error) {
	if g.Databases == nil {
		return nil, fmt.Errorf("databases are not initialized")
	}
	if name == config.MysqlAISystem && g.Databases.AISystem != nil {
		return g.Databases.AISystem, nil
	}
	if db := g.Databases.Named[name]; db != nil {
		return db, nil
	}
	return nil, fmt.Errorf("connection %q is not registered (MYSQL_CONNECTIONS)", name)
}
