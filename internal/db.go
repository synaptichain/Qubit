package internal

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq" // Driver de PostgreSQL
)

type Database struct {
	Connection *sql.DB
}

type Transaction struct {
	From   string
	To     string
	Amount int64
}

// InitDB inicializa la base de datos PostgreSQL y crea las tablas necesarias.
func InitDB(connectionString string) (*Database, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	queries := []string{
		`CREATE TABLE IF NOT EXISTS blocks (
			id SERIAL PRIMARY KEY,
			block_index INTEGER NOT NULL,
			timestamp TIMESTAMP NOT NULL,
			transactions JSONB NOT NULL,
			hash TEXT NOT NULL,
			prev_hash TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS balances (
			account TEXT PRIMARY KEY,
			balance BIGINT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS transactions (
			id SERIAL PRIMARY KEY,
			from_account TEXT NOT NULL,
			to_account TEXT NOT NULL,
			amount BIGINT NOT NULL,
			timestamp TIMESTAMP NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS pending_transactions (
			id SERIAL PRIMARY KEY,
			from_account TEXT NOT NULL,
			to_account TEXT NOT NULL,
			amount BIGINT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS wasm_contracts (
			id TEXT PRIMARY KEY,
			owner TEXT NOT NULL,
			wasm_code BYTEA NOT NULL
		);`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return nil, fmt.Errorf("error creando tabla: %w", err)
		}
	}

	return &Database{Connection: db}, nil
}

// SaveBlock guarda un bloque en la base de datos.
func (d *Database) SaveBlock(block Block) error {
	blockData, err := json.Marshal(block.Transactions)
	if err != nil {
		return fmt.Errorf("error serializando transacciones: %w", err)
	}

	_, err = d.Connection.Exec(
		"INSERT INTO blocks (block_index, timestamp, transactions, hash, prev_hash) VALUES ($1, $2, $3, $4, $5)",
		block.Index, block.Timestamp, string(blockData), block.Hash, block.PrevHash,
	)
	if err != nil {
		return fmt.Errorf("error guardando bloque en la base de datos: %w", err)
	}

	fmt.Printf("Bloque #%d guardado en la base de datos con éxito\n", block.Index)
	return nil
}

// LoadBlocks carga todos los bloques desde la base de datos.
func (d *Database) LoadBlocks() ([]Block, error) {
	rows, err := d.Connection.Query("SELECT block_index, timestamp, transactions, hash, prev_hash FROM blocks ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocks []Block
	for rows.Next() {
		var block Block
		var transactionsJSON string

		if err := rows.Scan(&block.Index, &block.Timestamp, &transactionsJSON, &block.Hash, &block.PrevHash); err != nil {
			return nil, err
		}

		// Deserializar las transacciones desde JSON
		if err := json.Unmarshal([]byte(transactionsJSON), &block.Transactions); err != nil {
			return nil, fmt.Errorf("error deserializando transacciones: %w", err)
		}

		blocks = append(blocks, block)
	}

	fmt.Printf("Cargados %d bloques desde la base de datos\n", len(blocks))
	return blocks, nil
}

// SaveBalance guarda o actualiza el saldo de una cuenta.
func (d *Database) SaveBalance(account string, balance int64) error {
	_, err := d.Connection.Exec(
		"INSERT INTO balances (account, balance) VALUES ($1, $2) ON CONFLICT (account) DO UPDATE SET balance = $2",
		account, balance,
	)
	if err != nil {
		return fmt.Errorf("error guardando saldo: %w", err)
	}

	fmt.Printf("Saldo de la cuenta %s actualizado a %d\n", account, balance)
	return nil
}

// GetBalance obtiene el saldo de una cuenta.
func (d *Database) GetBalance(account string) (int64, error) {
	var balance int64
	err := d.Connection.QueryRow("SELECT balance FROM balances WHERE account = $1", account).Scan(&balance)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return balance, err
}

// AccountExists verifica si una cuenta existe en la base de datos.
func (d *Database) AccountExists(account string) (bool, error) {
	var exists bool
	err := d.Connection.QueryRow("SELECT EXISTS(SELECT 1 FROM balances WHERE account = $1)", account).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error verificando existencia de cuenta: %w", err)
	}
	return exists, nil
}

// GetAllAccounts obtiene todas las cuentas de la base de datos.
func (d *Database) GetAllAccounts() ([]string, error) {
	query := `SELECT account FROM balances`
	rows, err := d.Connection.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error al obtener todas las cuentas: %w", err)
	}
	defer rows.Close()

	var accounts []string
	for rows.Next() {
		var account string
		if err := rows.Scan(&account); err != nil {
			return nil, fmt.Errorf("error al escanear cuenta: %w", err)
		}
		accounts = append(accounts, account)
	}

	fmt.Printf("Cargadas %d cuentas desde la base de datos\n", len(accounts))
	return accounts, nil
}

// SaveTransaction guarda una transacción en la base de datos.
func (d *Database) SaveTransaction(from, to string, amount int64, timestamp string) error {
	_, err := d.Connection.Exec(
		"INSERT INTO transactions (from_account, to_account, amount, timestamp) VALUES ($1, $2, $3, $4)",
		from, to, amount, timestamp,
	)
	if err != nil {
		return fmt.Errorf("error guardando transacción: %w", err)
	}

	fmt.Printf("Transacción guardada: de %s a %s por %d\n", from, to, amount)
	return nil
}

// LoadTransactions carga todas las transacciones desde la base de datos.
func (d *Database) LoadTransactions() ([]Transaction, error) {
	rows, err := d.Connection.Query("SELECT from_account, to_account, amount FROM transactions ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var t Transaction
		if err := rows.Scan(&t.From, &t.To, &t.Amount); err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}

	fmt.Printf("Cargadas %d transacciones desde la base de datos\n", len(transactions))
	return transactions, nil
}

// AddPendingTransaction agrega una transacción pendiente.
func (d *Database) AddPendingTransaction(from, to string, amount int64) error {
	query := `INSERT INTO pending_transactions (from_account, to_account, amount) VALUES ($1, $2, $3)`
	_, err := d.Connection.Exec(query, from, to, amount)
	if err != nil {
		return fmt.Errorf("error añadiendo transacción pendiente: %w", err)
	}

	fmt.Printf("Transacción pendiente añadida: de %s a %s por %d\n", from, to, amount)
	return nil
}

// GetPendingTransactions carga todas las transacciones pendientes.
func (d *Database) GetPendingTransactions() ([]Transaction, error) {
	query := `SELECT from_account, to_account, amount FROM pending_transactions`
	rows, err := d.Connection.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error al obtener transacciones pendientes: %w", err)
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var t Transaction
		if err := rows.Scan(&t.From, &t.To, &t.Amount); err != nil {
			return nil, fmt.Errorf("error al escanear transacción pendiente: %w", err)
		}
		transactions = append(transactions, t)
	}

	fmt.Printf("Cargadas %d transacciones pendientes desde la base de datos\n", len(transactions))
	return transactions, nil
}

// ClearPendingTransactions limpia todas las transacciones pendientes.
func (d *Database) ClearPendingTransactions() error {
	query := `DELETE FROM pending_transactions`
	_, err := d.Connection.Exec(query)
	if err != nil {
		return fmt.Errorf("error al limpiar transacciones pendientes: %w", err)
	}

	fmt.Println("Transacciones pendientes eliminadas de la base de datos")
	return nil
}

// UpdateBalances actualiza los saldos de las cuentas al realizar una transacción.
func (d *Database) UpdateBalances(from, to string, amount int64) error {
	timestamp := time.Now().Format(time.RFC3339)

	fromBalance, err := d.GetBalance(from)
	if err != nil {
		return fmt.Errorf("error obteniendo saldo de origen: %w", err)
	}

	toBalance, err := d.GetBalance(to)
	if err != nil {
		return fmt.Errorf("error obteniendo saldo de destino: %w", err)
	}

	if fromBalance < amount {
		return fmt.Errorf("saldo insuficiente en la cuenta %s", from)
	}

	err = d.SaveBalance(from, fromBalance-amount)
	if err != nil {
		return fmt.Errorf("error actualizando saldo de origen: %w", err)
	}

	err = d.SaveBalance(to, toBalance+amount)
	if err != nil {
		return fmt.Errorf("error actualizando saldo de destino: %w", err)
	}

	err = d.SaveTransaction(from, to, amount, timestamp)
	if err != nil {
		return fmt.Errorf("error guardando transacción: %w", err)
	}

	fmt.Printf("Transacción completada: de %s a %s por %d\n", from, to, amount)
	return nil
}

// SaveWASMContract guarda un contrato WASM en la base de datos.
func (d *Database) SaveWASMContract(contract WASMContract) error {
	_, err := d.Connection.Exec(
		"INSERT INTO wasm_contracts (id, owner, wasm_code) VALUES ($1, $2, $3)",
		contract.ID, contract.Owner, contract.WASMCode,
	)
	if err != nil {
		return fmt.Errorf("error guardando contrato WASM: %w", err)
	}

	fmt.Printf("Contrato WASM %s guardado en la base de datos con éxito\n", contract.ID)
	return nil
}

// LoadWASMContract carga un contrato WASM desde la base de datos por su ID.
func (d *Database) LoadWASMContract(id string) (*WASMContract, error) {
	var contract WASMContract

	err := d.Connection.QueryRow(
		"SELECT id, owner, wasm_code FROM wasm_contracts WHERE id = $1",
		id,
	).Scan(&contract.ID, &contract.Owner, &contract.WASMCode)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("error cargando contrato WASM: %w", err)
	}

	return &contract, nil
}
