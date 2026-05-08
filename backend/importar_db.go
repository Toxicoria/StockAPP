//go:build ignore

package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"

	_ "github.com/lib/pq"
)

const (
	dbReset  = "\033[0m"
	dbRed    = "\033[1;31m"
	dbGreen  = "\033[1;32m"
	dbYellow = "\033[1;33m"
	dbCyan   = "\033[1;36m"
)

func main() {

	connStr := "postgres://admin_dev:password_dev@localhost:5432/stock_db?sslmode=disable"
	rutaArchivoLimpio := "./productos_limpios.csv"

	fmt.Printf("%s[INFO]%s Conectando a PostgreSQL...\n", dbCyan, dbReset)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Printf("%s[ERROR]%s Error al abrir la conexion: %v\n", dbRed, dbReset, err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Printf("%s[ERROR]%s No se pudo hacer ping a la DB, revisar que el container este encendido: %v\n", dbRed, dbReset, err)
		os.Exit(1)
	}
	fmt.Printf("%s[OK]%s Conexion establecida correctamente.\n", dbGreen, dbReset)

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
		fmt.Printf("%s[ERROR]%s Error al crear la tabla: %v\n", dbRed, dbReset, err)
		os.Exit(1)
	}
	fmt.Printf("%s[OK]%s Tabla 'productos' lista.\n", dbGreen, dbReset)

	// 2. ABRIR EL ARCHIVO CSV LIMPIO
	archivo, err := os.Open(rutaArchivoLimpio)
	if err != nil {
		fmt.Printf("%s[ERROR]%s Error al abrir %s. Ya ejecutaste test_csv.go?: %v\n", dbRed, dbReset, rutaArchivoLimpio, err)
		os.Exit(1)
	}
	defer archivo.Close()

	lector := csv.NewReader(archivo)
	lector.Comma = '|'

	// Descartar encabezados
	if _, err = lector.Read(); err != nil {
		fmt.Printf("%s[ERROR]%s Error leyendo encabezados: %v\n", dbRed, dbReset, err)
		os.Exit(1)
	}

	// 3. INICIAR LA TRANSACCIÓN MASIVA
	fmt.Printf("%s[INFO]%s Iniciando inyeccion masiva de datos...\n", dbCyan, dbReset)
	tx, err := db.Begin()
	if err != nil {
		fmt.Printf("%s[ERROR]%s Error al iniciar transaccion: %v\n", dbRed, dbReset, err)
		os.Exit(1)
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
		fmt.Printf("%s[ERROR]%s Error al preparar la consulta: %v\n", dbRed, dbReset, err)
		os.Exit(1)
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
			fmt.Printf("%s[WARN]%s Error leyendo linea: %v\n", dbYellow, dbReset, err)
			continue
		}

		_, err = stmt.Exec(linea[0], linea[1], linea[2], linea[3], linea[4], linea[5])
		if err != nil {
			fmt.Printf("%s[WARN]%s Error al insertar producto %s: %v\n", dbYellow, dbReset, linea[0], err)
			continue
		}
		contadorInsertados++
	}

	// 5. CONFIRMAR (COMMIT) EN EL DISCO
	if err = tx.Commit(); err != nil {
		fmt.Printf("%s[ERROR]%s Error al hacer commit: %v\n", dbRed, dbReset, err)
		os.Exit(1)
	}

	fmt.Println("\n=================================================")
	fmt.Printf("%s[OK]%s MIGRACION COMPLETADA CON EXITO\n", dbGreen, dbReset)
	fmt.Println("=================================================")
	fmt.Printf("Total procesado/verificado: %d productos\n", contadorInsertados)
	fmt.Println("=================================================")
}
