package main

import (
	"fmt"
	"os"
	"time"

	"github.com/micro/go-micro"
	_ "github.com/micro/go-plugins/registry/kubernetes"
	"github.com/micro/services/portfolio/portfolio-value-tracking/handler"
	proto "github.com/micro/services/portfolio/portfolio-value-tracking/proto"
	"github.com/micro/services/portfolio/portfolio-value-tracking/storage/postgres"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
)

func main() {
	// Create The Service
	service := micro.NewService(
		micro.Name("kytra-v1-portfolio-value-tracking"),
		micro.Version("latest"),
	)
	service.Init()

	// Connect to the Database (Postgres)
	db, err := postgres.New(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
	)
	if err != nil {
		panic(errors.Wrap(err, "Could not connect to the database"))
	}
	defer db.Close()

	// Register to Service Discovery
	h := handler.New(db, service.Client())
	proto.RegisterPortfolioValueTrackingHandler(service.Server(), h)

	c := cron.New(cron.WithLocation(time.UTC))
	c.AddFunc("*/15 * * * *", h.RecordValuations)
	c.Start()
	defer c.Stop()

	if err := service.Run(); err != nil {
		fmt.Println(err)
	}
}
