package database

import (
	"context"
	"errors"
	"time"

	"github.com/threefoldtech/tf-pinning-service/pinning-api/models"
	"gorm.io/gorm"
)

type pins struct {
	db *gorm.DB
}

func New() PinsRepository {
	return &pins{
		db: DB,
	}
}

func (r *pins) Insert(ctx context.Context, pinStatus models.PinStatus) error {
	pin := PinDTO{}
	pin.FromEntity(pinStatus)
	// get user from context
	// pin.UserID = ctx.Value("userid")
	r.db.Create(&pin)
	return nil
}

func (r *pins) Patch(ctx context.Context, id string, fields map[string]interface{}) error {
	tx := r.db.Model(&PinDTO{}).Where("uuid = ?", id).Updates(fields)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (r *pins) FindByID(ctx context.Context, id string) (models.PinStatus, error) {
	var pin PinDTO
	tx := r.db.First(&pin, "uuid = ?", id)
	if tx.RowsAffected == 0 {
		return models.PinStatus{}, errors.New("id not exists")
	}
	return pin.ToEntity(), nil
}

func (r *pins) Find(
	ctx context.Context,
	cids, statuses []string,
	name string,
	before, after time.Time,
	match string,
	limit int,
) (models.PinResults, error) {
	var pins []PinDTO

	queryDB := r.db
	if len(cids) != 0 {
		// cids_list := strings.Split(cids, ",")
		queryDB = queryDB.Where("cid IN ?", cids)
	}
	if name != "" {
		queryDB = queryDB.Where("name = ?", name)
	}
	if len(statuses) != 0 {
		// status_list := strings.Split(status, ",")
		queryDB = queryDB.Where("status IN ?", statuses)
	}
	// TODO: before, after
	queryDB = queryDB.Limit(limit)
	tx := queryDB.Find(&pins)
	count := tx.RowsAffected
	if tx.Error != nil {
		return models.PinResults{}, tx.Error
	}
	var filterd_pins []models.PinStatus
	for _, pin := range pins {
		pin_status := pin.ToEntity()
		filterd_pins = append(filterd_pins, pin_status)
	}

	return models.PinResults{Count: int32(count), Results: filterd_pins}, nil // TODO: check type PinResults.Count
}

func (r *pins) Delete(ctx context.Context, id string) error {
	tx := r.db.Where("uuid = ?", id).Delete(&PinDTO{})
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}
