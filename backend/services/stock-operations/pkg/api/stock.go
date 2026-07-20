package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"stock-operations/db"
	"stock-operations/pkg/logger"
)

var errProveedorAjeno = errors.New("proveedor ajeno")

type itemStock struct {
	IDProducto  string  `json:"id_producto"`
	Descripcion string  `json:"descripcion"`
	Marca       string  `json:"marca"`
	Cantidad    float64 `json:"cantidad_disponible"`
	Precio      float64 `json:"precio_venta"`
	StockMinimo float64 `json:"stock_minimo"`
	IDProveedor *int    `json:"id_proveedor"`
	Proveedor   string  `json:"proveedor"`
}

// listarStockHandler devuelve el inventario del negocio del usuario
// autenticado (el id_negocio sale del token, no de la URL: un negocio
// nunca puede ver el stock de otro). GET /api/stock
func listarStockHandler(w http.ResponseWriter, r *http.Request) {
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
		   LEFT JOIN proveedores pr ON pr.id_proveedor = s.id_proveedor AND pr.id_negocio = s.id_negocio
		  WHERE s.id_negocio = $1
		  ORDER BY p.productos_descripcion`, negocioDe(r))
	if err != nil {
		logger.Error("stock: error consultando inventario: %v", err)
		responderError(w, http.StatusInternalServerError, "error consultando el inventario")
		return
	}
	defer filas.Close()

	items := []itemStock{}
	for filas.Next() {
		var i itemStock
		if err := filas.Scan(&i.IDProducto, &i.Descripcion, &i.Marca, &i.Cantidad, &i.Precio,
			&i.StockMinimo, &i.IDProveedor, &i.Proveedor); err != nil {
			logger.Error("stock: error leyendo fila: %v", err)
			responderError(w, http.StatusInternalServerError, "error leyendo el inventario")
			return
		}
		items = append(items, i)
	}
	responderJSON(w, http.StatusOK, items)
}

// Punteros para distinguir "no enviado" de "enviado en 0":
// {"cantidad_disponible": 0} pone el stock en cero, omitirlo lo deja como está.
// "delta" es el ajuste rápido de la pantalla Stock (botones + y −): suma o
// resta sobre el valor actual y lo puede usar cualquier rol; los campos
// absolutos siguen reservados a admin y no se mezclan con delta.
type cuerpoStock struct {
	Cantidad    *float64 `json:"cantidad_disponible"`
	Precio      *float64 `json:"precio_venta"`
	StockMinimo *float64 `json:"stock_minimo"`
	IDProveedor *int     `json:"id_proveedor"`
	Delta       *float64 `json:"delta"`
}

// actualizarStockHandler crea o actualiza la fila de stock de un producto
// para el negocio del usuario. PUT /api/stock/{id_producto}
func actualizarStockHandler(w http.ResponseWriter, r *http.Request) {
	idProducto := r.PathValue("id_producto")
	var cuerpo cuerpoStock
	if err := json.NewDecoder(r.Body).Decode(&cuerpo); err != nil {
		responderError(w, http.StatusBadRequest, "cuerpo JSON inválido")
		return
	}

	hayAbsolutos := cuerpo.Cantidad != nil || cuerpo.Precio != nil ||
		cuerpo.StockMinimo != nil || cuerpo.IDProveedor != nil

	// Modo delta: ajuste concurrente seguro (dos cajas no se pisan porque
	// la suma la hace la base, no el cliente). Cualquier rol autenticado.
	if cuerpo.Delta != nil {
		if hayAbsolutos {
			responderError(w, http.StatusBadRequest, "delta no se puede combinar con valores absolutos")
			return
		}
		var cantidad float64
		err := db.DB.QueryRow(
			`UPDATE stock_interno
			    SET cantidad_disponible  = GREATEST(cantidad_disponible + $3::numeric, 0),
			        ultima_actualizacion = CURRENT_TIMESTAMP
			  WHERE id_negocio = $1 AND id_producto = $2
			  RETURNING cantidad_disponible`,
			negocioDe(r), idProducto, cuerpo.Delta).Scan(&cantidad)
		if err != nil {
			responderError(w, http.StatusNotFound, "el producto no está en el inventario")
			return
		}
		responderJSON(w, http.StatusOK, map[string]any{
			"id_producto": idProducto, "cantidad_disponible": cantidad,
		})
		return
	}

	// Modo absoluto: setear precio/cantidad/mínimo/proveedor. Solo admin.
	if rolDe(r) != "admin" {
		responderError(w, http.StatusForbidden, "solo un admin puede modificar el stock")
		return
	}
	if !hayAbsolutos {
		responderError(w, http.StatusBadRequest, "hay que enviar delta o algún campo a modificar")
		return
	}
	if (cuerpo.Cantidad != nil && *cuerpo.Cantidad < 0) ||
		(cuerpo.Precio != nil && *cuerpo.Precio < 0) ||
		(cuerpo.StockMinimo != nil && *cuerpo.StockMinimo < 0) {
		responderError(w, http.StatusBadRequest, "los valores no pueden ser negativos")
		return
	}

	// UPSERT sobre UNIQUE (id_negocio, id_producto): crea la fila si el
	// producto todavía no está en el inventario; COALESCE conserva el valor
	// actual cuando un campo no se envía. El ::numeric es obligatorio: sin él
	// Postgres deduce integer por el "0" y rebota decimales como 850.50.
	tx, err := db.DB.Begin()
	if err != nil {
		logger.Error("stock: error abriendo transacción: %v", err)
		responderError(w, http.StatusInternalServerError, "no se pudo guardar el stock")
		return
	}
	defer tx.Rollback()
	if err := validarProveedorTx(tx, negocioDe(r), cuerpo.IDProveedor); err != nil {
		responderError(w, http.StatusBadRequest, "el proveedor no pertenece a este negocio")
		return
	}

	_, err = tx.Exec(
		`INSERT INTO stock_interno (id_negocio, id_producto, cantidad_disponible, precio_venta, stock_minimo, id_proveedor)
		 VALUES ($1, $2, COALESCE($3::numeric, 0), COALESCE($4::numeric, 0), COALESCE($5::numeric, 0), $6)
		 ON CONFLICT (id_negocio, id_producto) DO UPDATE
		    SET cantidad_disponible  = COALESCE($3::numeric, stock_interno.cantidad_disponible),
		        precio_venta         = COALESCE($4::numeric, stock_interno.precio_venta),
		        stock_minimo         = COALESCE($5::numeric, stock_interno.stock_minimo),
		        id_proveedor         = COALESCE($6::int, stock_interno.id_proveedor),
		        ultima_actualizacion = CURRENT_TIMESTAMP`,
		negocioDe(r), idProducto, cuerpo.Cantidad, cuerpo.Precio, cuerpo.StockMinimo, cuerpo.IDProveedor)
	if err != nil {
		logger.Error("stock: error en upsert de %s: %v", idProducto, err)
		// Lo más probable: la FK rebotó un código que no existe en el catálogo
		responderError(w, http.StatusBadRequest, "no se pudo guardar — verificá que el producto exista en el catálogo")
		return
	}
	if err := tx.Commit(); err != nil {
		logger.Error("stock: error confirmando upsert de %s: %v", idProducto, err)
		responderError(w, http.StatusInternalServerError, "no se pudo guardar el stock")
		return
	}

	responderJSON(w, http.StatusOK, map[string]string{"estado": "ok", "id_producto": idProducto})
}
