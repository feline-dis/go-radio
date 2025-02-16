package main

import (
	"fmt"
	"net/http"

	"github.com/feline-dis/go-radio/internal/controller"
	"github.com/feline-dis/go-radio/internal/download"
	"github.com/feline-dis/go-radio/internal/ingest"
	"github.com/feline-dis/go-radio/internal/orchestrator"
	"github.com/feline-dis/go-radio/internal/picker"
)

type ServerConfig struct {
	Host       string
	Port       int
	InjestPath string
	CachePath  string
	MaxWorkers int
}

type Server struct {
	config          ServerConfig
	downloadService *download.DownloadService
	dataService     *ingest.DataService
	router          *http.ServeMux
}

func NewServer(config ServerConfig) *Server {
	router := http.NewServeMux()

	dataService := ingest.NewDataService(config.InjestPath, config.MaxWorkers)
	dataService.Start()

	if err := dataService.Ingest(); err != nil {
		fmt.Printf("failed to ingest: %v", err)
	}

	dataService.WaitForJobs()
	dataService.Stop()

	fmt.Println("ingest complete")

	downloadService := download.NewDownloadService(config.CachePath, config.MaxWorkers)
	pickerService := picker.NewPickerService(dataService)

	webSocketController := controller.NewWebsocketController()
	webSocketController.RegisterRoutes(router)

	fileController := controller.NewFileController(router, downloadService, dataService)
	fileController.RegisterRoutes()

	orc := orchestrator.NewOrchestrator(downloadService, pickerService, webSocketController)
	orc.Start()

	return &Server{
		config:          config,
		downloadService: downloadService,
		dataService:     dataService,
		router:          router,
	}
}

func (s *Server) Start() {
	fmt.Println("starting server")

	http.ListenAndServe(fmt.Sprintf(":%d", s.config.Port), s.router)
}

func main() {
	config := ServerConfig{
		Host:       "localhost",
		Port:       8080,
		InjestPath: "./ingest",
		CachePath:  "./cache",
		MaxWorkers: 4,
	}

	NewServer(config).Start()
}
