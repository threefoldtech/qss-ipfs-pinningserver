package database

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"gorm.io/gorm"
)

type users struct {
	db *gorm.DB
}

func NewUsersRepository() UsersRepository {
	return &users{
		db: DB,
	}
}

func (u *users) Insert(ctx context.Context, access_token string) error {
	hash := sha256.Sum256([]byte(access_token))
	hex_string := hex.EncodeToString(hash[:])
	user := User{
		AccessToken: hex_string,
	}
	u.db.Create(&user)
	return nil
}

func (u *users) Patch(ctx context.Context, id string, fields map[string]interface{}) error {
	tx := u.db.Model(&User{}).Where("id = ?", id).Updates(fields)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (u *users) FindByID(ctx context.Context, id string) (User, error) {
	var user User
	tx := u.db.First(&user, "id = ?", id)
	if tx.RowsAffected == 0 {
		return User{}, errors.New("id not exists")
	}
	return user, nil
}
func (u *users) FindByTokenHash(ctx context.Context, hash string) (User, error) {
	var user User
	tx := u.db.First(&user, "access_token = ?", hash)
	if tx.RowsAffected == 0 {
		return User{}, errors.New("AccessToken not exists")
	}
	return user, nil
}
func (u *users) Delete(ctx context.Context, id string) error {
	tx := u.db.Where("id = ?", id).Delete(&PinDTO{})
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}
