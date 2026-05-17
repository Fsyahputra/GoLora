package periphIO

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

// isOrangePiOne checks if the current platform is specifically an Orange Pi One.
func isOrangePiOne() bool {
	// Check for specific Orange Pi One identification in device tree
	if _, err := os.Stat("/proc/device-tree/model"); err == nil {
		model, err := os.ReadFile("/proc/device-tree/model")
		// Matches "Xunlong Orange Pi One" exactly as shown in cat /proc/device-tree/model
		if err == nil && strings.Contains(string(model), "Orange Pi One") {
			return true
		}
	}

	return false
}

// getWiringPiPin maps GPIO name to WiringPi pin number based on Orange Pi H3 table.
func getWiringPiPin(pinName string) string {
	// Clean the name (e.g., "GPIO6" -> "6", "PA6" -> "6")
	name := strings.TrimPrefix(pinName, "GPIO")
	
	// Internal mapping table based on Orange Pi H3 GPIO map
	mapping := map[string]string{
		"12":  "0",  // SDA.0 (PA12)
		"11":  "1",  // SCL.0 (PA11)
		"6":   "2",  // PA6
		"13":  "3",  // TXD.3 (PA13)
		"14":  "4",  // RXD.3 (PA14)
		"1":   "5",  // RXD.2 (PA1)
		"110": "6",  // PD14
		"0":   "7",  // TXD.2 (PA0)
		"3":   "8",  // CTS.2 (PA3)
		"68":  "9",  // PC4 (PC04)
		"71":  "10", // PC7 (PC07)
		"64":  "11", // MOSI.0 (PC64?) - Based on table physical 19
		"65":  "12", // MISO.0
		"2":   "13", // RTS.2 (PA2)
		"66":  "14", // SCLK.0
		"67":  "15", // CE.0
		"21":  "16", // PA21
		"19":  "17", // SDA.1 (PA19)
		"18":  "18", // SCL.1 (PA18)
		"7":   "19", // PA07
		"8":   "20", // PA08
		"200": "21", // RTS.1 (PG200?) - Based on table physical 31
		"9":   "22", // PA09
		"10":  "23", // PA10
		"201": "24", // CTS.1
		"20":  "25", // PA20
		"198": "26", // TXD.1
		"199": "27", // RXD.1
	}

	if wpi, ok := mapping[name]; ok {
		return wpi
	}

	// If no mapping found, return original name and hope for the best
	return name
}

// applyGPIOInWorkaround runs 'gpio mode <n> in' specifically for Orange Pi One stability.
func applyGPIOInWorkaround(pinName string) {
	if !isOrangePiOne() {
		return
	}

	wpi := getWiringPiPin(pinName)
	log.Printf("[OrangePiOneWorkaround] Setting GPIO %s (wPi %s) to mode IN via subprocess", pinName, wpi)
	
	cmd := exec.Command("gpio", "mode", wpi, "in")
	if err := cmd.Run(); err != nil {
		log.Printf("[OrangePiOneWorkaround] Warning: failed to execute 'gpio mode %s in': %v", wpi, err)
	}
}
