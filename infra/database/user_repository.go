package database

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

func (m *MySQLDB) FindUserByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := m.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return User{}, ErrNotFound
	}
	return user, err
}
