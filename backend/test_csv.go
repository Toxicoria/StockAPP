//go:build ignore

package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
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

	fmt.Println("🚀 Iniciando pipeline de procesamiento de datos...")

	// 1. LEER CARPETA DE ENTRADA
	archivos, err := os.ReadDir(rutaCarpetaEntrada)
	if err != nil {
		log.Fatalf("❌ No se pudo leer la carpeta de entrada %s: %v", rutaCarpetaEntrada, err)
	}

	// 2. EXTRAER Y LIMPIAR DATOS
	for _, archivoInfo := range archivos {
		if filepath.Ext(archivoInfo.Name()) != ".csv" {
			continue
		}

		rutaCompleta := filepath.Join(rutaCarpetaEntrada, archivoInfo.Name())
		fmt.Printf("📄 Procesando: %s\n", archivoInfo.Name())

		archivo, err := os.Open(rutaCompleta)
		if err != nil {
			log.Printf("⚠️ Error al abrir %s: %v", archivoInfo.Name(), err)
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
	fmt.Println("\n💾 Generando archivo maestro unificado...")
	archivoSalida, err := os.Create(rutaArchivoSalida)
	if err != nil {
		log.Fatalf("❌ Error fatal al crear el archivo de salida: %v", err)
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
		log.Fatalf("Error al escribir encabezados: %v", err)
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
			log.Printf("⚠️ Error al escribir el producto %s: %v", codigo, err)
		}
	}

	// Asegurarnos de que todo se guarde
	escritor.Flush()
	if err := escritor.Error(); err != nil {
		log.Fatalf("❌ Error al finalizar la escritura: %v", err)
	}

	// 4. REPORTE FINAL
	fmt.Println("\n=================================================")
	fmt.Println("✅ OPERACIÓN COMPLETADA CON ÉXITO")
	fmt.Println("=================================================")
	fmt.Printf("📦 Total de productos unificados: %d\n", len(catalogoUnico))
	fmt.Printf("🗑️  Falsos EAN descartados:       %d\n", productosDescartadosPorEAN)
	fmt.Printf("🔁 Duplicados eliminados:         %d\n", productosDuplicados)
	fmt.Printf("📁 Archivo generado en:           %s\n", rutaArchivoSalida)
	fmt.Println("=================================================")
}
