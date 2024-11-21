package models

import (
	"time"

	"gorm.io/gorm"
)

type Address struct {
	StreetName string
	Locality   string
	State      string
	Pincode    string
}

type CartItem struct {
	gorm.Model
	UserID       string `gorm:"index"`
	ProductID    string
	RestaurantID string
	ProductName  string
	Description  string
	Category     string
	Price        float64
	Quantity     int32
}

type Order struct {
	gorm.Model
	OrderID           string `gorm:"type:varchar(36);uniqueIndex"`
	UserID            string `gorm:"type:varchar(36);index"`
	RestaurantID      string `gorm:"type:varchar(36);index"`
	RestaurantName    string `gorm:"type:varchar(255)"`
	RestaurantPhone   uint64
	StreetName        string // Flattened address fields
	Locality          string
	State             string
	Pincode           string
	TotalAmount       float64
	OrderStatus       string `gorm:"type:varchar(20)"`
	CreatedAt         time.Time
	DeliveryAddressID string      `gorm:"type:varchar(36)"`
	CancelReason      string      `gorm:"type:varchar(255)"`
	OrderItems        []OrderItem `gorm:"foreignKey:OrderID;references:OrderID"`
}

type OrderItem struct {
	gorm.Model
	OrderID     string `gorm:"type:varchar(36);index"`
	ProductID   string `gorm:"type:varchar(36)"`
	ProductName string `gorm:"type:varchar(255)"`
	Description string `gorm:"type:text"`
	Category    string `gorm:"type:varchar(100)"`
	Price       float64
	Quantity    int32
}
