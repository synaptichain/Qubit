package internal

import (
	"errors"
	"fmt"
)

// TokenSupply maneja el suministro de tokens en la blockchain.
type TokenSupply struct {
	TotalSupply int64            // Suministro total de tokens
	Balance     map[string]int64 // Saldos por cuenta
}

// NewTokenSupply inicializa el suministro de tokens.
func NewTokenSupply(totalSupply int64, initialHolder string) *TokenSupply {
	return &TokenSupply{
		TotalSupply: totalSupply,
		Balance: map[string]int64{
			initialHolder: totalSupply,
		},
	}
}

// Mint crea nuevos tokens y los asigna a una cuenta específica.
func (ts *TokenSupply) Mint(to string, amount int64) error {
	if amount <= 0 {
		return errors.New("el monto a acuñar debe ser mayor que cero")
	}

	// Incrementa el suministro total y asigna los tokens a la cuenta.
	ts.TotalSupply += amount
	ts.Balance[to] += amount
	fmt.Printf("Se han acuñado %d tokens para la cuenta %s\n", amount, to)
	return nil
}

// Transfer realiza una transferencia de tokens entre cuentas.
func (ts *TokenSupply) Transfer(from, to string, amount int64) error {
	if amount <= 0 {
		return errors.New("el monto de la transferencia debe ser mayor que cero")
	}

	if ts.Balance[from] < amount {
		return errors.New("saldo insuficiente")
	}

	ts.Balance[from] -= amount
	ts.Balance[to] += amount
	fmt.Printf("Transferencia de %d tokens de %s a %s realizada con éxito\n", amount, from, to)
	return nil
}

// GetBalance obtiene el saldo de una cuenta específica.
func (ts *TokenSupply) GetBalance(account string) int64 {
	return ts.Balance[account]
}
