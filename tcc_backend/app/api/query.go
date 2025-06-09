package api

import (
	"github.com/kaiquemsa/nlp-sql-backend/app/internal/services"

	"github.com/gofiber/fiber/v2"
)

type QueryHandler struct {
	service *services.QueryService
}

func NewQueryHandler(service *services.QueryService) *QueryHandler {
	return &QueryHandler{
		service: service,
	}
}

type QueryRequest struct {
	Question string `json:"question"`
	History string `json:"history"`
	Uuid string `json:"uuid"`
}

type QueryResponse struct {
	SQL         string      `json:"sql"`
	Data        interface{} `json:"data,omitempty"`
	Explanation string      `json:"explanation"`
}

func (h *QueryHandler) HandleQuery(c *fiber.Ctx) error {
	var req QueryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Erro ao processar requisição",
		})
	}

	sql, data, err := h.service.ProcessQuery(req.Question, req.History, req.Uuid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(QueryResponse{
		SQL:         sql,
		Data:        data,
		Explanation: "Consulta gerada com base na sua pergunta",
	})
}
