package utils

import (
	"errors"
	"strconv"
	"strings"
)

// ParseRut parses a RUT string (e.g., "12.345.678-k", "12345678-9", "12345678")
// Returns the numeric body and the normalized DV (uppercase).
func ParseRut(rutStr string) (int, string, error) {
	// 1. Limpiar puntos y espacios
	cleanRut := strings.ReplaceAll(rutStr, ".", "")
	cleanRut = strings.TrimSpace(cleanRut)

	var bodyStr, dvStr string

	// 2. Separar por guión
	if strings.Contains(cleanRut, "-") {
		parts := strings.Split(cleanRut, "-")
		if len(parts) != 2 {
			return 0, "", errors.New("invalid rut format")
		}
		bodyStr = parts[0]
		dvStr = parts[1]
	} else {
		// Si no hay guión, asumimos que el último caracter es el DV
		// (A menos que sea muy corto, pero validaremos al convertir)
		if len(cleanRut) < 2 {
			return 0, "", errors.New("rut too short")
		}
		bodyStr = cleanRut[:len(cleanRut)-1]
		dvStr = cleanRut[len(cleanRut)-1:]
	}

	// 3. Convertir cuerpo a número
	rutBody, err := strconv.Atoi(bodyStr)
	if err != nil {
		return 0, "", errors.New("invalid rut body")
	}

	// 4. Normalizar DV
	dvStr = strings.ToUpper(dvStr)
	
	// Validar que el DV sea un dígito o 'K'
	if len(dvStr) != 1 {
		return 0, "", errors.New("invalid dv length")
	}
	// (Opcional: Aquí se podría validar el algoritmo del módulo 11)

	return rutBody, dvStr, nil
}

// NormalizeDv ensures the DV is uppercase
func NormalizeDv(dv string) string {
	return strings.ToUpper(dv)
}
