package repository

import (
	"errors"
	"strings"

	"fleetify/models"

	"gorm.io/gorm"
)

type VehicleRepository struct {
	db *gorm.DB
}

func NewVehicleRepository(db *gorm.DB) *VehicleRepository {
	return &VehicleRepository{db: db}
}

func (r *VehicleRepository) FindAll() ([]models.Vehicle, error) {
	var vehicles []models.Vehicle
	err := r.db.Order("license_plate asc").Find(&vehicles).Error
	return vehicles, err
}

func (r *VehicleRepository) FindByID(id uint64) (*models.Vehicle, error) {
	var vehicle models.Vehicle
	if err := r.db.First(&vehicle, id).Error; err != nil {
		return nil, err
	}
	return &vehicle, nil
}

func (r *VehicleRepository) FindOrCreate(licensePlate, model string) (*models.Vehicle, error) {
	plate := strings.TrimSpace(licensePlate)
	if plate == "" {
		return nil, errors.New("license plate is required")
	}

	var vehicle models.Vehicle
	err := r.db.Where("license_plate = ?", plate).First(&vehicle).Error
	if err == nil {
		return &vehicle, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	modelName := strings.TrimSpace(model)
	if modelName == "" {
		modelName = "Tidak diketahui"
	}
	vehicle = models.Vehicle{LicensePlate: plate, Model: modelName}
	if err := r.db.Create(&vehicle).Error; err != nil {
		return nil, err
	}
	return &vehicle, nil
}
