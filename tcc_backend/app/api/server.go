package api

import (
	"github.com/kaiquemsa/nlp-sql-backend/app/config"
	"github.com/kaiquemsa/nlp-sql-backend/app/internal/services"
	"github.com/kaiquemsa/nlp-sql-backend/app/internal/supabase"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/timeout"
    "time"
)

type Server struct {
	app    *fiber.App
	config *config.Config
}

func NewServer(cfg *config.Config) *Server {
	app := fiber.New()

	// Timeout de 5 minutos para todas as rotas
	app.Use(timeout.New(func(c *fiber.Ctx) error {
		return c.Next()
	}, 5*time.Minute))

	// Middlewares
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,OPTIONS",
		AllowHeaders: "Content-Type",
	}))

	// Serviços
	// openAIService := services.NewOpenAIService(cfg.OpenAIKey)
	geminiService := services.NewGeminiService()
	supabaseService := supabase.NewSupabaseService(cfg.SupabaseURL, cfg.SupabaseKey)
	queryService := services.NewQueryService(geminiService, supabaseService)

	// Handlers
	queryHandler := NewQueryHandler(queryService)

	// Rotas (agora separado)
	RegisterRoutes(app, queryHandler, supabaseService)

	return &Server{
		app:    app,
		config: cfg,
	}
}

func (s *Server) Start() error {
	return s.app.Listen(":" + s.config.Port)
}
