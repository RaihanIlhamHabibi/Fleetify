package repository

import (
	"errors"
	"strings"

	"fleetify/models"

	"gorm.io/gorm"
)

type MasterItemRepository struct {
	db *gorm.DB
}

func NewMasterItemRepository(db *gorm.DB) *MasterItemRepository {
	return &MasterItemRepository{db: db}
}

func (r *MasterItemRepository) FindAll() ([]models.MasterItem, error) {
	var items []models.MasterItem
	err := r.db.Order("item_name asc").Find(&items).Error
	return items, err
}

func (r *MasterItemRepository) FindByIDs(ids []uint64) ([]models.MasterItem, error) {
	var items []models.MasterItem
	if len(ids) == 0 {
		return items, nil
	}
	err := r.db.Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *MasterItemRepository) FindOrCreate(itemName string, itemType models.ItemType, price float64) (*models.MasterItem, error) {
	name := strings.TrimSpace(itemName)
	if name == "" {
		return nil, errors.New("item name is required")
	}
	if itemType != models.ItemPart && itemType != models.ItemService {
		return nil, errors.New("type must be PART or SERVICE")
	}
	if price <= 0 {
		return nil, errors.New("price must be greater than 0")
	}

	var item models.MasterItem
	err := r.db.Where("item_name = ? AND type = ?", name, itemType).First(&item).Error
	if err == nil {
		return &item, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	item = models.MasterItem{ItemName: name, Type: itemType, Price: price}
	if err := r.db.Create(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}
