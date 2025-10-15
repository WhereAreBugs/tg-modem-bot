package dbus_mbim

import (
	"fmt"

	"github.com/godbus/dbus/v5"
)

// getModemProperty 获取 modem 的一个属性
func (e *DBusMBIMEngine) getModemProperty(iface, propName string) (dbus.Variant, error) {
	modemObj := e.Conn.Object(mmService, e.modemPath)
	return modemObj.GetProperty(fmt.Sprintf("%s.%s", iface, propName))
}
