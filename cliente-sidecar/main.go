package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"tailscale.com/tsnet"
)

func main() {
	// 1. Necesitamos una Auth Key para que este cliente se una a tu red
	authKey := os.Getenv("TS_AUTHKEY")
	if authKey == "" {
		log.Fatal("❌ Error: Falta la variable de entorno TS_AUTHKEY")
	}

	// 2. Inicializar el motor embebido de Tailscale (tsnet)
	s := &tsnet.Server{
		Hostname: "cliente-stock-app", // Así se verá el cliente en tu panel
		AuthKey:  authKey,
		Dir:      "./tsnet-state", // Guarda las llaves en una carpeta local para no pedir login cada vez
	}
	defer s.Close()

	fmt.Println("⏳ Conectando al búnker de Tailscale...")

	// 3. Crear un cliente HTTP que viaja exclusivamente por el túnel seguro
	tsClient := s.HTTPClient()

	// 4. Crear el servidor local para que Svelte le hable (El Proxy)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Permitir que Svelte (localhost) hable con este puerto
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Solo interceptar peticiones a la API
		if r.URL.Path == "/api/ping" {
			// Redirigir la petición al contenedor del servidor usando su nombre en Tailscale
			targetURL := "http://stock-server-api:8080" + r.URL.Path

			// Preparar el paquete
			req, err := http.NewRequest(r.Method, targetURL, r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Disparar por el túnel
			resp, err := tsClient.Do(req)
			if err != nil {
				http.Error(w, "El búnker no responde: "+err.Error(), http.StatusBadGateway)
				return
			}
			defer resp.Body.Close()

			// Devolver la respuesta exacta al Svelte
			w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
			w.WriteHeader(resp.StatusCode)
			io.Copy(w, resp.Body)
		} else {
			http.NotFound(w, r)
		}
	})

	fmt.Println("👻 Sidecar activo. Escuchando a Svelte en http://localhost:9090")
	log.Fatal(http.ListenAndServe(":9090", nil))
}
