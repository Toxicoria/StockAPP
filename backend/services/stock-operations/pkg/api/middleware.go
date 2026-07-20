package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"stock-operations/db"
)

// Clave propia para el contexto: evita colisiones con otras librerías
// que también guarden valores en el contexto del request.
type claveContexto string

const claveClaims claveContexto = "claims"

// responderJSON serializa cualquier dato como JSON con el código HTTP dado.
func responderJSON(w http.ResponseWriter, codigo int, dato any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(codigo)
	json.NewEncoder(w).Encode(dato)
}

// responderError devuelve un JSON {"error": "..."} — mismo formato para
// todos los errores, así el frontend los maneja de una sola manera.
func responderError(w http.ResponseWriter, codigo int, mensaje string) {
	responderJSON(w, codigo, map[string]string{"error": mensaje})
}

// conCORS agrega los headers para peticiones cross-origin, responde el
// preflight OPTIONS y suma headers de seguridad básicos. Normalmente el que
// da la cara es stock-api, pero esto cubre el acceso directo en desarrollo
// y en K8s (stock-operations sigue alcanzable dentro del cluster).
func conCORS(siguiente http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		// Defensa en profundidad: la API solo devuelve JSON, nunca HTML, pero
		// estos headers evitan que un navegador lo interprete distinto ante
		// una respuesta de error mal tipada o un endpoint mal configurado.
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		siguiente(w, r)
	}
}

// La sesión se valida contra stock-api (introspección): stock-operations no
// conoce el secreto JWT — el dueño de usuarios y sesiones es stock-api.
// El cache de 60 s evita pagar un roundtrip HTTP por cada request (los
// tokens viven 15 minutos, así que la ventana de revocación es corta igual).
var (
	clienteSesiones = &http.Client{Timeout: 5 * time.Second}
	cacheSesiones   sync.Map // token → sesionCacheada
)

type sesionCacheada struct {
	claims map[string]any
	expira time.Time
}

func urlStockAPI() string {
	return db.Env("STOCK_API_URL", "http://localhost:3000")
}

// validarSesion consulta /internal/session de stock-api y devuelve los claims.
// El segundo valor es el código HTTP a devolver si la validación falla.
func validarSesion(token string) (map[string]any, int) {
	if cacheada, ok := cacheSesiones.Load(token); ok {
		sesion := cacheada.(sesionCacheada)
		if time.Now().Before(sesion.expira) {
			return sesion.claims, 0
		}
		cacheSesiones.Delete(token)
	}

	req, err := http.NewRequest(http.MethodGet, urlStockAPI()+"/internal/session", nil)
	if err != nil {
		return nil, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := clienteSesiones.Do(req)
	if err != nil {
		return nil, http.StatusServiceUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, http.StatusUnauthorized
	}
	var claims map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&claims); err != nil {
		return nil, http.StatusServiceUnavailable
	}

	cacheSesiones.Store(token, sesionCacheada{claims: claims, expira: time.Now().Add(60 * time.Second)})
	return claims, 0
}

// conAuth exige un access token válido (header "Authorization: Bearer <token>")
// y deja los claims del usuario en el contexto para que el handler los lea.
func conAuth(siguiente http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		crudo, hayToken := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !hayToken {
			responderError(w, http.StatusUnauthorized, "falta el header Authorization: Bearer <token>")
			return
		}

		claims, codigoError := validarSesion(crudo)
		if claims == nil {
			switch codigoError {
			case http.StatusServiceUnavailable:
				responderError(w, codigoError, "no se pudo validar la sesión — stock-api no responde")
			default:
				responderError(w, http.StatusUnauthorized, "token inválido o vencido")
			}
			return
		}

		ctx := context.WithValue(r.Context(), claveClaims, claims)
		siguiente(w, r.WithContext(ctx))
	}
}

// claimsDe recupera los claims que dejó conAuth en el contexto.
func claimsDe(r *http.Request) map[string]any {
	claims, _ := r.Context().Value(claveClaims).(map[string]any)
	return claims
}

// negocioDe devuelve el id_negocio del usuario autenticado.
// Los números en JSON llegan como float64, por eso la conversión doble.
func negocioDe(r *http.Request) int {
	negocio, _ := claimsDe(r)["negocio"].(float64)
	return int(negocio)
}

// rolDe devuelve el rol ('admin' o 'cajero') del usuario autenticado.
func rolDe(r *http.Request) string {
	rol, _ := claimsDe(r)["rol"].(string)
	return rol
}

// usuarioDe devuelve el id_usuario del usuario autenticado (claim "sub").
func usuarioDe(r *http.Request) int {
	usuario, _ := claimsDe(r)["sub"].(float64)
	return int(usuario)
}
