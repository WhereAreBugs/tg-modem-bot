package dbus_mbim

import (
	"errors"
	"fmt"
	"log"
	"tg_modem/engine"
	"tg_modem/engine/at"

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
	atHandler *at.Handler
}

func (e *DBusMBIMEngine) SetATHandler(handler interface{}) {
	e.atHandler = handler.(*at.Handler)
}

// Init 初始化 D-Bus 连接并查找第一个可用的调制解调器
func (e *DBusMBIMEngine) Init() error {
	var err error
	e.Conn, err = dbus.SystemBus()
	if err != nil {
		return fmt.Errorf("无法连接到系统 D-Bus: %w", err)
	}
	if err := e.setupSignalPolling(); err != nil {
		log.Printf("WARN: Could not setup signal polling: %v. Detailed signal info may be unavailable.", err)
	}

	// findModem 会执行查找逻辑，包含错误修正
	modemPath, err := e.findActiveModem()
	if err != nil {
		return fmt.Errorf("引擎初始化失败: %w", err)
	}

	e.modemPath = modemPath
	fmt.Printf("使用调制解调器: %s\n", e.modemPath)
	return nil
}

// findModem 查找调制解调器对象
func (e *DBusMBIMEngine) findModem() (dbus.ObjectPath, error) {
	obj := e.Conn.Object(mmService, mmPath)

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

func (e *DBusMBIMEngine) findActiveModem() (dbus.ObjectPath, error) {
	obj := e.Conn.Object(mmService, mmPath)
	var managedObjects map[dbus.ObjectPath]map[string]map[string]dbus.Variant
	err := obj.Call(objectManagerIface+".GetManagedObjects", 0).Store(&managedObjects)
	if err != nil {
		return "", fmt.Errorf("调用 GetManagedObjects 失败: %w", err)
	}

	log.Println("正在扫描所有调制解调器以查找活动设备...")
	for path, interfaces := range managedObjects {
		// 检查这是否是一个 Modem 对象
		if modemData, ok := interfaces[modemIface]; ok {
			// 检查 Modem 的状态
			if stateVar, ok := modemData["State"]; ok {
				if state, ok := stateVar.Value().(int32); ok {
					// 状态 8 (Registered) 和 11 (Connected) 是我们想要的
					log.Printf("发现 Modem: %s, 状态: %d", path, state)
					if state == 8 || state == 11 {
						return path, nil // 找到了！
					}
				}
			}
		}
	}

	return "", errors.New("未找到任何已连接或已注册的调制解调器")
}

func (e *DBusMBIMEngine) setupSignalPolling() error {
	obj := e.Conn.Object(mmService, mmPath)
	var modemPaths []dbus.ObjectPath
	// We need to find *any* modem path to call Setup on its Signal interface.
	// It seems to be a global setting for the modem hardware.
	err := obj.Call(objectManagerIface+".GetManagedObjects", 0).Store(make(map[dbus.ObjectPath]map[string]map[string]dbus.Variant))
	if err != nil {
		return err
	}

	// Use the first modem found just for this call
	if len(modemPaths) > 0 {
		modemObj := e.Conn.Object(mmService, modemPaths[0])
		return modemObj.Call("org.freedesktop.ModemManager1.Modem.Signal.Setup", 0, uint32(1)).Store()
	}
	return errors.New("no modems found to setup signal polling")
}
