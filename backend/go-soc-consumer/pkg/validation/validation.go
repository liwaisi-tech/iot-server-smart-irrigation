package validation

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidateMACAddress validates the MAC address format
// It supports both colon (:) and dash (-) separators, but they must be consistent
// Example valid formats: "01:23:45:67:89:AB" or "01-23-45-67-89-AB"
func ValidateMACAddress(macAddress string) error {
	if macAddress == "" {
		return fmt.Errorf("mac address is required")
	}

	// Normalize to uppercase for consistency
	macAddress = strings.ToUpper(strings.TrimSpace(macAddress))

	// Check for consistent separator (either all colons or all dashes)
	hasColon := strings.Contains(macAddress, ":")
	hasDash := strings.Contains(macAddress, "-")

	if hasColon && hasDash {
		return fmt.Errorf("invalid mac address format: mixed separators (use either colons or dashes)")
	}

	// MAC address pattern: XX:XX:XX:XX:XX:XX or XX-XX-XX-XX-XX-XX
	macPattern := `^([0-9A-F]{2}[:-]){5}([0-9A-F]{2})$`
	matched, err := regexp.MatchString(macPattern, macAddress)
	if err != nil {
		return fmt.Errorf("error validating mac address: %w", err)
	}

	if !matched {
		return fmt.Errorf("invalid mac address format: %s (expected format: XX:XX:XX:XX:XX:XX or XX-XX-XX-XX-XX-XX)", macAddress)
	}

	return nil
}
