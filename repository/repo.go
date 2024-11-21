package repository

import (
	"errors"

	"github.com/liju-github/FoodBuddyMicroserviceOrderCart/models"
	"gorm.io/gorm"
)

type OrderCartRepository interface {
	// Cart operations
	AddToCart(item *models.CartItem) error
	GetCartItems(userID, restaurantID string) ([]models.CartItem, error)
	GetAllUserCarts(userID string) (map[string][]models.CartItem, error)
	UpdateCartItemQuantity(userID, restaurantID, productID string, quantity int32) error
	RemoveFromCart(userID, restaurantID, productID string) error
	ClearCart(userID, restaurantID string) error

	// Order operations
	CreateOrder(order *models.Order) error
	GetAllOrders(userID string) ([]models.Order, error)
	GetOrderByID(orderID string) (*models.Order, error)
	UpdateOrderStatus(orderID, status string) error
	GetRestaurantOrders(restaurantID string, status string) ([]models.Order, error)
	UpdateOrderCancellation(orderID, reason string) error
}

type orderCartRepo struct {
	db *gorm.DB
}

func NewOrderCartRepository(db *gorm.DB) OrderCartRepository {
	return &orderCartRepo{db: db}
}

// Cart operations implementation
func (r *orderCartRepo) AddToCart(item *models.CartItem) error {
	var existingItem models.CartItem
	result := r.db.Where("user_id = ? AND restaurant_id = ? AND product_id = ?", item.UserID, item.RestaurantID, item.ProductID).First(&existingItem)

	if result.Error == nil {
		// Update existing item quantity
		existingItem.Quantity += item.Quantity
		return r.db.Save(&existingItem).Error
	}

	return r.db.Create(item).Error
}

func (r *orderCartRepo) GetCartItems(userID, restaurantID string) ([]models.CartItem, error) {
	var items []models.CartItem
	result := r.db.Where("user_id = ? AND restaurant_id = ?", userID, restaurantID).Find(&items)
	return items, result.Error
}

func (r *orderCartRepo) GetAllUserCarts(userID string) (map[string][]models.CartItem, error) {
	var items []models.CartItem
	result := r.db.Where("user_id = ?", userID).Find(&items)
	if result.Error != nil {
		return nil, result.Error
	}

	// Group items by restaurant
	cartsByRestaurant := make(map[string][]models.CartItem)
	for _, item := range items {
		cartsByRestaurant[item.RestaurantID] = append(cartsByRestaurant[item.RestaurantID], item)
	}
	return cartsByRestaurant, nil
}

func (r *orderCartRepo) UpdateCartItemQuantity(userID, restaurantID, productID string, quantity int32) error {
	result := r.db.Model(&models.CartItem{}).
		Where("user_id = ? AND restaurant_id = ? AND product_id = ?", userID, restaurantID, productID).
		Update("quantity", quantity)

	if result.RowsAffected == 0 {
		return errors.New("cart item not found")
	}
	return result.Error
}

func (r *orderCartRepo) RemoveFromCart(userID, restaurantID, productID string) error {
	result := r.db.Where("user_id = ? AND restaurant_id = ? AND product_id = ?", userID, restaurantID, productID).
		Delete(&models.CartItem{})

	if result.RowsAffected == 0 {
		return errors.New("cart item not found")
	}
	return result.Error
}

func (r *orderCartRepo) ClearCart(userID, restaurantID string) error {
	return r.db.Where("user_id = ? AND restaurant_id = ?", userID, restaurantID).Delete(&models.CartItem{}).Error
}

// Order operations implementation
func (r *orderCartRepo) CreateOrder(order *models.Order) error {
	return r.db.Create(order).Error
}

func (r *orderCartRepo) GetAllOrders(userID string) ([]models.Order, error) {
	var orders []models.Order
	err := r.db.Preload("OrderItems").Where("user_id = ?", userID).Find(&orders).Error
	return orders, err
}

func (r *orderCartRepo) GetOrderByID(orderID string) (*models.Order, error) {
	var order models.Order
	err := r.db.Preload("OrderItems").Where("order_id = ?", orderID).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderCartRepo) UpdateOrderStatus(orderID, status string) error {
	result := r.db.Model(&models.Order{}).
		Where("order_id = ?", orderID).
		Update("order_status", status)

	if result.RowsAffected == 0 {
		return errors.New("order not found")
	}
	return result.Error
}

func (r *orderCartRepo) GetRestaurantOrders(restaurantID string, status string) ([]models.Order, error) {
	var orders []models.Order
	query := r.db.Preload("OrderItems").Where("restaurant_id = ?", restaurantID)
	if status != "" {
		query = query.Where("order_status = ?", status)
	}
	err := query.Find(&orders).Error
	return orders, err
}

func (r *orderCartRepo) UpdateOrderCancellation(orderID, reason string) error {
	result := r.db.Model(&models.Order{}).
		Where("order_id = ?", orderID).
		Updates(map[string]interface{}{
			"order_status":  "CANCELLED",
			"cancel_reason": reason,
		})

	if result.RowsAffected == 0 {
		return errors.New("order not found")
	}
	return result.Error
}
