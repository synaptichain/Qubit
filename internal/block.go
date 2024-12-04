package internal

import (
	"encoding/hex"
	"fmt"
	"time"

	"golang.org/x/crypto/blake2b"
)

// Block representa un bloque en la blockchain.
type Block struct {
	Index     int        // Índice del bloque
	Timestamp string     // Fecha y hora de creación
	Data      string     // Datos del bloque
	Hash      string     // Hash del bloque
	PrevHash  string     // Hash del bloque anterior
	Contracts []Contract // Contratos asociados al bloque
}

// CalculateHash genera el hash del bloque utilizando BLAKE2b.
func (b *Block) CalculateHash() string {
	record := fmt.Sprint(b.Index) + b.Timestamp + b.Data + b.PrevHash

	// Crear el hash con BLAKE2b
	hash := blake2b.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

// NewBlock crea un nuevo bloque con los datos proporcionados.
func NewBlock(index int, data string, prevHash string) Block {
	block := Block{
		Index:     index,
		Timestamp: time.Now().String(),
		Data:      data,
		PrevHash:  prevHash,
	}
	block.Hash = block.CalculateHash()
	return block
}

// AddContract agrega un contrato al bloque.
func (b *Block) AddContract(contract Contract) {
	b.Contracts = append(b.Contracts, contract)
}
