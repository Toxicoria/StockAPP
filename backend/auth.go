package main

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	duracionAccessToken  = 15 * time.Minute
	duracionRefreshToken = 30 * 24 * time.Hour // 30 días, igual que esquema.sql
)

// secretoJWT lee la clave de firma desde el entorno. El default solo sirve
// para desarrollo local; en producción JWT_SECRET es obligatoria.
func secretoJWT() []byte {
	return []byte(getEnv("JWT_SECRET", "secreto_solo_para_desarrollo"))
}

type credencialesLogin struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	Dispositivo string `json:"dispositivo"` // Ej: "PC-Caja1" — opcional
}

type respuestaTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Nombre       string `json:"nombre"`
	Rol          string `json:"rol"`
}

// loginHandler valida email + contraseña contra la tabla usuarios y entrega
// el par de tokens: access (JWT, 15 min) y refresh (opaco, 30 días).
func loginHandler(w http.ResponseWriter, r *http.Request) {
	var cred credencialesLogin
	if err := json.NewDecoder(r.Body).Decode(&cred); err != nil {
		responderError(w, http.StatusBadRequest, "cuerpo JSON inválido")
		return
	}
	if cred.Email == "" || cred.Password == "" {
		responderError(w, http.StatusBadRequest, "email y password son obligatorios")
		return
	}

	var (
		idUsuario, idNegocio int
		nombre, hash, rol    string
	)
	err := DB.QueryRow(
		`SELECT id_usuario, id_negocio, nombre, password_hash, rol
		   FROM usuarios
		  WHERE email = $1`, cred.Email,
	).Scan(&idUsuario, &idNegocio, &nombre, &hash, &rol)

	// Mismo mensaje para "email no existe" y "contraseña incorrecta":
	// no revelar cuáles emails están registrados.
	if errors.Is(err, sql.ErrNoRows) {
		responderError(w, http.StatusUnauthorized, "credenciales inválidas")
		return
	}
	if err != nil {
		LogError("login: error consultando usuarios: %v", err)
		responderError(w, http.StatusInternalServerError, "error interno")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(cred.Password)) != nil {
		responderError(w, http.StatusUnauthorized, "credenciales inválidas")
		return
	}

	tokens, err := emitirTokens(idUsuario, idNegocio, nombre, rol, cred.Dispositivo)
	if err != nil {
		LogError("login: error emitiendo tokens: %v", err)
		responderError(w, http.StatusInternalServerError, "error interno")
		return
	}
	responderJSON(w, http.StatusOK, tokens)
}

type cuerpoRefresh struct {
	RefreshToken string `json:"refresh_token"`
}

// refreshHandler canjea un refresh token vigente por un par nuevo (rotación:
// el token usado queda invalidado siempre, incluso si ya estaba vencido).
func refreshHandler(w http.ResponseWriter, r *http.Request) {
	var cuerpo cuerpoRefresh
	if err := json.NewDecoder(r.Body).Decode(&cuerpo); err != nil || cuerpo.RefreshToken == "" {
		responderError(w, http.StatusBadRequest, "falta refresh_token en el cuerpo")
		return
	}

	var (
		idFila, idUsuario, idNegocio int
		dispositivo, nombre, rol     string
		expira                       time.Time
	)
	err := DB.QueryRow(
		`SELECT rt.id, rt.dispositivo, rt.expira_en, u.id_usuario, u.id_negocio, u.nombre, u.rol
		   FROM refresh_tokens rt
		   JOIN usuarios u ON u.id_usuario = rt.id_usuario
		  WHERE rt.token_hash = $1`, hashDeToken(cuerpo.RefreshToken),
	).Scan(&idFila, &dispositivo, &expira, &idUsuario, &idNegocio, &nombre, &rol)

	if errors.Is(err, sql.ErrNoRows) {
		responderError(w, http.StatusUnauthorized, "refresh token desconocido")
		return
	}
	if err != nil {
		LogError("refresh: error consultando tokens: %v", err)
		responderError(w, http.StatusInternalServerError, "error interno")
		return
	}

	if _, err := DB.Exec(`DELETE FROM refresh_tokens WHERE id = $1`, idFila); err != nil {
		LogError("refresh: error rotando el token: %v", err)
		responderError(w, http.StatusInternalServerError, "error interno")
		return
	}

	if time.Now().After(expira) {
		responderError(w, http.StatusUnauthorized, "la sesión venció, hay que iniciar sesión de nuevo")
		return
	}

	tokens, err := emitirTokens(idUsuario, idNegocio, nombre, rol, dispositivo)
	if err != nil {
		LogError("refresh: error emitiendo tokens: %v", err)
		responderError(w, http.StatusInternalServerError, "error interno")
		return
	}
	responderJSON(w, http.StatusOK, tokens)
}

// logoutHandler invalida el refresh token del dispositivo. El access token
// sigue siendo válido hasta que venza (15 min máximo) — es el costo de que
// sea stateless.
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	var cuerpo cuerpoRefresh
	if err := json.NewDecoder(r.Body).Decode(&cuerpo); err != nil || cuerpo.RefreshToken == "" {
		responderError(w, http.StatusBadRequest, "falta refresh_token en el cuerpo")
		return
	}
	if _, err := DB.Exec(
		`DELETE FROM refresh_tokens WHERE token_hash = $1`, hashDeToken(cuerpo.RefreshToken),
	); err != nil {
		LogError("logout: error borrando el token: %v", err)
		responderError(w, http.StatusInternalServerError, "error interno")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// emitirTokens genera el JWT de acceso y un refresh token nuevo, y guarda
// el SHA-256 del refresh en la DB (el token en crudo solo viaja al cliente).
func emitirTokens(idUsuario, idNegocio int, nombre, rol, dispositivo string) (*respuestaTokens, error) {
	ahora := time.Now()
	claims := jwt.MapClaims{
		"sub":     idUsuario,
		"negocio": idNegocio,
		"nombre":  nombre,
		"rol":     rol,
		"iat":     ahora.Unix(),
		"exp":     ahora.Add(duracionAccessToken).Unix(),
	}
	access, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secretoJWT())
	if err != nil {
		return nil, err
	}

	crudo := make([]byte, 32)
	if _, err := rand.Read(crudo); err != nil {
		return nil, err
	}
	refresh := hex.EncodeToString(crudo)

	_, err = DB.Exec(
		`INSERT INTO refresh_tokens (id_usuario, token_hash, dispositivo, expira_en)
		 VALUES ($1, $2, $3, $4)`,
		idUsuario, hashDeToken(refresh), dispositivo, ahora.Add(duracionRefreshToken),
	)
	if err != nil {
		return nil, err
	}

	return &respuestaTokens{
		AccessToken:  access,
		RefreshToken: refresh,
		Nombre:       nombre,
		Rol:          rol,
	}, nil
}

func hashDeToken(token string) string {
	suma := sha256.Sum256([]byte(token))
	return hex.EncodeToString(suma[:])
}
