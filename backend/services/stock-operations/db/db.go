package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"

	"stock-operations/pkg/logger"
)

var DB *sql.DB

// Iniciar abre y verifica la conexión a PostgreSQL. Si la base no responde,
// el programa termina acá con un error claro.
func Iniciar() {
	// Leemos las credenciales desde variables de entorno.
	host := Env("DB_HOST", "localhost")
	port := Env("DB_PORT", "5432")
	user := Env("DB_USER", "admin_dev")
	password := Env("DB_PASSWORD", "password_dev")
	dbname := Env("DB_NAME", "stock_db")

	// El "connection string" es la URL que describe cómo conectarse.
	// Formato estándar de PostgreSQL.
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		logger.Error("Error al configurar la conexión a la DB: %v", err)
		os.Exit(1)
	}

	if err = DB.Ping(); err != nil {
		logger.Error("No se pudo conectar a PostgreSQL: %v", err)
		os.Exit(1)
	}

	logger.OK("Conexión a PostgreSQL establecida.")
}

// Env lee una variable de entorno y devuelve un valor por defecto si no
// existe. Esto permite correr el backend tanto en Docker (con variables
// seteadas) como localmente (con los valores default para desarrollo).
func Env(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
