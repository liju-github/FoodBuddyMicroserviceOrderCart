package clients

import (
	"fmt"

	restaurantPb "github.com/liju-github/CentralisedFoodbuddyMicroserviceProto/Restaurant"
	"google.golang.org/grpc"
)

func NewRestaurantClient() (restaurantPb.RestaurantServiceClient, error) {
	conn, err := grpc.Dial("localhost:50052", grpc.WithInsecure()) // Use appropriate address and options
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RestaurantService: %w", err)
	}
	return restaurantPb.NewRestaurantServiceClient(conn), nil
}
