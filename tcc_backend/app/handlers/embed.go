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

	// tenta achar o Python em vários locais possíveis
	possiblePaths := []string{
		"/app/venv/bin/python",                                            // runtime Railway
		filepath.Join(baseDir, "..", "venv", "bin", "python"),             // sobe um nível (quando binário está em /app/tcc_backend)
		filepath.Join(baseDir, "venv", "bin", "python"),                   // local build
		filepath.Join(baseDir, "app", "python", ".venv", "bin", "python"), // estrutura antiga
	}

	var pythonExe string
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			pythonExe = path
			break
		}
	}

	// debug pra ver qual caminho foi pego
	fmt.Println("🔍 Usando Python em:", pythonExe)

	if pythonExe == "" {
		pythonExe = "python3"
	}

	cmd := exec.Command(pythonExe, scriptPath)
	cmd.Stdin = bytes.NewBufferString(text)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("erro ao executar embed.py: %v\nSaída: %s", err, string(output))
	}
	fmt.Println("🧩 Saída bruta do Python:")
	fmt.Println(string(output))

	var embedding []float32
	if err := json.Unmarshal(output, &embedding); err != nil {
		return nil, fmt.Errorf("erro ao parsear o resultado do Python: %v", err)
	}
	return embedding, nil
}
