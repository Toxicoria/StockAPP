//go:build ignore

package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	csvReset  = "\033[0m"
	csvRed    = "\033[1;31m"
	csvGreen  = "\033[1;32m"
	csvYellow = "\033[1;33m"
	csvCyan   = "\033[1;36m"
)

// Estructura para guardar temporalmente en memoria
type Producto struct {
	Nombre   string
	Cantidad string
	Unidad   string
	Marca    string
}

func main() {
	// =================================================================
	// 📂 CONFIGURACIÓN DE RUTAS
	// =================================================================
	rutaCarpetaEntrada := "./lote_csv"
	rutaArchivoSalida := "./productos_limpios.csv"
	// =================================================================

	// Mapeo de columnas basado en tus CSV originales
	const (
		colIDProducto   = 3
		colProductosEAN = 4
		colNombre       = 5
		colCantidad     = 6
		colUnidad       = 7
		colMarca        = 8
	)

	catalogoUnico := make(map[string]Producto)
	productosDescartadosPorEAN := 0
	productosDuplicados := 0

	fmt.Printf("%s[INFO]%s Iniciando pipeline de procesamiento de datos...\n", csvCyan, csvReset)

	// 1. LEER CARPETA DE ENTRADA
	archivos, err := os.ReadDir(rutaCarpetaEntrada)
	if err != nil {
		fmt.Printf("%s[ERROR]%s No se pudo leer la carpeta de entrada %s: %v\n", csvRed, csvReset, rutaCarpetaEntrada, err)
		os.Exit(1)
	}

	// 2. EXTRAER Y LIMPIAR DATOS
	for _, archivoInfo := range archivos {
		if filepath.Ext(archivoInfo.Name()) != ".csv" {
			continue
		}

		rutaCompleta := filepath.Join(rutaCarpetaEntrada, archivoInfo.Name())
		fmt.Printf("%s[INFO]%s Procesando: %s\n", csvCyan, csvReset, archivoInfo.Name())

		archivo, err := os.Open(rutaCompleta)
		if err != nil {
			fmt.Printf("%s[WARN]%s Error al abrir %s: %v\n", csvYellow, csvReset, archivoInfo.Name(), err)
			continue
		}

		lector := csv.NewReader(archivo)
		lector.Comma = '|'

		// Descartar la primera línea (encabezados)
		_, err = lector.Read()
		if err != nil {
			archivo.Close()
			continue
		}

		for {
			linea, err := lector.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				continue
			}

			// Regla de descarte: EAN inválido
			if linea[colProductosEAN] == "0" {
				productosDescartadosPorEAN++
				continue
			}

			// Regla de saneamiento: Normalizar marca
			marcaLimpiada := strings.TrimSpace(linea[colMarca])
			if marcaLimpiada == "" {
				marcaLimpiada = "Sin Marca"
			}

			codigoBarras := strings.TrimSpace(linea[colIDProducto])

			// Control de duplicados en memoria
			if _, existe := catalogoUnico[codigoBarras]; existe {
				productosDuplicados++
				continue
			}

			// Guardar registro limpio
			catalogoUnico[codigoBarras] = Producto{
				Nombre:   strings.TrimSpace(linea[colNombre]),
				Cantidad: strings.TrimSpace(linea[colCantidad]),
				Unidad:   strings.TrimSpace(linea[colUnidad]),
				Marca:    marcaLimpiada,
			}
		}
		archivo.Close()
	}

	// 3. EXPORTAR EL NUEVO CATÁLOGO LIMPIO
	fmt.Printf("\n%s[INFO]%s Generando archivo maestro unificado...\n", csvCyan, csvReset)
	archivoSalida, err := os.Create(rutaArchivoSalida)
	if err != nil {
		fmt.Printf("%s[ERROR]%s Error fatal al crear el archivo de salida: %v\n", csvRed, csvReset, err)
		os.Exit(1)
	}
	defer archivoSalida.Close()

	escritor := csv.NewWriter(archivoSalida)
	escritor.Comma = '|' // Mantenemos el formato pipe

	// Escribir los nuevos encabezados
	encabezados := []string{
		"id_producto",
		"productos_ean",
		"productos_descripcion",
		"productos_cantidad_presentacion",
		"productos_unidad_medida_presentacion",
		"productos_marca",
	}

	if err := escritor.Write(encabezados); err != nil {
		fmt.Printf("%s[ERROR]%s Error al escribir encabezados: %v\n", csvRed, csvReset, err)
		os.Exit(1)
	}

	// Volcar el catálogo desde la memoria RAM al disco duro
	for codigo, prod := range catalogoUnico {
		fila := []string{
			codigo,
			"1", // Todo el catálogo exportado ya tiene EAN verificado
			prod.Nombre,
			prod.Cantidad,
			prod.Unidad,
			prod.Marca,
		}
		if err := escritor.Write(fila); err != nil {
			fmt.Printf("%s[WARN]%s Error al escribir el producto %s: %v\n", csvYellow, csvReset, codigo, err)
		}
	}

	// Asegurarnos de que todo se guarde
	escritor.Flush()
	if err := escritor.Error(); err != nil {
		fmt.Printf("%s[ERROR]%s Error al finalizar la escritura: %v\n", csvRed, csvReset, err)
		os.Exit(1)
	}

	// 4. REPORTE FINAL
	fmt.Println("\n=================================================")
	fmt.Printf("%s[OK]%s OPERACION COMPLETADA CON EXITO\n", csvGreen, csvReset)
	fmt.Println("=================================================")
	fmt.Printf("Total de productos unificados: %d\n", len(catalogoUnico))
	fmt.Printf("Falsos EAN descartados:        %d\n", productosDescartadosPorEAN)
	fmt.Printf("Duplicados eliminados:         %d\n", productosDuplicados)
	fmt.Printf("Archivo generado en:           %s\n", rutaArchivoSalida)
	fmt.Println("=================================================")
}
