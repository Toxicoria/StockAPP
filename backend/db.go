package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func IniciarDB() {
	// Leemos las credenciales desde variables de entorno.
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "admin_dev")
	password := getEnv("DB_PASSWORD", "password_dev")
	dbname := getEnv("DB_NAME", "stock_db")

	// El "connection string" es la URL que describe cómo conectarse.
	// Formato estándar de PostgreSQL.
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		LogError("%s[ERROR]%s Error al configurar la conexión a la DB: %v", colorRed, colorReset, err)
		os.Exit(1)
	}

	if err = DB.Ping(); err != nil {
		LogError("%s[ERROR]%s No se pudo conectar a PostgreSQL: %v", colorRed, colorReset, err)
		os.Exit(1)
	}

	LogOK("%s[OK]%s Conexión a PostgreSQL establecida.\n", colorGreen, colorReset)
}

// getEnv lee una variable de entorno y devuelve un valor por defecto si no existe.
// Esto nos permite correr el backend tanto en Docker (con variables seteadas)
// como localmente (con los valores default para desarrollo).
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
