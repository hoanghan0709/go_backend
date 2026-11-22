package database

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(dsn, sqlitePath string) (*gorm.DB, error) {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		if sqlitePath == "" {
			return nil, errors.New("DATABASE_URL is required when DATABASE_PATH is unset")
		}
		return InitDB(sqlitePath)
	}

	// Driver errors during initialization can contain the connection string.
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil, errors.New("could not connect to PostgreSQL; check DATABASE_URL and database availability")
	}
	db = db.Session(&gorm.Session{Logger: logger.Default})
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		return tx.AutoMigrate(getModels()...)
	}); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("migrate PostgreSQL schema: %w", err)
	}
	return db, nil
}

func InitDB(filepath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(filepath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// SQLite chỉ có một writer; một connection giúp các transaction ghi tuần tự,
	// đồng thời PRAGMA foreign_keys áp dụng nhất quán cho mọi truy vấn.
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)
	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		return nil, err
	}
	if err := db.Exec("PRAGMA busy_timeout = 5000").Error; err != nil {
		return nil, err
	}
	if err := autoMigrate(db); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	return db, nil
}
