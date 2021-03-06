package main

import (
	"fmt"
	"os"
	"time"

	"github.com/micro/go-micro"
	_ "github.com/micro/go-plugins/registry/kubernetes"
	iex "github.com/micro/services/portfolio/helpers/iex-cloud"
	"github.com/micro/services/portfolio/stock-earnings/handler"
	proto "github.com/micro/services/portfolio/stock-earnings/proto"
	"github.com/micro/services/portfolio/stock-earnings/storage/postgres"
	"github.com/robfig/cron/v3"
)

func main() {
	// Create The Service
	service := micro.NewService(
		micro.Name("kytra-v1-stock-earnings"),
		micro.Version("latest"),
	)
	service.Init()

	// Initialize IEX Package
	iex, err := iex.New(os.Getenv("IEX_TOKEN"))
	if err != nil {
		fmt.Printf("Could not initiate iex package: %v\n.", err)
		os.Exit(2)
	}

	// Connect to the Database (Postgres)
	db, err := postgres.New(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
	)
	if err != nil {
		fmt.Printf("Could not connect to DB: %v\n.", err)
		os.Exit(2)
	}
	defer db.Close()

	// Register to Service Discovery
	h := handler.New(iex, db, service.Client())

	// Setup ticker to fetch earnings from IEX
	c := cron.New(cron.WithLocation(time.UTC))
	c.AddFunc("0 0 * * 0", h.FetchEarnings)
	c.Start()
	defer c.Stop()

	proto.RegisterStockEarningsHandler(service.Server(), h)

	if err := service.Run(); err != nil {
		fmt.Println(err)
	}
}
