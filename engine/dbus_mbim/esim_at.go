package dbus_mbim

import (
	"errors"
	"strings"
)

// GetEsimICCID retrieves the eSIM ICCID via AT commands.
func (e *DBusMBIMEngine) GetEsimICCID() (string, error) {
	if e.atHandler == nil {
		return "", errors.New("AT command handler not configured for this engine")
	}
	return e.atHandler.GetICCID()
}

func (e *DBusMBIMEngine) GetEsimEID() (string, error) {
	if e.atHandler == nil {
		return "", errors.New("AT command handler not configured for this engine")
	}
	return e.atHandler.GetEID()
}
func (e *DBusMBIMEngine) GetEsimPower() (string, error) {
	if e.atHandler == nil {
		return "", errors.New("AT command handler not configured for this engine")
	}

	power, err := e.atHandler.GetEsimPower()
	if err != nil {
		return "", err
	}
	powerList := strings.Split(power, ",")
	var build strings.Builder
	if powerList[0] == "0" {
		build.WriteString("ESIM模块启用")
	} else if powerList[0] == "1" {
		build.WriteString("ESIM模块禁用")
	}
	if powerList[1] == "0" {
		build.WriteString(",SKU_based 0")
	} else if powerList[1] == "1" {
		build.WriteString(",SKU_based 1")
	}
	if powerList[2] == "0" {
		build.WriteString(",IMSI_based 0")
	} else if powerList[2] == "1" {
		build.WriteString(",IMSI_based 1")
	}
	return strings.TrimSpace(build.String()), nil
}
func (e *DBusMBIMEngine) SetEsimPower(power bool) (string, error) {
	if e.atHandler == nil {
		return "", errors.New("AT command handler not configured for this engine")
	}
	return e.atHandler.SetEsimPower(power)
}
func (e *DBusMBIMEngine) GetEsimStatus() (string, error) {
	if e.atHandler == nil {
		return "", errors.New("AT command handler not configured for this engine")
	}
	status, err := e.atHandler.GetStatus()
	if status == "1" {
		return "ESIM已启用", err
	} else {
		return "ESIM未启用或未知", err
	}
}
