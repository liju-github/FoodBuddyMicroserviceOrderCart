package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	orderCartPb "github.com/liju-github/CentralisedFoodbuddyMicroserviceProto/OrderCart"
	restaurantPb "github.com/liju-github/CentralisedFoodbuddyMicroserviceProto/Restaurant"
	userPb "github.com/liju-github/CentralisedFoodbuddyMicroserviceProto/User"
	clients "github.com/liju-github/FoodBuddyMicroserviceOrderCart/clients"
	"github.com/liju-github/FoodBuddyMicroserviceOrderCart/models"
	"github.com/liju-github/FoodBuddyMicroserviceOrderCart/repository"
)

type OrderCartService struct {
	orderCartPb.UnimplementedOrderCartServiceServer
	repo repository.OrderCartRepository
}

func NewOrderCartService(repo repository.OrderCartRepository) *OrderCartService {
	return &OrderCartService{repo: repo}
}

// Cart Operations
// AddProductToCart adds a product to the user's cart, but only if the
// restaurant has sufficient stock.
func (s *OrderCartService) AddProductToCart(ctx context.Context, req *orderCartPb.AddProductToCartRequest) (*orderCartPb.AddProductToCartResponse, error) {
	restaurantClient, err := clients.NewRestaurantClient()
	if err != nil {
		return nil, err
	}

	// Get product details
	productReq := &restaurantPb.GetProductByIDRequest{
		ProductId: req.ProductId,
	}
	productResp, err := restaurantClient.GetProductByID(ctx, productReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get product details: %w", err)
	}

	if productResp.Product == nil {
		return &orderCartPb.AddProductToCartResponse{
			Message: "Product not found",
		}, nil
	}

	// Check if restaurant is banned
	banStatusReq := &restaurantPb.CheckRestaurantBanStatusRequest{
		RestaurantId: productResp.Product.RestaurantId,
	}
	banStatus, err := restaurantClient.CheckRestaurantBanStatus(ctx, banStatusReq)
	if err != nil {
		return nil, fmt.Errorf("failed to check restaurant status: %w", err)
	}
	if banStatus.IsBanned {
		return &orderCartPb.AddProductToCartResponse{
			Message: fmt.Sprintf("Restaurant is currently unavailable. Reason: %s", banStatus.Reason),
		}, nil
	}

	// Create cart item with additional details
	cartItem := &models.CartItem{
		UserID:       req.UserId,
		ProductID:    req.ProductId,
		RestaurantID: productResp.Product.RestaurantId,
		ProductName:  productResp.Product.Name,
		Description:  productResp.Product.Description,
		Category:     productResp.Product.Category,
		Price:        productResp.Product.Price,
		Quantity:     req.Quantity,
	}

	err = s.repo.AddToCart(cartItem)
	if err != nil {
		return nil, fmt.Errorf("failed to add to cart: %w", err)
	}

	return &orderCartPb.AddProductToCartResponse{
		Message: "Product added to cart successfully",
	}, nil
}

// GetCartItems returns the items in the user's cart, as well as the total cost of all items in the cart.
func (s *OrderCartService) GetCartItems(ctx context.Context, req *orderCartPb.GetCartItemsRequest) (*orderCartPb.GetCartItemsResponse, error) {
	items, err := s.repo.GetCartItems(req.UserId, req.RestaurantId)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart items: %w", err)
	}

	var cartItems []*orderCartPb.CartItem
	var totalAmount float64

	for _, item := range items {
		cartItems = append(cartItems, &orderCartPb.CartItem{
			ProductId:    item.ProductID,
			RestaurantId: item.RestaurantID,
			ProductName:  item.ProductName,
			Description:  item.Description,
			Category:     item.Category,
			Price:        item.Price,
			Quantity:     item.Quantity,
		})
		totalAmount += item.Price * float64(item.Quantity)
	}

	return &orderCartPb.GetCartItemsResponse{
		Items:       cartItems,
		TotalAmount: totalAmount,
		Message:     "Cart items retrieved successfully",
	}, nil
}

// GetCartByRestaurant returns items in the user's cart for a specific restaurant
func (s *OrderCartService) GetCartByRestaurant(ctx context.Context, req *orderCartPb.GetCartByRestaurantRequest) (*orderCartPb.GetCartByRestaurantResponse, error) {
	items, err := s.repo.GetCartItems(req.UserId, req.RestaurantId)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart items: %w", err)
	}

	var totalAmount float64
	var cartItems []*orderCartPb.CartItem

	for _, item := range items {
		cartItems = append(cartItems, &orderCartPb.CartItem{
			ProductId:   item.ProductID,
			ProductName: item.ProductName,
			Description: item.Description,
			Category:    item.Category,
			Price:       item.Price,
			Quantity:    item.Quantity,
		})
		totalAmount += item.Price * float64(item.Quantity)
	}

	return &orderCartPb.GetCartByRestaurantResponse{
		Items:       cartItems,
		TotalAmount: totalAmount,
	}, nil
}

// GetAllCarts returns all cart items for a user, grouped by restaurant
func (s *OrderCartService) GetAllCarts(ctx context.Context, req *orderCartPb.GetAllCartsRequest) (*orderCartPb.GetAllCartsResponse, error) {
	cartsByRestaurant, err := s.repo.GetAllUserCarts(req.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed to get user carts: %w", err)
	}

	var response []*orderCartPb.RestaurantCart
	for restaurantID, items := range cartsByRestaurant {
		var cartItems []*orderCartPb.CartItem
		var totalAmount float64

		for _, item := range items {
			cartItems = append(cartItems, &orderCartPb.CartItem{
				ProductId:   item.ProductID,
				ProductName: item.ProductName,
				Description: item.Description,
				Category:    item.Category,
				Price:       item.Price,
				Quantity:    item.Quantity,
			})
			totalAmount += item.Price * float64(item.Quantity)
		}

		response = append(response, &orderCartPb.RestaurantCart{
			RestaurantId: restaurantID,
			Items:        cartItems,
			TotalAmount:  totalAmount,
		})
	}

	return &orderCartPb.GetAllCartsResponse{
		Carts: response,
	}, nil
}

// IncrementProductQuantity increments the quantity of the product in the user's cart. If the product is not found in the cart, the method does nothing.
//
// The method returns an error if the operation fails.
func (s *OrderCartService) IncrementProductQuantity(ctx context.Context, req *orderCartPb.IncrementProductQuantityRequest) (*orderCartPb.IncrementProductQuantityResponse, error) {
	items, err := s.repo.GetCartItems(req.UserId, req.RestaurantId)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart items: %w", err)
	}

	for _, item := range items {
		if item.ProductID == req.ProductId {
			err := s.repo.UpdateCartItemQuantity(req.UserId, req.RestaurantId, req.ProductId, item.Quantity+1)
			if err != nil {
				return nil, fmt.Errorf("failed to increment quantity: %w", err)
			}
			break
		}
	}

	return &orderCartPb.IncrementProductQuantityResponse{
		Message: "Product quantity incremented successfully",
	}, nil
}

// DecrementProductQuantity decrements the quantity of the product in the user's cart. If the product is not found in the cart, the method does nothing.
//
// The method returns an error if the operation fails.
func (s *OrderCartService) DecrementProductQuantity(ctx context.Context, req *orderCartPb.DecrementProductQuantityRequest) (*orderCartPb.DecrementProductQuantityResponse, error) {
	items, err := s.repo.GetCartItems(req.UserId, req.RestaurantId)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart items: %w", err)
	}

	itemFound := false
	for _, item := range items {
		if item.ProductID == req.ProductId {
			itemFound = true
			if item.Quantity > 1 {
				err := s.repo.UpdateCartItemQuantity(req.UserId, req.RestaurantId, req.ProductId, item.Quantity-1)
				if err != nil {
					return nil, fmt.Errorf("failed to decrement quantity: %w", err)
				}
			} else {
				err := s.repo.RemoveFromCart(req.UserId, req.RestaurantId, req.ProductId)
				if err != nil {
					return nil, fmt.Errorf("failed to remove item: %w", err)
				}
			}
			break
		}
	}

	if !itemFound {
		return &orderCartPb.DecrementProductQuantityResponse{
			Message: "Product not found in cart",
		}, nil
	}

	return &orderCartPb.DecrementProductQuantityResponse{
		Message: "Product quantity decremented successfully",
	}, nil
}

// RemoveProductFromCart removes a product from the user's cart.
//
// The method returns an error if the operation fails.
func (s *OrderCartService) RemoveProductFromCart(ctx context.Context, req *orderCartPb.RemoveProductFromCartRequest) (*orderCartPb.RemoveProductFromCartResponse, error) {
	err := s.repo.RemoveFromCart(req.UserId, req.RestaurantId, req.ProductId)
	if err != nil {
		return nil, fmt.Errorf("failed to remove product from cart: %w", err)
	}

	return &orderCartPb.RemoveProductFromCartResponse{
		Message: "Product removed from cart successfully",
	}, nil
}

// ClearCart clears all items from the user's cart.
//
// The method returns an error if the operation fails.
func (s *OrderCartService) ClearCart(ctx context.Context, req *orderCartPb.ClearCartRequest) (*orderCartPb.ClearCartResponse, error) {
	err := s.repo.ClearCart(req.UserId, req.RestaurantId)
	if err != nil {
		return nil, fmt.Errorf("failed to clear cart: %w", err)
	}

	return &orderCartPb.ClearCartResponse{
		Message: "Cart cleared successfully",
	}, nil
}

// Order Operations
// PlaceOrderByRestID processes an order for a specified restaurant by filtering
// the user's cart items, calculating the total amount, creating a new order,
// and clearing the relevant cart items. It returns the order ID if successful.
//
// This method performs the following steps:
// 1. Retrieves the user's cart items.
// 2. Filters the items by the provided restaurant ID and calculates the total cost.
// 3. Creates a new order with the filtered items and a "PENDING" status.
// 4. Removes the processed items from the user's cart.
//
// The method returns an error if any of the operations fail or if no items
// match the specified restaurant ID.
func (s *OrderCartService) PlaceOrderByRestID(ctx context.Context, req *orderCartPb.PlaceOrderByRestIDRequest) (*orderCartPb.PlaceOrderByRestIDResponse, error) {
	restaurantClient, err := clients.NewRestaurantClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create restaurant client: %w", err)
	}

	// Get restaurant details
	restaurantReq := &restaurantPb.GetRestaurantByIDRequest{
		RestaurantId: req.RestaurantId,
	}
	restaurantResp, err := restaurantClient.GetRestaurantByID(ctx, restaurantReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get restaurant details: %w", err)
	}

	// Check restaurant ban status
	banStatusReq := &restaurantPb.CheckRestaurantBanStatusRequest{
		RestaurantId: req.RestaurantId,
	}
	banStatus, err := restaurantClient.CheckRestaurantBanStatus(ctx, banStatusReq)
	if err != nil {
		return nil, fmt.Errorf("failed to check restaurant status: %w", err)
	}
	if banStatus.IsBanned {
		return nil, fmt.Errorf("restaurant is banned: %s", banStatus.Reason)
	}

	// Get cart items
	cartItems, err := s.repo.GetCartItems(req.UserId, req.RestaurantId)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart items: %w", err)
	}

	if len(cartItems) == 0 {
		return nil, fmt.Errorf("cart is empty for restaurant %s", req.RestaurantId)
	}

	// Calculate total amount and create order items
	var totalAmount float64
	var orderItems []models.OrderItem

	for _, item := range cartItems {
		// Get latest product details
		productReq := &restaurantPb.GetProductByIDRequest{
			ProductId: item.ProductID,
		}
		productResp, err := restaurantClient.GetProductByID(ctx, productReq)
		if err != nil {
			return nil, fmt.Errorf("failed to get product details for %s: %w", item.ProductID, err)
		}

		if productResp.Product == nil {
			return nil, fmt.Errorf("product %s not found", item.ProductID)
		}

		// Check stock
		if productResp.Product.Stock < item.Quantity {
			return nil, fmt.Errorf("insufficient stock for product %s: available %d, required %d",
				item.ProductName, productResp.Product.Stock, item.Quantity)
		}

		// Create order item
		orderItem := models.OrderItem{
			ProductID:   item.ProductID,
			ProductName: productResp.Product.Name,
			Description: productResp.Product.Description,
			Category:    productResp.Product.Category,
			Price:       productResp.Product.Price,
			Quantity:    item.Quantity,
		}
		orderItems = append(orderItems, orderItem)
		totalAmount += productResp.Product.Price * float64(item.Quantity)

		// Decrease stock
		decrementReq := &restaurantPb.DecrementProductStockByValueByValueRequest{
			ProductId:    item.ProductID,
			RestaurantId: req.RestaurantId,
			Value:        item.Quantity,
		}
		_, err = restaurantClient.DecrementProductStockByValue(ctx, decrementReq)
		if err != nil {
			return nil, fmt.Errorf("failed to update stock for product %s: %w", item.ProductID, err)
		}
	}

	// Create order
	order := &models.Order{
		OrderID:           fmt.Sprintf("order_%s", uuid.New().String()),
		UserID:            req.UserId,
		RestaurantID:      req.RestaurantId,
		RestaurantName:    restaurantResp.RestaurantName,
		RestaurantPhone:   restaurantResp.PhoneNumber,
		TotalAmount:       totalAmount,
		OrderStatus:       "PENDING",
		CreatedAt:         time.Now(),
		OrderItems:        orderItems,
		DeliveryAddressID: req.DeliveryAddressId,
	}

	// Get delivery address details and validate
	userClient, err := clients.NewUserClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create user client: %w", err)
	}

	validateAddressReq := &userPb.ValidateUserAddressRequest{
		UserId:    req.UserId,
		AddressId: req.DeliveryAddressId,
	}
	validateAddressResp, err := userClient.ValidateUserAddress(ctx, validateAddressReq)
	if err != nil {
		return nil, fmt.Errorf("failed to validate delivery address: %w", err)
	}

	if !validateAddressResp.IsValid {
		return nil, fmt.Errorf("invalid delivery address: %s", validateAddressResp.Message)
	}

	// Update order with address details
	order.StreetName = validateAddressResp.Address.StreetName
	order.Locality = validateAddressResp.Address.Locality
	order.State = validateAddressResp.Address.State
	order.Pincode = validateAddressResp.Address.Pincode

	// Save order
	err = s.repo.CreateOrder(order)
	if err != nil {
		// Attempt to rollback stock changes
		for _, item := range orderItems {
			incrementReq := &restaurantPb.IncremenentProductStockByValueRequest{
				ProductId:    item.ProductID,
				RestaurantId: req.RestaurantId,
				Value:        item.Quantity,
			}
			_, rollbackErr := restaurantClient.IncremenentProductStockByValue(ctx, incrementReq)
			if rollbackErr != nil {
				// Log rollback error but return original error
				log.Printf("Failed to rollback stock for product %s: %v", item.ProductID, rollbackErr)
			}
		}
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Convert order items to protobuf format
	var orderItemsPb []*orderCartPb.OrderItem
	for _, item := range order.OrderItems {
		orderItemsPb = append(orderItemsPb, &orderCartPb.OrderItem{
			ProductId:   item.ProductID,
			ProductName: item.ProductName,
			Description: item.Description,
			Category:    item.Category,
			Price:       item.Price,
			Quantity:    item.Quantity,
		})
	}

	// Convert order to protobuf format
	orderPb := &orderCartPb.Order{
		OrderId:      order.OrderID,
		UserId:       order.UserID,
		RestaurantId: order.RestaurantID,
		Items:        orderItemsPb,
		TotalAmount:  order.TotalAmount,
		OrderStatus:  order.OrderStatus,
		CreatedAt:    order.CreatedAt.Format(time.RFC3339),
		DeliveryAddress: &orderCartPb.Address{
			StreetName: order.StreetName,
			Locality:   order.Locality,
			State:      order.State,
			Pincode:    order.Pincode,
		},
	}

	// Clear cart
	err = s.repo.ClearCart(req.UserId, req.RestaurantId)
	if err != nil {
		log.Printf("Failed to clear cart for user %s: %v", req.UserId, err)
		// Continue with order placement even if cart clearing fails
	}

	return &orderCartPb.PlaceOrderByRestIDResponse{
		Success: true,
		Order:   orderPb,
		OrderId: order.OrderID,
		Message: "Order placed successfully",
	}, nil
}

// GetOrderDetailsAll retrieves all orders for a specified user ID.
//
// The method returns an error if the operation fails.
func (s *OrderCartService) GetOrderDetailsAll(ctx context.Context, req *orderCartPb.GetOrderDetailsAllRequest) (*orderCartPb.GetOrderDetailsAllResponse, error) {
	orders, err := s.repo.GetAllOrders(req.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	var pbOrders []*orderCartPb.Order
	for _, order := range orders {
		var orderItems []*orderCartPb.OrderItem
		for _, item := range order.OrderItems {
			orderItems = append(orderItems, &orderCartPb.OrderItem{
				ProductId:   item.ProductID,
				ProductName: item.ProductName,
				Description: item.Description,
				Category:    item.Category,
				Price:       item.Price,
				Quantity:    item.Quantity,
			})
		}

		pbOrders = append(pbOrders, &orderCartPb.Order{
			OrderId:      order.OrderID,
			UserId:       order.UserID,
			RestaurantId: order.RestaurantID,
			Items:        orderItems,
			TotalAmount:  order.TotalAmount,
			OrderStatus:  order.OrderStatus,
			CreatedAt:    order.CreatedAt.Format(time.RFC3339),
			DeliveryAddress: &orderCartPb.Address{
				StreetName: order.StreetName,
				Locality:   order.Locality,
				State:      order.State,
				Pincode:    order.Pincode,
			},
		})
	}

	return &orderCartPb.GetOrderDetailsAllResponse{
		Orders:  pbOrders,
		Message: "Orders retrieved successfully",
	}, nil
}

// GetOrderDetailsByID retrieves an order by ID.
//
// The method returns an error if the operation fails or if the order is not found.
func (s *OrderCartService) GetOrderDetailsByID(ctx context.Context, req *orderCartPb.GetOrderDetailsByIDRequest) (*orderCartPb.GetOrderDetailsByIDResponse, error) {
	order, err := s.repo.GetOrderByID(req.OrderId)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	var orderItems []*orderCartPb.OrderItem
	for _, item := range order.OrderItems {
		orderItems = append(orderItems, &orderCartPb.OrderItem{
			ProductId:   item.ProductID,
			ProductName: item.ProductName,
			Description: item.Description,
			Category:    item.Category,
			Price:       item.Price,
			Quantity:    item.Quantity,
		})
	}

	pbOrder := &orderCartPb.Order{
		OrderId:      order.OrderID,
		UserId:       order.UserID,
		RestaurantId: order.RestaurantID,
		Items:        orderItems,
		TotalAmount:  order.TotalAmount,
		OrderStatus:  order.OrderStatus,
		CreatedAt:    order.CreatedAt.Format(time.RFC3339),
		DeliveryAddress: &orderCartPb.Address{
			StreetName: order.StreetName,
			Locality:   order.Locality,
			State:      order.State,
			Pincode:    order.Pincode,
		},
	}

	return &orderCartPb.GetOrderDetailsByIDResponse{
		Order:   pbOrder,
		Message: "Order details retrieved successfully",
	}, nil
}

// CancelOrder cancels an order by ID if the user is authorized and the order is in the "PENDING" status.
//
// The method returns an error if the operation fails or if the order is not found, or if the user is not authorized to cancel the order,
// or if the order is not in the "PENDING" status.
func (s *OrderCartService) CancelOrder(ctx context.Context, req *orderCartPb.CancelOrderRequest) (*orderCartPb.CancelOrderResponse, error) {
	order, err := s.repo.GetOrderByID(req.OrderId)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if order.UserID != req.UserId {
		return &orderCartPb.CancelOrderResponse{
			Success: false,
			Message: "Unauthorized to cancel this order",
		}, nil
	}

	if order.OrderStatus != "PENDING" {
		return &orderCartPb.CancelOrderResponse{
			Success: false,
			Message: "Order cannot be cancelled in current status",
		}, nil
	}

	err = s.repo.UpdateOrderStatus(req.OrderId, "CANCELLED")
	if err != nil {
		return nil, fmt.Errorf("failed to cancel order: %w", err)
	}

	return &orderCartPb.CancelOrderResponse{
		Success: true,
		Message: "Order cancelled successfully",
	}, nil
}

// UpdateOrderStatus updates the status of an order.
//
// The method validates that the order exists and belongs to the specified restaurant
// before updating its status. Returns an error if the operation fails.
func (s *OrderCartService) UpdateOrderStatus(ctx context.Context, req *orderCartPb.UpdateOrderStatusRequest) (*orderCartPb.UpdateOrderStatusResponse, error) {
	// Get the order to validate ownership
	order, err := s.repo.GetOrderByID(req.OrderId)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Validate restaurant ownership
	if order.RestaurantID != req.RestaurantId {
		return &orderCartPb.UpdateOrderStatusResponse{
			Success: false,
			Message: "Unauthorized: Order does not belong to this restaurant",
		}, nil
	}

	// Update order status
	err = s.repo.UpdateOrderStatus(req.OrderId, req.NewStatus)
	if err != nil {
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	return &orderCartPb.UpdateOrderStatusResponse{
		Success: true,
		Message: fmt.Sprintf("Order status updated to %s successfully", req.NewStatus),
	}, nil
}

// OrderCart Service - Simple Order Confirmation
func (s *OrderCartService) ConfirmOrder(ctx context.Context, req *orderCartPb.ConfirmOrderRequest) (*orderCartPb.ConfirmOrderResponse, error) {
	// Update the order status to CONFIRMED
	updateResp, err := s.UpdateOrderStatus(ctx, &orderCartPb.UpdateOrderStatusRequest{
		OrderId:      req.OrderId,
		RestaurantId: req.RestaurantId,
		NewStatus:    "CONFIRMED",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to confirm order: %w", err)
	}

	if !updateResp.Success {
		return &orderCartPb.ConfirmOrderResponse{
			Success:     false,
			Message:     updateResp.Message,
			OrderStatus: "PENDING",
		}, nil
	}

	return &orderCartPb.ConfirmOrderResponse{
		Success:     true,
		Message:     "Order confirmed successfully",
		OrderStatus: "CONFIRMED",
	}, nil
}

// GetRestaurantOrders retrieves all orders for a specified restaurant ID.
//
// The method returns an error if the operation fails.
func (s *OrderCartService) GetRestaurantOrders(ctx context.Context, req *orderCartPb.GetRestaurantOrdersRequest) (*orderCartPb.GetRestaurantOrdersResponse, error) {
	orders, err := s.repo.GetRestaurantOrders(req.RestaurantId, req.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to get restaurant orders: %w", err)
	}

	var pbOrders []*orderCartPb.Order
	for _, order := range orders {
		var orderItems []*orderCartPb.OrderItem
		for _, item := range order.OrderItems {
			orderItems = append(orderItems, &orderCartPb.OrderItem{
				ProductId:   item.ProductID,
				ProductName: item.ProductName,
				Description: item.Description,
				Category:    item.Category,
				Price:       item.Price,
				Quantity:    item.Quantity,
			})
		}

		pbOrders = append(pbOrders, &orderCartPb.Order{
			OrderId:      order.OrderID,
			UserId:       order.UserID,
			RestaurantId: order.RestaurantID,
			Items:        orderItems,
			TotalAmount:  order.TotalAmount,
			OrderStatus:  order.OrderStatus,
			CreatedAt:    order.CreatedAt.Format(time.RFC3339),
			DeliveryAddress: &orderCartPb.Address{
				StreetName: order.StreetName,
				Locality:   order.Locality,
				State:      order.State,
				Pincode:    order.Pincode,
			},
		})
	}

	fmt.Println(pbOrders, "pbOrders")

	return &orderCartPb.GetRestaurantOrdersResponse{
		Orders:  pbOrders,
		Message: "Orders retrieved successfully",
	}, nil
}
