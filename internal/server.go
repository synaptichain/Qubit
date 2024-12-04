package internal

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
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

// Start inicia el servidor HTTP.
func (s *Server) Start(port int) {
	router := mux.NewRouter()

	// Rutas de la API
	router.HandleFunc("/blocks", s.GetBlocks).Methods("GET")
	router.HandleFunc("/balances/{account}", s.GetBalance).Methods("GET") // Ruta para obtener el saldo de una cuenta
	router.HandleFunc("/transactions", s.GetTransactions).Methods("GET")
	router.HandleFunc("/generate-address", s.GenerateAddress).Methods("GET") // Ruta para generar dirección
	router.HandleFunc("/transactions", s.AddTransaction).Methods("POST")
	router.HandleFunc("/transactions/validate", s.ValidateTransaction).Methods("POST")
	router.HandleFunc("/contracts", s.AddContract).Methods("POST")
	router.HandleFunc("/contracts/{id}/execute", s.ExecuteContract).Methods("POST")
	router.HandleFunc("/tokens/create", s.CreateToken).Methods("POST") // Nueva ruta para crear tokens

	// Servir el archivo swagger.json
	router.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		currentDir, _ := os.Getwd()
		http.ServeFile(w, r, currentDir+"/docs/swagger.json")
	}).Methods("GET")

	// Servir Swagger UI
	swaggerUIDir := http.Dir("./swagger-ui")
	router.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", http.FileServer(swaggerUIDir)))

	fmt.Printf("Servidor escuchando en el puerto %d\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), router)
}

// GenerateAddress genera una nueva dirección y devuelve la clave privada asociada.
func (s *Server) GenerateAddress(w http.ResponseWriter, r *http.Request) {
	address, privKey, err := GetAddress(s.DB) // Pasamos la base de datos aquí
	if err != nil {
		http.Error(w, "Error generando dirección", http.StatusInternalServerError)
		return
	}

	// Serializar la clave privada en hexadecimal
	privKeyHex := hex.EncodeToString(privKey.D.Bytes())

	// Responder con la dirección y la clave privada
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

	json.NewEncoder(w).Encode(blocks)
}

// GetBalance maneja la solicitud para obtener el saldo de una cuenta.
func (s *Server) GetBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	account := vars["account"]

	// Obtener el balance de la base de datos o de alguna estructura que maneje los saldos
	balance, err := s.DB.GetBalance(account)
	if err != nil {
		http.Error(w, "Error obteniendo el saldo", http.StatusInternalServerError)
		return
	}

	// Enviar el balance como respuesta JSON
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

	// Validar cuentas
	fromExists, _ := s.DB.AccountExists(payload.From)
	toExists, _ := s.DB.AccountExists(payload.To)

	if !fromExists {
		http.Error(w, "La cuenta de origen no existe", http.StatusBadRequest)
		return
	}
	if !toExists {
		http.Error(w, "La cuenta de destino no existe", http.StatusBadRequest)
		return
	}

	// Validar saldo suficiente
	fromBalance, _ := s.DB.GetBalance(payload.From)
	if fromBalance < payload.Amount {
		http.Error(w, "Saldo insuficiente en la cuenta de origen", http.StatusBadRequest)
		return
	}

	// Registrar la transacción
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	err := s.DB.SaveTransaction(payload.From, payload.To, payload.Amount, timestamp)
	if err != nil {
		http.Error(w, "Error registrando la transacción", http.StatusInternalServerError)
		return
	}

	// Actualizar balances
	s.TokenSupply.Transfer(payload.From, payload.To, payload.Amount)
	s.DB.SaveBalance(payload.From, s.TokenSupply.GetBalance(payload.From))
	s.DB.SaveBalance(payload.To, s.TokenSupply.GetBalance(payload.To))

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Transacción registrada exitosamente",
	})
}

// ValidateTransaction valida una transacción sin ejecutarla.
func (s *Server) ValidateTransaction(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		From   string `json:"from"`
		To     string `json:"to"`
		Amount int64  `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}

	// Validar cuentas
	fromExists, _ := s.DB.AccountExists(payload.From)
	toExists, _ := s.DB.AccountExists(payload.To)

	if !fromExists {
		http.Error(w, "La cuenta de origen no existe", http.StatusBadRequest)
		return
	}
	if !toExists {
		http.Error(w, "La cuenta de destino no existe", http.StatusBadRequest)
		return
	}

	// Validar saldo suficiente
	fromBalance, _ := s.DB.GetBalance(payload.From)
	if fromBalance < payload.Amount {
		http.Error(w, "Saldo insuficiente en la cuenta de origen", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "La transacción es válida",
	})
}

// AddContract maneja la solicitud para registrar un nuevo contrato.
func (s *Server) AddContract(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		ID     string                 `json:"id"`
		Owner  string                 `json:"owner"`
		Code   string                 `json:"code"`
		Params map[string]interface{} `json:"params"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}

	// Crear el contrato
	contract := NewContract(payload.ID, payload.Owner, payload.Code, payload.Params)

	// Asociar el contrato al último bloque
	lastBlock := &s.Blockchain.Blocks[len(s.Blockchain.Blocks)-1]
	lastBlock.AddContract(*contract)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Contrato registrado exitosamente",
	})
}

// ExecuteContract maneja la solicitud para ejecutar un contrato.
func (s *Server) ExecuteContract(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contractID := vars["id"]

	// Buscar el contrato
	var contract *Contract
	for _, block := range s.Blockchain.Blocks {
		for _, c := range block.Contracts {
			if c.ID == contractID {
				contract = &c
				break
			}
		}
	}

	if contract == nil {
		http.Error(w, "Contrato no encontrado", http.StatusNotFound)
		return
	}

	// Leer la entrada desde la solicitud
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Entrada inválida para el contrato", http.StatusBadRequest)
		return
	}

	// Crear el entorno para ejecutar el contrato
	env := map[string]interface{}{
		"transfer": func(from, to string, amount int64) {
			fmt.Printf("Ejecutando transferencia de %d tokens de %s a %s\n", amount, from, to)
			if err := s.TokenSupply.Transfer(from, to, amount); err != nil {
				panic(err)
			}
		},
		"mint": func(to string, amount int64) {
			fmt.Printf("Acuñando %d tokens para %s\n", amount, to)
			s.TokenSupply.Mint(to, amount)
		},
	}

	// Depurar el código registrado
	fmt.Println("Código del contrato registrado:", contract.Compiled)

	// Ejecutar el contrato
	result, err := contract.Execute(input, env)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error ejecutando el contrato: %s", err.Error()), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": result,
	})
}

// CreateToken permite crear nuevos tokens.
func (s *Server) CreateToken(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Name   string `json:"name"`
		Symbol string `json:"symbol"`
		Amount int64  `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}

	// Crear tokens
	s.TokenSupply.Mint(payload.Name, payload.Amount)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Se crearon %d tokens para el proyecto %s", payload.Amount, payload.Name),
	})
}
