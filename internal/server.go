package internal

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Server struct {
	DB          *Database
	Blockchain  *Blockchain
	TokenSupply *TokenSupply
}

// NewServer inicializa un servidor con la base de datos, blockchain y token supply.
func NewServer(db *Database, bc *Blockchain, ts *TokenSupply) *Server {
	return &Server{
		DB:          db,
		Blockchain:  bc,
		TokenSupply: ts,
	}
}

// Start inicia el servidor HTTP y lanza el proceso de minería.
func (s *Server) Start(port int) {
	router := mux.NewRouter()

	// Rutas de la API
	router.HandleFunc("/blocks", s.GetBlocks).Methods("GET")
	router.HandleFunc("/balances/{account}", s.GetBalance).Methods("GET")
	router.HandleFunc("/transactions", s.GetTransactions).Methods("GET")
	router.HandleFunc("/generate-address", s.GenerateAddress).Methods("GET")
	router.HandleFunc("/transactions", s.AddTransaction).Methods("POST")
	router.HandleFunc("/wasm-contracts", s.AddWASMContract).Methods("POST")
	router.HandleFunc("/execute-wasm", s.ExecuteWASMContract).Methods("POST")
	router.HandleFunc("/wasm-contracts", s.AddWASMContract).Methods("POST")

	// Servir el archivo swagger.json
	router.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "/home/tuxz/blockchain-go/docs/swagger.json")
	}).Methods("GET")

	// Servir Swagger UI
	swaggerUIDir := http.Dir("/home/tuxz/blockchain-go/swagger-ui")
	router.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", http.FileServer(swaggerUIDir)))

	// Lanzar el proceso de minería en un goroutine
	go s.StartMining()

	// Agregar código de depuración para imprimir las rutas registradas
	fmt.Println("Rutas registradas:")
	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, err := route.GetPathTemplate()
		if err != nil {
			path = "no disponible"
		}
		methods, err := route.GetMethods()
		if err != nil {
			methods = []string{"no disponible"}
		}
		fmt.Printf("Ruta: %s, Métodos: %v\n", path, methods)
		return nil
	})

	fmt.Printf("Servidor escuchando en el puerto %d\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), router); err != nil {
		fmt.Printf("Error iniciando el servidor: %s\n", err)
	}

}

// StartMining procesa transacciones pendientes y genera bloques periódicamente.
func (s *Server) StartMining() {
	for {
		time.Sleep(10 * time.Second) // Intervalo de minería
		pendingTxs, err := s.DB.GetPendingTransactions()
		if err != nil {
			fmt.Printf("Error al obtener transacciones pendientes: %s\n", err)
			continue
		}
		if len(pendingTxs) == 0 {
			fmt.Println("No hay transacciones pendientes para minar.")
			continue
		}

		// Crear un nuevo bloque con las transacciones pendientes
		newBlock := s.Blockchain.AddBlock(pendingTxs, time.Now().UTC().Format(time.RFC3339))
		err = s.DB.SaveBlock(*newBlock)
		if err != nil {
			fmt.Printf("Error al guardar el bloque: %s\n", err)
			continue
		}

		// Actualizar saldos en la base de datos
		for _, tx := range pendingTxs {
			err = s.DB.UpdateBalances(tx.From, tx.To, tx.Amount)
			if err != nil {
				fmt.Printf("Error al actualizar saldos: %s\n", err)
			}
		}

		// Limpiar las transacciones pendientes
		err = s.DB.ClearPendingTransactions()
		if err != nil {
			fmt.Printf("Error al limpiar transacciones pendientes: %s\n", err)
		}

		fmt.Printf("Bloque minado: #%d con %d transacciones\n", newBlock.Index, len(newBlock.Transactions))
	}
}

// GenerateAddress genera una nueva dirección y devuelve la clave privada asociada.
func (s *Server) GenerateAddress(w http.ResponseWriter, r *http.Request) {
	address, privKey, err := GetAddress(s.DB, 1000) // Saldo inicial configurable
	if err != nil {
		http.Error(w, "Error generando dirección", http.StatusInternalServerError)
		return
	}

	privKeyHex := hex.EncodeToString(privKey.D.Bytes())
	response := map[string]string{
		"address":     address,
		"private_key": privKeyHex,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetBlocks maneja la solicitud para obtener todos los bloques.
func (s *Server) GetBlocks(w http.ResponseWriter, r *http.Request) {
	blocks, err := s.DB.LoadBlocks()
	if err != nil {
		http.Error(w, "Error cargando bloques", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(blocks)
}

// GetBalance maneja la solicitud para obtener el saldo de una cuenta.
func (s *Server) GetBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	account := vars["account"]

	balance, err := s.DB.GetBalance(account)
	if err != nil {
		http.Error(w, "Error obteniendo el saldo", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"account": account,
		"balance": balance,
	})
}

// GetTransactions maneja una solicitud para obtener todas las transacciones.
func (s *Server) GetTransactions(w http.ResponseWriter, r *http.Request) {
	transactions, err := s.DB.LoadTransactions()
	if err != nil {
		http.Error(w, "Error cargando transacciones", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

// AddTransaction maneja una solicitud para registrar una nueva transacción.
func (s *Server) AddTransaction(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		From   string `json:"from"`
		To     string `json:"to"`
		Amount int64  `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}

	// Validar que las cuentas existan
	fromExists, err := s.DB.AccountExists(payload.From)
	if err != nil || !fromExists {
		http.Error(w, "La cuenta origen no existe", http.StatusBadRequest)
		return
	}

	toExists, err := s.DB.AccountExists(payload.To)
	if err != nil || !toExists {
		http.Error(w, "La cuenta destino no existe", http.StatusBadRequest)
		return
	}

	err = s.DB.AddPendingTransaction(payload.From, payload.To, payload.Amount)
	if err != nil {
		http.Error(w, "Error añadiendo transacción pendiente", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Transacción añadida a la cola",
	})
}

// AddWASMContract maneja la solicitud para registrar un nuevo contrato WASM.
func (s *Server) AddWASMContract(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		ID       string `json:"id"`
		Owner    string `json:"owner"`
		WASMCode []byte `json:"wasm_code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}

	wasmContract := WASMContract{
		ID:       payload.ID,
		Owner:    payload.Owner,
		WASMCode: payload.WASMCode,
	}

	err := s.DB.SaveWASMContract(wasmContract)
	if err != nil {
		http.Error(w, "Error guardando contrato WASM", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Contrato WASM registrado exitosamente",
	})
}

// ExecuteWASMContract maneja la ejecución de un contrato WASM.
func (s *Server) ExecuteWASMContract(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		ID    string `json:"id"`
		Input []byte `json:"input"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}

	contract, err := s.DB.LoadWASMContract(payload.ID)
	if err != nil || contract == nil {
		http.Error(w, "Contrato WASM no encontrado", http.StatusNotFound)
		return
	}

	result, err := contract.Execute(payload.Input)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error ejecutando contrato WASM: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"result": result,
	})
}
