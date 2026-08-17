package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"coldchain-alert/internal/auth"
	"coldchain-alert/internal/httpapi"
	"coldchain-alert/internal/service"
	"coldchain-alert/internal/simulator"
	"coldchain-alert/internal/store"
)

func main() {
	address := flag.String("addr", ":8080", "HTTP listen address")
	database := flag.String("db", "./data/coldchain.db", "bbolt database path")
	flag.Parse()
	storage, err := store.Open(*database)
	if err != nil {
		log.Fatal(err)
	}
	defer storage.Close()
	devices := service.NewDeviceService(storage)
	alerts := service.NewAlertService(storage)
	dashboard := service.NewDashboardService(storage)
	events := service.NewEventService(storage)
	audits := service.NewAuditService(storage)
	reports := service.NewReportService(storage)
	workflow := service.NewWorkflowService(devices, alerts, events, audits)
	if count, countErr := storage.Count(store.BucketZones); countErr != nil {
		log.Fatal(countErr)
	} else if count == 0 {
		if _, seedErr := simulator.Seed(workflow); seedErr != nil {
			log.Fatal(seedErr)
		}
	}
	manager := auth.NewManager()
	api := httpapi.New(manager, devices, alerts, dashboard, events, workflow, audits, reports)
	server := &http.Server{Addr: *address, Handler: api.Handler()}
	fmt.Printf("coldchain-alert listening on %s\n", *address)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
