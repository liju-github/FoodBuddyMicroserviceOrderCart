package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	orderCartPb "github.com/liju-github/CentralisedFoodbuddyMicroserviceProto/OrderCart"
	"github.com/liju-github/FoodBuddyMicroserviceOrderCart/configs"
	"github.com/liju-github/FoodBuddyMicroserviceOrderCart/db"
	"github.com/liju-github/FoodBuddyMicroserviceOrderCart/repository"
	"github.com/liju-github/FoodBuddyMicroserviceOrderCart/service"
)

func main() {
	// Load configuration
	config := config.LoadConfig()

	// Database connection
	dbConn, err := db.Connect(
		config.DBUser,
		config.DBPassword,
		config.DBHost,
		config.DBPort,
		config.DBName,
	)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize repository
	repo := repository.NewOrderCartRepository(dbConn)

	// Initialize service
	svc := service.NewOrderCartService(repo)

	// Initialize gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", config.ORDERCARTGRPCPORT))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	orderCartPb.RegisterOrderCartServiceServer(grpcServer, svc)

	log.Printf("Starting OrderCart gRPC server on port %s", config.ORDERCARTGRPCPORT)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
