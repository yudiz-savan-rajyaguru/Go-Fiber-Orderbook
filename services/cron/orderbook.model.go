package cron

import (
	"time"

	"gorm.io/gorm"
)

type OrderBookModel struct {
	gorm.Model
	YesUID    string    `json:"yes_Uid" gorm:"not null" validate:"required"`
	NoUID     string    `json:"no_Uid" gorm:"not null" validate:"required"`
	EventID   string    `json:"eid" gorm:"not null" validate:"required"`
	YesPrice  float32   `json:"yes_Price" gorm:"not null" validate:"required"`
	NoPrice   float32   `json:"no_Price" gorm:"not null" validate:"required"`
	YesQty    int       `json:"yes_Qty" gorm:"not null" validate:"required"`
	NoQty     int       `json:"no_Qty" gorm:"not null" validate:"required"`
	CreatedAt time.Time `json:"created_at" gorm:"not null"`
}

func (OrderBookModel) TableName() string {
	return "orderbook"
}
