package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"sort"

	"stock-operations/db"
	"stock-operations/pkg/logger"
)

// La zona horaria del negocio. fecha_hora se guarda en UTC (el contenedor
// corre en UTC); "hoy" se resuelve siempre en SQL con la doble conversión
// AT TIME ZONE — una sola conversión sobre un timestamp naive invierte el
// sentido y corre las ventas de la noche al día equivocado.
const zonaHoraria = "America/Argentina/Buenos_Aires"

type itemVenta struct {
	IDProducto string  `json:"id_producto"`
	Cantidad   float64 `json:"cantidad"`
}

type cuerpoVenta struct {
	MetodoPago string      `json:"metodo_pago"`
	Items      []itemVenta `json:"items"`
}

var metodosDePago = map[string]bool{"efectivo": true, "transferencia": true, "tarjeta": true}

// crearVentaHandler cobra un ticket: descuenta el stock y deja la venta
// anotada con sus detalles, todo en una transacción. El precio lo fija el
// servidor (el vigente en stock_interno), nunca el cliente.
// POST /api/ventas
func crearVentaHandler(w http.ResponseWriter, r *http.Request) {
	var cuerpo cuerpoVenta
	if err := json.NewDecoder(r.Body).Decode(&cuerpo); err != nil {
		responderError(w, http.StatusBadRequest, "cuerpo JSON inválido")
		return
	}
	if !metodosDePago[cuerpo.MetodoPago] {
		responderError(w, http.StatusBadRequest, "metodo_pago debe ser efectivo, transferencia o tarjeta")
		return
	}
	if len(cuerpo.Items) == 0 {
		responderError(w, http.StatusBadRequest, "el ticket está vacío")
		return
	}
	for _, item := range cuerpo.Items {
		if item.Cantidad <= 0 {
			responderError(w, http.StatusBadRequest, "las cantidades deben ser mayores a cero")
			return
		}
	}

	tx, err := db.DB.Begin()
	if err != nil {
		logger.Error("ventas: error abriendo transacción: %v", err)
		responderError(w, http.StatusInternalServerError, "no se pudo registrar la venta")
		return
	}
	defer tx.Rollback()

	negocio := negocioDe(r)

	// Primero se fijan los precios y se descuenta el stock. FOR UPDATE evita
	// que dos cajas cobren el mismo producto a la vez y pisen el descuento.
	// El stock puede quedar en 0 pero la venta no se frena: en el mostrador
	// no se rebota una cobranza porque el sistema diga que no hay.
	type lineaVenta struct {
		id       string
		cantidad float64
		precio   float64
	}
	// Consolidar y ordenar evita el deadlock A→B / B→A entre dos tickets
	// concurrentes. También deja un único detalle por producto.
	cantidades := make(map[string]float64, len(cuerpo.Items))
	for _, item := range cuerpo.Items {
		cantidades[item.IDProducto] += item.Cantidad
	}
	ids := make([]string, 0, len(cantidades))
	for id := range cantidades {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	lineas := make([]lineaVenta, 0, len(ids))
	total := 0.0
	for _, idProducto := range ids {
		cantidad := cantidades[idProducto]
		var precio float64
		err := tx.QueryRow(
			`SELECT precio_venta FROM stock_interno
			  WHERE id_negocio = $1 AND id_producto = $2
			  FOR UPDATE`, negocio, idProducto).Scan(&precio)
		if err == sql.ErrNoRows {
			responderError(w, http.StatusNotFound, "el producto "+idProducto+" no está en el inventario")
			return
		}
		if err != nil {
			logger.Error("ventas: error leyendo precio de %s: %v", idProducto, err)
			responderError(w, http.StatusInternalServerError, "no se pudo registrar la venta")
			return
		}

		if _, err := tx.Exec(
			`UPDATE stock_interno
			    SET cantidad_disponible  = GREATEST(cantidad_disponible - $3::numeric, 0),
			        ultima_actualizacion = CURRENT_TIMESTAMP
			  WHERE id_negocio = $1 AND id_producto = $2`,
			negocio, idProducto, cantidad); err != nil {
			logger.Error("ventas: error descontando stock de %s: %v", idProducto, err)
			responderError(w, http.StatusInternalServerError, "no se pudo registrar la venta")
			return
		}

		lineas = append(lineas, lineaVenta{idProducto, cantidad, precio})
		total += precio * cantidad
	}

	var idVenta int
	if err := tx.QueryRow(
		`INSERT INTO ventas (id_negocio, id_usuario, total_venta, metodo_pago)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id_venta`,
		negocio, usuarioDe(r), total, cuerpo.MetodoPago).Scan(&idVenta); err != nil {
		logger.Error("ventas: error insertando venta: %v", err)
		responderError(w, http.StatusInternalServerError, "no se pudo registrar la venta")
		return
	}

	for _, l := range lineas {
		if _, err := tx.Exec(
			`INSERT INTO detalles_venta (id_venta, id_producto, cantidad_llevada, precio_unitario_cobrado, subtotal)
			 VALUES ($1, $2, $3, $4, $5)`,
			idVenta, l.id, l.cantidad, l.precio, l.precio*l.cantidad); err != nil {
			logger.Error("ventas: error insertando detalle de %s: %v", l.id, err)
			responderError(w, http.StatusInternalServerError, "no se pudo registrar la venta")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		logger.Error("ventas: error confirmando venta: %v", err)
		responderError(w, http.StatusInternalServerError, "no se pudo registrar la venta")
		return
	}

	responderJSON(w, http.StatusCreated, map[string]any{
		"id_venta":    idVenta,
		"total_venta": total,
		"metodo_pago": cuerpo.MetodoPago,
	})
}

type ventaListada struct {
	ID         int     `json:"id_venta"`
	Hora       string  `json:"hora"`
	Detalle    string  `json:"detalle"`
	MetodoPago string  `json:"metodo_pago"`
	Total      float64 `json:"total_venta"`
	Factura    *string `json:"factura"`
}

// listarVentasHandler devuelve la caja de un día (por defecto, hoy en la
// zona horaria del negocio): totales arriba y la lista venta por venta.
// GET /api/ventas?fecha=2026-07-18
func listarVentasHandler(w http.ResponseWriter, r *http.Request) {
	fecha := r.URL.Query().Get("fecha") // vacío → hoy (lo resuelve el COALESCE en SQL)

	filas, err := db.DB.Query(
		`SELECT v.id_venta,
		        to_char(v.fecha_hora AT TIME ZONE 'UTC' AT TIME ZONE '`+zonaHoraria+`', 'HH24:MI'),
		        COALESCE(STRING_AGG(
		            COALESCE(p.productos_descripcion, d.id_producto) ||
		            CASE WHEN d.cantidad_llevada > 1 THEN ' ×' || d.cantidad_llevada::int ELSE '' END,
		            ', ' ORDER BY d.id_detalle), ''),
		        v.metodo_pago,
		        v.total_venta,
		        v.factura
		   FROM ventas v
		   LEFT JOIN detalles_venta d ON d.id_venta = v.id_venta
		   LEFT JOIN productos p ON p.id_producto = d.id_producto
		  WHERE v.id_negocio = $1
		    AND (v.fecha_hora AT TIME ZONE 'UTC' AT TIME ZONE '`+zonaHoraria+`')::date =
		        COALESCE(NULLIF($2, '')::date, (now() AT TIME ZONE '`+zonaHoraria+`')::date)
		  GROUP BY v.id_venta
		  ORDER BY v.fecha_hora DESC`, negocioDe(r), fecha)
	if err != nil {
		logger.Error("ventas: error consultando el día: %v", err)
		responderError(w, http.StatusInternalServerError, "error consultando las ventas")
		return
	}
	defer filas.Close()

	ventas := []ventaListada{}
	totalDia := 0.0
	for filas.Next() {
		var v ventaListada
		if err := filas.Scan(&v.ID, &v.Hora, &v.Detalle, &v.MetodoPago, &v.Total, &v.Factura); err != nil {
			logger.Error("ventas: error leyendo fila: %v", err)
			responderError(w, http.StatusInternalServerError, "error leyendo las ventas")
			return
		}
		ventas = append(ventas, v)
		totalDia += v.Total
	}

	promedio := 0.0
	if len(ventas) > 0 {
		promedio = totalDia / float64(len(ventas))
	}
	responderJSON(w, http.StatusOK, map[string]any{
		"total_dia": totalDia,
		"cantidad":  len(ventas),
		"promedio":  promedio,
		"ventas":    ventas,
	})
}
