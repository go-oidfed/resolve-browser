package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-oidfed/resolve-browser/internal/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(map[string]interface{}{
				"error": err.Error(),
			})
		},
	})

	app.Use(logger.New())
	app.Use(recover.New())

	execPath, _ := os.Executable()
	execDir := filepath.Dir(execPath)
	staticPath := filepath.Join(execDir, "frontend", "static")
	
	if _, err := os.Stat(staticPath); os.IsNotExist(err) {
		cwd, _ := os.Getwd()
		staticPath = filepath.Join(cwd, "frontend", "static")
	}

	app.Use("/", filesystem.New(filesystem.Config{
		Root:   http.Dir(staticPath),
		Browse: false,
		MaxAge: 3600,
	}))

	api := app.Group("/api")
	api.Post("/resolve", handlers.ResolveHandler)
	api.Post("/resolve/preview", handlers.ResolvePreviewHandler)

	log.Println("Starting server on :8080")
	log.Println("Serving static files from:", staticPath)
	if err := app.Listen(":8080"); err != nil {
		log.Fatal(err)
	}
}
