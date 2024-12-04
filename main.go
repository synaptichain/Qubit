package main

import (
	"blockchain-go/internal"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	// Importar el paquete CORS
)

// Función para manejar la petición GET /generate-address
func GenerateAddressHandler(w http.ResponseWriter, r *http.Request, db *internal.Database) {
	// Llamamos a la función GetAddress del paquete internal
	address, privateKey, err := internal.GetAddress(db)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error al generar la dirección: %s", err), http.StatusInternalServerError)
		return
	}

	// Enviar la respuesta como JSON
	response := map[string]string{
		"address":    address,
		"privateKey": hex.EncodeToString(privateKey.D.Bytes()), // Convertir la clave privada a hexadecimal
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Función para manejar la petición GET /all-accounts
func GetAllAccountsHandler(w http.ResponseWriter, r *http.Request, db *internal.Database) {
	// Obtener todas las cuentas de la base de datos
	accounts, err := db.GetAllAccounts()
	if err != nil {
		http.Error(w, "Error al obtener las cuentas", http.StatusInternalServerError)
		return
	}

	// Crear una lista para las cuentas y sus balances
	var accountsWithBalance []map[string]interface{}

	// Iterar sobre las cuentas y obtener su balance
	for _, account := range accounts {
		balance, err := db.GetBalance(account) // Obtener el balance de cada cuenta
		if err != nil {
			http.Error(w, fmt.Sprintf("Error obteniendo balance para %s: %s", account, err), http.StatusInternalServerError)
			return
		}

		// Añadir la cuenta y su balance a la lista
		accountsWithBalance = append(accountsWithBalance, map[string]interface{}{
			"address": account,
			"balance": balance,
		})
	}

	// Enviar la respuesta como JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accountsWithBalance)
}

// Función para manejar la petición GET /blocks
func GetBlocksHandler(w http.ResponseWriter, r *http.Request, db *internal.Database) {
	// Cargar todos los bloques desde la base de datos
	blocks, err := db.LoadBlocks()
	if err != nil {
		http.Error(w, "Error al cargar los bloques", http.StatusInternalServerError)
		return
	}

	// Enviar la respuesta como JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(blocks)
}

// Función para manejar la petición GET /balance
func GetBalanceHandler(w http.ResponseWriter, r *http.Request, db *internal.Database) {
	account := r.URL.Query().Get("account")
	if account == "" {
		http.Error(w, "Cuenta no especificada", http.StatusBadRequest)
		return
	}

	// Obtener el saldo de la cuenta
	balance, err := db.GetBalance(account)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error al obtener el saldo: %s", err), http.StatusInternalServerError)
		return
	}

	// Enviar la respuesta como JSON
	response := map[string]interface{}{
		"account": account,
		"balance": balance,
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

	// Decodificar el cuerpo de la solicitud
	if err := json.NewDecoder(r.Body).Decode(&transferRequest); err != nil {
		http.Error(w, fmt.Sprintf("Error al decodificar la solicitud: %s", err), http.StatusBadRequest)
		return
	}

	// Transferir el dinero
	timestamp := fmt.Sprintf("%d", (r.ContentLength / 1000)) // Timestamp aproximado para la transferencia
	err := db.Transfer(transferRequest.From, transferRequest.To, transferRequest.Amount, timestamp)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error en la transferencia: %s", err), http.StatusInternalServerError)
		return
	}

	// Responder que la transferencia fue exitosa
	response := map[string]string{
		"status": "transferencia exitosa",
		"from":   transferRequest.From,
		"to":     transferRequest.To,
		"amount": fmt.Sprintf("%d", transferRequest.Amount),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Inicializar la base de datos
	db, err := internal.InitDB("blockchain.db")
	if err != nil {
		log.Fatalf("Error inicializando la base de datos: %s\n", err)
	}
	defer db.Connection.Close()

	// Crear una nueva blockchain
	bc := internal.NewBlockchain()

	// Cargar bloques existentes desde la base de datos
	blocks, err := db.LoadBlocks()
	if err != nil {
		log.Fatalf("Error cargando bloques: %s\n", err)
	}

	// Si no hay bloques, crear el bloque génesis
	if len(blocks) == 0 {
		fmt.Println("No se encontraron bloques, creando bloque génesis...")
		bc.AddBlock("Bloque génesis")

		// Guardar el bloque génesis
		genesisBlock := bc.Blocks[0]
		err := db.SaveBlock(genesisBlock)
		if err != nil {
			log.Fatalf("Error guardando el bloque génesis: %s\n", err)
		}
		fmt.Println("Bloque génesis creado y guardado.")
	} else {
		// Si ya existen bloques, cargarlos en la blockchain
		bc.Blocks = blocks
	}

	// Agregar un nuevo bloque
	bc.AddBlock("Nuevo bloque para probar BLAKE2")

	// Obtener el último bloque y guardarlo en la base de datos
	newBlock := bc.Blocks[len(bc.Blocks)-1]
	err = db.SaveBlock(newBlock)
	if err != nil {
		log.Fatalf("Error guardando el nuevo bloque: %s\n", err)
	}
	fmt.Println("Nuevo bloque agregado y guardado.")

	// Manejar balances
	tokenSupply := internal.NewTokenSupply(100_000_000, "cuenta1")
	err = db.SaveBalance("cuenta1", tokenSupply.GetBalance("cuenta1"))
	if err != nil {
		log.Fatalf("Error guardando el balance inicial: %s\n", err)
	}

	// Servir swagger.json desde la carpeta docs
	http.Handle("/swagger.json", http.StripPrefix("/", http.FileServer(http.Dir("./docs"))))

	// Servir los archivos estáticos de Swagger UI desde la carpeta swagger-ui
	http.Handle("/swagger-ui/", http.StripPrefix("/swagger-ui/", http.FileServer(http.Dir("./swagger-ui"))))

	// Registrar las rutas para la blockchain
	http.HandleFunc("/generate-address", func(w http.ResponseWriter, r *http.Request) {
		GenerateAddressHandler(w, r, db)
	})
	http.HandleFunc("/all-accounts", func(w http.ResponseWriter, r *http.Request) {
		GetAllAccountsHandler(w, r, db)
	})
	http.HandleFunc("/blocks", func(w http.ResponseWriter, r *http.Request) {
		GetBlocksHandler(w, r, db)
	})
	http.HandleFunc("/balance", func(w http.ResponseWriter, r *http.Request) {
		GetBalanceHandler(w, r, db)
	})
	http.HandleFunc("/transfer", func(w http.ResponseWriter, r *http.Request) {
		TransferHandler(w, r, db)
	})

	// Iniciar el servidor estándar de Go en el puerto 8080
	fmt.Println("Servidor escuchando en el puerto 8080...")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Error iniciando el servidor: %s\n", err)
	}
}
