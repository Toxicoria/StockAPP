package api

import (
	"net/http"
	"strconv"

	"stock-operations/db"
	"stock-operations/pkg/logger"
)

type diaSemana struct {
	Fecha string  `json:"fecha"`
	Total float64 `json:"total"`
}

type topProducto struct {
	IDProducto  string  `json:"id_producto"`
	Descripcion string  `json:"descripcion"`
	Cantidad    float64 `json:"cantidad"`
}

// resumenHandler arma la vista del dueño: cómo viene el día y la semana,
// qué se vende más y cuántos productos hay que reponer. Solo admin.
// GET /api/resumen
func resumenHandler(w http.ResponseWriter, r *http.Request) {
	if rolDe(r) != "admin" {
		responderError(w, http.StatusForbidden, "solo un admin puede ver el resumen")
		return
	}
	negocio := negocioDe(r)

	// Hoy: total, cantidad y ticket promedio.
	var hoyTotal float64
	var hoyCantidad int
	if err := db.DB.QueryRow(
		`SELECT COALESCE(SUM(total_venta), 0), COUNT(*)
		   FROM ventas
		  WHERE id_negocio = $1
		    AND (fecha_hora AT TIME ZONE 'UTC' AT TIME ZONE '`+zonaHoraria+`')::date =
		        (now() AT TIME ZONE '`+zonaHoraria+`')::date`, negocio).
		Scan(&hoyTotal, &hoyCantidad); err != nil {
		logger.Error("resumen: error consultando hoy: %v", err)
		responderError(w, http.StatusInternalServerError, "error consultando el resumen")
		return
	}
	hoyPromedio := 0.0
	if hoyCantidad > 0 {
		hoyPromedio = hoyTotal / float64(hoyCantidad)
	}

	// Semana: los últimos 7 días siempre completos (generate_series mete
	// los días sin ventas con total 0), hoy al final.
	filas, err := db.DB.Query(
		`SELECT to_char(d.dia, 'YYYY-MM-DD'), COALESCE(SUM(v.total_venta), 0)
		   FROM generate_series(
		            (now() AT TIME ZONE '`+zonaHoraria+`')::date - 6,
		            (now() AT TIME ZONE '`+zonaHoraria+`')::date,
		            '1 day') AS d(dia)
		   LEFT JOIN ventas v
		          ON v.id_negocio = $1
		         AND (v.fecha_hora AT TIME ZONE 'UTC' AT TIME ZONE '`+zonaHoraria+`')::date = d.dia
		  GROUP BY d.dia
		  ORDER BY d.dia`, negocio)
	if err != nil {
		logger.Error("resumen: error consultando la semana: %v", err)
		responderError(w, http.StatusInternalServerError, "error consultando el resumen")
		return
	}
	defer filas.Close()

	semana := []diaSemana{}
	semanaTotal := 0.0
	for filas.Next() {
		var d diaSemana
		if err := filas.Scan(&d.Fecha, &d.Total); err != nil {
			logger.Error("resumen: error leyendo día: %v", err)
			responderError(w, http.StatusInternalServerError, "error leyendo el resumen")
			return
		}
		semana = append(semana, d)
		semanaTotal += d.Total
	}

	// Top 5 de la semana por unidades vendidas.
	filasTop, err := db.DB.Query(
		`SELECT d.id_producto, COALESCE(p.productos_descripcion, d.id_producto), SUM(d.cantidad_llevada)
		   FROM detalles_venta d
		   JOIN ventas v ON v.id_venta = d.id_venta
		   LEFT JOIN productos p ON p.id_producto = d.id_producto
		  WHERE v.id_negocio = $1
		    AND (v.fecha_hora AT TIME ZONE 'UTC' AT TIME ZONE '`+zonaHoraria+`')::date >=
		        (now() AT TIME ZONE '`+zonaHoraria+`')::date - 6
		  GROUP BY d.id_producto, p.productos_descripcion
		  ORDER BY SUM(d.cantidad_llevada) DESC
		  LIMIT 5`, negocio)
	if err != nil {
		logger.Error("resumen: error consultando el top: %v", err)
		responderError(w, http.StatusInternalServerError, "error consultando el resumen")
		return
	}
	defer filasTop.Close()

	top := []topProducto{}
	for filasTop.Next() {
		var t topProducto
		if err := filasTop.Scan(&t.IDProducto, &t.Descripcion, &t.Cantidad); err != nil {
			logger.Error("resumen: error leyendo top: %v", err)
			responderError(w, http.StatusInternalServerError, "error leyendo el resumen")
			return
		}
		top = append(top, t)
	}

	// Cuántos productos están para reponer (mínimo configurado y alcanzado).
	var alertas int
	if err := db.DB.QueryRow(
		`SELECT COUNT(*) FROM stock_interno
		  WHERE id_negocio = $1 AND stock_minimo > 0 AND cantidad_disponible <= stock_minimo`,
		negocio).Scan(&alertas); err != nil {
		logger.Error("resumen: error contando alertas: %v", err)
		responderError(w, http.StatusInternalServerError, "error consultando el resumen")
		return
	}

	responderJSON(w, http.StatusOK, map[string]any{
		"hoy": map[string]any{
			"total":    hoyTotal,
			"cantidad": hoyCantidad,
			"promedio": hoyPromedio,
		},
		"semana_total":  semanaTotal,
		"semana":        semana,
		"top_productos": top,
		"alertas":       alertas,
	})
}

// frecuentesHandler devuelve los productos más vendidos de los últimos 30
// días (la grilla "Los de siempre" de la pantalla Vender). Si todavía no
// hay ventas, salen los del inventario actualizados más recientemente.
// GET /api/stock/frecuentes?limite=12
func frecuentesHandler(w http.ResponseWriter, r *http.Request) {
	limite, err := strconv.Atoi(r.URL.Query().Get("limite"))
	if err != nil || limite <= 0 || limite > 50 {
		limite = 12
	}

	filas, err := db.DB.Query(
		`SELECT s.id_producto,
		        COALESCE(p.productos_descripcion, ''),
		        COALESCE(p.productos_marca, ''),
		        s.cantidad_disponible,
		        s.precio_venta,
		        COALESCE(s.stock_minimo, 0),
		        s.id_proveedor,
		        COALESCE(pr.nombre, '')
		   FROM stock_interno s
		   JOIN productos p ON p.id_producto = s.id_producto
		   LEFT JOIN proveedores pr ON pr.id_proveedor = s.id_proveedor
		   LEFT JOIN (
		        SELECT d.id_producto, SUM(d.cantidad_llevada) AS vendidos
		          FROM detalles_venta d
		          JOIN ventas v ON v.id_venta = d.id_venta
		         WHERE v.id_negocio = $1
		           AND v.fecha_hora >= now() - interval '30 days'
		         GROUP BY d.id_producto
		   ) ult ON ult.id_producto = s.id_producto
		  WHERE s.id_negocio = $1
		  ORDER BY COALESCE(ult.vendidos, 0) DESC, s.ultima_actualizacion DESC
		  LIMIT $2`, negocioDe(r), limite)
	if err != nil {
		logger.Error("frecuentes: error consultando: %v", err)
		responderError(w, http.StatusInternalServerError, "error consultando los frecuentes")
		return
	}
	defer filas.Close()

	items := []itemStock{}
	for filas.Next() {
		var i itemStock
		if err := filas.Scan(&i.IDProducto, &i.Descripcion, &i.Marca, &i.Cantidad, &i.Precio,
			&i.StockMinimo, &i.IDProveedor, &i.Proveedor); err != nil {
			logger.Error("frecuentes: error leyendo fila: %v", err)
			responderError(w, http.StatusInternalServerError, "error leyendo los frecuentes")
			return
		}
		items = append(items, i)
	}
	responderJSON(w, http.StatusOK, items)
}
