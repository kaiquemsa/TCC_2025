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
	// caminho do executável Go (binário)
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter caminho do executável: %v", err)
	}
	baseDir := filepath.Dir(exePath)

	// caminho do script
	scriptPath := filepath.Join(baseDir, "app", "python", "embed.py")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		// fallback para execução local (go run a partir do diretório do código)
		scriptPath = filepath.Join(".", "python", "embed.py")
	}

	// // Python do venv criado no build (na raiz do app)
	// venvPython := filepath.Join(baseDir, "venv", "bin", "python")

	// pythonExe := venvPython
	// if _, err := os.Stat(pythonExe); os.IsNotExist(err) {
	// 	// fallback: python3 do sistema (ex.: ambiente local sem venv)
	// 	pythonExe = "python3"
	// }

	// Caminhos possíveis para o Python
	possiblePaths := []string{
		"/app/venv/bin/python",                                            // Railway runtime
		filepath.Join(baseDir, "venv", "bin", "python"),                   // Local build
		filepath.Join(baseDir, "app", "python", ".venv", "bin", "python"), // antigo
	}

	var pythonExe string
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			pythonExe = path
			break
		}
	}

	if pythonExe == "" {
		pythonExe = "python3" // fallback local
	}
	fmt.Println("Usando Python em:", pythonExe)

	cmd := exec.Command(pythonExe, scriptPath)
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
