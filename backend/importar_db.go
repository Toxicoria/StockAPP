//go:build ignore

package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://admin_dev:password_dev@127.0.0.1:5432/stock_db?sslmode=disable"
	}
	rutaArchivoLimpio := "./productos_limpios.csv"

	fmt.Println("Conectando PostgreSQL de docker")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error al abrir la conexion: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("No se pudo hacer ping a la DB, revisar que el conteiner este encendido: %v", err)
	}
	fmt.Println("Conexion establecida correctamente.")

	queryCrearTabla := `
	CREATE TABLE IF NOT EXISTS productos (
		id_producto VARCHAR(100) PRIMARY KEY,
		productos_ean VARCHAR(10),
		productos_descripcion VARCHAR(255),
		productos_cantidad_presentacion VARCHAR(50),
		productos_unidad_medida_presentacion VARCHAR(50),
		productos_marca VARCHAR(100)
	);`

	if _, err = db.Exec(queryCrearTabla); err != nil {
		log.Fatalf("Error al crear la tabla: %v", err)
	}
	fmt.Println("Tabla 'productos' lista.")

	// 2. ABRIR EL ARCHIVO CSV LIMPIO
	archivo, err := os.Open(rutaArchivoLimpio)
	if err != nil {
		log.Fatalf("Error al abrir %s. ¿Ya ejecutaste test_csv.go?: %v", rutaArchivoLimpio, err)
	}
	defer archivo.Close()

	lector := csv.NewReader(archivo)
	lector.Comma = '|'

	// Descartar encabezados
	if _, err = lector.Read(); err != nil {
		log.Fatalf("Error leyendo encabezados: %v", err)
	}

	// 3. INICIAR LA TRANSACCIÓN MASIVA
	fmt.Println("🚀 Iniciando inyección masiva de datos...")
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Error al iniciar transacción: %v", err)
	}

	// Preparamos la consulta con ON CONFLICT para no duplicar si corres el script 2 veces
	queryInsert := `
		INSERT INTO productos (
			id_producto, productos_ean, productos_descripcion, 
			productos_cantidad_presentacion, productos_unidad_medida_presentacion, productos_marca
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id_producto) DO NOTHING;
	`
	stmt, err := tx.Prepare(queryInsert)
	if err != nil {
		log.Fatalf("Error al preparar la consulta: %v", err)
	}
	defer stmt.Close()

	contadorInsertados := 0

	// 4. LEER E INSERTAR BUCLE
	for {
		linea, err := lector.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error leyendo línea: %v", err)
			continue
		}

		_, err = stmt.Exec(linea[0], linea[1], linea[2], linea[3], linea[4], linea[5])
		if err != nil {
			log.Printf("Error al insertar producto %s: %v", linea[0], err)
			continue
		}
		contadorInsertados++
	}

	// 5. CONFIRMAR (COMMIT) EN EL DISCO
	if err = tx.Commit(); err != nil {
		log.Fatalf("Error al hacer commit: %v", err)
	}

	fmt.Println("\n=================================================")
	fmt.Println("MIGRACIÓN COMPLETADA CON ÉXITO")
	fmt.Println("=================================================")
	fmt.Printf("Total procesado/verificado: %d productos\n", contadorInsertados)
	fmt.Println("=================================================")
}
