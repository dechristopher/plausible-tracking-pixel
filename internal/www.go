package internal

import (
	// stdlib
	"os"

	// third-party
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

const logFormat = "${ip} ${header:x-forwarded-for} ${header:x-real-ip} ${header:fly-client-ip} " +
	"[${time}] ${pid} ${locals:requestid} \"${method} ${path} ${protocol}\" " +
	"${status} ${latency} \"${referrer}\" \"${ua}\"\n"

// WireHandlers builds all http routes
// into the fiber app context
func WireHandlers(r *fiber.App, corsAllowedOrigins string) {
	// recover from panics
	r.Use(recover.New())

	r.Use(requestid.New())

	// Configure CORS
	r.Use(cors.New(cors.Config{
		// TODO: programmatically create CORS origins
		AllowOrigins: corsAllowedOrigins,
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// STDOUT request logger
	r.Use(logger.New(logger.Config{
		// For more options, see the Config section
		TimeZone:   "local",
		TimeFormat: "2006-01-02T15:04:05-0700",
		Format:     logFormat,
		Output:     os.Stdout,
	}))
}
