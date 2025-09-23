package api

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"log"
	"strings"

	"chelbit/excelms/internal/excel"
	"chelbit/excelms/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

type Server struct {
	app       *fiber.App
	authToken string
}

func NewServer(authToken string) *fiber.App {
	app := fiber.New(fiber.Config{
		BodyLimit: 50 * 1024 * 1024,
	})
	app.Use(logger.New())
	s := &Server{app: app, authToken: authToken}
	s.registerRoutes()
	return s.app
}

func (s *Server) registerRoutes() {
	s.app.Get("/health", s.healthCheck)
	apiGroup := s.app.Group("/api")
	if s.authToken != "" {
		apiGroup.Use(s.authMiddleware)
	}
	apiGroup.Post("/parse", s.handleImport)
	apiGroup.Post("/export", s.handleExport)
}

func (s *Server) authMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing authorization header"})
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Malformed authorization header"})
	}
	if parts[1] != s.authToken {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
	}
	return c.Next()
}

func (s *Server) healthCheck(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
}

func (s *Server) handleImport(c *fiber.Ctx) error {
	schemaJSON := c.FormValue("schema")
	if schemaJSON == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "schema field is required"})
	}

	var schema models.ImportSchema
	if err := json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "failed to parse schema JSON", "details": err.Error()})
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file field is required"})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to open uploaded file"})
	}
	defer file.Close()

	c.Set("Content-Type", "application/x-ndjson")
	c.Set("Transfer-Encoding", "chunked")

	// Use a stream writer for memory efficiency
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		importer := excel.NewImporter(file, schema, w)
		if err := importer.Process(); err != nil {
			// We can't write a JSON error here as the headers are already sent.
			// We can log it. The client will see a broken stream.
			log.Printf("Error during import processing: %v", err)
		}
		if err := w.Flush(); err != nil {
			log.Printf("Error flushing stream writer: %v", err)
		}
	})

	return nil
}

func (s *Server) handleExport(c *fiber.Ctx) error {
	var request models.ExportRequest
	var templateBytes []byte

	contentType := c.Get("Content-Type")
	if strings.Contains(contentType, "multipart/form-data") {
		requestJSON := c.FormValue("request")
		if err := json.Unmarshal([]byte(requestJSON), &request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "failed to parse request JSON from form", "details": err.Error()})
		}
		if templateFile, err := c.FormFile("template"); err == nil {
			file, err := templateFile.Open()
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to open template file"})
			}
			defer file.Close()
			buf := new(bytes.Buffer)
			if _, err := buf.ReadFrom(file); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to read template file to buffer"})
			}
			templateBytes = buf.Bytes()
		}
	} else { // Assume application/json
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "failed to parse JSON body", "details": err.Error()})
		}
	}

	if request.TemplateBase64 != "" {
		decoded, err := base64.StdEncoding.DecodeString(request.TemplateBase64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "failed to decode templateBase64"})
		}
		templateBytes = decoded
	}

	exporter := excel.NewExporter(request, templateBytes)
	resultBytes, err := exporter.Process()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate excel file", "details": err.Error()})
	}

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", "attachment; filename=export.xlsx")
	return c.Send(resultBytes)
}
