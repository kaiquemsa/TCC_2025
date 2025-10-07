package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	OpenAIKey   string
	SupabaseURL string
	SupabaseKey string
}

func Load() (*Config, error) {
	// Tenta carregar o .env, mas ignora erro se não existir
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	cfg := &Config{
		Port:        port,
		OpenAIKey:   os.Getenv("OPENAI_API_KEY"),
		SupabaseURL: os.Getenv("SUPABASE_URL"),
		SupabaseKey: os.Getenv("SUPABASE_ANON_KEY"),
	}

	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func validateConfig(cfg *Config) error {
	if cfg.OpenAIKey == "" {
		return fmt.Errorf("OPENAI_API_KEY não configurada")
	}
	if cfg.SupabaseURL == "" {
		return fmt.Errorf("SUPABASE_URL não configurada")
	}
	if cfg.SupabaseKey == "" {
		return fmt.Errorf("SUPABASE_ANON_KEY não configurada")
	}
	return nil
}
