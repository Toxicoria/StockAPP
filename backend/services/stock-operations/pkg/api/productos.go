package api

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"

	"stock-operations/db"
	"stock-operations/pkg/logger"
)

// validarProveedorTx impide que una fila de inventario de un negocio apunte
// a un proveedor de otro. La FK por sí sola solo comprueba existencia global.
func validarProveedorTx(tx *sql.Tx, idNegocio int, idProveedor *int) error {
	if idProveedor == nil {
		return nil
	}
	var existe bool
	if err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM proveedores WHERE id_proveedor = $1 AND id_negocio = $2)`, *idProveedor, idNegocio).Scan(&existe); err != nil {
		return err
	}
	if !existe {
		return errProveedorAjeno
	}
	return nil
}

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

	filas, err := db.DB.Query(
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
		logger.Error("productos: error consultando el catálogo: %v", err)
		responderError(w, http.StatusInternalServerError, "error consultando el catálogo")
		return
	}
	defer filas.Close()

	productos := []producto{} // slice vacío (no nil) para que el JSON sea [] y no null
	for filas.Next() {
		var p producto
		if err := filas.Scan(&p.ID, &p.Descripcion, &p.Cantidad, &p.Unidad, &p.Marca); err != nil {
			logger.Error("productos: error leyendo fila: %v", err)
			responderError(w, http.StatusInternalServerError, "error leyendo el catálogo")
			return
		}
		productos = append(productos, p)
	}
	responderJSON(w, http.StatusOK, productos)
}

// El "producto" que ve la UI es catálogo + inventario: la descripción vive en
// productos (catálogo global, PK = código de barras) y el resto en stock_interno.
type cuerpoProducto struct {
	Descripcion  *string  `json:"descripcion"`
	CodigoBarras *string  `json:"codigo_barras"`
	Precio       *float64 `json:"precio_venta"`
	Cantidad     *float64 `json:"cantidad_disponible"`
	StockMinimo  *float64 `json:"stock_minimo"`
	IDProveedor  *int     `json:"id_proveedor"`
}

// crearProductoHandler da de alta un producto en el inventario del negocio.
// Si el código no existe en el catálogo global lo agrega; si ya existe, no lo
// toca. Sin código de barras se genera un id local "LOC-xxxxxxxx" (producto
// propio del negocio, ej. torta casera). Solo admin. POST /api/productos
func crearProductoHandler(w http.ResponseWriter, r *http.Request) {
	if rolDe(r) != "admin" {
		responderError(w, http.StatusForbidden, "solo un admin puede crear productos")
		return
	}

	var cuerpo cuerpoProducto
	if err := json.NewDecoder(r.Body).Decode(&cuerpo); err != nil {
		responderError(w, http.StatusBadRequest, "cuerpo JSON inválido")
		return
	}
	if cuerpo.Descripcion == nil || *cuerpo.Descripcion == "" {
		responderError(w, http.StatusBadRequest, "falta la descripción del producto")
		return
	}
	if (cuerpo.Precio != nil && *cuerpo.Precio < 0) ||
		(cuerpo.Cantidad != nil && *cuerpo.Cantidad < 0) ||
		(cuerpo.StockMinimo != nil && *cuerpo.StockMinimo < 0) {
		responderError(w, http.StatusBadRequest, "los valores no pueden ser negativos")
		return
	}

	idProducto := ""
	if cuerpo.CodigoBarras != nil {
		idProducto = *cuerpo.CodigoBarras
	}
	if idProducto == "" {
		aleatorio := make([]byte, 4)
		rand.Read(aleatorio)
		idProducto = "LOC-" + hex.EncodeToString(aleatorio)
	}

	tx, err := db.DB.Begin()
	if err != nil {
		logger.Error("productos: error abriendo transacción: %v", err)
		responderError(w, http.StatusInternalServerError, "no se pudo crear el producto")
		return
	}
	defer tx.Rollback()
	if err := validarProveedorTx(tx, negocioDe(r), cuerpo.IDProveedor); err != nil {
		responderError(w, http.StatusBadRequest, "el proveedor no pertenece a este negocio")
		return
	}

	// Catálogo global: solo se inserta si el código no existía (D1); una
	// descripción ya cargada por otro negocio no se pisa al crear.
	if _, err := tx.Exec(
		`INSERT INTO productos (id_producto, productos_descripcion)
		 VALUES ($1, $2)
		 ON CONFLICT (id_producto) DO NOTHING`, idProducto, *cuerpo.Descripcion); err != nil {
		logger.Error("productos: error insertando %s en catálogo: %v", idProducto, err)
		responderError(w, http.StatusInternalServerError, "no se pudo crear el producto")
		return
	}

	if _, err := tx.Exec(
		`INSERT INTO stock_interno (id_negocio, id_producto, cantidad_disponible, precio_venta, stock_minimo, id_proveedor)
		 VALUES ($1, $2, COALESCE($3::numeric, 0), COALESCE($4::numeric, 0), COALESCE($5::numeric, 0), $6)
		 ON CONFLICT (id_negocio, id_producto) DO UPDATE
		    SET cantidad_disponible  = COALESCE($3::numeric, stock_interno.cantidad_disponible),
		        precio_venta         = COALESCE($4::numeric, stock_interno.precio_venta),
		        stock_minimo         = COALESCE($5::numeric, stock_interno.stock_minimo),
		        id_proveedor         = COALESCE($6::int, stock_interno.id_proveedor),
		        ultima_actualizacion = CURRENT_TIMESTAMP`,
		negocioDe(r), idProducto, cuerpo.Cantidad, cuerpo.Precio, cuerpo.StockMinimo, cuerpo.IDProveedor); err != nil {
		logger.Error("productos: error en inventario de %s: %v", idProducto, err)
		responderError(w, http.StatusBadRequest, "no se pudo guardar — verificá el proveedor")
		return
	}

	if err := tx.Commit(); err != nil {
		logger.Error("productos: error confirmando alta de %s: %v", idProducto, err)
		responderError(w, http.StatusInternalServerError, "no se pudo crear el producto")
		return
	}

	responderJSON(w, http.StatusCreated, map[string]string{"id_producto": idProducto})
}

// editarProductoHandler actualiza campos sueltos (punteros: lo no enviado no
// se toca). La descripción edita el catálogo global — aceptado mientras haya
// un solo negocio (riesgo D1). Solo admin. PUT /api/productos/{id_producto}
func editarProductoHandler(w http.ResponseWriter, r *http.Request) {
	if rolDe(r) != "admin" {
		responderError(w, http.StatusForbidden, "solo un admin puede editar productos")
		return
	}

	idProducto := r.PathValue("id_producto")
	var cuerpo cuerpoProducto
	if err := json.NewDecoder(r.Body).Decode(&cuerpo); err != nil {
		responderError(w, http.StatusBadRequest, "cuerpo JSON inválido")
		return
	}
	if (cuerpo.Precio != nil && *cuerpo.Precio < 0) ||
		(cuerpo.Cantidad != nil && *cuerpo.Cantidad < 0) ||
		(cuerpo.StockMinimo != nil && *cuerpo.StockMinimo < 0) {
		responderError(w, http.StatusBadRequest, "los valores no pueden ser negativos")
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		logger.Error("productos: error abriendo transacción: %v", err)
		responderError(w, http.StatusInternalServerError, "no se pudo editar el producto")
		return
	}
	defer tx.Rollback()
	if err := validarProveedorTx(tx, negocioDe(r), cuerpo.IDProveedor); err != nil {
		responderError(w, http.StatusBadRequest, "el proveedor no pertenece a este negocio")
		return
	}

	resultado, err := tx.Exec(
		`UPDATE stock_interno
		    SET cantidad_disponible  = COALESCE($3::numeric, cantidad_disponible),
		        precio_venta         = COALESCE($4::numeric, precio_venta),
		        stock_minimo         = COALESCE($5::numeric, stock_minimo),
		        id_proveedor         = COALESCE($6::int, id_proveedor),
		        ultima_actualizacion = CURRENT_TIMESTAMP
		  WHERE id_negocio = $1 AND id_producto = $2`,
		negocioDe(r), idProducto, cuerpo.Cantidad, cuerpo.Precio, cuerpo.StockMinimo, cuerpo.IDProveedor)
	if err != nil {
		logger.Error("productos: error editando %s: %v", idProducto, err)
		responderError(w, http.StatusInternalServerError, "no se pudo editar el producto")
		return
	}
	if filas, _ := resultado.RowsAffected(); filas == 0 {
		responderError(w, http.StatusNotFound, "el producto no está en el inventario")
		return
	}

	if cuerpo.Descripcion != nil && *cuerpo.Descripcion != "" {
		if _, err := tx.Exec(
			`UPDATE productos SET productos_descripcion = $2 WHERE id_producto = $1`,
			idProducto, *cuerpo.Descripcion); err != nil {
			logger.Error("productos: error editando descripción de %s: %v", idProducto, err)
			responderError(w, http.StatusInternalServerError, "no se pudo editar el producto")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		logger.Error("productos: error confirmando edición de %s: %v", idProducto, err)
		responderError(w, http.StatusInternalServerError, "no se pudo editar el producto")
		return
	}

	responderJSON(w, http.StatusOK, map[string]string{"estado": "ok", "id_producto": idProducto})
}

// eliminarProductoHandler saca el producto del inventario del negocio; el
// catálogo global y el historial de ventas quedan intactos (las FK de
// detalles_venta apuntan a productos, no a stock_interno). Solo admin.
// DELETE /api/productos/{id_producto}
func eliminarProductoHandler(w http.ResponseWriter, r *http.Request) {
	if rolDe(r) != "admin" {
		responderError(w, http.StatusForbidden, "solo un admin puede eliminar productos")
		return
	}

	idProducto := r.PathValue("id_producto")
	resultado, err := db.DB.Exec(
		`DELETE FROM stock_interno WHERE id_negocio = $1 AND id_producto = $2`,
		negocioDe(r), idProducto)
	if err != nil {
		logger.Error("productos: error eliminando %s: %v", idProducto, err)
		responderError(w, http.StatusInternalServerError, "no se pudo eliminar el producto")
		return
	}
	if filas, _ := resultado.RowsAffected(); filas == 0 {
		responderError(w, http.StatusNotFound, "el producto no está en el inventario")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
