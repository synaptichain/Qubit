# Qubit: A Blockchain Implementation

**Qubit** is a modular blockchain project designed for experimentation and learning, supporting smart contracts and WASM integration. It uses PostgreSQL for persistence and offers a RESTful API for interaction.

## Features
- **Blockchain Core**: Supports transaction processing, block creation, and mining.
- **Consensus Mechanism**: Work-in-progress (PoST or Tendermint-based).
- **Smart Contracts**:
  - **WASM Integration**: Upload and execute WebAssembly contracts.
  - **Secure Execution**: Contract isolation using Wasmer.
- **Persistence**: Data storage using PostgreSQL.
- **RESTful API**:
  - Manage accounts, balances, and transactions.
  - Upload and query smart contracts.
  - Access blockchain statistics.
- **Swagger Documentation**: Comprehensive API documentation.

## Project Structure
```
├── internal
│   ├── blockchain.go       # Blockchain core functionality
│   ├── block.go            # Block structure and utilities
│   ├── wasm_executor.go    # WASM contract execution
│   ├── db.go               # PostgreSQL database integration
│   ├── server.go           # HTTP server and API handlers
├── wasm_lib
│   ├── src
│   │   ├── lib.rs          # Rust library for WASM smart contracts
│   └── Cargo.toml          # Rust project configuration
├── docs
│   └── swagger.json        # Swagger API documentation
├── README.md               # Project documentation
└── main.go                 # Application entry point
```

## Setup

### Prerequisites
- **Go** (1.20+)
- **Rust** (with `wasm32-unknown-unknown` target)
- **PostgreSQL**

### Installation
1. **Clone the repository**:
   ```bash
   git clone https://github.com/synaptichain/Qubit.git
   cd Qubit
   ```

2. **Set up PostgreSQL**:
   - Create a database:
     ```sql
     CREATE DATABASE blockchain_db;
     ```
   - Update the connection string in `internal/db.go`:
     ```go
     "postgres://<username>:<password>@localhost:5432/blockchain_db"
     ```

3. **Compile the WASM library**:
   ```bash
   cd wasm_lib
   cargo build --release --target wasm32-unknown-unknown
   ```

4. **Run the application**:
   ```bash
   go run main.go
   ```

## API Endpoints
- **GET** `/blocks` - Retrieve all blocks.
- **GET** `/balances/{account}` - Retrieve account balance.
- **POST** `/transactions` - Add a new transaction.
- **GET** `/generate-address` - Generate a new address.
- **POST** `/wasm-contracts` - Upload a WASM smart contract.

## Smart Contracts
Contracts are written in Rust and compiled to WASM. Use the `serde-wasm-bindgen` library for data serialization.

Example WASM smart contract:
```rust
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

#[wasm_bindgen]
pub fn execute(input: JsValue) -> JsValue {
    let input_data: InputData = serde_wasm_bindgen::from_value(input).unwrap();
    let output_data = OutputData {
        result: input_data.value * 2,
    };
    serde_wasm_bindgen::to_value(&output_data).unwrap()
}

#[derive(Serialize, Deserialize)]
struct InputData {
    value: i32,
}

#[derive(Serialize, Deserialize)]
struct OutputData {
    result: i32,
}
```

## Roadmap
- Implement Tendermint for consensus.
- Create a GUI-based contract management tool.
- Enhance performance with indexing and caching.

## Contributing
Feel free to fork the repository, open issues, or submit pull requests.

## License
[MIT License](LICENSE)

 
