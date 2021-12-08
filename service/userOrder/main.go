package main

import (
	"github.com/micro/go-micro/util/log"
	"github.com/micro/go-micro"
	"bj38web/service/userOrder/handler"

	userOrder "bj38web/service/userOrder/proto/userOrder"
	"github.com/micro/go-plugins/registry/consul"
	"bj38web/service/userOrder/model"
)

func main() {
	// New Service
	consulReg := consul.NewRegistry()

	service := micro.NewService(
		micro.Name("go.micro.srv.userOrder"),
		micro.Version("latest"),
		micro.Registry(consulReg),
		micro.Address("192.168.28.128:9986"),
	)

	// Initialise service
	service.Init()
	model.InitDb()

	// Register Handler
	userOrder.RegisterUserOrderHandler(service.Server(), new(handler.UserOrder))


	// Run service
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
