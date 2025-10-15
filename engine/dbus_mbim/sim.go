package dbus_mbim

// SwitchSim 切换SIM卡槽
func (e *DBusMBIMEngine) SwitchSim(slot uint32) error {
	modemObj := e.Conn.Object(mmService, e.modemPath)
	return modemObj.Call(modemIface+".SetCurrentSlots", 0, []uint32{slot}).Store()
}
