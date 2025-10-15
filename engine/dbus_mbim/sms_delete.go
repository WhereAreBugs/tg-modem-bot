package dbus_mbim

import "github.com/godbus/dbus/v5"

func (e *DBusMBIMEngine) DeleteSms(path dbus.ObjectPath) error {
	modemObj := e.Conn.Object(mmService, e.modemPath)
	return modemObj.Call(messagingIface+".Delete", 0, path).Store()
}
