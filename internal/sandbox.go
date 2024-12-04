package internal

import (
	"fmt"

	"github.com/dop251/goja"
)

// Sandbox ejecuta un script JavaScript en un entorno seguro.
func Sandbox(script string, env map[string]interface{}) (string, error) {
	vm := goja.New()

	// Pasar las funciones y variables de entorno a la máquina virtual.
	for key, value := range env {
		vm.Set(key, value)
	}

	// Ejecutar el script.
	result, err := vm.RunString(script)
	if err != nil {
		return "", fmt.Errorf("error ejecutando script: %w", err)
	}

	return result.String(), nil
}
