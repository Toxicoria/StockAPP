package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"

	"stock-operations/pkg/logger"
)

var DB *sql.DB

// Iniciar abre y verifica la conexión a PostgreSQL con reintentos automáticos.
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

	for intento := 1; intento <= 15; intento++ {
		err = DB.Ping()
		if err == nil {
			logger.OK("Conexión a PostgreSQL establecida.")
			// Asegurar esquema base de dispositivos y claves si no existen
			_, _ = DB.Exec(`
				ALTER TABLE negocios ADD COLUMN IF NOT EXISTS max_dispositivos INT NOT NULL DEFAULT 4;
				ALTER TABLE negocios ADD COLUMN IF NOT EXISTS ts_auth_key TEXT;
				CREATE TABLE IF NOT EXISTS dispositivos_cliente (
					id_dispositivo SERIAL PRIMARY KEY,
					id_negocio INT NOT NULL REFERENCES negocios(id_negocio) ON DELETE CASCADE,
					device_id VARCHAR(100) NOT NULL,
					nombre_dispositivo VARCHAR(100) NOT NULL,
					tipo_dispositivo VARCHAR(20) DEFAULT 'desktop',
					fecha_registro TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					ultima_conexion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					activo BOOLEAN DEFAULT TRUE,
					UNIQUE (id_negocio, device_id)
				);
				CREATE INDEX IF NOT EXISTS idx_dispositivos_negocio_activo ON dispositivos_cliente (id_negocio, activo);
			`)
			return
		}
		logger.Warn("Esperando a PostgreSQL (intento %d/15)...", intento)
		time.Sleep(2 * time.Second)
	}

	logger.Error("No se pudo conectar a PostgreSQL tras varios intentos: %v", err)
	os.Exit(1)
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
