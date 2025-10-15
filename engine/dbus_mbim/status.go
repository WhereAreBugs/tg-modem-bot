package dbus_mbim

import (
	"errors"
	"fmt"
	"github.com/godbus/dbus/v5"
	"log"
	"strings"
)

func (e *DBusMBIMEngine) GetStatus() (string, error) {
	if !e.modemPath.IsValid() {
		return "", errors.New("engine not initialized or modem path is invalid")
	}

	modemObj := e.Conn.Object(mmService, e.modemPath)
	var builder strings.Builder

	// --- 1. Modem State & Network Info ---
	builder.WriteString("ℹ️ *Modem State*\n")

	// Get basic state from the Modem object
	stateVar, _ := modemObj.GetProperty(modemIface + ".State")
	if state, ok := stateVar.Value().(int32); ok {
		stateMap := map[int32]string{
			-1: "Failed", 0: "Unknown", 1: "Initializing", 2: "Locked", 3: "Disabled",
			4: "Disabling", 5: "Enabling", 6: "Enabled", 7: "Searching", 8: "Registered",
			9: "Disconnecting", 10: "Connecting", 11: "Connected",
		}
		builder.WriteString(fmt.Sprintf("`State:` %s\n", stateMap[state]))
	}

	// Get registration state from the Modem object
	regStateVar, _ := modemObj.GetProperty(modemIface + ".RegistrationState")
	if regState, ok := regStateVar.Value().(uint32); ok {
		regStateMap := map[uint32]string{0: "Idle", 1: "Home", 2: "Searching", 3: "Denied", 4: "Unknown", 5: "Roaming"}
		builder.WriteString(fmt.Sprintf("`Registration:` %s\n", regStateMap[regState]))
	}

	// *** 修正核心：从 SIM 对象获取运营商信息 ***
	// 1. 从 Modem 对象获取 SIM 对象的路径
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
		} else if opName, ok := opNameVar.Value().(string); ok && opName != "" {
			builder.WriteString(fmt.Sprintf("`Operator:` %s\n", opName))
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

	// --- 3. Data Connection ---
	builder.WriteString("\n🌐 *Data Connection*\n")
	ipAddress := e.findIPAddress()
	if ipAddress != "" {
		builder.WriteString(fmt.Sprintf("`IPv4 Address:` %s\n", ipAddress))
	} else {
		builder.WriteString("`Status:` Not connected or no IP assigned\n")
	}

	return builder.String(), nil
}

// findIPAddress robustly finds the IPv4 address by checking all available bearers.
func (e *DBusMBIMEngine) findIPAddress() string {
	modemObj := e.Conn.Object(mmService, e.modemPath)
	bearersVar, err := modemObj.GetProperty(modemIface + ".Bearers")
	if err != nil {
		log.Printf("ERROR: Could not get bearers list: %v", err)
		return ""
	}

	bearerPaths, ok := bearersVar.Value().([]dbus.ObjectPath)
	if !ok || len(bearerPaths) == 0 {
		return ""
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

		if ip4Map, ok := ip4ConfigVar.Value().(map[string]dbus.Variant); ok {
			if addrVar, ok := ip4Map["address"]; ok {
				if ip := addrVar.Value().(string); ip != "" {
					return ip // Found it!
				}
			}
		}
	}

	return ""
}
