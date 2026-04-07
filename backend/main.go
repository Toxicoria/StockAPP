package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func pingHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Habilitar CORS: Permitimos que cualquier frontend consulte esta ruta
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	// 2. Armar la respuesta en formato JSON
	respuesta := map[string]string{
		"estado":  "online",
		"mensaje": "¡Conexión exitosa desde Svelte hasta tu API segura en Go! 🚀",
	}

	// 3. Enviar la respuesta
	json.NewEncoder(w).Encode(respuesta)
}

func main() {
	http.HandleFunc("/api/ping", pingHandler)
	print("print de test")
	fmt.Println("🚀 Servidor Go escuchando en el puerto 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

// test
