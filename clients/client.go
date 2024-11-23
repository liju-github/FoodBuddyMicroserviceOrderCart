package clients

import (
	"fmt"

	restaurantPb "github.com/liju-github/CentralisedFoodbuddyMicroserviceProto/Restaurant"
	userPb "github.com/liju-github/CentralisedFoodbuddyMicroserviceProto/User"
	config "github.com/liju-github/FoodBuddyMicroserviceOrderCart/configs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewRestaurantClient() (restaurantPb.RestaurantServiceClient, error) {
	conn, err := grpc.NewClient("localhost:"+config.LoadConfig().RESTAURANTGRPCPORT, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RestaurantService: %w", err)
	}
	return restaurantPb.NewRestaurantServiceClient(conn), nil
}

func NewUserClient() (userPb.UserServiceClient, error) {
	conn, err := grpc.NewClient("localhost:"+config.LoadConfig().USERGRPCPORT, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to UserService: %w", err)
	}
	return userPb.NewUserServiceClient(conn), nil

}
