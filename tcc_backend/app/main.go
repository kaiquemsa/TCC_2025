package main

import (
	"log"

	"github.com/kaiquemsa/nlp-sql-backend/app/api"
	"github.com/kaiquemsa/nlp-sql-backend/app/config"
)

func main() {
	// Carrega configurações
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Erro ao carregar configurações: %v", err)
	}

	// Inicializa e executa o servidor
	server := api.NewServer(cfg)
	log.Fatal(server.Start())
}
