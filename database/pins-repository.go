package database

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/threefoldtech/tf-pinning-service/pinning-api/models"
	"gorm.io/gorm"
)

type pins struct {
	db *gorm.DB
	sync.Mutex
}

func NewPinsRepository() PinsRepository {
	return &pins{
		db: DB,
	}
}

func (r *pins) InsertOrGet(ctx context.Context, user_id uint, pinStatus models.PinStatus) (models.PinStatus, error) {
	pin := PinDTO{}
	pin.FromEntity(pinStatus)
	// get user from context
	pin.UserID = user_id
	uuid := pin.UUID
	pin.UUID = ""
	// pin_test := PinDTO{}
	//tx := r.db.Where(PinDTO{UserID: pin.UserID, Cid: pin.Cid}).First(&pin_test)

	r.db.Where("user_id = ? AND cid = ?", pin.UserID, pin.Cid).Attrs(PinDTO{UUID: uuid}).FirstOrCreate(&pin)

	return pin.ToEntity(), nil
}

func (r *pins) Patch(ctx context.Context, user_id uint, id string, fields map[string]interface{}) error {
	tx := r.db.Debug().Model(&PinDTO{}).Where("uuid = ? AND user_id = ?", id, user_id).Updates(fields)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (r *pins) FindByID(ctx context.Context, user_id uint, id string) (models.PinStatus, error) {
	var pin PinDTO
	tx := r.db.First(&pin, "uuid = ? AND user_id = ?", id, user_id)
	if tx.RowsAffected == 0 {
		return models.PinStatus{}, errors.New("id not exists")
	}
	return pin.ToEntity(), nil
}

func (r *pins) Find(
	ctx context.Context,
	user_id uint,
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
	if user_id != 0 {
		queryDB = queryDB.Where("user_id = ?", user_id)
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

func (r *pins) Delete(ctx context.Context, user_id uint, id string) error {
	tx := r.db.Where("uuid = ? AND user_id = ?", id, user_id).Delete(&PinDTO{})
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (r *pins) CIDRefrenceCount(ctx context.Context, cid string) int64 {
	var count int64
	r.db.Model(&PinDTO{}).Where("cid = ?", cid).Count(&count)
	return count

}

func (r *pins) FindByStatus(ctx context.Context, statuses []string) ([]PinDTO, error) {
	var pins []PinDTO

	queryDB := r.db

	if len(statuses) != 0 {
		queryDB = queryDB.Where("status IN ?", statuses)
	}

	tx := queryDB.Find(&pins)
	if tx.Error != nil {
		return []PinDTO{}, tx.Error
	}
	return pins, nil
}
