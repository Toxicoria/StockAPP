package api

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"

	"stock-operations/db"
)

// La clave fiscal de ARCA (y cualquier otro secreto del negocio) se guarda
// cifrada, nunca en texto plano — es la credencial de acceso al portal
// impositivo real del negocio. AES-256-GCM con la clave derivada por SHA-256
// de CONFIG_SECRET_KEY: así cualquier passphrase sirve como clave de 32
// bytes, sin exigirle un formato particular (mismo patrón que JWT_SECRET).
func claveCifrado() [32]byte {
	return sha256.Sum256([]byte(db.Env("CONFIG_SECRET_KEY", "clave_local_dev_no_usar_en_produccion")))
}

// cifrar devuelve nonce+ciphertext listos para guardar en una columna BYTEA.
func cifrar(texto string) ([]byte, error) {
	clave := claveCifrado()
	bloque, err := aes.NewCipher(clave[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(bloque)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, []byte(texto), nil), nil
}

// descifrar es de uso interno exclusivo (por ejemplo, el futuro cliente WSAA
// de ARCA) — nunca se expone el resultado por la API.
func descifrar(datos []byte) (string, error) {
	clave := claveCifrado()
	bloque, err := aes.NewCipher(clave[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(bloque)
	if err != nil {
		return "", err
	}
	tamanoNonce := gcm.NonceSize()
	if len(datos) < tamanoNonce {
		return "", errors.New("datos cifrados inválidos")
	}
	nonce, texto := datos[:tamanoNonce], datos[tamanoNonce:]
	claro, err := gcm.Open(nil, nonce, texto, nil)
	if err != nil {
		return "", err
	}
	return string(claro), nil
}
