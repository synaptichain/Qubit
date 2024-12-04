package internal

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

// CompileContract compila un contrato inteligente Solidity usando solc.
func CompileContract(source string) (string, error) {
	// Guardar el código del contrato en un archivo temporal.
	tempFile := "temp.sol"
	err := os.WriteFile(tempFile, []byte(source), 0644)
	if err != nil {
		return "", fmt.Errorf("error creando archivo temporal: %w", err)
	}

	// Ejecutar el compilador Solidity.
	cmd := exec.Command("solc", "--bin", tempFile)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error compilando contrato: %s", stderr.String())
	}

	// Obtener el binario del contrato compilado.
	return out.String(), nil
}
