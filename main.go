package main

import (
	// stdlib
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	// package
	"github.com/dechristopher/plausible-tracking-pixel/internal"

	// third-party
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"

	//go:embed static/*
	static embed.FS

	// Plausible event API endpoint
	flagEventEndpoint *string
	// allowed CORS origins
	flagCorsOrigins *string
	// name of header that contains the visitor's IP
	flagVisitorIPHeader *string
	// listen address for the HTTP server
	flagAddress *string
)

func init() {
	flagEventEndpoint = flag.String("endpoint", "localhost:8000/api/event", "Plausible event API endpoint")
	flagCorsOrigins = flag.String("corsOrigins", "*", "Comma-separated list of CORS origins")
	flagVisitorIPHeader = flag.String("visitorIPHeader", "Fly-Client-IP", "Name of header that contains the visitor's IP")
	flagAddress = flag.String("address", ":8080", "Listening address")
	flag.Parse()

	if date == "unknown" {
		date = time.Now().Format("2006-01-02")
	}
}

func main() {
	_ = godotenv.Load()

	if os.Getenv("ENDPOINT") != "" {
		*flagEventEndpoint = os.Getenv("ENDPOINT")
	}

	if os.Getenv("CORS_ORIGINS") != "" {
		*flagCorsOrigins = os.Getenv("CORS_ORIGINS")
	}

	if os.Getenv("VISITOR_IP_HEADER") != "" {
		*flagVisitorIPHeader = os.Getenv("VISITOR_IP_HEADER")
	}

	if os.Getenv("ADDRESS") != "" {
		*flagAddress = os.Getenv("ADDRESS")
	}

	log.Printf("dchr/ptp: [%s-%s %s] init {endpoint: %s, address: %s}",
		version, commit, date, *flagEventEndpoint, *flagAddress)

	r := fiber.New(fiber.Config{
		ServerHeader:          "dchr/ptp",
		CaseSensitive:         true,
		ErrorHandler:          nil,
		DisableStartupMessage: true,
	})

	internal.WireHandlers(r, *flagCorsOrigins)

	controller := &internal.PixelController{
		EventEndpoint:   *flagEventEndpoint,
		VisitorIPHeader: *flagVisitorIPHeader,
		FS:              static,
	}
	r.Get("/", controller.Handler)

	// Graceful shutdown with SIGINT
	// SIGTERM and others will hard kill
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		_ = <-c
		log.Printf("dchr/ptp: shutdown")
		_ = r.Shutdown()
	}()

	// listen for connections on primary listening port
	if err := r.Listen(fmt.Sprintf("%s", *flagAddress)); err != nil {
		log.Println(err)
	}

	// Exit cleanly
	log.Printf("dchr/ptp: exit")
	os.Exit(0)
}
