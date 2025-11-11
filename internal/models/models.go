package models

import (
	"time"

	"github.com/google/uuid"
)

type Organization struct {
	ID           uuid.UUID     `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name         string        `gorm:"type:varchar(255)"`
	Type         string        `gorm:"type:varchar(50)"`
	BIN          string        `gorm:"type:varchar(32)"`
	HeadFullName string        `gorm:"type:varchar(255)"`
	Address      string        `gorm:"type:varchar(255)"`
	Phone        string        `gorm:"type:varchar(32)"`
	ParentOrgID  *uuid.UUID    `gorm:"type:uuid"`
	ParentOrg    *Organization `gorm:"foreignKey:ParentOrgID;constraint:OnDelete:SET NULL"`
	IsActive     bool          `gorm:"default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (Organization) TableName() string {
	return "organizations"
}

type User struct {
	ID             uuid.UUID     `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Phone          string        `gorm:"type:varchar(32);uniqueIndex"`
	Role           string        `gorm:"type:varchar(50)"`
	Login          *string       `gorm:"type:varchar(64)"`
	PasswordHash   *string       `gorm:"type:varchar(255)"`
	OrganizationID *uuid.UUID    `gorm:"type:uuid"`
	Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:SET NULL"`
	DriverID       *uuid.UUID    `gorm:"type:uuid"`
	Driver         *Driver       `gorm:"foreignKey:DriverID;constraint:OnDelete:SET NULL"`
	IsActive       bool          `gorm:"default:true"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (User) TableName() string {
	return "users"
}

type Driver struct {
	ID           uuid.UUID     `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	ContractorID *uuid.UUID    `gorm:"type:uuid"`
	Contractor   *Organization `gorm:"foreignKey:ContractorID;constraint:OnDelete:SET NULL"`
	FullName     string        `gorm:"type:varchar(255)"`
	IIN          string        `gorm:"type:varchar(32)"`
	BirthYear    int           `gorm:"type:int"`
	Phone        string        `gorm:"type:varchar(32)"`
	IsActive     bool          `gorm:"default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (Driver) TableName() string {
	return "drivers"
}

type Vehicle struct {
	ID           uuid.UUID     `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	ContractorID *uuid.UUID    `gorm:"type:uuid"`
	Contractor   *Organization `gorm:"foreignKey:ContractorID;constraint:OnDelete:SET NULL"`
	PlateNumber  string        `gorm:"type:varchar(32);uniqueIndex"`
	Brand        string        `gorm:"type:varchar(64)"`
	Model        string        `gorm:"type:varchar(64)"`
	Color        string        `gorm:"type:varchar(64)"`
	Year         int           `gorm:"type:int"`
	BodyVolumeM3 float64       `gorm:"type:decimal(10,2)"`
	DriverID     *uuid.UUID    `gorm:"type:uuid"`
	Driver       *Driver       `gorm:"foreignKey:DriverID;constraint:OnDelete:SET NULL"`
	IsActive     bool          `gorm:"default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (Vehicle) TableName() string {
	return "vehicles"
}

