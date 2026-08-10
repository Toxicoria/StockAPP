package api

import "net/http"

// Router arma el mux con todas las rutas de la API. El patrón "MÉTODO /ruta"
// (Go 1.22+) rechaza automáticamente los métodos incorrectos con 405.
func Router() *http.ServeMux {
	mux := http.NewServeMux()

	// Públicas: sin autenticación.
	mux.HandleFunc("GET /api/ping", conCORS(pingHandler))
	// Registro inicial del negocio (solo funciona si no hay negocios aún).
	mux.HandleFunc("GET /api/registro", conCORS(negocioRegistradoHandler))
	mux.HandleFunc("POST /api/registro", conCORS(registroHandler))

	// Protegidas: exigen "Authorization: Bearer <access token>"
	mux.HandleFunc("GET /api/productos", conCORS(conAuth(listarProductosHandler)))
	mux.HandleFunc("POST /api/productos", conCORS(conAuth(crearProductoHandler)))
	mux.HandleFunc("PUT /api/productos/{id_producto}", conCORS(conAuth(editarProductoHandler)))
	mux.HandleFunc("DELETE /api/productos/{id_producto}", conCORS(conAuth(eliminarProductoHandler)))
	mux.HandleFunc("GET /api/stock", conCORS(conAuth(listarStockHandler)))
	mux.HandleFunc("PUT /api/stock/{id_producto}", conCORS(conAuth(actualizarStockHandler)))
	mux.HandleFunc("GET /api/stock/frecuentes", conCORS(conAuth(frecuentesHandler)))
	mux.HandleFunc("GET /api/proveedores", conCORS(conAuth(listarProveedoresHandler)))
	mux.HandleFunc("POST /api/proveedores", conCORS(conAuth(crearProveedorHandler)))
	mux.HandleFunc("POST /api/proveedores/{id_proveedor}/aumento", conCORS(conAuth(aplicarAumentoHandler)))
	mux.HandleFunc("POST /api/ventas", conCORS(conAuth(crearVentaHandler)))
	mux.HandleFunc("GET /api/ventas", conCORS(conAuth(listarVentasHandler)))
	mux.HandleFunc("POST /api/ventas/{id_venta}/factura", conCORS(conAuth(facturarVentaHandler)))
	mux.HandleFunc("GET /api/facturas", conCORS(conAuth(listarFacturasHandler)))
	mux.HandleFunc("GET /api/negocio", conCORS(conAuth(negocioHandler)))
	mux.HandleFunc("PUT /api/negocio", conCORS(conAuth(editarNegocioHandler)))
	mux.HandleFunc("GET /api/resumen", conCORS(conAuth(resumenHandler)))
	mux.HandleFunc("GET /api/usuarios", conCORS(conAuth(listarUsuariosHandler)))
	mux.HandleFunc("POST /api/usuarios", conCORS(conAuth(crearUsuarioHandler)))
	mux.HandleFunc("PUT /api/usuarios/{id_usuario}", conCORS(conAuth(editarUsuarioHandler)))
	mux.HandleFunc("DELETE /api/usuarios/{id_usuario}", conCORS(conAuth(eliminarUsuarioHandler)))

	// Preflight CORS: el mux con patrones "MÉTODO /ruta" respondería 405 a
	// OPTIONS si no hubiera un catch-all; conCORS corta con 204.
	mux.HandleFunc("OPTIONS /api/", conCORS(func(w http.ResponseWriter, r *http.Request) {}))

	return mux
}

// pingHandler — endpoint de prueba para verificar que el servidor responde.
func pingHandler(w http.ResponseWriter, r *http.Request) {
	responderJSON(w, http.StatusOK, map[string]string{
		"estado":  "online",
		"mensaje": "API funcionando ✅",
	})
}
