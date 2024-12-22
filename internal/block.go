package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// Block representa un bloque en la blockchain.
type Block struct {
	Index         int
	Timestamp     string
	Transactions  []Transaction
	PrevHash      string
	Hash          string
	MetadataRef   string // Referencia a metadatos almacenados en la base de datos PostgreSQL
	WASMContracts []WASMContract
}

// NewBlock crea un nuevo bloque.
func NewBlock(index int, transactions []Transaction, prevHash, metadataRef string) *Block {
	block := &Block{
		Index:        index,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		Transactions: transactions,
		PrevHash:     prevHash,
		MetadataRef:  metadataRef,
	}
	block.Hash = block.CalculateHash()
	return block
}

// CalculateHash calcula el hash de un bloque basado en su contenido.
func (b *Block) CalculateHash() string {
	transactionsHash := b.HashTransactions()
	record := fmt.Sprintf("%d%s%s%s%s", b.Index, b.Timestamp, transactionsHash, b.PrevHash, b.MetadataRef)
	h := sha256.New()
	h.Write([]byte(record))
	return hex.EncodeToString(h.Sum(nil))
}

// HashTransactions calcula el hash de todas las transacciones en un bloque.
func (b *Block) HashTransactions() string {
	h := sha256.New()
	for _, tx := range b.Transactions {
		txRecord := fmt.Sprintf("%s%s%d", tx.From, tx.To, tx.Amount)
		h.Write([]byte(txRecord))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// AddWASMContract agrega un contrato WASM al bloque.
func (b *Block) AddWASMContract(contract WASMContract) {
	b.WASMContracts = append(b.WASMContracts, contract)
}

// Validate verifica la integridad del bloque comparando su hash con el calculado.
func (b *Block) Validate() bool {
	calculatedHash := b.CalculateHash()
	return b.Hash == calculatedHash
}

// Serialize convierte el bloque en un formato que se pueda almacenar.
func (b *Block) Serialize() (map[string]interface{}, error) {
	wasmContractsJSON, err := json.Marshal(b.WASMContracts)
	if err != nil {
		return nil, fmt.Errorf("error serializando contratos WASM: %w", err)
	}

	return map[string]interface{}{
		"index":          b.Index,
		"timestamp":      b.Timestamp,
		"transactions":   b.Transactions,
		"prev_hash":      b.PrevHash,
		"hash":           b.Hash,
		"metadata_ref":   b.MetadataRef,
		"wasm_contracts": string(wasmContractsJSON),
	}, nil
}

// Deserialize reconstruye un bloque a partir de datos almacenados.
func Deserialize(data map[string]interface{}) (*Block, error) {
	var transactions []Transaction
	transactionsData, _ := json.Marshal(data["transactions"])
	if err := json.Unmarshal(transactionsData, &transactions); err != nil {
		return nil, fmt.Errorf("error deserializando transacciones: %w", err)
	}

	var wasmContracts []WASMContract
	if data["wasm_contracts"] != nil {
		wasmContractsData, _ := json.Marshal(data["wasm_contracts"])
		if err := json.Unmarshal(wasmContractsData, &wasmContracts); err != nil {
			return nil, fmt.Errorf("error deserializando contratos WASM: %w", err)
		}
	}

	return &Block{
		Index:         int(data["index"].(float64)),
		Timestamp:     data["timestamp"].(string),
		Transactions:  transactions,
		PrevHash:      data["prev_hash"].(string),
		Hash:          data["hash"].(string),
		MetadataRef:   data["metadata_ref"].(string),
		WASMContracts: wasmContracts,
	}, nil
}
