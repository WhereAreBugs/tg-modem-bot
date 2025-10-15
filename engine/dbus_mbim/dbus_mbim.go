package dbus_mbim

import (
	"errors"
	"fmt"
	"tg_modem/engine"

	"github.com/godbus/dbus/v5"
)

// D-Bus 常量
const (
	mmService          = "org.freedesktop.ModemManager1"
	mmPath             = "/org/freedesktop/ModemManager1"
	objectManagerIface = "org.freedesktop.DBus.ObjectManager"
	modemIface         = "org.freedesktop.ModemManager1.Modem"
)

func init() {
	engine.Register("dbus_mbim", &DBusMBIMEngine{})
}

// DBusMBIMEngine 通过 D-Bus 与 ModemManager 交互
type DBusMBIMEngine struct {
	Conn      *dbus.Conn
	modemPath dbus.ObjectPath
}

// Init 初始化 D-Bus 连接并查找第一个可用的调制解调器
func (e *DBusMBIMEngine) Init() error {
	var err error
	e.Conn, err = dbus.SystemBus()
	if err != nil {
		return fmt.Errorf("无法连接到系统 D-Bus: %w", err)
	}

	// findModem 会执行查找逻辑，包含错误修正
	modemPath, err := e.findModem()
	if err != nil {
		return fmt.Errorf("引擎初始化失败: %w", err)
	}

	e.modemPath = modemPath
	fmt.Printf("使用调制解调器: %s\n", e.modemPath)
	return nil
}

// findModem 使用正确的方法查找调制解调器对象
func (e *DBusMBIMEngine) findModem() (dbus.ObjectPath, error) {
	obj := e.Conn.Object(mmService, mmPath)

	// *** 这是关键的修正 ***
	// GetManagedObjects 方法在 org.freedesktop.DBus.ObjectManager 接口上
	var managedObjects map[dbus.ObjectPath]map[string]map[string]dbus.Variant
	err := obj.Call(objectManagerIface+".GetManagedObjects", 0).Store(&managedObjects)
	if err != nil {
		return "", fmt.Errorf("调用 GetManagedObjects 失败: %w", err)
	}

	// 遍历所有对象，找到实现了 Modem 接口的那个
	for path, interfaces := range managedObjects {
		if _, ok := interfaces[modemIface]; ok {
			return path, nil // 找到了第一个 modem
		}
	}

	return "", errors.New("未找到任何调制解调器")
}
func (e *DBusMBIMEngine) GetModemPath() dbus.ObjectPath {
	return e.modemPath
}
