package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"stock-operations/db"
	"stock-operations/pkg/logger"

	"github.com/lib/pq"
)

type proveedor struct {
	ID                  int        `json:"id_proveedor"`
	Nombre              string     `json:"nombre"`
	CantidadProductos   int        `json:"cantidad_productos"`
	UltimaActualizacion *time.Time `json:"ultima_actualizacion_precios"`
}

// listarProveedoresHandler devuelve los proveedores del negocio con cuántos
// productos del inventario tiene asignados cada uno. GET /api/proveedores
func listarProveedoresHandler(w http.ResponseWriter, r *http.Request) {
	filas, err := db.DB.Query(
		`SELECT pr.id_proveedor,
		        pr.nombre,
		        COUNT(s.id_stock)::int,
		        pr.ultima_actualizacion_precios
		   FROM proveedores pr
		   LEFT JOIN stock_interno s
		          ON s.id_proveedor = pr.id_proveedor AND s.id_negocio = pr.id_negocio
		  WHERE pr.id_negocio = $1
		  GROUP BY pr.id_proveedor
		  ORDER BY pr.nombre`, negocioDe(r))
	if err != nil {
		logger.Error("proveedores: error consultando: %v", err)
		responderError(w, http.StatusInternalServerError, "error consultando los proveedores")
		return
	}
	defer filas.Close()

	proveedores := []proveedor{}
	for filas.Next() {
		var p proveedor
		if err := filas.Scan(&p.ID, &p.Nombre, &p.CantidadProductos, &p.UltimaActualizacion); err != nil {
			logger.Error("proveedores: error leyendo fila: %v", err)
			responderError(w, http.StatusInternalServerError, "error leyendo los proveedores")
			return
		}
		proveedores = append(proveedores, p)
	}
	responderJSON(w, http.StatusOK, proveedores)
}

// crearProveedorHandler da de alta un proveedor del negocio. Solo admin.
// POST /api/proveedores  body: {"nombre": "Dulces del Valle"}
func crearProveedorHandler(w http.ResponseWriter, r *http.Request) {
	if rolDe(r) != "admin" {
		responderError(w, http.StatusForbidden, "solo un admin puede crear proveedores")
		return
	}

	var cuerpo struct {
		Nombre string `json:"nombre"`
	}
	if err := json.NewDecoder(r.Body).Decode(&cuerpo); err != nil || cuerpo.Nombre == "" {
		responderError(w, http.StatusBadRequest, "falta el nombre del proveedor")
		return
	}

	var id int
	err := db.DB.QueryRow(
		`INSERT INTO proveedores (id_negocio, nombre) VALUES ($1, $2) RETURNING id_proveedor`,
		negocioDe(r), cuerpo.Nombre).Scan(&id)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			responderError(w, http.StatusConflict, "ya existe un proveedor con ese nombre")
			return
		}
		logger.Error("proveedores: error creando %q: %v", cuerpo.Nombre, err)
		responderError(w, http.StatusInternalServerError, "no se pudo crear el proveedor")
		return
	}
	responderJSON(w, http.StatusCreated, proveedor{ID: id, Nombre: cuerpo.Nombre})
}

// aplicarAumentoHandler actualiza de una sola vez el precio de todos los
// productos del proveedor. El porcentaje puede ser negativo (baja de precios)
// y el resultado se redondea a la decena, como se hace en el mostrador.
// Solo admin. POST /api/proveedores/{id_proveedor}/aumento  body: {"porcentaje": 12.5}
func aplicarAumentoHandler(w http.ResponseWriter, r *http.Request) {
	if rolDe(r) != "admin" {
		responderError(w, http.StatusForbidden, "solo un admin puede cambiar precios")
		return
	}

	idProveedor, err := strconv.Atoi(r.PathValue("id_proveedor"))
	if err != nil {
		responderError(w, http.StatusBadRequest, "id de proveedor inválido")
		return
	}

	var cuerpo struct {
		Porcentaje float64 `json:"porcentaje"`
	}
	if err := json.NewDecoder(r.Body).Decode(&cuerpo); err != nil {
		responderError(w, http.StatusBadRequest, "cuerpo JSON inválido")
		return
	}
	if cuerpo.Porcentaje == 0 || cuerpo.Porcentaje < -500 || cuerpo.Porcentaje > 500 {
		responderError(w, http.StatusBadRequest, "el porcentaje debe ser distinto de 0 y estar entre -500 y 500")
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		logger.Error("proveedores: error abriendo transacción: %v", err)
		responderError(w, http.StatusInternalServerError, "no se pudo aplicar el aumento")
		return
	}
	defer tx.Rollback()

	// Marca la fecha del aumento y de paso verifica que el proveedor sea
	// del negocio del token: si no hay fila, es ajeno o no existe → 404.
	var nombre string
	err = tx.QueryRow(
		`UPDATE proveedores
		    SET ultima_actualizacion_precios = CURRENT_TIMESTAMP
		  WHERE id_proveedor = $1 AND id_negocio = $2
		  RETURNING nombre`, idProveedor, negocioDe(r)).Scan(&nombre)
	if err != nil {
		responderError(w, http.StatusNotFound, "proveedor no encontrado")
		return
	}

	resultado, err := tx.Exec(
		`UPDATE stock_interno
		    SET precio_venta = ROUND(precio_venta * (1 + $3::numeric / 100) / 10) * 10,
		        ultima_actualizacion = CURRENT_TIMESTAMP
		  WHERE id_negocio = $1 AND id_proveedor = $2`,
		negocioDe(r), idProveedor, cuerpo.Porcentaje)
	if err != nil {
		logger.Error("proveedores: error aplicando aumento de %v%% a %q: %v", cuerpo.Porcentaje, nombre, err)
		responderError(w, http.StatusInternalServerError, "no se pudo aplicar el aumento")
		return
	}
	actualizados, _ := resultado.RowsAffected()

	if err := tx.Commit(); err != nil {
		logger.Error("proveedores: error confirmando aumento: %v", err)
		responderError(w, http.StatusInternalServerError, "no se pudo aplicar el aumento")
		return
	}

	responderJSON(w, http.StatusOK, map[string]any{
		"proveedor":              nombre,
		"porcentaje":             cuerpo.Porcentaje,
		"productos_actualizados": actualizados,
	})
}
