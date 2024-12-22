package internal

import (
	"fmt"
)

// Blockchain representa una cadena de bloques.
type Blockchain struct {
	Blocks []*Block
}

// NewBlockchain crea una nueva blockchain con un bloque génesis.
func NewBlockchain(metadataRef string) *Blockchain {
	genesisBlock := NewBlock(0, []Transaction{}, "", metadataRef)
	return &Blockchain{
		Blocks: []*Block{genesisBlock},
	}
}

// AddBlock agrega un nuevo bloque a la cadena utilizando transacciones.
func (bc *Blockchain) AddBlock(transactions []Transaction, metadataRef string) *Block {
	prevBlock := bc.Blocks[len(bc.Blocks)-1]
	newBlock := NewBlock(prevBlock.Index+1, transactions, prevBlock.Hash, metadataRef)
	bc.Blocks = append(bc.Blocks, newBlock)
	return newBlock
}

// FindWASMContractByID busca un contrato WASM en la blockchain.
func (bc *Blockchain) FindWASMContractByID(id string) *WASMContract {
	for _, block := range bc.Blocks {
		for _, contract := range block.WASMContracts {
			if contract.ID == id {
				return &contract
			}
		}
	}
	return nil
}

// AddWASMContract agrega un contrato WASM a un bloque reciente.
func (bc *Blockchain) AddWASMContract(contract WASMContract) *Block {
	latestBlock := bc.Blocks[len(bc.Blocks)-1]
	latestBlock.AddWASMContract(contract)
	latestBlock.Hash = latestBlock.CalculateHash()
	return latestBlock
}

// IsValid verifica la integridad de la cadena de bloques.
func (bc *Blockchain) IsValid() bool {
	if len(bc.Blocks) == 0 || bc.Blocks[0].Hash != bc.Blocks[0].CalculateHash() {
		fmt.Println("Error: El bloque génesis tiene un hash inválido")
		return false
	}

	for i := 1; i < len(bc.Blocks); i++ {
		currentBlock := bc.Blocks[i]
		prevBlock := bc.Blocks[i-1]

		if currentBlock.Hash != currentBlock.CalculateHash() {
			fmt.Printf("Error: Bloque %d tiene un hash inválido\n", currentBlock.Index)
			return false
		}

		if currentBlock.PrevHash != prevBlock.Hash {
			fmt.Printf("Error: Bloque %d no está correctamente vinculado al bloque anterior\n", currentBlock.Index)
			return false
		}
	}
	return true
}

// LoadBlockchain carga todos los bloques desde una base de datos.
func (bc *Blockchain) LoadBlockchain(db *Database) error {
	blocks, err := db.LoadBlocks()
	if err != nil {
		return fmt.Errorf("error al cargar bloques desde la base de datos: %w", err)
	}
	bc.Blocks = make([]*Block, len(blocks))
	for i, block := range blocks {
		bc.Blocks[i] = &block
	}
	return nil
}

// SaveBlockchain guarda la cadena completa en la base de datos.
func (bc *Blockchain) SaveBlockchain(db *Database) error {
	for _, block := range bc.Blocks {
		if err := db.SaveBlock(*block); err != nil {
			return fmt.Errorf("error al guardar bloque %d en la base de datos: %w", block.Index, err)
		}
	}
	return nil
}
