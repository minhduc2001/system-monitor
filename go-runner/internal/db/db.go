package db

import (
	"fmt"
	"log"

	"go-runner/internal/config"
	"go-runner/internal/project"
	"go-runner/internal/system"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "modernc.org/sqlite"
)

func InitDB(cfg *config.Config) *gorm.DB {
	var db *gorm.DB
	var err error

	// Configure GORM logger
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	switch cfg.Database.Driver {
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(cfg.Database.Path), gormConfig)
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
			cfg.Database.Host,
			cfg.Database.Username,
			cfg.Database.Password,
			cfg.Database.DBName,
			cfg.Database.Port,
			cfg.Database.SSLMode,
		)
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.Database.Username,
			cfg.Database.Password,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.DBName,
		)
		db, err = gorm.Open(mysql.Open(dsn), gormConfig)
	default:
		log.Fatalf("unsupported database driver: %s", cfg.Database.Driver)
	}

	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Auto migrate schemas
	if err := db.AutoMigrate(
		&project.ProjectGroup{}, 
		&project.Project{},
		&system.SystemMetrics{},
		&system.SystemAlert{},
		&system.SystemConfig{},
	); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	log.Printf("✅ Database connected successfully (%s)", cfg.Database.Driver)
	return db
}
