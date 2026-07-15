package main

import (
	"encoding/json"
	"net/http"
)

type itemStock struct {
	IDProducto  string  `json:"id_producto"`
	Descripcion string  `json:"descripcion"`
	Marca       string  `json:"marca"`
	Cantidad    float64 `json:"cantidad_disponible"`
	Precio      float64 `json:"precio_venta"`
}

// listarStockHandler devuelve el inventario del negocio del usuario
// autenticado (el id_negocio sale del token, no de la URL: un negocio
// nunca puede ver el stock de otro). GET /api/stock
func listarStockHandler(w http.ResponseWriter, r *http.Request) {
	filas, err := DB.Query(
		`SELECT s.id_producto,
		        COALESCE(p.productos_descripcion, ''),
		        COALESCE(p.productos_marca, ''),
		        s.cantidad_disponible,
		        s.precio_venta
		   FROM stock_interno s
		   JOIN productos p ON p.id_producto = s.id_producto
		  WHERE s.id_negocio = $1
		  ORDER BY p.productos_descripcion`, negocioDe(r))
	if err != nil {
		LogError("stock: error consultando inventario: %v", err)
		responderError(w, http.StatusInternalServerError, "error consultando el inventario")
		return
	}
	defer filas.Close()

	items := []itemStock{}
	for filas.Next() {
		var i itemStock
		if err := filas.Scan(&i.IDProducto, &i.Descripcion, &i.Marca, &i.Cantidad, &i.Precio); err != nil {
			LogError("stock: error leyendo fila: %v", err)
			responderError(w, http.StatusInternalServerError, "error leyendo el inventario")
			return
		}
		items = append(items, i)
	}
	responderJSON(w, http.StatusOK, items)
}

// Punteros para distinguir "no enviado" de "enviado en 0":
// {"cantidad_disponible": 0} pone el stock en cero, omitirlo lo deja como está.
type cuerpoStock struct {
	Cantidad *float64 `json:"cantidad_disponible"`
	Precio   *float64 `json:"precio_venta"`
}

// actualizarStockHandler crea o actualiza la fila de stock de un producto
// para el negocio del usuario. Solo rol admin. PUT /api/stock/{id_producto}
func actualizarStockHandler(w http.ResponseWriter, r *http.Request) {
	if rolDe(r) != "admin" {
		responderError(w, http.StatusForbidden, "solo un admin puede modificar el stock")
		return
	}

	idProducto := r.PathValue("id_producto")
	var cuerpo cuerpoStock
	if err := json.NewDecoder(r.Body).Decode(&cuerpo); err != nil {
		responderError(w, http.StatusBadRequest, "cuerpo JSON inválido")
		return
	}
	if cuerpo.Cantidad == nil && cuerpo.Precio == nil {
		responderError(w, http.StatusBadRequest, "hay que enviar cantidad_disponible y/o precio_venta")
		return
	}
	if (cuerpo.Cantidad != nil && *cuerpo.Cantidad < 0) || (cuerpo.Precio != nil && *cuerpo.Precio < 0) {
		responderError(w, http.StatusBadRequest, "los valores no pueden ser negativos")
		return
	}

	// UPSERT sobre UNIQUE (id_negocio, id_producto): crea la fila si el
	// producto todavía no está en el inventario; COALESCE conserva el valor
	// actual cuando un campo no se envía. El ::numeric es obligatorio: sin él
	// Postgres deduce integer por el "0" y rebota decimales como 850.50.
	_, err := DB.Exec(
		`INSERT INTO stock_interno (id_negocio, id_producto, cantidad_disponible, precio_venta)
		 VALUES ($1, $2, COALESCE($3::numeric, 0), COALESCE($4::numeric, 0))
		 ON CONFLICT (id_negocio, id_producto) DO UPDATE
		    SET cantidad_disponible  = COALESCE($3::numeric, stock_interno.cantidad_disponible),
		        precio_venta         = COALESCE($4::numeric, stock_interno.precio_venta),
		        ultima_actualizacion = CURRENT_TIMESTAMP`,
		negocioDe(r), idProducto, cuerpo.Cantidad, cuerpo.Precio)
	if err != nil {
		LogError("stock: error en upsert de %s: %v", idProducto, err)
		// Lo más probable: la FK rebotó un código que no existe en el catálogo
		responderError(w, http.StatusBadRequest, "no se pudo guardar — verificá que el producto exista en el catálogo")
		return
	}

	responderJSON(w, http.StatusOK, map[string]string{"estado": "ok", "id_producto": idProducto})
}
