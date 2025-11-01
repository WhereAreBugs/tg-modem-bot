package dbus_mbim

import (
	"errors"
	"fmt"
	"log"

	"github.com/godbus/dbus/v5"
)

const simpleIface = "org.freedesktop.ModemManager1.Modem.Simple"
const bearerIface = "org.freedesktop.ModemManager1.Bearer"

// SetData 开启或关闭移动数据
func (e *DBusMBIMEngine) SetData(enable bool) error {
	modemObj := e.Conn.Object(mmService, e.modemPath)

	var status map[string]dbus.Variant
	err := modemObj.Call(simpleIface+".GetStatus", 0).Store(&status)
	if err != nil {
		return fmt.Errorf("无法调用 GetStatus: %w", err)
	}

	// Extract state from the returned dictionary.
	stateVar, ok := status["state"]
	if !ok {
		return errors.New("GetStatus 响应中未找到 'state'")
	}
	// State is a uint32 according to the documentation (MMModemState).
	state, ok := stateVar.Value().(uint32)
	if !ok {
		return fmt.Errorf("state 的类型不是 uint32, 而是 %T", stateVar.Value())
	}
	isConnected := (state == 11) // MM_MODEM_STATE_CONNECTED

	if enable && !isConnected {
		var bearerPath dbus.ObjectPath
		err := modemObj.Call(simpleIface+".Connect", 0, map[string]dbus.Variant{}).Store(&bearerPath)
		if err != nil {
			return fmt.Errorf("开启数据连接失败 (Connect call failed): %w", err)
		}
		log.Printf("数据连接已开启, Bearer: %s", bearerPath)
	} else if !enable && isConnected {
		// The bearer to disconnect is also returned by GetStatus.
		bearerPath := e.findActiveBearerForDisconnect()
		if bearerPath == "" {
			return errors.New("找不到有效的 bearer 来断开连接")
		}
		// The Disconnect method is on the Simple interface and takes the bearer path as an argument.
		err := modemObj.Call(simpleIface+".Disconnect", 0, dbus.ObjectPath(bearerPath)).Store()
		if err != nil {
			return fmt.Errorf("关闭数据连接失败 (Disconnect call failed): %w", err)
		}
		log.Printf("数据连接已关闭, Bearer: %s", bearerPath)
	}

	return nil
}

func (e *DBusMBIMEngine) findActiveBearerForDisconnect() dbus.ObjectPath {
	modemObj := e.Conn.Object(mmService, e.modemPath)
	var status map[string]dbus.Variant
	err := modemObj.Call(simpleIface+".GetStatus", 0).Store(&status)
	if err == nil {
		if bearerVar, ok := status["bearer"]; ok {
			if bearerPath, ok := bearerVar.Value().(dbus.ObjectPath); ok && bearerPath.IsValid() {
				return bearerPath
			}
		}
	}
	return ""
}
