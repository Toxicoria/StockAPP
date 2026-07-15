package main

import (
	"net/http"
	"time"
)

func main() {
	// 1. Conectarse a la base de datos antes de levantar el servidor.
	//    Si la DB no responde, el programa termina aquí con un error claro.
	IniciarDB()

	// 2. Registrar las rutas. El patrón "MÉTODO /ruta" (Go 1.22+) rechaza
	//    automáticamente los métodos incorrectos con 405.

	// Públicas
	http.HandleFunc("GET /api/ping", conCORS(pingHandler))
	http.HandleFunc("POST /api/login", conCORS(loginHandler))
	http.HandleFunc("POST /api/refresh", conCORS(refreshHandler))
	http.HandleFunc("POST /api/logout", conCORS(logoutHandler))

	// Protegidas: exigen "Authorization: Bearer <access token>"
	http.HandleFunc("GET /api/productos", conCORS(conAuth(listarProductosHandler)))
	http.HandleFunc("GET /api/stock", conCORS(conAuth(listarStockHandler)))
	http.HandleFunc("PUT /api/stock/{id_producto}", conCORS(conAuth(actualizarStockHandler)))

	// 3. Servidor con timeouts: evita que una conexión colgada quede
	//    ocupando recursos para siempre.
	servidor := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	LogInfo("Servidor Go escuchando en el puerto 8080...")
	if err := servidor.ListenAndServe(); err != nil {
		LogError("El servidor terminó con error: %v", err)
	}
}

// pingHandler — endpoint de prueba para verificar que el servidor responde.
func pingHandler(w http.ResponseWriter, r *http.Request) {
	responderJSON(w, http.StatusOK, map[string]string{
		"estado":  "online",
		"mensaje": "API funcionando ✅",
	})
}
