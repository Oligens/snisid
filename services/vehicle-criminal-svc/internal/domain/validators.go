package domain

import (
	"fmt"
	"regexp"
)

var (
	plateRegex   = regexp.MustCompile(`^[A-Z]{1,3}[-\s]?[0-9]{3,6}[A-Z]?$`)
	vinRegex     = regexp.MustCompile(`^[A-HJ-NPR-Z0-9]{17}$`)
	statePlateRe = regexp.MustCompile(`^SE[-\s]?[0-9]{4,6}$`)
)

func ValidatePlateNumber(plate string) error {
	if plate == "" {
		return fmt.Errorf("numéro de plaque requis")
	}
	if !plateRegex.MatchString(plate) {
		return fmt.Errorf("format de plaque invalide: %s (attendu: XX-1234 ou XXX-123456)", plate)
	}
	return nil
}

func ValidateVIN(vin string) error {
	if vin == "" {
		return fmt.Errorf("NIV/VIN requis")
	}
	if len(vin) != 17 {
		return fmt.Errorf("NIV/VIN doit contenir exactement 17 caractères, reçu: %d", len(vin))
	}
	if !vinRegex.MatchString(vin) {
		return fmt.Errorf("NIV/VIN contient des caractères invalides: %s", vin)
	}
	return nil
}

func IsStatePlate(plate string) bool {
	return statePlateRe.MatchString(plate)
}

func NormalizePlate(plate string) string {
	result := make([]byte, 0, len(plate))
	for _, c := range plate {
		if c == ' ' || c == '-' {
			continue
		}
		result = append(result, byte(c))
	}
	return string(result)
}
