package seeder

import (
	"log"

	"fleetify/models"

	"gorm.io/gorm"
)

func Run(db *gorm.DB) {
	var count int64
	db.Model(&models.User{}).Count(&count)
	if count > 0 {
		log.Println("seeder skipped: data already exists")
		return
	}

	users := []models.User{
		{Username: "advisor_sa", Role: models.RoleSA},
		{Username: "manager_approval", Role: models.RoleApproval},
	}
	vehicles := []models.Vehicle{
		{LicensePlate: "B 1234 XYZ", Model: "Toyota Avanza"},
		{LicensePlate: "B 5678 ABC", Model: "Honda Brio"},
		{LicensePlate: "B 9012 DEF", Model: "Mitsubishi Xpander"},
	}
	items := []models.MasterItem{
		{ItemName: "Oli Mesin 1L", Type: models.ItemPart, Price: 85000},
		{ItemName: "Filter Oli", Type: models.ItemPart, Price: 45000},
		{ItemName: "Kampas Rem Depan", Type: models.ItemPart, Price: 320000},
		{ItemName: "Jasa Ganti Oli", Type: models.ItemService, Price: 75000},
		{ItemName: "Jasa Service Berkala", Type: models.ItemService, Price: 250000},
	}

	if err := db.Create(&users).Error; err != nil {
		log.Fatalf("seed users: %v", err)
	}
	if err := db.Create(&vehicles).Error; err != nil {
		log.Fatalf("seed vehicles: %v", err)
	}
	if err := db.Create(&items).Error; err != nil {
		log.Fatalf("seed master_items: %v", err)
	}
	log.Println("seeder completed successfully")
}
