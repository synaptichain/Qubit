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
func GetAddress(db *Database) (string, *ecdsa.PrivateKey, error) {
	// Generar una clave privada
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", nil, err
	}

	// Crear un hash Blake2b de la clave pública
	pubKey := append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...)
	address := blake2b.Sum256(pubKey)

	// Asignar saldo inicial a la nueva cuenta (1000 tokens)
	accountAddress := hex.EncodeToString(address[:])

	// Verificar si la cuenta ya existe, si no, guardar el saldo
	exists, err := db.AccountExists(accountAddress)
	if err != nil {
		return "", nil, err
	}

	if !exists {
		fmt.Println("Guardando saldo para la nueva dirección:", accountAddress)
		err = db.SaveBalance(accountAddress, 1000)
		if err != nil {
			return "", nil, err
		}
	}

	// Devolver la dirección como hexadecimal y la clave privada
	return accountAddress, privateKey, nil
}
