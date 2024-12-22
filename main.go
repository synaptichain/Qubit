package main

import (
	"blockchain-go/internal"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// Función para manejar la petición GET /generate-address
func GenerateAddressHandler(w http.ResponseWriter, r *http.Request, db *internal.Database) {
	const initialBalance = 1000 // Define el saldo inicial predeterminado
	address, privateKey, err := internal.GetAddress(db, initialBalance)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error al generar la dirección: %s", err), http.StatusInternalServerError)
		return
	}

	privateKeyHex := hex.EncodeToString(privateKey.D.Bytes())

	response := map[string]string{
		"address":    address,
		"privateKey": privateKeyHex,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Función para manejar la petición GET /blocks
func GetBlocksHandler(w http.ResponseWriter, r *http.Request, db *internal.Database) {
	blocks, err := db.LoadBlocks()
	if err != nil {
		http.Error(w, "Error al cargar los bloques", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(blocks)
}

// Función para manejar la petición GET /transactions
func GetTransactionsHandler(w http.ResponseWriter, r *http.Request, db *internal.Database) {
	transactions, err := db.LoadTransactions()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error al cargar transacciones: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

// Función para manejar la petición GET /stats
func GetStatsHandler(w http.ResponseWriter, r *http.Request, db *internal.Database) {
	blocks, err := db.LoadBlocks()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error al cargar bloques: %s", err), http.StatusInternalServerError)
		return
	}

	transactions, err := db.LoadTransactions()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error al cargar transacciones: %s", err), http.StatusInternalServerError)
		return
	}

	accounts, err := db.GetAllAccounts()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error al cargar cuentas: %s", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"total_blocks":       len(blocks),
		"total_transactions": len(transactions),
		"total_accounts":     len(accounts),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Función para manejar la petición POST /transfer
func TransferHandler(w http.ResponseWriter, r *http.Request, db *internal.Database) {
	var transferRequest struct {
		From   string `json:"from"`
		To     string `json:"to"`
		Amount int64  `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&transferRequest); err != nil {
		http.Error(w, fmt.Sprintf("Error al decodificar la solicitud: %s", err), http.StatusBadRequest)
		return
	}

	err := db.AddPendingTransaction(transferRequest.From, transferRequest.To, transferRequest.Amount)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error al agregar la transacción: %s", err), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"status": "Transacción añadida a la cola",
		"from":   transferRequest.From,
		"to":     transferRequest.To,
		"amount": fmt.Sprintf("%d", transferRequest.Amount),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Función para manejar la petición GET /balance
func GetBalanceHandler(w http.ResponseWriter, r *http.Request, db *internal.Database) {
	account := r.URL.Query().Get("account")
	if account == "" {
		http.Error(w, "Cuenta no especificada", http.StatusBadRequest)
		return
	}

	balance, err := db.GetBalance(account)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error al obtener el saldo: %s", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"account": account,
		"balance": balance,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Ciclo de minado automático
func startMiningLoop(db *internal.Database, bc *internal.Blockchain) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		pendingTransactions, err := db.GetPendingTransactions()
		if err != nil {
			fmt.Printf("Error al obtener transacciones pendientes: %s\n", err)
			continue
		}

		if len(pendingTransactions) == 0 {
			fmt.Println("No hay transacciones pendientes para minar.")
			continue
		}

		newBlock := bc.AddBlock(pendingTransactions, time.Now().String())
		fmt.Printf("Bloque %d minado con %d transacciones.\n", newBlock.Index, len(pendingTransactions))

		if err := db.SaveBlock(*newBlock); err != nil {
			fmt.Printf("Error al guardar el bloque: %s\n", err)
			continue
		}

		for _, tx := range pendingTransactions {
			err = db.UpdateBalances(tx.From, tx.To, tx.Amount)
			if err != nil {
				fmt.Printf("Error al actualizar saldo: %s\n", err)
			}
		}

		if err := db.ClearPendingTransactions(); err != nil {
			fmt.Printf("Error al limpiar transacciones pendientes: %s\n", err)
		}
	}
}

func main() {
	db, err := internal.InitDB("postgres://postgres:170280@localhost:5432/blockchain_db")
	if err != nil {
		log.Fatalf("Error inicializando la base de datos: %s\n", err)
	}
	defer db.Connection.Close()

	bc := internal.NewBlockchain("Genesis Hash or Data")

	blocks, err := db.LoadBlocks()
	if err != nil {
		log.Fatalf("Error cargando bloques: %s\n", err)
	}

	if len(blocks) == 0 {
		fmt.Println("No se encontraron bloques, creando bloque génesis...")
		bc.AddBlock([]internal.Transaction{}, "Genesis Hash")
		err := db.SaveBlock(*bc.Blocks[0])
		if err != nil {
			log.Fatalf("Error guardando el bloque génesis: %s\n", err)
		}
		fmt.Println("Bloque génesis creado y guardado.")
	} else {
		for _, block := range blocks {
			bc.Blocks = append(bc.Blocks, &block)
		}
	}

	go startMiningLoop(db, bc)

	router := mux.NewRouter()
	router.HandleFunc("/generate-address", func(w http.ResponseWriter, r *http.Request) {
		GenerateAddressHandler(w, r, db)
	}).Methods("GET")
	router.HandleFunc("/blocks", func(w http.ResponseWriter, r *http.Request) {
		GetBlocksHandler(w, r, db)
	}).Methods("GET")
	router.HandleFunc("/transactions", func(w http.ResponseWriter, r *http.Request) {
		GetTransactionsHandler(w, r, db)
	}).Methods("GET")
	router.HandleFunc("/transfer", func(w http.ResponseWriter, r *http.Request) {
		TransferHandler(w, r, db)
	}).Methods("POST")
	router.HandleFunc("/balance", func(w http.ResponseWriter, r *http.Request) {
		GetBalanceHandler(w, r, db)
	}).Methods("GET")
	router.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		GetStatsHandler(w, r, db)
	}).Methods("GET")

	// Rutas relacionadas con Swagger
	router.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "/home/tuxz/blockchain-go/docs/swagger.json")
	}).Methods("GET")
	swaggerUIDir := http.Dir("/home/tuxz/blockchain-go/swagger-ui")
	router.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", http.FileServer(swaggerUIDir)))

	// Listar todas las rutas registradas
	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, _ := route.GetPathTemplate()
		methods, _ := route.GetMethods()
		fmt.Printf("Ruta registrada: %s, Métodos: %v\n", path, methods)
		return nil
	})

	fmt.Println("Servidor escuchando en el puerto 8080...")
	err = http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatalf("Error iniciando el servidor: %s\n", err)
	}
}
