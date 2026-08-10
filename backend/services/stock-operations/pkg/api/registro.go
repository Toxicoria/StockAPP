package api

import (
	"encoding/json"
	"net/http"

	"stock-operations/db"
	"stock-operations/pkg/logger"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// cuerpoRegistro contiene los datos del onboarding inicial:
// crea el primer negocio y su usuario admin de una sola vez.
type cuerpoRegistro struct {
	// Negocio
	NombreNegocio string `json:"nombre_negocio"`
	Direccion     string `json:"direccion"`
	Cuit          string `json:"cuit"`
	NombreDueno   string `json:"nombre_dueno"`
	Telefono      string `json:"telefono"`
	EmailNegocio  string `json:"email_negocio"`
	// Admin inicial
	NombreAdmin string `json:"nombre_admin"`
	EmailAdmin  string `json:"email_admin"`
	Password    string `json:"password"`
}

// registroHandler crea un negocio nuevo junto con su primer usuario admin.
// Solo funciona si NO existe ningún negocio todavía (primer arranque del servidor).
// POST /api/registro  — ruta pública, sin autenticación.
func registroHandler(w http.ResponseWriter, r *http.Request) {
	var cuerpo cuerpoRegistro
	if err := json.NewDecoder(r.Body).Decode(&cuerpo); err != nil {
		responderError(w, http.StatusBadRequest, "cuerpo JSON inválido")
		return
	}

	// Validaciones básicas
	if cuerpo.NombreNegocio == "" || cuerpo.EmailAdmin == "" || cuerpo.Password == "" || cuerpo.NombreAdmin == "" {
		responderError(w, http.StatusBadRequest, "nombre_negocio, nombre_admin, email_admin y password son obligatorios")
		return
	}
	if len(cuerpo.Password) < 8 {
		responderError(w, http.StatusBadRequest, "la contraseña debe tener al menos 8 caracteres")
		return
	}

	// Verificar que todavía no haya negocios (evita que cualquiera llame a este endpoint en producción)
	var cantidad int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM negocios`).Scan(&cantidad); err != nil {
		logger.Error("registro: error chequeando negocios existentes: %v", err)
		responderError(w, http.StatusInternalServerError, "error al verificar el estado del servidor")
		return
	}
	if cantidad > 0 {
		responderError(w, http.StatusConflict, "el servidor ya tiene un negocio registrado")
		return
	}

	// Hash de la contraseña
	hash, err := bcrypt.GenerateFromPassword([]byte(cuerpo.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("registro: error hasheando password: %v", err)
		responderError(w, http.StatusInternalServerError, "error interno")
		return
	}

	// Insertar negocio y usuario admin en una transacción
	tx, err := db.DB.Begin()
	if err != nil {
		logger.Error("registro: error abriendo transacción: %v", err)
		responderError(w, http.StatusInternalServerError, "error interno")
		return
	}
	defer tx.Rollback()

	var idNegocio int
	err = tx.QueryRow(
		`INSERT INTO negocios (nombre_negocio, direccion, cuit, nombre_dueno, telefono, email_negocio)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id_negocio`,
		cuerpo.NombreNegocio,
		cuerpo.Direccion,
		cuerpo.Cuit,
		cuerpo.NombreDueno,
		cuerpo.Telefono,
		cuerpo.EmailNegocio,
	).Scan(&idNegocio)
	if err != nil {
		logger.Error("registro: error creando negocio: %v", err)
		responderError(w, http.StatusInternalServerError, "no se pudo crear el negocio")
		return
	}

	_, err = tx.Exec(
		`INSERT INTO usuarios (id_negocio, nombre, email, password_hash, rol)
		 VALUES ($1, $2, $3, $4, 'admin')`,
		idNegocio, cuerpo.NombreAdmin, cuerpo.EmailAdmin, string(hash),
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			responderError(w, http.StatusConflict, "ya existe un usuario con ese email")
			return
		}
		logger.Error("registro: error creando usuario admin: %v", err)
		responderError(w, http.StatusInternalServerError, "no se pudo crear el usuario admin")
		return
	}

	if err := tx.Commit(); err != nil {
		logger.Error("registro: error confirmando transacción: %v", err)
		responderError(w, http.StatusInternalServerError, "error interno")
		return
	}

	logger.Info("registro: nuevo negocio #%d (%s) con admin %s", idNegocio, cuerpo.NombreNegocio, cuerpo.EmailAdmin)
	responderJSON(w, http.StatusCreated, map[string]any{
		"id_negocio":     idNegocio,
		"nombre_negocio": cuerpo.NombreNegocio,
		"email_admin":    cuerpo.EmailAdmin,
	})
}

// negocioRegistradoHandler informa si ya existe al menos un negocio en la base.
// El frontend lo usa al arrancar para saber si tiene que mostrar el wizard.
// GET /api/registro  — ruta pública, sin autenticación.
func negocioRegistradoHandler(w http.ResponseWriter, r *http.Request) {
	var cantidad int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM negocios`).Scan(&cantidad); err != nil {
		logger.Error("registro: error verificando negocios: %v", err)
		responderError(w, http.StatusInternalServerError, "error verificando el estado")
		return
	}
	responderJSON(w, http.StatusOK, map[string]bool{"registrado": cantidad > 0})
}
