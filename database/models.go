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
	PinDTOs     []PinDTO
	// Email       string
	// Account		string // polka account id ? if so should be unique filed
}

type PinDTO struct {
	UUID      string `gorm:"primarykey"`
	CreatedAt int
	UpdatedAt int
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Status    Status         `gorm:"embedded"`
	Cid       string         // `gorm:"index:cidUserId,unique"`
	Name      string
	UserID    uint // `gorm:"index:cidUserId,unique"`
	DagSize   int
}

func (p *PinDTO) ToEntity() models.PinStatus {
	return models.PinStatus{
		Requestid: p.UUID,
		Status:    models.Status(p.Status),
		Created:   time.Unix(int64(p.CreatedAt), 0),
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
	p.CreatedAt = int(ps.Created.Unix())
}
