package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"stock-operations/db"
	"stock-operations/pkg/logger"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type negocioAdminItem struct {
	IDNegocio      int    `json:"id_negocio"`
	NombreNegocio  string `json:"nombre_negocio"`
	Direccion      string `json:"direccion"`
	CUIT           string `json:"cuit"`
	NombreDueno    string `json:"nombre_dueno"`
	Telefono       string `json:"telefono"`
	EmailNegocio   string `json:"email_negocio"`
	FechaAlta      string `json:"fecha_alta"`
	IDDueno        int    `json:"id_dueno"`
	NombreUsuario  string `json:"nombre_usuario"`
	Usuario        string `json:"usuario"`
	EmailUsuario   string `json:"email_usuario"`
	PerfilCompleto  bool   `json:"perfil_completo"`
	MaxDispositivos int    `json:"max_dispositivos"`
	TSAuthKey       string `json:"ts_auth_key"`
}

type reqAdminLogin struct {
	Usuario  string `json:"usuario"`
	Password string `json:"password"`
}

type respAdminLogin struct {
	IDAdmin int    `json:"id_admin"`
	Usuario string `json:"usuario"`
	Nombre  string `json:"nombre"`
	Email   string `json:"email"`
}

// adminLoginHandler autentica credenciales de SuperAdmin contra la tabla super_admins.
// POST /api/admin/login
func adminLoginHandler(w http.ResponseWriter, r *http.Request) {
	var body reqAdminLogin
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		responderError(w, http.StatusBadRequest, "cuerpo de solicitud inválido")
		return
	}

	usr := strings.TrimSpace(body.Usuario)
	pass := strings.TrimSpace(body.Password)

	if usr == "" || pass == "" {
		responderError(w, http.StatusBadRequest, "el usuario y la contraseña son obligatorios")
		return
	}

	var idAdmin int
	var usuarioDB, emailDB, hashDB, nombreDB string
	err := db.DB.QueryRow(`
		SELECT id_admin, usuario, COALESCE(email, ''), password_hash, nombre
		  FROM super_admins
		 WHERE LOWER(usuario) = LOWER($1)
	`, usr).Scan(&idAdmin, &usuarioDB, &emailDB, &hashDB, &nombreDB)

	if err != nil {
		logger.Warn("Intento fallido de login SuperAdmin (usuario no encontrado): %s", usr)
		responderError(w, http.StatusUnauthorized, "usuario o contraseña incorrectos")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashDB), []byte(pass)); err != nil {
		logger.Warn("Intento fallido de login SuperAdmin (clave incorrecta): %s", usr)
		responderError(w, http.StatusUnauthorized, "usuario o contraseña incorrectos")
		return
	}

	responderJSON(w, http.StatusOK, respAdminLogin{
		IDAdmin: idAdmin,
		Usuario: usuarioDB,
		Nombre:  nombreDB,
		Email:   emailDB,
	})
}

// adminListarNegociosHandler devuelve todos los negocios registrados en la plataforma.
// GET /api/admin/negocios
func adminListarNegociosHandler(w http.ResponseWriter, r *http.Request) {
	filas, err := db.DB.Query(`
		SELECT n.id_negocio,
		       n.nombre_negocio,
		       COALESCE(n.direccion, ''),
		       COALESCE(n.cuit, ''),
		       COALESCE(n.nombre_dueno, ''),
		       COALESCE(n.telefono, ''),
		       COALESCE(n.email_negocio, ''),
		       to_char(n.fecha_alta, 'DD/MM/YYYY HH24:MI'),
		       COALESCE(u.id_usuario, 0),
		       COALESCE(u.nombre, ''),
		       COALESCE(u.usuario, ''),
		       COALESCE(u.email, ''),
		       COALESCE(n.max_dispositivos, 4),
		       COALESCE(n.ts_auth_key, '')
		  FROM negocios n
		  LEFT JOIN usuarios u ON u.id_negocio = n.id_negocio AND u.rol = 'dueño'
		 ORDER BY n.id_negocio DESC
	`)
	if err != nil {
		logger.Error("admin: error consultando negocios: %v", err)
		responderError(w, http.StatusInternalServerError, "error consultando negocios")
		return
	}
	defer filas.Close()

	negocios := []negocioAdminItem{}
	for filas.Next() {
		var item negocioAdminItem
		if err := filas.Scan(
			&item.IDNegocio,
			&item.NombreNegocio,
			&item.Direccion,
			&item.CUIT,
			&item.NombreDueno,
			&item.Telefono,
			&item.EmailNegocio,
			&item.FechaAlta,
			&item.IDDueno,
			&item.NombreUsuario,
			&item.Usuario,
			&item.EmailUsuario,
			&item.MaxDispositivos,
			&item.TSAuthKey,
		); err != nil {
			logger.Error("admin: error leyendo fila de negocio: %v", err)
			responderError(w, http.StatusInternalServerError, "error leyendo negocios")
			return
		}

		// Consideramos perfil completo si completó los datos esenciales del negocio (fantasía, dueño, dirección y teléfono)
		item.PerfilCompleto = (strings.TrimSpace(item.NombreNegocio) != "" &&
			!strings.HasPrefix(item.NombreNegocio, "Negocio de ") &&
			strings.TrimSpace(item.NombreDueno) != "" &&
			strings.TrimSpace(item.Direccion) != "" &&
			strings.TrimSpace(item.Telefono) != "")
		negocios = append(negocios, item)
	}

	responderJSON(w, http.StatusOK, negocios)
}

type cuerpoAdminCrearNegocio struct {
	NombreNegocio string `json:"nombre_negocio"`
	Usuario       string `json:"usuario"`
	Email         string `json:"email"`
	Password      string `json:"password"`
	NombreDueno   string `json:"nombre_dueno"`
}

// adminCrearNegocioClienteHandler crea un nuevo negocio y su usuario dueño con datos mínimos (usuario + contraseña + email).
// POST /api/admin/negocios
func adminCrearNegocioClienteHandler(w http.ResponseWriter, r *http.Request) {
	var cuerpo cuerpoAdminCrearNegocio
	if err := json.NewDecoder(r.Body).Decode(&cuerpo); err != nil {
		responderError(w, http.StatusBadRequest, "cuerpo JSON inválido")
		return
	}

	cuerpo.Usuario = strings.TrimSpace(cuerpo.Usuario)
	cuerpo.Email = strings.TrimSpace(cuerpo.Email)
	cuerpo.NombreNegocio = strings.TrimSpace(cuerpo.NombreNegocio)
	cuerpo.NombreDueno = strings.TrimSpace(cuerpo.NombreDueno)

	if (cuerpo.Usuario == "" && cuerpo.Email == "") || cuerpo.Password == "" {
		responderError(w, http.StatusBadRequest, "el usuario/email y la contraseña son obligatorios")
		return
	}
	if cuerpo.Usuario == "" {
		cuerpo.Usuario = strings.Split(cuerpo.Email, "@")[0]
	}
	if len(cuerpo.Password) < 4 {
		responderError(w, http.StatusBadRequest, "la contraseña debe tener al menos 4 caracteres")
		return
	}

	if cuerpo.NombreNegocio == "" {
		cuerpo.NombreNegocio = "Negocio de " + cuerpo.Usuario
	}
	if cuerpo.NombreDueno == "" {
		cuerpo.NombreDueno = cuerpo.Usuario
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cuerpo.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("admin: error generando hash: %v", err)
		responderError(w, http.StatusInternalServerError, "error procesando contraseña")
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		logger.Error("admin: error abriendo transacción: %v", err)
		responderError(w, http.StatusInternalServerError, "error interno")
		return
	}
	defer tx.Rollback()

	var idNegocio int
	err = tx.QueryRow(`
		INSERT INTO negocios (nombre_negocio, nombre_dueno, email_negocio)
		VALUES ($1, $2, NULLIF($3, ''))
		RETURNING id_negocio
	`, cuerpo.NombreNegocio, cuerpo.NombreDueno, cuerpo.Email).Scan(&idNegocio)
	if err != nil {
		logger.Error("admin: error creando negocio: %v", err)
		responderError(w, http.StatusInternalServerError, "no se pudo crear el negocio")
		return
	}

	var idUsuario int
	permisosCompleto := `["vender", "stock", "ventas", "resumen", "productos", "precios", "facturas", "usuarios", "configuracion"]`

	err = tx.QueryRow(`
		INSERT INTO usuarios (id_negocio, nombre, usuario, email, password_hash, rol, permisos)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5, 'dueño', $6::jsonb)
		RETURNING id_usuario
	`, idNegocio, cuerpo.NombreDueno, cuerpo.Usuario, cuerpo.Email, string(hash), permisosCompleto).Scan(&idUsuario)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			responderError(w, http.StatusConflict, "el nombre de usuario o correo ya está registrado")
			return
		}
		logger.Error("admin: error creando usuario dueño: %v", err)
		responderError(w, http.StatusInternalServerError, "no se pudo crear el usuario dueño")
		return
	}

	if err := tx.Commit(); err != nil {
		logger.Error("admin: error confirmando transacción: %v", err)
		responderError(w, http.StatusInternalServerError, "error guardando el negocio")
		return
	}

	logger.Info("admin: creado nuevo negocio #%d (%s) con dueño usuario=%s", idNegocio, cuerpo.NombreNegocio, cuerpo.Usuario)
	responderJSON(w, http.StatusCreated, map[string]any{
		"id_negocio":     idNegocio,
		"id_usuario":     idUsuario,
		"nombre_negocio": cuerpo.NombreNegocio,
		"usuario":        cuerpo.Usuario,
		"nombre_dueno":   cuerpo.NombreDueno,
	})
}

type cuerpoAdminCambiarPassword struct {
	Password string `json:"password"`
}

// adminCambiarPasswordHandler permite al admin resetear la clave de cualquier usuario.
// PUT /api/admin/usuarios/{id_usuario}/password
func adminCambiarPasswordHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id_usuario")
	idUsuario, err := strconv.Atoi(idStr)
	if err != nil || idUsuario <= 0 {
		responderError(w, http.StatusBadRequest, "ID de usuario inválido")
		return
	}

	var cuerpo cuerpoAdminCambiarPassword
	if err := json.NewDecoder(r.Body).Decode(&cuerpo); err != nil || cuerpo.Password == "" {
		responderError(w, http.StatusBadRequest, "contraseña inválida")
		return
	}

	if len(cuerpo.Password) < 4 {
		responderError(w, http.StatusBadRequest, "la contraseña debe tener al menos 4 caracteres")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cuerpo.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("admin: error hasheando clave: %v", err)
		responderError(w, http.StatusInternalServerError, "error cambiando contraseña")
		return
	}

	res, err := db.DB.Exec(`UPDATE usuarios SET password_hash = $1 WHERE id_usuario = $2`, string(hash), idUsuario)
	if err != nil {
		logger.Error("admin: error actualizando clave de usuario %d: %v", idUsuario, err)
		responderError(w, http.StatusInternalServerError, "no se pudo actualizar la contraseña")
		return
	}

	afectadas, _ := res.RowsAffected()
	if afectadas == 0 {
		responderError(w, http.StatusNotFound, "usuario no encontrado")
		return
	}

	responderJSON(w, http.StatusOK, map[string]any{
		"mensaje": "contraseña actualizada correctamente",
	})
}

type reqAdminEditarNegocio struct {
	NombreNegocio string `json:"nombre_negocio"`
	Direccion     string `json:"direccion"`
	CUIT          string `json:"cuit"`
	Telefono      string `json:"telefono"`
	EmailNegocio  string `json:"email_negocio"`
	NombreDueno   string `json:"nombre_dueno"`
	Usuario       string `json:"usuario"`
	EmailUsuario  string `json:"email_usuario"`
	Password      string `json:"password"`
}

// adminEditarNegocioHandler actualiza los datos comerciales y de dueño de un negocio.
// PUT /api/admin/negocios/{id_negocio}
func adminEditarNegocioHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id_negocio")
	idNegocio, err := strconv.Atoi(idStr)
	if err != nil || idNegocio <= 0 {
		responderError(w, http.StatusBadRequest, "ID de negocio inválido")
		return
	}

	var req reqAdminEditarNegocio
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responderError(w, http.StatusBadRequest, "cuerpo de solicitud inválido")
		return
	}

	req.NombreNegocio = strings.TrimSpace(req.NombreNegocio)
	req.Usuario = strings.TrimSpace(req.Usuario)
	req.EmailUsuario = strings.TrimSpace(req.EmailUsuario)
	req.NombreDueno = strings.TrimSpace(req.NombreDueno)
	req.Direccion = strings.TrimSpace(req.Direccion)
	req.CUIT = strings.TrimSpace(req.CUIT)
	req.Telefono = strings.TrimSpace(req.Telefono)
	req.EmailNegocio = strings.TrimSpace(req.EmailNegocio)
	req.Password = strings.TrimSpace(req.Password)

	tx, err := db.DB.Begin()
	if err != nil {
		responderError(w, http.StatusInternalServerError, "error iniciando transacción")
		return
	}
	defer tx.Rollback()

	// Actualizar datos del negocio
	resNeg, err := tx.Exec(`
		UPDATE negocios
		   SET nombre_negocio = COALESCE(NULLIF($1, ''), nombre_negocio),
		       direccion = $2,
		       cuit = $3,
		       telefono = $4,
		       email_negocio = $5,
		       nombre_dueno = $6
		 WHERE id_negocio = $7
	`, req.NombreNegocio, req.Direccion, req.CUIT, req.Telefono, req.EmailNegocio, req.NombreDueno, idNegocio)

	if err != nil {
		logger.Error("admin: error actualizando negocio %d: %v", idNegocio, err)
		responderError(w, http.StatusInternalServerError, "error actualizando negocio")
		return
	}

	afectadas, _ := resNeg.RowsAffected()
	if afectadas == 0 {
		responderError(w, http.StatusNotFound, "negocio no encontrado")
		return
	}

	// Actualizar usuario dueño asociado al negocio
	if req.Usuario != "" || req.EmailUsuario != "" || req.NombreDueno != "" {
		_, err := tx.Exec(`
			UPDATE usuarios
			   SET usuario = COALESCE(NULLIF($1, ''), usuario),
			       email = NULLIF($2, ''),
			       nombre = COALESCE(NULLIF($3, ''), nombre)
			 WHERE id_negocio = $4 AND rol = 'dueño'
		`, req.Usuario, req.EmailUsuario, req.NombreDueno, idNegocio)

		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
				responderError(w, http.StatusConflict, "el nombre de usuario ya está registrado en este negocio")
				return
			}
			logger.Error("admin: error actualizando usuario dueño del negocio %d: %v", idNegocio, err)
			responderError(w, http.StatusInternalServerError, "error actualizando usuario dueño")
			return
		}
	}

	// Cambiar contraseña del dueño si se especificó una nueva
	if req.Password != "" {
		if len(req.Password) < 4 {
			responderError(w, http.StatusBadRequest, "la contraseña debe tener al menos 4 caracteres")
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			responderError(w, http.StatusInternalServerError, "error procesando contraseña")
			return
		}

		_, err = tx.Exec(`
			UPDATE usuarios
			   SET password_hash = $1
			 WHERE id_negocio = $2 AND rol = 'dueño'
		`, string(hash), idNegocio)

		if err != nil {
			logger.Error("admin: error cambiando clave de dueño del negocio %d: %v", idNegocio, err)
			responderError(w, http.StatusInternalServerError, "error cambiando contraseña")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		responderError(w, http.StatusInternalServerError, "error guardando cambios")
		return
	}

	responderJSON(w, http.StatusOK, map[string]any{
		"mensaje": "negocio y usuario actualizados correctamente",
	})
}
