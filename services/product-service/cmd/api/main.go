package main

import (
	"log"
	"net"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	productv1 "uop-ms/gen/product/v1"
	"uop-ms/services/product-service/internal/app/config"
	"uop-ms/services/product-service/internal/app/db"
	"uop-ms/services/product-service/internal/core"
	productgrpc "uop-ms/services/product-service/internal/grpc"
	"uop-ms/services/product-service/internal/product"
	"uop-ms/services/product-service/internal/routes"
)

func main() {
	cfg := config.Load()

	gdb := db.Connect(cfg.MySQLDSN)

	// migrate
	if err := gdb.AutoMigrate(&product.Product{}); err != nil {
		log.Fatal(err)
	}

	// Domain
	store := product.NewStore(gdb)
	svc := product.NewService(store)
	h := product.NewHandler(svc)

	// gRPC Server(Internal)
	grpcServer := grpc.NewServer()

	productGrpcServer := productgrpc.NewServer(store)
	productv1.RegisterProductServiceServer(grpcServer, productGrpcServer)

	go func() {
		lis, err := net.Listen("tcp", ":9090")
		if err != nil {
			log.Fatal("failed to listen for gRPC:", err)
		}

		log.Println("product-service gRPC listening on :9090")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	//REST Server (External)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(core.ErrorHandler())

	routes.Register(r)
	h.Register(r)

	addr := ":" + cfg.Port
	log.Println("product-service listening on", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
