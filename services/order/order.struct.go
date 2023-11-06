package order

import (
	"time"

	"github.com/google/uuid"
	"github.com/opinion-trading/helper"
)

type RequestBody struct {
	UserID       string                `json:"userId" validate:"required"`
	EventID      string                `json:"eventId" validate:"required"`
	Price        float32               `json:"price" validate:"required,lte=9.5,gte=0.5"`
	Qty          int16                 `json:"qty" validate:"required,gte=1"`
	Side         helper.Sides          `json:"side" validate:"required,uppercase"`
	PurchaseFlag helper.FlagOfPurchase `json:"purchaseFlag" validate:"required,uppercase"`
}

type PlacedOrder struct {
	YesOrderId uuid.UUID `json:"yes_orderId"`
	NoOrderId  uuid.UUID `json:"no_orderId"`
	YesUserId  string    `json:"yes_uid"`
	NoUserId   string    `json:"no_uid"`
	EventId    string    `json:"eid"`
	YesPrice   float32   `json:"yes_price"`
	NoPrice    float32   `json:"no_price"`
	YesQty     int16     `json:"yes_qty"`
	NoQty      int16     `json:"no_qty"`
	CreatedAt  time.Time `json:"created_at" default:"time.Now()"`
}

type Order struct {
	ID     uuid.UUID
	UserId string
	Price  float32
	Qty    int16
}

type DataStruct struct {
	Price int16
	Data  Order
}

type User struct {
	UserId  uint
	Balance int64
}

type Event struct {
	EventId string
	Name    string
}

type Orderbook struct {
	BidYes []Order
	BidNo  []Order
}
