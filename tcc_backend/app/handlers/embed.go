package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func GenerateEmbeddingLocal(text string) ([]float32, error) {
	// Caminho do executável (usado no Render)
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter caminho do executável: %v", err)
	}

	// Diretório base do binário
	baseDir := filepath.Dir(exePath)
	scriptPath := filepath.Join(baseDir, "app", "python", "embed.py")

	// Se não existir (caso local, go run .), tenta caminho alternativo
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		// Caminho relativo ao projeto local
		scriptPath = filepath.Join(".", "python", "embed.py")
	}

	cmd := exec.Command("python3", scriptPath)
	cmd.Stdin = bytes.NewBufferString(text)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("erro ao executar embed.py: %v\nSaída: %s", err, string(output))
	}

	var embedding []float32
	if err := json.Unmarshal(output, &embedding); err != nil {
		return nil, fmt.Errorf("erro ao parsear o resultado do Python: %v", err)
	}

	return embedding, nil
}
