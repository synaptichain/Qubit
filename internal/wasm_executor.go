package internal

import (
	"errors"
	"fmt"
	"time"

	"github.com/wasmerio/wasmer-go/wasmer"
)

// WASMContract representa un contrato en formato WASM.
type WASMContract struct {
	ID       string
	Owner    string
	WASMCode []byte
	Logs     []string // Registro de eventos durante la ejecución
}

// NewWASMContract crea un nuevo contrato WASM.
func NewWASMContract(id, owner string, wasmCode []byte) *WASMContract {
	return &WASMContract{
		ID:       id,
		Owner:    owner,
		WASMCode: wasmCode,
		Logs:     []string{},
	}
}

// Log agrega un mensaje al registro del contrato.
func (c *WASMContract) Log(message string) {
	c.Logs = append(c.Logs, message)
}

// Execute ejecuta el contrato WASM con los parámetros dados.
func (c *WASMContract) Execute(input []byte) ([]byte, error) {
	c.Log("Iniciando la ejecución del contrato WASM.")

	// Crear un nuevo motor WASM y un store.
	engine := wasmer.NewEngine()
	store := wasmer.NewStore(engine)

	// Compilar el código WASM.
	module, err := wasmer.NewModule(store, c.WASMCode)
	if err != nil {
		c.Log(fmt.Sprintf("Error al compilar el contrato WASM: %s", err))
		return nil, fmt.Errorf("error al compilar el contrato WASM: %w", err)
	}

	// Crear un nuevo ambiente WASM.
	importObject := wasmer.NewImportObject()
	instance, err := wasmer.NewInstance(module, importObject)
	if err != nil {
		c.Log(fmt.Sprintf("Error al crear la instancia WASM: %s", err))
		return nil, fmt.Errorf("error al crear la instancia WASM: %w", err)
	}

	// Buscar la función `execute` en el contrato.
	executeFunc, err := instance.Exports.GetFunction("execute")
	if err != nil {
		c.Log("La función 'execute' no está definida en el contrato WASM.")
		return nil, errors.New("la función 'execute' no está definida en el contrato WASM")
	}

	// Ejecutar con un límite de tiempo.
	resultChan := make(chan []byte)
	errorChan := make(chan error)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				errorChan <- fmt.Errorf("pánico durante la ejecución del contrato WASM: %v", r)
			}
		}()

		// Llamar a la función `execute` con la entrada proporcionada.
		result, err := executeFunc(input)
		if err != nil {
			errorChan <- fmt.Errorf("error al ejecutar el contrato WASM: %w", err)
			return
		}

		// Validar el tipo del resultado antes de convertirlo.
		if byteResult, ok := result.([]byte); ok {
			resultChan <- byteResult
		} else {
			errorChan <- fmt.Errorf("resultado inesperado al ejecutar el contrato WASM")
		}
	}()

	select {
	case result := <-resultChan:
		c.Log("Contrato WASM ejecutado exitosamente.")
		return result, nil
	case err := <-errorChan:
		c.Log(fmt.Sprintf("Error durante la ejecución del contrato WASM: %s", err))
		return nil, err
	case <-time.After(5 * time.Second): // Límite de tiempo de 5 segundos
		c.Log("Ejecución del contrato WASM terminada por tiempo excedido.")
		return nil, errors.New("la ejecución del contrato WASM excedió el tiempo permitido")
	}
}
