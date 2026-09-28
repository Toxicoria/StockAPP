package db

import "testing"

func TestEnvObligatoriaRechazaValorFaltanteOVacio(t *testing.T) {
	t.Setenv("STOCKAPP_TEST_SECRET", "")
	if _, err := EnvObligatoria("STOCKAPP_TEST_SECRET"); err == nil {
		t.Fatal("se esperaba error con una variable vacía")
	}

	t.Setenv("STOCKAPP_TEST_SECRET", "   ")
	if _, err := EnvObligatoria("STOCKAPP_TEST_SECRET"); err == nil {
		t.Fatal("se esperaba error con una variable en blanco")
	}
}

func TestEnvObligatoriaDevuelveElSecretoConfigurado(t *testing.T) {
	t.Setenv("STOCKAPP_TEST_SECRET", "valor-seguro")
	value, err := EnvObligatoria("STOCKAPP_TEST_SECRET")
	if err != nil {
		t.Fatalf("EnvObligatoria: %v", err)
	}
	if value != "valor-seguro" {
		t.Fatalf("valor inesperado: %q", value)
	}
}
