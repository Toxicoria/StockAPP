//go:build integration

package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"stock-operations/db"

	_ "github.com/lib/pq"
)

func TestBootstrapConcurrenteCreaUnSoloDueno(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL no está configurada")
	}

	testDB, err := sql.Open("postgres", databaseURL)
	if err != nil {
		t.Fatalf("abrir PostgreSQL: %v", err)
	}
	t.Cleanup(func() { testDB.Close() })
	if err := testDB.Ping(); err != nil {
		t.Fatalf("conectar a PostgreSQL: %v", err)
	}

	_, archivoActual, _, _ := runtime.Caller(0)
	esquemaPath := filepath.Join(filepath.Dir(archivoActual), "..", "..", "model", "esquema.sql")
	esquema, err := os.ReadFile(esquemaPath)
	if err != nil {
		t.Fatalf("leer esquema: %v", err)
	}
	if _, err := testDB.Exec(string(esquema)); err != nil {
		t.Fatalf("aplicar esquema: %v", err)
	}
	if _, err := testDB.Exec(`TRUNCATE TABLE negocios CASCADE`); err != nil {
		t.Fatalf("limpiar base de prueba: %v", err)
	}

	dbAnterior := db.DB
	db.DB = testDB
	t.Cleanup(func() { db.DB = dbAnterior })

	servidor := httptest.NewServer(Router())
	t.Cleanup(servidor.Close)

	respInicial, err := http.Get(servidor.URL + "/api/registro")
	if err != nil {
		t.Fatalf("consultar estado inicial: %v", err)
	}
	assertJSONRegistro(t, respInicial, false)

	cuerpo := map[string]string{
		"nombre_negocio": "Almacén de prueba",
		"nombre_dueno":   "Dueña de prueba",
		"nombre_admin":   "Dueña de prueba",
		"email_admin":    "duena@example.com",
		"password":       "password-seguro",
	}
	payload, err := json.Marshal(cuerpo)
	if err != nil {
		t.Fatalf("serializar registro: %v", err)
	}

	inicio := make(chan struct{})
	estados := make(chan int, 2)
	errores := make(chan error, 2)
	var grupo sync.WaitGroup
	for range 2 {
		grupo.Add(1)
		go func() {
			defer grupo.Done()
			<-inicio
			resp, err := http.Post(servidor.URL+"/api/registro", "application/json", bytes.NewReader(payload))
			if err != nil {
				errores <- err
				return
			}
			defer resp.Body.Close()
			io.Copy(io.Discard, resp.Body)
			estados <- resp.StatusCode
		}()
	}
	close(inicio)
	grupo.Wait()
	close(estados)
	close(errores)

	for err := range errores {
		t.Errorf("registrar negocio: %v", err)
	}
	conteoEstados := map[int]int{}
	for estado := range estados {
		conteoEstados[estado]++
	}
	if conteoEstados[http.StatusCreated] != 1 || conteoEstados[http.StatusConflict] != 1 {
		t.Fatalf("se esperaba un 201 y un 409; se obtuvo %v", conteoEstados)
	}

	var negocios, usuarios int
	var rol string
	if err := testDB.QueryRow(`SELECT COUNT(*) FROM negocios`).Scan(&negocios); err != nil {
		t.Fatalf("contar negocios: %v", err)
	}
	if err := testDB.QueryRow(`SELECT COUNT(*), MIN(rol) FROM usuarios`).Scan(&usuarios, &rol); err != nil {
		t.Fatalf("consultar usuario inicial: %v", err)
	}
	if negocios != 1 || usuarios != 1 || rol != "dueño" {
		t.Fatalf("bootstrap inconsistente: negocios=%d usuarios=%d rol=%q", negocios, usuarios, rol)
	}

	respFinal, err := http.Get(servidor.URL + "/api/registro")
	if err != nil {
		t.Fatalf("consultar estado final: %v", err)
	}
	assertJSONRegistro(t, respFinal, true)
}

func assertJSONRegistro(t *testing.T, resp *http.Response, esperado bool) {
	t.Helper()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("estado de registro: esperado 200, recibido %d", resp.StatusCode)
	}
	var datos struct {
		Registrado bool `json:"registrado"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&datos); err != nil {
		t.Fatalf("decodificar estado de registro: %v", err)
	}
	if datos.Registrado != esperado {
		t.Fatalf("registrado: esperado %v, recibido %v", esperado, datos.Registrado)
	}
}
