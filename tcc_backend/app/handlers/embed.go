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
	wd, _ := os.Getwd()
	scriptPath := filepath.Join(wd, "python", "embed.py")
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
