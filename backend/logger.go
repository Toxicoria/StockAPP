package main

import "fmt"

// Los códigos ANSI son secuencias especiales que la terminal interpreta como
// instrucciones de formato en lugar de imprimirlas como texto.
// Formato: \033[ = inicio de secuencia, número = comando, m = fin de secuencia.
const (
	colorReset  = "\033[0m"  // Vuelve al color por defecto
	colorRed    = "\033[31m" // Texto rojo
	colorGreen  = "\033[32m" // Texto verde
	colorYellow = "\033[33m" // Texto amarillo
	colorCyan   = "\033[36m" // Texto cyan
	colorBold   = "\033[1m"  // Texto en negrita
)

// LogError imprime un mensaje de error en rojo: [ERROR] mensaje
func LogError(msg string, args ...any) {
	// Sprintf formatea el mensaje. Si hay args extra los concatena.
	texto := fmt.Sprintf(msg, args...)
	fmt.Printf("%s[ERROR]%s %s\n", colorRed+colorBold, colorReset, texto)
}

// LogOK imprime un mensaje de éxito en verde: [OK] mensaje
func LogOK(msg string, args ...any) {
	texto := fmt.Sprintf(msg, args...)
	fmt.Printf("%s[OK]%s %s\n", colorGreen+colorBold, colorReset, texto)
}

// LogInfo imprime un mensaje informativo en cyan: [INFO] mensaje
func LogInfo(msg string, args ...any) {
	texto := fmt.Sprintf(msg, args...)
	fmt.Printf("%s[INFO]%s %s\n", colorCyan+colorBold, colorReset, texto)
}

// LogWarn imprime una advertencia en amarillo: [WARN] mensaje
func LogWarn(msg string, args ...any) {
	texto := fmt.Sprintf(msg, args...)
	fmt.Printf("%s[WARN]%s %s\n", colorYellow+colorBold, colorReset, texto)
}
