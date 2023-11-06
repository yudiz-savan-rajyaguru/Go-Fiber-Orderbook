package trade

import (
	"time"

	"github.com/opinion-trading/helper"
)

type MatchResult struct {
	Code      uint
	Data      *TradeRequestBody
	Orderbook *Orderbook
	Error     error
	Msg       string
}

type TradeRequestBody struct {
	UID   string                `json:"uid" validate:"required"`
	EID   string                `json:"eid" validate:"required"`
	Price float32               `json:"price" validate:"required,lte=9.5,gte=0.5"`
	Qty   int                   `json:"qty" validate:"required,gte=1"`
	Side  helper.Sides          `json:"side" validate:"required,uppercase"`
	Flag  helper.FlagOfPurchase `json:"flag" validate:"required,uppercase"`
}

type UpdateTradeRequestBody struct {
	UID      string                `json:"uid" validate:"required"`
	EID      string                `json:"eid" validate:"required"`
	PrePrice float32               `json:"prePrice" validate:"required,lte=9.5,gte=0.5"`
	PreQty   int                   `json:"preQty" validate:"required,gte=1"`
	NewPrice float32               `json:"newPrice" validate:"required,lte=9.5,gte=0.5"`
	NewQty   int                   `json:"newQty" validate:"required,gte=1"`
	Side     helper.Sides          `json:"side" validate:"required,uppercase"`
	Flag     helper.FlagOfPurchase `json:"flag" validate:"required,uppercase"`
}

type Orderbook struct {
	YES_UID   string    `json:"yes_uid"`
	NO_UID    string    `json:"no_uid"`
	EID       string    `json:"eid"`
	YES_Price float32   `json:"yes_price"`
	NO_Price  float32   `json:"no_price"`
	YES_Qty   int       `json:"yes_qty"`
	NO_Qty    int       `json:"no_qty"`
	CreatedAt time.Time `json:"created_at" default:"time.Now()"`
}
