package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"stock-operations/db"
	"stock-operations/pkg/logger"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// usuario representa un usuario del negocio, tal como se lo devuelve al frontend.
// La contraseña nunca se incluye en la respuesta.
type usuario struct {
	ID        int       `json:"id_usuario"`
	Nombre    string    `json:"nombre"`
	Email     string    `json:"email"`
	Rol       string    `json:"rol"`
	FechaAlta time.Time `json:"fecha_alta"`
}

// listarUsuariosHandler devuelve todos los usuarios del negocio autenticado.
// Solo dueño. GET /api/usuarios
func listarUsuariosHandler(w http.ResponseWriter, r *http.Request) {
	if rolDe(r) != "dueño" {
		responderError(w, http.StatusForbidden, "solo el dueño puede gestionar usuarios")
		return
	}

	filas, err := db.DB.Query(
		`SELECT id_usuario, nombre, email, rol, fecha_alta
		   FROM usuarios
		  WHERE id_negocio = $1
		  ORDER BY fecha_alta`, negocioDe(r))
	if err != nil {
		logger.Error("usuarios: error listando: %v", err)
		responderError(w, http.StatusInternalServerError, "error consultando los usuarios")
		return
	}
	defer filas.Close()

	usuarios := []usuario{}
	for filas.Next() {
		var u usuario
		if err := filas.Scan(&u.ID, &u.Nombre, &u.Email, &u.Rol, &u.FechaAlta); err != nil {
			logger.Error("usuarios: error leyendo fila: %v", err)
			responderError(w, http.StatusInternalServerError, "error leyendo los usuarios")
			return
		}
		usuarios = append(usuarios, u)
	}
	responderJSON(w, http.StatusOK, usuarios)
}

// cuerpoCrearUsuario son los datos para crear un nuevo empleado.
type cuerpoCrearUsuario struct {
	Nombre   string `json:"nombre"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// crearUsuarioHandler crea un nuevo cajero para el negocio. El dueño solo
// puede crear usuarios con rol 'cajero'. Solo dueño.
// POST /api/usuarios  body: {"nombre": "...", "email": "...", "password": "..."}
func crearUsuarioHandler(w http.ResponseWriter, r *http.Request) {
	if rolDe(r) != "dueño" {
		responderError(w, http.StatusForbidden, "solo el dueño puede crear usuarios")
		return
	}

	var cuerpo cuerpoCrearUsuario
	if err := json.NewDecoder(r.Body).Decode(&cuerpo); err != nil {
		responderError(w, http.StatusBadRequest, "cuerpo JSON inválido")
		return
	}
	if cuerpo.Nombre == "" || cuerpo.Email == "" || cuerpo.Password == "" {
		responderError(w, http.StatusBadRequest, "nombre, email y password son obligatorios")
		return
	}
	if len(cuerpo.Password) < 8 {
		responderError(w, http.StatusBadRequest, "la contraseña debe tener al menos 8 caracteres")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cuerpo.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("usuarios: error hasheando password: %v", err)
		responderError(w, http.StatusInternalServerError, "error interno")
		return
	}

	var nuevo usuario
	err = db.DB.QueryRow(
		`INSERT INTO usuarios (id_negocio, nombre, email, password_hash, rol)
		 VALUES ($1, $2, $3, $4, 'cajero')
		 RETURNING id_usuario, nombre, email, rol, fecha_alta`,
		negocioDe(r), cuerpo.Nombre, cuerpo.Email, string(hash),
	).Scan(&nuevo.ID, &nuevo.Nombre, &nuevo.Email, &nuevo.Rol, &nuevo.FechaAlta)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			responderError(w, http.StatusConflict, "ya existe un usuario con ese email")
			return
		}
		logger.Error("usuarios: error creando %q: %v", cuerpo.Email, err)
		responderError(w, http.StatusInternalServerError, "no se pudo crear el usuario")
		return
	}

	logger.Info("usuarios: nuevo cajero #%d (%s) en negocio #%d", nuevo.ID, nuevo.Email, negocioDe(r))
	responderJSON(w, http.StatusCreated, nuevo)
}

// cuerpoEditarUsuario son los datos editables de un usuario (sin contraseña).
type cuerpoEditarUsuario struct {
	Nombre string `json:"nombre"`
	Email  string `json:"email"`
}

// editarUsuarioHandler modifica el nombre y/o email de un usuario del negocio.
// Solo dueño. No permite editar la contraseña (eso sería otro endpoint).
// PUT /api/usuarios/{id_usuario}  body: {"nombre": "...", "email": "..."}
func editarUsuarioHandler(w http.ResponseWriter, r *http.Request) {
	if rolDe(r) != "dueño" {
		responderError(w, http.StatusForbidden, "solo el dueño puede editar usuarios")
		return
	}

	idUsuario, err := strconv.Atoi(r.PathValue("id_usuario"))
	if err != nil {
		responderError(w, http.StatusBadRequest, "id de usuario inválido")
		return
	}

	var cuerpo cuerpoEditarUsuario
	if err := json.NewDecoder(r.Body).Decode(&cuerpo); err != nil {
		responderError(w, http.StatusBadRequest, "cuerpo JSON inválido")
		return
	}
	if cuerpo.Nombre == "" || cuerpo.Email == "" {
		responderError(w, http.StatusBadRequest, "nombre y email son obligatorios")
		return
	}

	// No puede editarse a sí mismo por esta vía (evita que el dueño se
	// cambie el email y se quede sin acceso por error).
	if idUsuario == usuarioDe(r) {
		responderError(w, http.StatusForbidden, "no podés editar tu propia cuenta desde acá")
		return
	}

	var editado usuario
	err = db.DB.QueryRow(
		`UPDATE usuarios
		    SET nombre = $1, email = $2
		  WHERE id_usuario = $3 AND id_negocio = $4
		  RETURNING id_usuario, nombre, email, rol, fecha_alta`,
		cuerpo.Nombre, cuerpo.Email, idUsuario, negocioDe(r),
	).Scan(&editado.ID, &editado.Nombre, &editado.Email, &editado.Rol, &editado.FechaAlta)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			responderError(w, http.StatusConflict, "ya existe un usuario con ese email")
			return
		}
		responderError(w, http.StatusNotFound, "usuario no encontrado")
		return
	}

	responderJSON(w, http.StatusOK, editado)
}

// eliminarUsuarioHandler borra un usuario del negocio. Solo dueño.
// No permite eliminarse a sí mismo (el negocio quedaría sin dueño).
// DELETE /api/usuarios/{id_usuario}
func eliminarUsuarioHandler(w http.ResponseWriter, r *http.Request) {
	if rolDe(r) != "dueño" {
		responderError(w, http.StatusForbidden, "solo el dueño puede eliminar usuarios")
		return
	}

	idUsuario, err := strconv.Atoi(r.PathValue("id_usuario"))
	if err != nil {
		responderError(w, http.StatusBadRequest, "id de usuario inválido")
		return
	}

	if idUsuario == usuarioDe(r) {
		responderError(w, http.StatusForbidden, "no podés eliminar tu propia cuenta")
		return
	}

	resultado, err := db.DB.Exec(
		`DELETE FROM usuarios WHERE id_usuario = $1 AND id_negocio = $2`,
		idUsuario, negocioDe(r))
	if err != nil {
		logger.Error("usuarios: error eliminando #%d: %v", idUsuario, err)
		responderError(w, http.StatusInternalServerError, "no se pudo eliminar el usuario")
		return
	}

	filas, _ := resultado.RowsAffected()
	if filas == 0 {
		responderError(w, http.StatusNotFound, "usuario no encontrado")
		return
	}

	logger.Info("usuarios: eliminado #%d del negocio #%d", idUsuario, negocioDe(r))
	responderJSON(w, http.StatusOK, map[string]string{"mensaje": "usuario eliminado"})
}
