package dbus_mbim

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/godbus/dbus/v5"
)

// formatDuration 将秒数转换为人类可读的 Dd Hh Mm Ss 格式
func formatDuration(totalSeconds uint32) string {
	if totalSeconds == 0 {
		return "0s"
	}
	d := totalSeconds / 86400
	h := (totalSeconds % 86400) / 3600
	m := (totalSeconds % 3600) / 60
	s := totalSeconds % 60

	var parts []string
	if d > 0 {
		parts = append(parts, fmt.Sprintf("%dd", d))
	}
	if h > 0 {
		parts = append(parts, fmt.Sprintf("%dh", h))
	}
	if m > 0 {
		parts = append(parts, fmt.Sprintf("%dm", m))
	}
	if s > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%ds", s))
	}
	return strings.Join(parts, " ")
}

// GetStatus queries the modem for detailed status information.
func (e *DBusMBIMEngine) GetStatus() (string, error) {
	if !e.modemPath.IsValid() {
		return "", errors.New("引擎未初始化或 modem path is invalid")
	}

	modemObj := e.Conn.Object(mmService, e.modemPath)
	var builder strings.Builder

	// --- 1. Modem State & Network Info ---
	builder.WriteString("ℹ️ *Modem State*\n")

	// Get basic state from the main Modem interface
	stateVar, _ := modemObj.GetProperty(modemIface + ".State")
	if state, ok := stateVar.Value().(int32); ok {
		stateMap := map[int32]string{
			-1: "Failed", 0: "Unknown", 1: "Initializing", 2: "Locked", 3: "Disabled",
			4: "Disabling", 5: "Enabling", 6: "Enabled", 7: "Searching", 8: "Registered",
			9: "Disconnecting", 10: "Connecting", 11: "Connected",
		}
		builder.WriteString(fmt.Sprintf("`State:` %s\n", stateMap[state]))
	}
	// Get eSIM EID from SIM object
	simPathVar, err := modemObj.GetProperty(modemIface + ".Sim")
	if err != nil || !simPathVar.Value().(dbus.ObjectPath).IsValid() {
		log.Printf("WARN: Could not get active SIM path from modem: %v", err)
	} else {
		simPath := simPathVar.Value().(dbus.ObjectPath)
		// 2. 创建 SIM 对象
		simObj := e.Conn.Object(mmService, simPath)
		// 3. 从 SIM 对象获取 OperatorName
		opNameVar, err := simObj.GetProperty("org.freedesktop.ModemManager1.Sim.OperatorName")
		if err != nil {
			log.Printf("DEBUG: Could not get OperatorName from SIM object: %v", err)
		}
		// Get operator name and registration state from the Modem3gpp interface
		if opname, ok := opNameVar.Value().(string); ok && opname != "" {
			builder.WriteString(fmt.Sprintf("Esim: %s\n", opname))
		}
	}
	opNameVar1, _ := modemObj.GetProperty("org.freedesktop.ModemManager1.Modem.Modem3gpp.OperatorName")
	if opName, ok := opNameVar1.Value().(string); ok && opName != "" {
		builder.WriteString(fmt.Sprintf("`Operator:` %s\n", opName))
	}
	regStateVar, _ := modemObj.GetProperty("org.freedesktop.ModemManager1.Modem.Modem3gpp.RegistrationState")
	if regState, ok := regStateVar.Value().(uint32); ok {
		regStateMap := map[uint32]string{0: "Idle", 1: "Home", 2: "Searching", 3: "Denied", 4: "Unknown", 5: "Roaming"}
		builder.WriteString(fmt.Sprintf("`Registration:` %s\n", regStateMap[regState]))
	}

	// Get access technology from the Modem object
	techVar, _ := modemObj.GetProperty(modemIface + ".AccessTechnologies")
	if tech, ok := techVar.Value().(uint32); ok {
		if techStr := accessTechToString(tech); techStr != "" {
			builder.WriteString(fmt.Sprintf("`Network Type:` %s\n", techStr))
		}
	}

	// --- 2. Signal Quality ---
	builder.WriteString("\n📶 *Signal Quality*\n")
	signalVar, _ := modemObj.GetProperty(modemIface + ".SignalQuality")
	if qualityTuple, ok := signalVar.Value().([]interface{}); ok && len(qualityTuple) > 0 {
		if quality, ok := qualityTuple[0].(uint32); ok {
			builder.WriteString(fmt.Sprintf("`Quality:` %d%%\n", quality))
		}
	}

	// Detailed Signal Metrics (robustly check if available)
	signalDetailsIface := modemIface + ".Signal"
	if lteVar, err := modemObj.GetProperty(signalDetailsIface + ".Lte"); err == nil && lteVar.Value() != nil {
		if lteMap, ok := lteVar.Value().(map[string]dbus.Variant); ok && len(lteMap) > 0 {
			builder.WriteString("*LTE Metrics:*\n")
			if v, ok := lteMap["rsrp"]; ok {
				builder.WriteString(fmt.Sprintf("`RSRP:` %.2f dBm\n", v.Value().(float64)))
			}
			if v, ok := lteMap["rsrq"]; ok {
				builder.WriteString(fmt.Sprintf("`RSRQ:` %.2f dB\n", v.Value().(float64)))
			}
			if v, ok := lteMap["sinr"]; ok {
				builder.WriteString(fmt.Sprintf("`SINR:` %.2f dB\n", v.Value().(float64)))
			}
			if v, ok := lteMap["rssi"]; ok {
				builder.WriteString(fmt.Sprintf("`RSSI:` %.2f dB\n", v.Value().(float64)))
			}
		}
	}

	if nrVar, err := modemObj.GetProperty(signalDetailsIface + ".Nr5g"); err == nil && nrVar.Value() != nil {
		if nrMap, ok := nrVar.Value().(map[string]dbus.Variant); ok && len(nrMap) > 0 {
			builder.WriteString("*5G NR Metrics:*\n")
			if v, ok := nrMap["rsrp"]; ok {
				builder.WriteString(fmt.Sprintf("`RSRP:` %.2f dBm\n", v.Value().(float64)))
			}
			if v, ok := nrMap["rsrq"]; ok {
				builder.WriteString(fmt.Sprintf("`RSRQ:` %.2f dB\n", v.Value().(float64)))
			}
			if v, ok := nrMap["sinr"]; ok {
				builder.WriteString(fmt.Sprintf("`SINR:` %.2f dB\n", v.Value().(float64)))
			}
			if v, ok := nrMap["rssi"]; ok {
				builder.WriteString(fmt.Sprintf("`RSSI:` %.2f dB\n", v.Value().(float64)))
			}
		}
	}

	// Try to get detailed LTE signal metrics
	lteVar, err := modemObj.GetProperty("org.freedesktop.ModemManager1.Modem.Signal.Lte")
	if err == nil {
		if lteMap, ok := lteVar.Value().(map[string]dbus.Variant); ok && len(lteMap) > 0 {
			// The key for S/N is "snr", not "sinr" based on mmcli output.
			if v, ok := lteMap["snr"]; ok {
				if snr, ok := v.Value().(float64); ok {
					builder.WriteString(fmt.Sprintf("`S/N (SINR):` %.2f dB\n", snr))
				}
			}
		}
	}

	// --- 3. Data Connection ---
	builder.WriteString("\n🌐 *Data Connection*\n")
	ipAddress, onlineDuration := e.findBearerInfo()
	if ipAddress != "" {
		builder.WriteString(fmt.Sprintf("`IPv4 Address:` %s\n", ipAddress))
		if onlineDuration != "" {
			builder.WriteString(fmt.Sprintf("`Online Duration:` %s\n", onlineDuration))
		}
	} else {
		builder.WriteString("`Status:` Not connected or no IP assigned\n")
	}

	return builder.String(), nil
}

func (e *DBusMBIMEngine) findBearerInfo() (string, string) {
	modemObj := e.Conn.Object(mmService, e.modemPath)
	bearersVar, err := modemObj.GetProperty(modemIface + ".Bearers")
	if err != nil {
		log.Printf("ERROR: Could not get bearers list: %v", err)
		return "", ""
	}

	bearerPaths, ok := bearersVar.Value().([]dbus.ObjectPath)
	if !ok || len(bearerPaths) == 0 {
		return "", ""
	}

	for _, bearerPath := range bearerPaths {
		if !bearerPath.IsValid() {
			continue
		}
		bearerObj := e.Conn.Object(mmService, bearerPath)
		ip4ConfigVar, err := bearerObj.GetProperty(bearerIface + ".Ip4Config")
		if err != nil || ip4ConfigVar.Value() == nil {
			continue
		}
		var ipAddress, onlineDuration string
		if ip4Map, ok := ip4ConfigVar.Value().(map[string]dbus.Variant); ok {
			if addrVar, ok := ip4Map["address"]; ok {
				if ip := addrVar.Value().(string); ip != "" {
					ipAddress, ok = addrVar.Value().(string)
					if !ok {
						fmt.Printf("Get IP error!")
					}
				}
			}
		}
		if ipAddress != "" {
			statsVar, _ := bearerObj.GetProperty(bearerIface + ".Stats")
			if statsMap, ok := statsVar.Value().(map[string]dbus.Variant); ok {
				if durationVar, ok := statsMap["duration"]; ok {
					onlineDuration = formatDuration(durationVar.Value().(uint32))
				}
			}
			return ipAddress, onlineDuration // Return the first valid bearer's info
		}
	}
	return "", ""
}

func accessTechToString(tech uint32) string {
	// Based on MM_MODEM_ACCESS_TECHNOLOGY enum
	if (tech & (1 << 15)) != 0 { // MM_MODEM_ACCESS_TECHNOLOGY_NR
		return "5G"
	}
	if (tech & (1 << 14)) != 0 { // MM_MODEM_ACCESS_TECHNOLOGY_LTE
		return "4G (LTE)"
	}
	if (tech & (1 << 13)) != 0 { // MM_MODEM_ACCESS_TECHNOLOGY_HSPAP
		return "3G (HSPA+)"
	}
	if (tech & (1 << 12)) != 0 { // MM_MODEM_ACCESS_TECHNOLOGY_HSUPA
		return "3G (HSUPA)"
	}
	if (tech & (1 << 11)) != 0 { // MM_MODEM_ACCESS_TECHNOLOGY_HSDPA
		return "3G (HSDPA)"
	}
	if (tech & (1 << 10)) != 0 { // MM_MODEM_ACCESS_TECHNOLOGY_UMTS
		return "3G (UMTS)"
	}
	if (tech & (1 << 5)) != 0 { // MM_MODEM_ACCESS_TECHNOLOGY_EDGE
		return "2G (EDGE)"
	}
	return "" // Return empty if unknown or lower tech
}
