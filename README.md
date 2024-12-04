# Blockchain Go

Blockchain Go es una implementación básica de una blockchain utilizando Go. El proyecto permite la creación de bloques, la gestión de cuentas y saldos, transferencias de tokens y la ejecución de contratos inteligentes.

## Tecnologías Utilizadas

- **Lenguaje**: Go
- **Base de Datos**: SQLite (para almacenar bloques, saldos y transacciones)
- **Bibliotecas**:
  - `goja`: Para ejecutar contratos inteligentes en JavaScript.
  - `github.com/rs/cors`: Para habilitar CORS y permitir acceso desde diferentes dominios.
  - `swagger-ui`: Para la documentación interactiva de la API.

## Funcionalidades

### API REST

La API REST permite interactuar con la blockchain a través de las siguientes rutas:

- **GET `/generate-address`**: Genera una nueva dirección y devuelve la clave privada asociada.
- **GET `/all-accounts`**: Obtiene todas las cuentas almacenadas en la base de datos junto con sus balances.
- **GET `/blocks`**: Obtiene todos los bloques almacenados en la blockchain.
- **GET `/balance?account={address}`**: Obtiene el saldo de una cuenta específica.
- **POST `/transfer`**: Realiza una transferencia de tokens entre dos cuentas.
- **POST `/contracts`**: Registra un contrato inteligente en la blockchain.
- **POST `/contracts/{id}/execute`**: Ejecuta un contrato inteligente dado su ID.

### Ejemplo de Uso

1. **Generar una nueva dirección**:
   ```bash
   curl -X GET http://localhost:8080/generate-address


2. **Consultar el saldo de una cuenta**:


curl -X GET "http://localhost:8080/balance?account=cuenta1"

3. **Realizar una transferencia**:

curl -X POST http://localhost:8080/transfer -d '{"from": "cuenta1", "to": "cuenta2", "amount": 500}'

Swagger UI
Documentación interactiva: La API está documentada usando Swagger UI.
Acceso: Una vez el servidor esté en funcionamiento, la documentación de la API puede ser accedida desde http://localhost:8080/swagger-ui/.
Instalación
Requisitos
Go (v1.18 o superior)
SQLite (automáticamente gestionado por el proyecto)
Pasos de Instalación


1. **Clonar el repositorio**:
git clone https://github.com/tu_usuario/blockchain-go.git
cd blockchain-go

2. **Instalar dependencias**: Si no tienes Go Modules activado, asegúrate de inicializarlo:

go mod init
go mod tidy

3. **Ejecutar el servidor**:

go run main.go

4. **Acceder a la API**:
 
 El servidor escuchará en http://localhost:8080. Puedes usar la interfaz de Swagger en http://localhost:8080/swagger-ui/ para interactuar con la API.

Contribuciones
Las contribuciones son bienvenidas. Si tienes alguna idea o mejora, por favor abre un Issue o envía un Pull Request.


License
Este proyecto está licenciado bajo la MIT License - ver el archivo LICENSE para más detalles.


### Qué incluye este README:
- **Descripción del proyecto**: Qué es lo que hace tu proyecto.
- **Tecnologías utilizadas**: Herramientas y bibliotecas empleadas.
- **Funcionamiento de la API**: Las rutas principales de la API y ejemplos de uso.
- **Instalación**: Instrucciones sobre cómo poner en marcha el proyecto localmente.
- **Contribuciones**: Cómo otros pueden contribuir a tu proyecto.
- **Licencia**: La licencia bajo la cual se distribuye el código (en este caso, MIT).

Si necesitas agregar más detalles o modificar algo, avísame y lo ajustamos.


 
