package api

import "testing"

func TestCifradoRequiereClaveConfigurada(t *testing.T) {
	t.Setenv("CONFIG_SECRET_KEY", "")
	if _, err := cifrar("secreto fiscal"); err == nil {
		t.Fatal("se esperaba error sin CONFIG_SECRET_KEY")
	}
}

func TestCifradoRechazaClaveCorta(t *testing.T) {
	t.Setenv("CONFIG_SECRET_KEY", "demasiado-corta")
	if _, err := cifrar("secreto fiscal"); err == nil {
		t.Fatal("se esperaba error con una CONFIG_SECRET_KEY corta")
	}
}

func TestCifradoRoundTrip(t *testing.T) {
	t.Setenv("CONFIG_SECRET_KEY", "clave-de-prueba-distinta-a-produccion")

	cifrado, err := cifrar("secreto fiscal")
	if err != nil {
		t.Fatalf("cifrar: %v", err)
	}
	claro, err := descifrar(cifrado)
	if err != nil {
		t.Fatalf("descifrar: %v", err)
	}
	if claro != "secreto fiscal" {
		t.Fatalf("round trip inválido: %q", claro)
	}
}
