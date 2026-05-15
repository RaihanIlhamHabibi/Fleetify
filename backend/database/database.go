package database

import (
	"fmt"
	"log"
	"time"

	"fleetify/config"
	"fleetify/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	var db *gorm.DB
	var err error
	for i := 0; i < 30; i++ {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Warn),
		})
		if err == nil {
			break
		}
		log.Printf("waiting for database... (%d/30)", i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Vehicle{},
		&models.MasterItem{},
		&models.MaintenanceReport{},
		&models.ReportItem{},
	); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return db, nil
}
