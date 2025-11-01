package at

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"go.bug.st/serial"
)

// Handler manages AT command communication over a serial port.
type Handler struct {
	portName string
}

// NewHandler creates a new AT command handler.
func NewHandler(portName string) *Handler {
	return &Handler{portName: portName}
}

// SendCommand sends an AT command and waits for a final response ("OK" or "ERROR").
func (h *Handler) SendCommand(cmd string) (string, error) {
	if h == nil {
		return "", errors.New("AT handler is not initialized")
	}

	port, err := serial.Open(h.portName, &serial.Mode{
		BaudRate: 115200,
	})
	if err != nil {
		return "", fmt.Errorf("failed to open serial port %s: %w", h.portName, err)
	}
	defer port.Close()

	// Set a read timeout
	port.SetReadTimeout(5 * time.Second)

	log.Printf("AT > %s", cmd)
	_, err = port.Write([]byte(cmd + "\r\n"))
	if err != nil {
		return "", fmt.Errorf("failed to write to serial port: %w", err)
	}

	var responseBuilder strings.Builder
	scanner := bufio.NewScanner(port)
	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("AT < %s", line)
		if line == "OK" {
			return responseBuilder.String(), nil
		}
		if strings.Contains(line, "ERROR") {
			return responseBuilder.String(), fmt.Errorf("AT command failed: %s", line)
		}
		if line != "" && line != cmd { // Exclude empty lines and command echo
			responseBuilder.WriteString(line + "\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return responseBuilder.String(), fmt.Errorf("error reading from serial port: %w", err)
	}

	return responseBuilder.String(), errors.New("AT command timed out, no OK/ERROR received")
}

// GetICCID retrieves the ICCID using the AT+CCID? command.
func (h *Handler) GetICCID() (string, error) {
	response, err := h.SendCommand("AT+CCID?")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(strings.TrimPrefix(response, "+CCID: ")), nil
}

func (h *Handler) GetStatus() (string, error) {
	response, err := h.SendCommand("AT+SIMTYPE?")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(strings.TrimPrefix(response, "+SIMTYPE: ")), nil
}

func (h *Handler) GetEID() (string, error) {
	response, err := h.SendCommand("AT+EID?")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(strings.TrimPrefix(response, "+EID: ")), nil
}
func (h *Handler) GetEsimPower() (string, error) {
	response, err := h.SendCommand("AT+GTESIMCFG?")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(strings.TrimPrefix(response, "+GTESIMCFG: ")), nil
}

func (h *Handler) SetEsimPower(power bool) (string, error) {
	if power == false { //关闭ESIM
		response, err := h.SendCommand("AT+GTESIMCFG=1,0,0")
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(strings.TrimPrefix(response, "+GTESIMCFG: ")), nil
	} else {
		response, err := h.SendCommand("AT+GTESIMCFG=0,0,0")
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(strings.TrimPrefix(response, "+GTESIMCFG: ")), nil
	}
}
