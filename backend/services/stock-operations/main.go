package main

import (
	"net/http"
	"time"

	"stock-operations/db"
	"stock-operations/pkg/api"
	"stock-operations/pkg/logger"
)

func main() {
	// 1. Conectarse a la base de datos antes de levantar el servidor.
	//    Si la DB no responde, el programa termina aquí con un error claro.
	db.Iniciar()

	// 2. Servidor con timeouts: evita que una conexión colgada quede
	//    ocupando recursos para siempre. Las rutas viven en pkg/api.
	servidor := &http.Server{
		Addr:         ":8080",
		Handler:      api.Router(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("Servidor Go escuchando en el puerto 8080...")
	if err := servidor.ListenAndServe(); err != nil {
		logger.Error("El servidor terminó con error: %v", err)
	}
}
