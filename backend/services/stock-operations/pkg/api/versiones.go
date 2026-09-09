package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"stock-operations/db"
	"stock-operations/pkg/logger"
)

// VersionDesktop representa una versión registrada de la app de escritorio.
type VersionDesktop struct {
	ID                int     `json:"id"`
	Version           string  `json:"version"`
	VersionMinima     string  `json:"version_minima"`
	EsObligatoria     bool    `json:"es_obligatoria"`
	Canal             string  `json:"canal"`
	Estado            string  `json:"estado"`
	URLWindows        string  `json:"url_windows"`
	FirmaWindows      string  `json:"firma_windows"`
	URLLinux          string  `json:"url_linux"`
	FirmaLinux        string  `json:"firma_linux"`
	NotasVersion      string  `json:"notas_version"`
	MotivoObligatoria string  `json:"motivo_obligatoria"`
	PublicadaEn       *string `json:"publicada_en"`
	CreadaEn          string  `json:"creada_en"`
	ActualizadaEn     string  `json:"actualizada_en"`
}

// reqCrearVersion payload para dar de alta una versión.
type reqCrearVersion struct {
	Version           string `json:"version"`
	VersionMinima     string `json:"version_minima"`
	EsObligatoria     bool   `json:"es_obligatoria"`
	Canal             string `json:"canal"`
	Estado            string `json:"estado"`
	URLWindows        string `json:"url_windows"`
	FirmaWindows      string `json:"firma_windows"`
	URLLinux          string `json:"url_linux"`
	FirmaLinux        string `json:"firma_linux"`
	NotasVersion      string `json:"notas_version"`
	MotivoObligatoria string `json:"motivo_obligatoria"`
}

// reqEditarVersion payload para actualizar propiedades de una versión.
type reqEditarVersion struct {
	VersionMinima     *string `json:"version_minima"`
	EsObligatoria     *bool   `json:"es_obligatoria"`
	Canal             *string `json:"canal"`
	Estado            *string `json:"estado"`
	URLWindows        *string `json:"url_windows"`
	FirmaWindows      *string `json:"firma_windows"`
	URLLinux          *string `json:"url_linux"`
	FirmaLinux        *string `json:"firma_linux"`
	NotasVersion      *string `json:"notas_version"`
	MotivoObligatoria *string `json:"motivo_obligatoria"`
}

// PlataformaTauri representa los datos de descarga para un target específico.
type PlataformaTauri struct {
	Signature string `json:"signature"`
	URL       string `json:"url"`
}

// CustomTauriPayload metadatos extendidos para la lógica de StockAPP.
type CustomTauriPayload struct {
	Obligatoria   bool   `json:"obligatoria"`
	VersionMinima string `json:"version_minima"`
	Motivo        string `json:"motivo,omitempty"`
}

// RespTauriUpdate estructura oficial esperada por @tauri-apps/plugin-updater.
type RespTauriUpdate struct {
	Version   string                     `json:"version"`
	Notes     string                     `json:"notes"`
	PubDate   string                     `json:"pub_date"`
	Platforms map[string]PlataformaTauri `json:"platforms"`
	Custom    CustomTauriPayload         `json:"custom"`
}

// parseSemVer extrae los componentes numéricos de una cadena de versión semántica.
func parseSemVer(v string) (int, int, int) {
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimSpace(v)
	partes := strings.Split(v, ".")
	var mayor, menor, parche int
	if len(partes) > 0 {
		mayor, _ = strconv.Atoi(partes[0])
	}
	if len(partes) > 1 {
		menor, _ = strconv.Atoi(partes[1])
	}
	if len(partes) > 2 {
		// Quita sufijos de pre-release (ej: "0-beta1")
		p := strings.Split(partes[2], "-")[0]
		p = strings.Split(p, "+")[0]
		parche, _ = strconv.Atoi(p)
	}
	return mayor, menor, parche
}

// esVersionMenor devuelve true si vActual es estrictamente menor que vReferencia.
func esVersionMenor(vActual, vReferencia string) bool {
	maj1, min1, pat1 := parseSemVer(vActual)
	maj2, min2, pat2 := parseSemVer(vReferencia)

	if maj1 != maj2 {
		return maj1 < maj2
	}
	if min1 != min2 {
		return min1 < min2
	}
	return pat1 < pat2
}

// desktopUpdateHandler responde las consultas de actualización del plugin de Tauri 2.
// GET /api/desktop/update/{target}/{current_version}
func desktopUpdateHandler(w http.ResponseWriter, r *http.Request) {
	target := strings.TrimSpace(r.PathValue("target"))
	if target == "" && r.PathValue("os") != "" {
		target = strings.TrimSpace(r.PathValue("os")) + "-" + strings.TrimSpace(r.PathValue("arch"))
	}
	currentVersion := strings.TrimSpace(r.PathValue("current_version"))

	if currentVersion == "" {
		responderError(w, http.StatusBadRequest, "la versión actual es obligatoria")
		return
	}

	// Buscar la versión más reciente publicada para el canal 'produccion'
	query := `
		SELECT id, version, version_minima, es_obligatoria,
		       COALESCE(url_windows, ''), COALESCE(firma_windows, ''),
		       COALESCE(url_linux, ''), COALESCE(firma_linux, ''),
		       COALESCE(notas_version, ''), COALESCE(motivo_obligatoria, ''),
		       COALESCE(publicada_en, creada_en)
		FROM versiones_desktop
		WHERE estado = 'publicada' AND canal = 'produccion'
		ORDER BY creada_en DESC
		LIMIT 1;
	`

	var id int
	var ver, verMin, urlWin, firmaWin, urlLin, firmaLin, notas, motivo string
	var esOblig bool
	var pubDate time.Time

	err := db.DB.QueryRowContext(r.Context(), query).Scan(
		&id, &ver, &verMin, &esOblig,
		&urlWin, &firmaWin, &urlLin, &firmaLin,
		&notas, &motivo, &pubDate,
	)

	if err == sql.ErrNoRows {
		// No hay ninguna versión publicada: responder 204 No Content
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err != nil {
		logger.Error("desktopUpdateHandler: error al consultar última versión: %v", err)
		responderError(w, http.StatusInternalServerError, "error al verificar actualizaciones")
		return
	}

	// Si la versión actual no es menor que la publicada, la app está al día -> HTTP 204 No Content
	if !esVersionMenor(currentVersion, ver) {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Determinar si la actualización es obligatoria/forzada:
	// - Si fue marcada como obligatoria explícitamente en el panel
	// - O si la versión actual es menor a la versión mínima soportada
	esForzada := esOblig || esVersionMenor(currentVersion, verMin)

	// Construir mapa de plataformas
	platforms := make(map[string]PlataformaTauri)

	if urlWin != "" {
		platforms["windows-x86_64"] = PlataformaTauri{
			Signature: firmaWin,
			URL:       urlWin,
		}
		// Alias por compatibilidad con variantes de target de Tauri
		platforms["windows-x86_64-nsis"] = PlataformaTauri{
			Signature: firmaWin,
			URL:       urlWin,
		}
		platforms["windows-x86_64-msi"] = PlataformaTauri{
			Signature: firmaWin,
			URL:       urlWin,
		}
	}

	if urlLin != "" {
		platforms["linux-x86_64"] = PlataformaTauri{
			Signature: firmaLin,
			URL:       urlLin,
		}
		platforms["linux-x86_64-appimage"] = PlataformaTauri{
			Signature: firmaLin,
			URL:       urlLin,
		}
	}

	resp := RespTauriUpdate{
		Version:   ver,
		Notes:     notas,
		PubDate:   pubDate.Format(time.RFC3339),
		Platforms: platforms,
		Custom: CustomTauriPayload{
			Obligatoria:   esForzada,
			VersionMinima: verMin,
			Motivo:        motivo,
		},
	}

	logger.Info("desktopUpdateHandler: entregando actualización %s a cliente %s (%s, forzada: %v)", ver, currentVersion, target, esForzada)
	responderJSON(w, http.StatusOK, resp)
}

// adminListarVersionesHandler devuelve todas las versiones registradas para el panel de administración.
// GET /api/admin/versiones
func adminListarVersionesHandler(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT id, version, version_minima, es_obligatoria, canal, estado,
		       COALESCE(url_windows, ''), COALESCE(firma_windows, ''),
		       COALESCE(url_linux, ''), COALESCE(firma_linux, ''),
		       COALESCE(notas_version, ''), COALESCE(motivo_obligatoria, ''),
		       to_char(publicada_en, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
		       to_char(creada_en, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
		       to_char(actualizada_en, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		FROM versiones_desktop
		ORDER BY creada_en DESC;
	`

	rows, err := db.DB.QueryContext(r.Context(), query)
	if err != nil {
		logger.Error("adminListarVersionesHandler: error al consultar versiones: %v", err)
		responderError(w, http.StatusInternalServerError, "error al consultar versiones")
		return
	}
	defer rows.Close()

	versiones := make([]VersionDesktop, 0)
	for rows.Next() {
		var v VersionDesktop
		var pubEn sql.NullString
		if err := rows.Scan(
			&v.ID, &v.Version, &v.VersionMinima, &v.EsObligatoria, &v.Canal, &v.Estado,
			&v.URLWindows, &v.FirmaWindows, &v.URLLinux, &v.FirmaLinux,
			&v.NotasVersion, &v.MotivoObligatoria, &pubEn, &v.CreadaEn, &v.ActualizadaEn,
		); err != nil {
			logger.Error("adminListarVersionesHandler: error al escanear fila: %v", err)
			continue
		}
		if pubEn.Valid {
			v.PublicadaEn = &pubEn.String
		}
		versiones = append(versiones, v)
	}

	responderJSON(w, http.StatusOK, map[string]any{"versiones": versiones})
}

// adminCrearVersionHandler da de alta una versión en la base de datos.
// POST /api/admin/versiones
func adminCrearVersionHandler(w http.ResponseWriter, r *http.Request) {
	var body reqCrearVersion
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		responderError(w, http.StatusBadRequest, "cuerpo de solicitud inválido")
		return
	}

	body.Version = strings.TrimSpace(body.Version)
	if body.Version == "" {
		responderError(w, http.StatusBadRequest, "el número de versión es obligatorio")
		return
	}
	if body.VersionMinima == "" {
		body.VersionMinima = "0.1.0"
	}
	if body.Canal == "" {
		body.Canal = "produccion"
	}
	if body.Estado == "" {
		body.Estado = "borrador"
	}

	var publicadaEn *time.Time
	if body.Estado == "publicada" {
		ahora := time.Now().UTC()
		publicadaEn = &ahora
	}

	query := `
		INSERT INTO versiones_desktop (
			version, version_minima, es_obligatoria, canal, estado,
			url_windows, firma_windows, url_linux, firma_linux,
			notas_version, motivo_obligatoria,
			publicada_en
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11,
			$12
		) RETURNING id, to_char(creada_en, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), to_char(actualizada_en, 'YYYY-MM-DD"T"HH24:MI:SS"Z"');
	`

	var id int
	var creadaEn, actualizadaEn string
	err := db.DB.QueryRowContext(r.Context(), query,
		body.Version, body.VersionMinima, body.EsObligatoria, body.Canal, body.Estado,
		body.URLWindows, body.FirmaWindows, body.URLLinux, body.FirmaLinux,
		body.NotasVersion, body.MotivoObligatoria, publicadaEn,
	).Scan(&id, &creadaEn, &actualizadaEn)

	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate key") {
			responderError(w, http.StatusConflict, "la versión ya existe")
			return
		}
		logger.Error("adminCrearVersionHandler: error al insertar versión: %v", err)
		responderError(w, http.StatusInternalServerError, "error al crear versión")
		return
	}

	logger.Info("adminCrearVersionHandler: versión %s creada exitosamente (ID: %d)", body.Version, id)
	responderJSON(w, http.StatusCreated, map[string]any{
		"mensaje":        "versión creada exitosamente",
		"id":             id,
		"version":        body.Version,
		"creada_en":      creadaEn,
		"actualizada_en": actualizadaEn,
	})
}

// adminEditarVersionHandler permite cambiar estado, obligatoriedad y notas de una versión.
// PUT /api/admin/versiones/{id}
func adminEditarVersionHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSpace(r.PathValue("id"))
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		responderError(w, http.StatusBadRequest, "identificador de versión inválido")
		return
	}

	var body reqEditarVersion
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		responderError(w, http.StatusBadRequest, "cuerpo de solicitud inválido")
		return
	}

	// Obtener valores actuales
	var vActual VersionDesktop
	queryGet := `
		SELECT id, version, version_minima, es_obligatoria, canal, estado,
		       COALESCE(url_windows, ''), COALESCE(firma_windows, ''),
		       COALESCE(url_linux, ''), COALESCE(firma_linux, ''),
		       COALESCE(notas_version, ''), COALESCE(motivo_obligatoria, '')
		FROM versiones_desktop WHERE id = $1;
	`
	err = db.DB.QueryRowContext(r.Context(), queryGet, id).Scan(
		&vActual.ID, &vActual.Version, &vActual.VersionMinima, &vActual.EsObligatoria,
		&vActual.Canal, &vActual.Estado, &vActual.URLWindows, &vActual.FirmaWindows,
		&vActual.URLLinux, &vActual.FirmaLinux, &vActual.NotasVersion, &vActual.MotivoObligatoria,
	)
	if err == sql.ErrNoRows {
		responderError(w, http.StatusNotFound, "versión no encontrada")
		return
	}
	if err != nil {
		responderError(w, http.StatusInternalServerError, "error al consultar versión")
		return
	}

	// Aplicar parches
	if body.VersionMinima != nil {
		vActual.VersionMinima = strings.TrimSpace(*body.VersionMinima)
	}
	if body.EsObligatoria != nil {
		vActual.EsObligatoria = *body.EsObligatoria
	}
	if body.Canal != nil {
		vActual.Canal = strings.TrimSpace(*body.Canal)
	}
	nuevoEstado := vActual.Estado
	if body.Estado != nil {
		nuevoEstado = strings.TrimSpace(*body.Estado)
	}
	if body.URLWindows != nil {
		vActual.URLWindows = strings.TrimSpace(*body.URLWindows)
	}
	if body.FirmaWindows != nil {
		vActual.FirmaWindows = strings.TrimSpace(*body.FirmaWindows)
	}
	if body.URLLinux != nil {
		vActual.URLLinux = strings.TrimSpace(*body.URLLinux)
	}
	if body.FirmaLinux != nil {
		vActual.FirmaLinux = strings.TrimSpace(*body.FirmaLinux)
	}
	if body.NotasVersion != nil {
		vActual.NotasVersion = strings.TrimSpace(*body.NotasVersion)
	}
	if body.MotivoObligatoria != nil {
		vActual.MotivoObligatoria = strings.TrimSpace(*body.MotivoObligatoria)
	}

	queryUpdate := `
		UPDATE versiones_desktop SET
			version_minima = $1,
			es_obligatoria = $2,
			canal = $3,
			estado = $4,
			url_windows = $5,
			firma_windows = $6,
			url_linux = $7,
			firma_linux = $8,
			notas_version = $9,
			motivo_obligatoria = $10,
			publicada_en = CASE
				WHEN $4::varchar = 'publicada' AND publicada_en IS NULL THEN CURRENT_TIMESTAMP
				ELSE publicada_en
			END,
			actualizada_en = CURRENT_TIMESTAMP
		WHERE id = $11;
	`

	_, err = db.DB.ExecContext(r.Context(), queryUpdate,
		vActual.VersionMinima, vActual.EsObligatoria, vActual.Canal, nuevoEstado,
		vActual.URLWindows, vActual.FirmaWindows, vActual.URLLinux, vActual.FirmaLinux,
		vActual.NotasVersion, vActual.MotivoObligatoria, id,
	)
	if err != nil {
		logger.Error("adminEditarVersionHandler: error al actualizar versión %d: %v", id, err)
		responderError(w, http.StatusInternalServerError, "error al actualizar versión")
		return
	}

	logger.Info("adminEditarVersionHandler: versión %d (%s) actualizada a estado '%s'", id, vActual.Version, nuevoEstado)
	responderJSON(w, http.StatusOK, map[string]string{"mensaje": "versión actualizada correctamente"})
}

// adminEliminarVersionHandler elimina una versión del catálogo.
// DELETE /api/admin/versiones/{id}
func adminEliminarVersionHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSpace(r.PathValue("id"))
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		responderError(w, http.StatusBadRequest, "identificador de versión inválido")
		return
	}

	res, err := db.DB.ExecContext(r.Context(), "DELETE FROM versiones_desktop WHERE id = $1", id)
	if err != nil {
		logger.Error("adminEliminarVersionHandler: error al eliminar versión %d: %v", id, err)
		responderError(w, http.StatusInternalServerError, "error al eliminar versión")
		return
	}

	afectados, _ := res.RowsAffected()
	if afectados == 0 {
		responderError(w, http.StatusNotFound, "versión no encontrada")
		return
	}

	logger.Info("adminEliminarVersionHandler: versión %d eliminada exitosamente", id)
	responderJSON(w, http.StatusOK, map[string]string{"mensaje": "versión eliminada correctamente"})
}

// webhookRegistrarVersionHandler permite al pipeline CI/CD registrar artefactos recién compilados.
// POST /api/admin/versiones/webhook
func webhookRegistrarVersionHandler(w http.ResponseWriter, r *http.Request) {
	secretEsperado := os.Getenv("CI_WEBHOOK_SECRET")
	secretHeader := r.Header.Get("X-Webhook-Secret")

	// Si está configurado el secreto en entorno, validar obligatoriamente
	if secretEsperado != "" && secretHeader != secretEsperado {
		responderError(w, http.StatusUnauthorized, "secreto de webhook inválido o ausente")
		return
	}

	var body reqCrearVersion
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		responderError(w, http.StatusBadRequest, "cuerpo de solicitud inválido")
		return
	}

	body.Version = strings.TrimSpace(body.Version)
	if body.Version == "" {
		responderError(w, http.StatusBadRequest, "el número de versión es obligatorio")
		return
	}
	if body.VersionMinima == "" {
		body.VersionMinima = "0.1.0"
	}
	if body.Canal == "" {
		body.Canal = "produccion"
	}

	// El pipeline registra la versión inicialmente como 'borrador' para revisión del SuperAdmin
	query := `
		INSERT INTO versiones_desktop (
			version, version_minima, canal, estado,
			url_windows, firma_windows, url_linux, firma_linux,
			notas_version
		) VALUES (
			$1, $2, $3, 'borrador',
			$4, $5, $6, $7,
			$8
		)
		ON CONFLICT (version) DO UPDATE SET
			url_windows = EXCLUDED.url_windows,
			firma_windows = EXCLUDED.firma_windows,
			url_linux = EXCLUDED.url_linux,
			firma_linux = EXCLUDED.firma_linux,
			notas_version = COALESCE(NULLIF(EXCLUDED.notas_version, ''), versiones_desktop.notas_version),
			actualizada_en = CURRENT_TIMESTAMP
		RETURNING id;
	`

	var id int
	err := db.DB.QueryRowContext(r.Context(), query,
		body.Version, body.VersionMinima, body.Canal,
		body.URLWindows, body.FirmaWindows, body.URLLinux, body.FirmaLinux,
		body.NotasVersion,
	).Scan(&id)

	if err != nil {
		logger.Error("webhookRegistrarVersionHandler: error al upsert versión: %v", err)
		responderError(w, http.StatusInternalServerError, "error al registrar versión desde CI/CD")
		return
	}

	logger.Info("webhookRegistrarVersionHandler: versión %s registrada exitosamente desde CI/CD (ID: %d)", body.Version, id)
	responderJSON(w, http.StatusOK, map[string]any{
		"mensaje": "versión registrada exitosamente desde CI/CD",
		"id":      id,
		"version": body.Version,
		"estado":  "borrador",
	})
}
