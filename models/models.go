package models

import (
	"time"

	"gorm.io/gorm"
)

type Address struct {
	StreetName string `gorm:"type:varchar(255)"`
	Locality   string `gorm:"type:varchar(255)"`
	State      string `gorm:"type:varchar(255)"`
	Pincode    string `gorm:"type:varchar(20)"`
}

type CartItem struct {
	gorm.Model
	UserID       string `gorm:"type:varchar(255);index"`
	ProductID    string `gorm:"type:varchar(255)"`
	RestaurantID string `gorm:"type:varchar(255);index"`
	ProductName  string `gorm:"type:varchar(255)"`
	Description  string `gorm:"type:text"`
	Category     string `gorm:"type:varchar(255)"`
	Price        float64
	Quantity     int32
}

type Order struct {
	gorm.Model
	OrderID           string `gorm:"type:varchar(255);uniqueIndex"`
	UserID            string `gorm:"type:varchar(255);index"`
	RestaurantID      string `gorm:"type:varchar(255);index"`
	RestaurantName    string `gorm:"type:varchar(255)"`
	RestaurantPhone   uint64
	StreetName        string `gorm:"type:varchar(255)"`
	Locality          string `gorm:"type:varchar(255)"`
	State             string `gorm:"type:varchar(255)"`
	Pincode           string `gorm:"type:varchar(20)"`
	TotalAmount       float64
	OrderStatus       string `gorm:"type:varchar(50)"`
	CreatedAt         time.Time
	DeliveryAddressID string      `gorm:"type:varchar(255)"`
	CancelReason      string      `gorm:"type:varchar(255)"`
	OrderItems        []OrderItem `gorm:"foreignKey:OrderID;references:OrderID"`
}

type OrderItem struct {
	gorm.Model
	OrderID     string `gorm:"type:varchar(255);index"`
	ProductID   string `gorm:"type:varchar(255)"`
	ProductName string `gorm:"type:varchar(255)"`
	Description string `gorm:"type:text"`
	Category    string `gorm:"type:varchar(255)"`
	Price       float64
	Quantity    int32
}
