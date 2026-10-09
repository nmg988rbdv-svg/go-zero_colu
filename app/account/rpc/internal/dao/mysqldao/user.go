package mysqldao

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type User struct {
	ID          int64  `gorm:"column:user_id;primaryKey;autoIncrement:false"`
	LegacyID    int64  `gorm:"column:legacy_id;index:idx_users_legacy_id"`
	Username    string `gorm:"column:username;size:64;uniqueIndex:uk_users_username;not null"`
	Password    string `gorm:"column:password;size:255;not null"`
	PhoneNumber int64  `gorm:"column:phone_number"`
	Status      int32  `gorm:"column:status;not null;default:1"`
	CreatedAt   int64  `gorm:"column:created_at"`
	UpdatedAt   int64  `gorm:"column:updated_at"`
}

func (User) TableName() string { return "users" }

type UserRepo struct{ db *gorm.DB }

func NewUserRepo(db *gorm.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) FindByUsername(ctx context.Context, username string) (*User, error) {
	var u User
	err := r.db.WithContext(ctx).Where("username = ?", username).Take(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) FindByID(ctx context.Context, userID int64) (*User, error) {
	var u User
	err := r.db.WithContext(ctx).Where("user_id = ? OR legacy_id = ?", userID, userID).Take(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
