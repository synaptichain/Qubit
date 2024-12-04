package internal

import (
	"errors"
	"fmt"

	"github.com/dop251/goja"
)

// Aquí inicializamos el mapa de balances para las direcciones
var balances = make(map[string]int64)

// Función para inicializar los balances de las direcciones
func initializeBalances() {
	// Inicializa balances para algunas direcciones de ejemplo
	balances["address1"] = 1000 // Dirección 1 con saldo inicial de 1000
	balances["address2"] = 500  // Dirección 2 con saldo inicial de 500
}

// Función para obtener el saldo de una dirección
func getBalance(address string) int64 {
	return balances[address]
}

// Función para hacer una transferencia
func transfer(from, to string, amount int64) error {
	// Verifica si la dirección 'from' tiene suficiente saldo
	if getBalance(from) < amount {
		return fmt.Errorf("Saldo insuficiente en la dirección %s", from)
	}

	// Realiza la transferencia (decrementa el saldo de 'from' y aumenta el saldo de 'to')
	balances[from] -= amount
	balances[to] += amount
	return nil
}

// Contract representa un contrato inteligente en la blockchain.
type Contract struct {
	ID       string                 // Identificador único del contrato
	Owner    string                 // Dueño del contrato
	Compiled string                 // Código del contrato (JavaScript)
	Params   map[string]interface{} // Parámetros iniciales del contrato
	IsActive bool                   // Estado del contrato
}

// NewContract crea un nuevo contrato inteligente.
func NewContract(id, owner, compiled string, params map[string]interface{}) *Contract {
	return &Contract{
		ID:       id,
		Owner:    owner,
		Compiled: compiled,
		Params:   params,
		IsActive: true,
	}
}

// Execute ejecuta un contrato inteligente en un entorno seguro.
func (c *Contract) Execute(input map[string]interface{}, env map[string]interface{}) (string, error) {
	if !c.IsActive {
		return "", errors.New("el contrato no está activo")
	}

	// Crear una nueva máquina virtual segura para ejecutar el código del contrato.
	vm := goja.New()

	// Registrar funciones del entorno (como mint y transfer).
	for key, value := range env {
		if err := vm.Set(key, value); err != nil {
			return "", fmt.Errorf("error configurando variable '%s' en la VM: %w", key, err)
		}
	}

	// Registrar parámetros iniciales del contrato.
	for key, value := range c.Params {
		if err := vm.Set(key, value); err != nil {
			return "", fmt.Errorf("error configurando parámetro '%s' en la VM: %w", key, err)
		}
	}

	// Registrar la entrada del usuario.
	if err := vm.Set("input", input); err != nil {
		return "", fmt.Errorf("error configurando entrada en la VM: %w", err)
	}

	// Registrar el código del contrato y envolverlo en una función 'execute'
	contractFunction := fmt.Sprintf("function execute(input) { %s }", c.Compiled)

	// Ejecutar el código del contrato.
	_, err := vm.RunString(contractFunction)
	if err != nil {
		return "", fmt.Errorf("error ejecutando el contrato: %w", err)
	}

	// Intentar obtener la función `execute` del contrato.
	executeFn, ok := goja.AssertFunction(vm.Get("execute"))
	if !ok {
		return "", errors.New("la función 'execute' no está definida en el contrato")
	}

	// Llamar a la función `execute` pasando `input` como parámetro.
	result, err := executeFn(goja.Undefined(), vm.ToValue(input))
	if err != nil {
		return "", fmt.Errorf("error ejecutando la función 'execute': %w", err)
	}

	// Retornar el resultado.
	return result.String(), nil
}
