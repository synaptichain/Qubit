package internal

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3" // Driver de SQLite
)

type Database struct {
	Connection *sql.DB
}

// InitDB inicializa la base de datos y crea las tablas necesarias.
func InitDB(dbFile string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}

	queries := []string{
		`CREATE TABLE IF NOT EXISTS blocks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			block_index INTEGER NOT NULL,
			timestamp TEXT NOT NULL,
			data TEXT NOT NULL,
			hash TEXT NOT NULL,
			prev_hash TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS balances (
			account TEXT PRIMARY KEY,
			balance INTEGER NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS transactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			from_account TEXT NOT NULL,
			to_account TEXT NOT NULL,
			amount INTEGER NOT NULL,
			timestamp TEXT NOT NULL
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
	_, err := d.Connection.Exec(
		"INSERT INTO blocks (block_index, timestamp, data, hash, prev_hash) VALUES (?, ?, ?, ?, ?)",
		block.Index, block.Timestamp, block.Data, block.Hash, block.PrevHash,
	)
	return err
}

// LoadBlocks carga todos los bloques desde la base de datos.
func (d *Database) LoadBlocks() ([]Block, error) {
	rows, err := d.Connection.Query("SELECT block_index, timestamp, data, hash, prev_hash FROM blocks ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocks []Block
	for rows.Next() {
		var block Block
		if err := rows.Scan(&block.Index, &block.Timestamp, &block.Data, &block.Hash, &block.PrevHash); err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}

	return blocks, nil
}

// SaveBalance guarda o actualiza el saldo de una cuenta.
func (d *Database) SaveBalance(account string, balance int64) error {
	_, err := d.Connection.Exec(
		"INSERT INTO balances (account, balance) VALUES (?, ?) ON CONFLICT(account) DO UPDATE SET balance=?",
		account, balance, balance,
	)
	return err
}

// GetBalance obtiene el saldo de una cuenta.
func (d *Database) GetBalance(account string) (int64, error) {
	var balance int64
	err := d.Connection.QueryRow("SELECT balance FROM balances WHERE account = ?", account).Scan(&balance)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return balance, err
}

// SaveTransaction guarda una transacción en la base de datos.
func (d *Database) SaveTransaction(from, to string, amount int64, timestamp string) error {
	_, err := d.Connection.Exec(
		"INSERT INTO transactions (from_account, to_account, amount, timestamp) VALUES (?, ?, ?, ?)",
		from, to, amount, timestamp,
	)
	return err
}

// LoadTransactions carga todas las transacciones desde la base de datos.
func (d *Database) LoadTransactions() ([]map[string]interface{}, error) {
	rows, err := d.Connection.Query("SELECT from_account, to_account, amount, timestamp FROM transactions ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []map[string]interface{}
	for rows.Next() {
		var from, to, timestamp string
		var amount int64
		if err := rows.Scan(&from, &to, &amount, &timestamp); err != nil {
			return nil, err
		}
		transactions = append(transactions, map[string]interface{}{
			"from":      from,
			"to":        to,
			"amount":    amount,
			"timestamp": timestamp,
		})
	}

	return transactions, nil
}

// AccountExists verifica si una cuenta existe en la base de datos.
func (d *Database) AccountExists(account string) (bool, error) {
	var exists bool
	err := d.Connection.QueryRow("SELECT EXISTS(SELECT 1 FROM balances WHERE account = ?)", account).Scan(&exists)
	return exists, err
}

// GetAllAccounts obtiene todas las cuentas (direcciones) de la base de datos.
func (d *Database) GetAllAccounts() ([]string, error) {
	rows, err := d.Connection.Query("SELECT account FROM balances")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []string
	for rows.Next() {
		var account string
		if err := rows.Scan(&account); err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

// Transferir dinero de una cuenta a otra, actualizando saldos y guardando la transacción.
func (d *Database) Transfer(from, to string, amount int64, timestamp string) error {
	// Verificar que ambas cuentas existen
	fromExists, err := d.AccountExists(from)
	if err != nil || !fromExists {
		return fmt.Errorf("la cuenta de origen no existe")
	}
	toExists, err := d.AccountExists(to)
	if err != nil || !toExists {
		return fmt.Errorf("la cuenta de destino no existe")
	}

	// Obtener los saldos actuales de ambas cuentas
	fromBalance, err := d.GetBalance(from)
	if err != nil {
		return fmt.Errorf("error al obtener saldo de la cuenta de origen: %w", err)
	}
	toBalance, err := d.GetBalance(to)
	if err != nil {
		return fmt.Errorf("error al obtener saldo de la cuenta de destino: %w", err)
	}

	// Verificar si la cuenta de origen tiene suficientes fondos
	if fromBalance < amount {
		return fmt.Errorf("saldo insuficiente en la cuenta de origen")
	}

	// Actualizar el saldo de ambas cuentas
	fromNewBalance := fromBalance - amount
	toNewBalance := toBalance + amount

	// Guardar los nuevos saldos en la base de datos
	if err := d.SaveBalance(from, fromNewBalance); err != nil {
		return fmt.Errorf("error al guardar saldo en la cuenta de origen: %w", err)
	}
	if err := d.SaveBalance(to, toNewBalance); err != nil {
		return fmt.Errorf("error al guardar saldo en la cuenta de destino: %w", err)
	}

	// Guardar la transacción en la base de datos
	if err := d.SaveTransaction(from, to, amount, timestamp); err != nil {
		return fmt.Errorf("error al guardar transacción: %w", err)
	}

	return nil
}
