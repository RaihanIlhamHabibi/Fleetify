package models

import "time"

type UserRole string

const (
	RoleSA       UserRole = "SA"
	RoleApproval UserRole = "APPROVAL"
)

type ItemType string

const (
	ItemPart    ItemType = "PART"
	ItemService ItemType = "SERVICE"
)

type ReportStatus string

const (
	StatusPendingApproval ReportStatus = "PENDING_APPROVAL"
	StatusApproved        ReportStatus = "APPROVED"
	StatusCompleted       ReportStatus = "COMPLETED"
)

type User struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  string    `gorm:"size:100;uniqueIndex;not null" json:"username"`
	Role      UserRole  `gorm:"type:enum('SA','APPROVAL');not null" json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

func (User) TableName() string { return "users" }

type Vehicle struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	LicensePlate string    `gorm:"size:20;uniqueIndex;not null" json:"license_plate"`
	Model        string    `gorm:"size:100;not null" json:"model"`
	CreatedAt    time.Time `json:"created_at"`
}

func (Vehicle) TableName() string { return "vehicles" }

type MasterItem struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ItemName  string    `gorm:"size:150;not null" json:"item_name"`
	Type      ItemType  `gorm:"type:enum('PART','SERVICE');not null" json:"type"`
	Price     float64   `gorm:"type:decimal(12,2);not null" json:"price"`
	CreatedAt time.Time `json:"created_at"`
}

func (MasterItem) TableName() string { return "master_items" }

type MaintenanceReport struct {
	ID           uint64       `gorm:"primaryKey;autoIncrement" json:"id"`
	VehicleID    uint64       `gorm:"not null" json:"vehicle_id"`
	CreatedBy    uint64       `gorm:"not null" json:"created_by"`
	Odometer     uint         `gorm:"not null" json:"odometer"`
	Complaint    string       `gorm:"type:text;not null" json:"complaint"`
	Status       ReportStatus `gorm:"type:enum('PENDING_APPROVAL','APPROVED','COMPLETED');default:PENDING_APPROVAL" json:"status"`
	InitialPhoto *string      `gorm:"size:500" json:"initial_photo"`
	ProofPhoto   *string      `gorm:"size:500" json:"proof_photo"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`

	Vehicle   Vehicle       `gorm:"foreignKey:VehicleID" json:"vehicle,omitempty"`
	Creator   User          `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Items     []ReportItem  `gorm:"foreignKey:ReportID" json:"items,omitempty"`
}

func (MaintenanceReport) TableName() string { return "maintenance_reports" }

type ReportItem struct {
	ID            uint64  `gorm:"primaryKey;autoIncrement" json:"id"`
	ReportID      uint64  `gorm:"not null" json:"report_id"`
	ItemID        uint64  `gorm:"not null" json:"item_id"`
	Quantity      uint    `gorm:"not null;default:1" json:"quantity"`
	PriceSnapshot float64 `gorm:"type:decimal(12,2);not null" json:"price_snapshot"`

	MasterItem MasterItem `gorm:"foreignKey:ItemID" json:"master_item,omitempty"`
}

func (ReportItem) TableName() string { return "report_items" }

type ReportListItem struct {
	ID            uint64       `json:"id"`
	SAName        string       `json:"sa_name"`
	LicensePlate  string       `json:"license_plate"`
	Status        ReportStatus `json:"status"`
	CreatedAt     time.Time    `json:"created_at"`
	Odometer      uint         `json:"odometer,omitempty"`
	Complaint     string       `json:"complaint,omitempty"`
	VehicleModel  string       `json:"vehicle_model,omitempty"`
}
