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
	OrderID         string `gorm:"uniqueIndex"`
	UserID          string `gorm:"index"`
	RestaurantID    string `gorm:"index"`
	RestaurantName  string
	RestaurantPhone uint64
	StreetName      string // Flattened address fields
	Locality        string
	State           string
	Pincode         string
	TotalAmount     float64
	OrderStatus     string
	CreatedAt       time.Time
	DeliveryAddress string
	CancelReason    string
	OrderItems      []OrderItem `gorm:"foreignKey:OrderID;references:OrderID"`
}

type OrderItem struct {
	gorm.Model
	OrderID     string `gorm:"index"`
	ProductID   string
	ProductName string
	Description string
	Category    string
	Price       float64
	Quantity    int32
}
