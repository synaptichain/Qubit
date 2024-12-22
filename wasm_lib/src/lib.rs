use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;
use serde_wasm_bindgen::{from_value, to_value};

#[derive(Serialize, Deserialize)]
pub struct InputData {
    pub key: String,
    pub value: String,
}

#[derive(Serialize, Deserialize)]
pub struct OutputData {
    pub success: bool,
    pub message: String,
}

#[wasm_bindgen]
pub fn process_data(input: JsValue) -> Result<JsValue, JsValue> {
    // Usa `serde-wasm-bindgen` para deserializar
    let input_data: InputData = from_value(input).map_err(|err| JsValue::from_str(&err.to_string()))?;

    // Procesa los datos
    let output_data = OutputData {
        success: true,
        message: format!("Key: {}, Value: {}", input_data.key, input_data.value),
    };

    // Usa `serde-wasm-bindgen` para serializar
    to_value(&output_data).map_err(|err| JsValue::from_str(&err.to_string()))
}

