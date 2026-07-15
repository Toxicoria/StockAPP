package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
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

// conCORS agrega los headers para peticiones cross-origin. El preflight
// OPTIONS lo responde el sidecar; esto cubre el acceso directo en desarrollo.
func conCORS(siguiente http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		siguiente(w, r)
	}
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

		token, err := jwt.Parse(crudo, func(t *jwt.Token) (any, error) {
			// Verificar el algoritmo evita el ataque clásico de "alg: none"
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("método de firma inesperado: %v", t.Header["alg"])
			}
			return secretoJWT(), nil
		})
		if err != nil || !token.Valid {
			responderError(w, http.StatusUnauthorized, "token inválido o vencido")
			return
		}

		claims, _ := token.Claims.(jwt.MapClaims)
		ctx := context.WithValue(r.Context(), claveClaims, claims)
		siguiente(w, r.WithContext(ctx))
	}
}

// claimsDe recupera los claims que dejó conAuth en el contexto.
func claimsDe(r *http.Request) jwt.MapClaims {
	claims, _ := r.Context().Value(claveClaims).(jwt.MapClaims)
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
