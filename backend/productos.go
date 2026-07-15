package main

import (
	"net/http"
	"strconv"
)

type producto struct {
	ID          string `json:"id_producto"`
	Descripcion string `json:"descripcion"`
	Cantidad    string `json:"cantidad_presentacion"`
	Unidad      string `json:"unidad_medida"`
	Marca       string `json:"marca"`
}

// listarProductosHandler busca en el catálogo maestro por código de barras
// exacto, o por descripción/marca parcial (para el buscador de la caja).
// GET /api/productos?buscar=coca&limite=20
func listarProductosHandler(w http.ResponseWriter, r *http.Request) {
	buscar := r.URL.Query().Get("buscar")
	limite, err := strconv.Atoi(r.URL.Query().Get("limite"))
	if err != nil || limite <= 0 || limite > 200 {
		limite = 50
	}

	filas, err := DB.Query(
		`SELECT id_producto,
		        COALESCE(productos_descripcion, ''),
		        COALESCE(productos_cantidad_presentacion, ''),
		        COALESCE(productos_unidad_medida_presentacion, ''),
		        COALESCE(productos_marca, '')
		   FROM productos
		  WHERE $1 = ''
		     OR id_producto = $1
		     OR productos_descripcion ILIKE '%' || $1 || '%'
		     OR productos_marca ILIKE '%' || $1 || '%'
		  ORDER BY productos_descripcion
		  LIMIT $2`, buscar, limite)
	if err != nil {
		LogError("productos: error consultando el catálogo: %v", err)
		responderError(w, http.StatusInternalServerError, "error consultando el catálogo")
		return
	}
	defer filas.Close()

	productos := []producto{} // slice vacío (no nil) para que el JSON sea [] y no null
	for filas.Next() {
		var p producto
		if err := filas.Scan(&p.ID, &p.Descripcion, &p.Cantidad, &p.Unidad, &p.Marca); err != nil {
			LogError("productos: error leyendo fila: %v", err)
			responderError(w, http.StatusInternalServerError, "error leyendo el catálogo")
			return
		}
		productos = append(productos, p)
	}
	responderJSON(w, http.StatusOK, productos)
}
