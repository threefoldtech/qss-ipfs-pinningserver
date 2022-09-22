package database

import (
	"time"

	"github.com/threefoldtech/tf-pinning-service/pinning-api/models"
	"gorm.io/gorm"
)

type Status string

// List of Status
const (
	QUEUED  Status = "queued"
	PINNING Status = "pinning"
	PINNED  Status = "pinned"
	FAILED  Status = "failed"
)

type User struct {
	gorm.Model
	AccessToken string
	// Email       string
	// Account		string // polka account id ? if so should be unique
}

type PinDTO struct {
	UUID      string `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Status    Status         `gorm:"embedded"`
	Cid       string
	Name      string
	UserID    uint
}

func (p *PinDTO) ToEntity() models.PinStatus {
	return models.PinStatus{
		Requestid: p.UUID,
		Status:    models.Status(p.Status),
		Created:   p.CreatedAt,
		Pin: models.Pin{
			Cid:  p.Cid,
			Name: p.Name,
		},
	}
}

func (p *PinDTO) FromEntity(ps models.PinStatus) {
	p.UUID = ps.Requestid
	p.Status = Status(ps.Status)
	p.Cid = ps.Pin.Cid
	p.Name = ps.Pin.Name
	p.CreatedAt = ps.Created
}
