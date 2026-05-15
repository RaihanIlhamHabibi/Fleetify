package repository

import (
	"errors"
	"fleetify/models"

	"gorm.io/gorm"
)

var (
	ErrReportNotFound      = errors.New("report not found")
	ErrInvalidStatusChange = errors.New("invalid status transition")
)

type ReportItemInput struct {
	ItemID   uint64
	Quantity uint
}

type CreateReportInput struct {
	VehicleID    uint64
	CreatedBy    uint64
	Odometer     uint
	Complaint    string
	InitialPhoto *string
	Items        []ReportItemInput
}

type ReportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

func (r *ReportRepository) Create(input CreateReportInput, priceMap map[uint64]float64) (*models.MaintenanceReport, error) {
	var report models.MaintenanceReport

	err := r.db.Transaction(func(tx *gorm.DB) error {
		report = models.MaintenanceReport{
			VehicleID:    input.VehicleID,
			CreatedBy:    input.CreatedBy,
			Odometer:     input.Odometer,
			Complaint:    input.Complaint,
			Status:       models.StatusPendingApproval,
			InitialPhoto: input.InitialPhoto,
		}
		if err := tx.Create(&report).Error; err != nil {
			return err
		}

		for _, item := range input.Items {
			price, ok := priceMap[item.ItemID]
			if !ok {
				return gorm.ErrRecordNotFound
			}
			row := models.ReportItem{
				ReportID:      report.ID,
				ItemID:        item.ItemID,
				Quantity:      item.Quantity,
				PriceSnapshot: price,
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r.FindByID(report.ID)
}

func (r *ReportRepository) FindByID(id uint64) (*models.MaintenanceReport, error) {
	var report models.MaintenanceReport
	err := r.db.Preload("Vehicle").Preload("Creator").Preload("Items.MasterItem").
		First(&report, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrReportNotFound
		}
		return nil, err
	}
	return &report, nil
}

func (r *ReportRepository) ListAll() ([]models.ReportListItem, error) {
	var rows []models.ReportListItem
	err := r.db.Table("maintenance_reports mr").
		Select(`mr.id, u.username AS sa_name, v.license_plate, mr.status, mr.created_at,
			mr.odometer, mr.complaint, v.model AS vehicle_model`).
		Joins("JOIN users u ON u.id = mr.created_by").
		Joins("JOIN vehicles v ON v.id = mr.vehicle_id").
		Order("mr.created_at DESC").
		Scan(&rows).Error
	return rows, err
}

func (r *ReportRepository) Approve(id uint64) (*models.MaintenanceReport, error) {
	return r.updateStatus(id, models.StatusPendingApproval, models.StatusApproved, nil)
}

func (r *ReportRepository) Complete(id uint64, proofPhoto string) (*models.MaintenanceReport, error) {
	return r.updateStatus(id, models.StatusApproved, models.StatusCompleted, &proofPhoto)
}

func (r *ReportRepository) updateStatus(id uint64, from, to models.ReportStatus, proofPhoto *string) (*models.MaintenanceReport, error) {
	var report models.MaintenanceReport
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&report, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrReportNotFound
			}
			return err
		}
		if report.Status != from {
			return ErrInvalidStatusChange
		}
		report.Status = to
		if proofPhoto != nil {
			report.ProofPhoto = proofPhoto
		}
		return tx.Save(&report).Error
	})
	if err != nil {
		return nil, err
	}
	return r.FindByID(id)
}
