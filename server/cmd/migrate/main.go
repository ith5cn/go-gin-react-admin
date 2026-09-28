// Command migrate inspects or applies template schema migrations independently of HTTP.
package main

import (
	"os"

	"server/database/migrations"
	gormInit "server/setup/gorm"
	loggerInit "server/setup/logger"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	_ = godotenv.Load()
	if err := loggerInit.Logger.Initialize(); err != nil {
		panic(err)
	}
	log := loggerInit.Logger.Get()
	defer log.Sync()
	if len(os.Args) != 2 || (os.Args[1] != "status" && os.Args[1] != "up") {
		log.Fatal("usage: migrate status|up")
	}
	db, err := gormInit.Gorm.Initialize()
	if err != nil {
		log.Fatal("connect failed", zap.Error(err))
	}
	if sqlDB, err := db.DB(); err == nil {
		defer sqlDB.Close()
	}
	if os.Args[1] == "up" {
		if err := migrations.Up(db); err != nil {
			log.Fatal("migration failed", zap.Error(err))
		}
	}
	plan, err := migrations.Plan(db)
	if err != nil {
		log.Fatal("migration status invalid", zap.Error(err))
	}
	records, err := migrations.Status(db)
	if err != nil {
		log.Fatal("read migration history failed", zap.Error(err))
	}
	log.Info("database migrations", zap.Any("applied", records), zap.Strings("pending", plan))
}
