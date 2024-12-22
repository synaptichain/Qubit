package internal

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/blake2b"
)

// GetAddress genera una dirección única, devuelve la clave privada asociada y asigna saldo inicial.
func GetAddress(db *Database, initialBalance int64) (string, *ecdsa.PrivateKey, error) {
	// Generar una clave privada
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", nil, err
	}

	// Crear un hash Blake2b de la clave pública
	pubKey := append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...)
	address := blake2b.Sum256(pubKey)

	// Convertir la dirección a formato hexadecimal
	accountAddress := hex.EncodeToString(address[:])

	// Verificar si la cuenta ya existe, si no, guardar el saldo
	exists, err := db.AccountExists(accountAddress)
	if err != nil {
		return "", nil, fmt.Errorf("error verificando si la cuenta existe: %w", err)
	}

	if !exists {
		fmt.Printf("Guardando saldo inicial de %d para la nueva dirección: %s\n", initialBalance, accountAddress)
		err = db.SaveBalance(accountAddress, initialBalance)
		if err != nil {
			return "", nil, fmt.Errorf("error guardando el saldo inicial: %w", err)
		}
	}

	// Devolver la dirección y la clave privada
	return accountAddress, privateKey, nil
}
