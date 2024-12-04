package internal

import (
	"fmt"
)

// Blockchain representa una cadena de bloques.
type Blockchain struct {
	Blocks []Block
}

// NewBlockchain crea una nueva blockchain con un bloque génesis.
func NewBlockchain() *Blockchain {
	genesisBlock := NewBlock(0, "Bloque Génesis", "")
	return &Blockchain{
		Blocks: []Block{genesisBlock},
	}
}

// AddBlock agrega un nuevo bloque a la cadena utilizando datos como entrada.
func (bc *Blockchain) AddBlock(data string) {
	prevBlock := bc.Blocks[len(bc.Blocks)-1]
	newBlock := NewBlock(prevBlock.Index+1, data, prevBlock.Hash)
	bc.Blocks = append(bc.Blocks, newBlock)
}

// AddBlockFromBlock agrega un bloque ya construido a la cadena.
func (bc *Blockchain) AddBlockFromBlock(block Block) {
	// Validar si el nuevo bloque es válido antes de agregarlo
	if block.Index != len(bc.Blocks) {
		fmt.Printf("Error: el índice del bloque (%d) no es consecutivo\n", block.Index)
		return
	}
	if block.PrevHash != bc.Blocks[len(bc.Blocks)-1].Hash {
		fmt.Printf("Error: el hash previo del bloque no coincide con el último hash de la cadena\n")
		return
	}
	bc.Blocks = append(bc.Blocks, block)
}

// IsValid verifica la integridad de la cadena de bloques.
func (bc *Blockchain) IsValid() bool {
	// Verificar el bloque génesis
	if len(bc.Blocks) == 0 || bc.Blocks[0].Hash != bc.Blocks[0].CalculateHash() {
		fmt.Println("Error: El bloque génesis tiene un hash inválido")
		return false
	}

	for i := 1; i < len(bc.Blocks); i++ {
		currentBlock := bc.Blocks[i]
		prevBlock := bc.Blocks[i-1]

		// Verifica el hash del bloque actual
		if currentBlock.Hash != currentBlock.CalculateHash() {
			fmt.Printf("Error: Bloque %d tiene un hash inválido\n", currentBlock.Index)
			return false
		}

		// Verifica la relación con el hash del bloque anterior
		if currentBlock.PrevHash != prevBlock.Hash {
			fmt.Printf("Error: Bloque %d no está correctamente vinculado al bloque anterior\n", currentBlock.Index)
			return false
		}
	}
	return true
}
