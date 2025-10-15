package dbus_mbim

import (
	"errors"
	"fmt"

	"github.com/godbus/dbus/v5"
)

const simpleIface = "org.freedesktop.ModemManager1.Modem.Simple"
const bearerIface = "org.freedesktop.ModemManager1.Bearer"

// SetData 开启或关闭移动数据
func (e *DBusMBIMEngine) SetData(enable bool) error {
	modemObj := e.Conn.Object(mmService, e.modemPath)

	props, err := modemObj.GetProperty(simpleIface + ".State")
	if err != nil {
		return fmt.Errorf("无法获取连接状态: %w", err)
	}

	state := props.Value().(int32)
	isConnected := (state == 11)

	if enable && !isConnected {
		var bearerPath dbus.ObjectPath
		err := modemObj.Call(simpleIface+".Connect", 0, map[string]dbus.Variant{}).Store(&bearerPath)
		if err != nil {
			return fmt.Errorf("开启数据连接失败: %w", err)
		}
	} else if !enable && isConnected {
		bearerVar, err := e.getModemProperty(simpleIface, "Bearer")
		if err != nil {
			return fmt.Errorf("无法找到 bearer: %w", err)
		}
		bearerPath := bearerVar.Value().(dbus.ObjectPath)
		if !bearerPath.IsValid() {
			return errors.New("找不到有效的 bearer 来断开连接")
		}
		bearerObj := e.Conn.Object(mmService, bearerPath)
		return bearerObj.Call(bearerIface+".Disconnect", 0).Store()
	}

	return nil
}
