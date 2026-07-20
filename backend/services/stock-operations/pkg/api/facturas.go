package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"stock-operations/db"
	"stock-operations/pkg/logger"
)

// facturarVentaHandler asigna a la venta el próximo número de comprobante
// del negocio. Numeración local por ahora: cuando se integre ARCA (WSFEv1)
// este mismo flujo pasará a pedir el CAE y completar cae/cae_vencimiento;
// el esquema ya está preparado. POST /api/ventas/{id_venta}/factura
func facturarVentaHandler(w http.ResponseWriter, r *http.Request) {
	idVenta, err := strconv.Atoi(r.PathValue("id_venta"))
	if err != nil {
		responderError(w, http.StatusBadRequest, "id de venta inválido")
		return
	}
	negocio := negocioDe(r)

	tx, err := db.DB.Begin()
	if err != nil {
		logger.Error("facturas: error abriendo transacción: %v", err)
		responderError(w, http.StatusInternalServerError, "no se pudo facturar")
		return
	}
	defer tx.Rollback()

	// El UPDATE ... RETURNING toma y reserva el número en un solo paso:
	// dos cajas facturando a la vez nunca obtienen el mismo. Si la venta
	// no se puede facturar, el rollback devuelve el número sin dejar hueco.
	var puntoVenta string
	var numero int
	if err := tx.QueryRow(
		`UPDATE negocios
		    SET proximo_numero_factura = proximo_numero_factura + 1
		  WHERE id_negocio = $1
		  RETURNING punto_venta, proximo_numero_factura - 1`, negocio).
		Scan(&puntoVenta, &numero); err != nil {
		logger.Error("facturas: error reservando número: %v", err)
		responderError(w, http.StatusInternalServerError, "no se pudo facturar")
		return
	}

	factura := fmt.Sprintf("C %s-%08d", puntoVenta, numero)
	resultado, err := tx.Exec(
		`UPDATE ventas
		    SET factura = $3, nro_comprobante = $4
		  WHERE id_venta = $1 AND id_negocio = $2 AND factura IS NULL`,
		idVenta, negocio, factura, numero)
	if err != nil {
		logger.Error("facturas: error facturando venta %d: %v", idVenta, err)
		responderError(w, http.StatusInternalServerError, "no se pudo facturar")
		return
	}
	if filas, _ := resultado.RowsAffected(); filas == 0 {
		// ¿La venta no existe (o es de otro negocio) o ya estaba facturada?
		var yaFacturada string
		err := tx.QueryRow(
			`SELECT COALESCE(factura, '') FROM ventas WHERE id_venta = $1 AND id_negocio = $2`,
			idVenta, negocio).Scan(&yaFacturada)
		if err == sql.ErrNoRows {
			responderError(w, http.StatusNotFound, "la venta no existe")
			return
		}
		responderError(w, http.StatusConflict, "la venta ya tiene factura "+yaFacturada)
		return
	}

	if err := tx.Commit(); err != nil {
		logger.Error("facturas: error confirmando factura de venta %d: %v", idVenta, err)
		responderError(w, http.StatusInternalServerError, "no se pudo facturar")
		return
	}

	responderJSON(w, http.StatusOK, map[string]any{
		"id_venta": idVenta,
		"factura":  factura,
	})
}

type facturaListada struct {
	IDVenta int     `json:"id_venta"`
	Factura string  `json:"factura"`
	Hora    string  `json:"hora"`
	Total   float64 `json:"total_venta"`
}

// listarFacturasHandler devuelve las ventas facturadas de un día (por
// defecto, hoy). Solo admin. GET /api/facturas?fecha=2026-07-18
func listarFacturasHandler(w http.ResponseWriter, r *http.Request) {
	if rolDe(r) != "admin" {
		responderError(w, http.StatusForbidden, "solo un admin puede ver la facturación")
		return
	}
	fecha := r.URL.Query().Get("fecha")

	filas, err := db.DB.Query(
		`SELECT id_venta,
		        factura,
		        to_char(fecha_hora AT TIME ZONE 'UTC' AT TIME ZONE '`+zonaHoraria+`', 'HH24:MI'),
		        total_venta
		   FROM ventas
		  WHERE id_negocio = $1
		    AND factura IS NOT NULL
		    AND (fecha_hora AT TIME ZONE 'UTC' AT TIME ZONE '`+zonaHoraria+`')::date =
		        COALESCE(NULLIF($2, '')::date, (now() AT TIME ZONE '`+zonaHoraria+`')::date)
		  ORDER BY fecha_hora DESC`, negocioDe(r), fecha)
	if err != nil {
		logger.Error("facturas: error consultando: %v", err)
		responderError(w, http.StatusInternalServerError, "error consultando las facturas")
		return
	}
	defer filas.Close()

	facturas := []facturaListada{}
	for filas.Next() {
		var f facturaListada
		if err := filas.Scan(&f.IDVenta, &f.Factura, &f.Hora, &f.Total); err != nil {
			logger.Error("facturas: error leyendo fila: %v", err)
			responderError(w, http.StatusInternalServerError, "error leyendo las facturas")
			return
		}
		facturas = append(facturas, f)
	}
	responderJSON(w, http.StatusOK, facturas)
}

// negocioHandler devuelve los datos del negocio del usuario autenticado —
// alimentan la barra de título, la card "Datos del negocio" y la pantalla
// de Configuración. La clave fiscal nunca viaja: solo se informa si hay una
// guardada, para que la UI muestre "configurada" sin exponer el valor.
// GET /api/negocio
func negocioHandler(w http.ResponseWriter, r *http.Request) {
	var nombre, direccion, cuit, puntoVenta string
	var hayClaveFiscal bool
	err := db.DB.QueryRow(
		`SELECT nombre_negocio, COALESCE(direccion, ''), COALESCE(cuit, ''), COALESCE(punto_venta, '0001'),
		        clave_fiscal_cifrada IS NOT NULL
		   FROM negocios WHERE id_negocio = $1`, negocioDe(r)).
		Scan(&nombre, &direccion, &cuit, &puntoVenta, &hayClaveFiscal)
	if err != nil {
		logger.Error("negocio: error consultando datos: %v", err)
		responderError(w, http.StatusInternalServerError, "error consultando el negocio")
		return
	}
	responderJSON(w, http.StatusOK, map[string]any{
		"nombre_negocio":           nombre,
		"direccion":                direccion,
		"cuit":                     cuit,
		"punto_venta":              puntoVenta,
		"clave_fiscal_configurada": hayClaveFiscal,
	})
}

// cuerpoNegocio usa punteros para distinguir "no enviado" de "enviado
// vacío" — igual que cuerpoStock. clave_fiscal es la excepción: un string
// vacío explícito borra el secreto guardado (para poder desconfigurarla).
type cuerpoNegocio struct {
	NombreNegocio *string `json:"nombre_negocio"`
	Direccion     *string `json:"direccion"`
	Cuit          *string `json:"cuit"`
	PuntoVenta    *string `json:"punto_venta"`
	ClaveFiscal   *string `json:"clave_fiscal"`
}

// editarNegocioHandler actualiza la configuración del negocio. Solo admin.
// PUT /api/negocio
func editarNegocioHandler(w http.ResponseWriter, r *http.Request) {
	if rolDe(r) != "admin" {
		responderError(w, http.StatusForbidden, "solo un admin puede editar la configuración")
		return
	}

	var cuerpo cuerpoNegocio
	if err := json.NewDecoder(r.Body).Decode(&cuerpo); err != nil {
		responderError(w, http.StatusBadRequest, "cuerpo JSON inválido")
		return
	}

	// nil = no tocar la clave guardada; "" = borrarla; cualquier otro
	// valor = cifrar y reemplazar.
	var claveCifradaNueva []byte
	tocarClave := cuerpo.ClaveFiscal != nil
	if tocarClave && *cuerpo.ClaveFiscal != "" {
		cifrada, err := cifrar(*cuerpo.ClaveFiscal)
		if err != nil {
			logger.Error("negocio: error cifrando la clave fiscal: %v", err)
			responderError(w, http.StatusInternalServerError, "no se pudo guardar la configuración")
			return
		}
		claveCifradaNueva = cifrada
	}

	_, err := db.DB.Exec(
		`UPDATE negocios
		    SET nombre_negocio = COALESCE($2, nombre_negocio),
		        direccion      = COALESCE($3, direccion),
		        cuit           = COALESCE($4, cuit),
		        punto_venta    = COALESCE($5, punto_venta),
		        clave_fiscal_cifrada = CASE WHEN $6 THEN $7::bytea ELSE clave_fiscal_cifrada END
		  WHERE id_negocio = $1`,
		negocioDe(r), cuerpo.NombreNegocio, cuerpo.Direccion, cuerpo.Cuit, cuerpo.PuntoVenta,
		tocarClave, claveCifradaNueva)
	if err != nil {
		logger.Error("negocio: error actualizando configuración: %v", err)
		responderError(w, http.StatusInternalServerError, "no se pudo guardar la configuración")
		return
	}

	responderJSON(w, http.StatusOK, map[string]string{"estado": "ok"})
}
