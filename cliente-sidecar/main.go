package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

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
	//    Reenvía cualquier ruta bajo /api/ — no hace falta tocar este archivo
	//    cada vez que el backend suma un endpoint nuevo.
	http.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		// Permitir que Svelte (localhost) hable con este puerto
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		// El navegador manda OPTIONS (preflight) antes de un POST/PUT con JSON
		// o con Authorization: se responde acá mismo, sin viajar por el túnel.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Redirigir la petición al contenedor del servidor usando su nombre en Tailscale
		targetURL := "http://stock-server-api:8080" + r.URL.Path
		if r.URL.RawQuery != "" {
			targetURL += "?" + r.URL.RawQuery
		}

		// Preparar el paquete
		req, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Reenviar los headers del cliente (Content-Type, Authorization, etc.)
		req.Header = r.Header.Clone()

		// Disparar por el túnel
		resp, err := tsClient.Do(req)
		if err != nil {
			http.Error(w, "El búnker no responde: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Devolver la respuesta exacta al Svelte
		copiarHeaders(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})

	fmt.Println("👻 Sidecar activo. Escuchando a Svelte en http://localhost:9090")
	log.Fatal(http.ListenAndServe(":9090", nil))
}

// copiarHeaders vuelca los headers de la respuesta del backend, sin pisar
// los CORS que el proxy ya seteó.
func copiarHeaders(destino, origen http.Header) {
	for clave, valores := range origen {
		if strings.HasPrefix(clave, "Access-Control-") {
			continue
		}
		for _, v := range valores {
			destino.Add(clave, v)
		}
	}
}
