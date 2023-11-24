package model

import (
	"gorm.io/gorm"
	"time"
)

// Account
// - uid
// - name
// - first_name
// - cid
// - created_at
// - updated_at
// - deleted_at

// Account Information
type Account struct {
	Uid       string `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Name      string
	Firstname string `gorm:"column:first_name"`
	Cid       string
	Age       int
}

func (Account) TableName() string {
	return "account"
}

// Login
// - uid (account_uid)
// - username
// - password
// - created_at
// - updated_at
// - deleted_at

// Login credential
type Login struct {
	AccountUid string `gorm:"primaryKey"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
	Username   string
	Password   string
}

func (Login) TableName() string {
	return "login"
}
