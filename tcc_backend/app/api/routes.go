package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kaiquemsa/nlp-sql-backend/app/crud"
	"github.com/kaiquemsa/nlp-sql-backend/app/internal/supabase"
)

func RegisterRoutes(app *fiber.App, queryHandler *QueryHandler, supabaseService *supabase.SupabaseService) {
	api := app.Group("/api")
	api.Post("/query", queryHandler.HandleQuery)
	api.Post("/save-chat", crud.SaveChatMessage(supabaseService))
	api.Get("/get-history", crud.GetHistory(supabaseService))
	api.Get("/get-embeddings", crud.GetEmbeddings(supabaseService))
	api.Put("/update-embedding/:id", crud.UpdateEmbedding(supabaseService))
	api.Delete("/delete-embedding/:id", crud.DeleteEmbedding(supabaseService))
	api.Get("/generate-embedding", crud.GenerateEmbeddingsFromStruct(supabaseService))
	api.Post("/exec-query", func(c *fiber.Ctx) error {
		query := string(c.Body())
		result, err := supabaseService.ExecuteQuery(query)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.JSON(result)
	})
}
