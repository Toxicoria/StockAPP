package api

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"stock-operations/db"
	"stock-operations/pkg/logger"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var reUsername = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// usuario representa un usuario del negocio, tal como se lo devuelve al frontend.
// La contraseña nunca se incluye en la respuesta.
type usuario struct {
	ID        int             `json:"id_usuario"`
	Nombre    string          `json:"nombre"`
	Usuario   *string         `json:"usuario"`
	Email     *string         `json:"email"`
	Rol       string          `json:"rol"`
	Permisos  json.RawMessage `json:"permisos"`
	FechaAlta time.Time       `json:"fecha_alta"`
}

// listarUsuariosHandler devuelve todos los usuarios del negocio autenticado.
// Solo dueño. GET /api/usuarios
func listarUsuariosHandler(w http.ResponseWriter, r *http.Request) {
	if rolDe(r) != "dueño" {
		responderError(w, http.StatusForbidden, "solo el dueño puede gestionar usuarios")
		return
	}

	filas, err := db.DB.Query(
		`SELECT id_usuario, nombre, usuario, email, rol, COALESCE(permisos, '["vender", "stock"]'::jsonb), fecha_alta
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
		if err := filas.Scan(&u.ID, &u.Nombre, &u.Usuario, &u.Email, &u.Rol, &u.Permisos, &u.FechaAlta); err != nil {
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
	Nombre   string          `json:"nombre"`
	Usuario  *string         `json:"usuario"`
	Email    *string         `json:"email"`
	Password string          `json:"password"`
	Permisos json.RawMessage `json:"permisos"`
}

// crearUsuarioHandler crea un nuevo cajero para el negocio. El dueño solo
// puede crear usuarios con rol 'cajero'. Solo dueño.
// POST /api/usuarios
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
	if cuerpo.Nombre == "" || cuerpo.Password == "" {
		responderError(w, http.StatusBadRequest, "el nombre del empleado y la contraseña son obligatorios")
		return
	}
	if len(cuerpo.Password) < 8 {
		responderError(w, http.StatusBadRequest, "la contraseña debe tener al menos 8 caracteres")
		return
	}

	// Validar usuario (username)
	if cuerpo.Usuario != nil && *cuerpo.Usuario != "" {
		if !reUsername.MatchString(*cuerpo.Usuario) {
			responderError(w, http.StatusBadRequest, "el nombre de usuario no puede contener espacios ni caracteres especiales")
			return
		}
	} else {
		cuerpo.Usuario = nil
	}

	// Limpiar email vacío
	if cuerpo.Email != nil && *cuerpo.Email == "" {
		cuerpo.Email = nil
	}

	// Permisos por defecto
	if len(cuerpo.Permisos) == 0 {
		cuerpo.Permisos = json.RawMessage(`["vender", "stock"]`)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cuerpo.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("usuarios: error hasheando password: %v", err)
		responderError(w, http.StatusInternalServerError, "error interno")
		return
	}

	var nuevo usuario
	err = db.DB.QueryRow(
		`INSERT INTO usuarios (id_negocio, nombre, usuario, email, password_hash, rol, permisos)
		 VALUES ($1, $2, $3, $4, $5, 'cajero', $6::jsonb)
		 RETURNING id_usuario, nombre, usuario, email, rol, permisos, fecha_alta`,
		negocioDe(r), cuerpo.Nombre, cuerpo.Usuario, cuerpo.Email, string(hash), string(cuerpo.Permisos),
	).Scan(&nuevo.ID, &nuevo.Nombre, &nuevo.Usuario, &nuevo.Email, &nuevo.Rol, &nuevo.Permisos, &nuevo.FechaAlta)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			responderError(w, http.StatusConflict, "ya existe un usuario con ese nombre de usuario en tu negocio")
			return
		}
		logger.Error("usuarios: error creando %q: %v", cuerpo.Nombre, err)
		responderError(w, http.StatusInternalServerError, "no se pudo crear el usuario")
		return
	}

	logger.Info("usuarios: nuevo cajero #%d (%s) en negocio #%d", nuevo.ID, nuevo.Nombre, negocioDe(r))
	responderJSON(w, http.StatusCreated, nuevo)
}

// cuerpoEditarUsuario son los datos editables de un usuario.
type cuerpoEditarUsuario struct {
	Nombre   string          `json:"nombre"`
	Usuario  *string         `json:"usuario"`
	Email    *string         `json:"email"`
	Permisos json.RawMessage `json:"permisos"`
}

// editarUsuarioHandler modifica el nombre, usuario, email y/o permisos de un usuario.
// PUT /api/usuarios/{id_usuario}
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
	if cuerpo.Nombre == "" {
		responderError(w, http.StatusBadRequest, "el nombre es obligatorio")
		return
	}

	if cuerpo.Usuario != nil && *cuerpo.Usuario != "" {
		if !reUsername.MatchString(*cuerpo.Usuario) {
			responderError(w, http.StatusBadRequest, "el nombre de usuario no puede contener espacios ni caracteres especiales")
			return
		}
	} else {
		cuerpo.Usuario = nil
	}

	if cuerpo.Email != nil && *cuerpo.Email == "" {
		cuerpo.Email = nil
	}
	if len(cuerpo.Permisos) == 0 {
		cuerpo.Permisos = json.RawMessage(`["vender", "stock"]`)
	}

	// No puede editarse a sí mismo por esta vía
	if idUsuario == usuarioDe(r) {
		responderError(w, http.StatusForbidden, "no podés editar tu propia cuenta desde acá")
		return
	}

	var editado usuario
	err = db.DB.QueryRow(
		`UPDATE usuarios
		    SET nombre = $1, usuario = $2, email = $3, permisos = $4::jsonb
		  WHERE id_usuario = $5 AND id_negocio = $6
		  RETURNING id_usuario, nombre, usuario, email, rol, permisos, fecha_alta`,
		cuerpo.Nombre, cuerpo.Usuario, cuerpo.Email, string(cuerpo.Permisos), idUsuario, negocioDe(r),
	).Scan(&editado.ID, &editado.Nombre, &editado.Usuario, &editado.Email, &editado.Rol, &editado.Permisos, &editado.FechaAlta)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			responderError(w, http.StatusConflict, "ya existe un usuario con ese nombre de usuario en tu negocio")
			return
		}
		responderError(w, http.StatusNotFound, "usuario no encontrado")
		return
	}

	responderJSON(w, http.StatusOK, editado)
}

type cuerpoCambiarPassword struct {
	Password string `json:"password"`
}

// cambiarPasswordUsuarioHandler permite al dueño cambiar la contraseña de un usuario de su negocio.
// PUT /api/usuarios/{id_usuario}/password
func cambiarPasswordUsuarioHandler(w http.ResponseWriter, r *http.Request) {
	if rolDe(r) != "dueño" {
		responderError(w, http.StatusForbidden, "solo el dueño puede cambiar contraseñas")
		return
	}

	idUsuario, err := strconv.Atoi(r.PathValue("id_usuario"))
	if err != nil {
		responderError(w, http.StatusBadRequest, "id de usuario inválido")
		return
	}

	var cuerpo cuerpoCambiarPassword
	if err := json.NewDecoder(r.Body).Decode(&cuerpo); err != nil {
		responderError(w, http.StatusBadRequest, "cuerpo JSON inválido")
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

	resultado, err := db.DB.Exec(
		`UPDATE usuarios SET password_hash = $1 WHERE id_usuario = $2 AND id_negocio = $3`,
		string(hash), idUsuario, negocioDe(r),
	)
	if err != nil {
		logger.Error("usuarios: error cambiando password usuario #%d: %v", idUsuario, err)
		responderError(w, http.StatusInternalServerError, "no se pudo actualizar la contraseña")
		return
	}
	filas, _ := resultado.RowsAffected()
	if filas == 0 {
		responderError(w, http.StatusNotFound, "usuario no encontrado")
		return
	}

	responderJSON(w, http.StatusOK, map[string]string{"mensaje": "contraseña actualizada correctamente"})
}

// eliminarUsuarioHandler borra un usuario del negocio. Solo dueño.
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
