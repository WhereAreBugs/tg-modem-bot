package engine

import "github.com/godbus/dbus/v5"

type SmsListResult struct {
	DisplayText string
	// Key: 用户看到的ID (e.g., "1", "2"), Value: 对应的D-Bus路径
	Messages map[string]dbus.ObjectPath
}

// Engine 定义了调制解调器控制引擎必须实现的方法
type Engine interface {
	Init() error
	GetStatus() (string, error)
	ListSms() (*SmsListResult, error)
	SendSms(recipient, text string) error
	SwitchSim(slot uint32) error
	SetData(enable bool) error
	DeleteSms(path dbus.ObjectPath) error
}

// 全局引擎注册表
var registry = make(map[string]Engine)

// Register 用于注册一个引擎实现
func Register(name string, e Engine) {
	registry[name] = e
}

// Get 用于获取一个已注册的引擎
func Get(name string) Engine {
	if e, ok := registry[name]; ok {
		return e
	}
	return nil
}
