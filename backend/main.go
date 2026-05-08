package main

import (
	"net/http"
)

func main() {
	// 1. Conectarse a la base de datos antes de levantar el servidor.
	//    Si la DB no responde, el programa termina aquí con un error claro.
	IniciarDB()

	// 2. Registrar las rutas (handlers).
	//    Cada línea mapea una URL a una función que maneja esa petición.
	http.HandleFunc("/api/ping", pingHandler)

	// Próximos endpoints de autenticación — los iremos agregando:
	// http.HandleFunc("/api/login", loginHandler)
	// http.HandleFunc("/api/logout", logoutHandler)
	// http.HandleFunc("/api/refresh", refreshHandler)

	LogInfo("Servidor Go escuchando en el puerto 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

// pingHandler — endpoint de prueba para verificar que el servidor responde.
func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"estado":"online","mensaje":"API funcionando ✅"}`))
}
