package automation

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/godbus/dbus/v5"
)

// AutomationParams 包含了启动自动化任务所需的所有依赖
type AutomationParams struct {
	Bot         *tgbotapi.BotAPI
	AdminChatID int64
	Conn        *dbus.Conn
	ModemPath   dbus.ObjectPath
}

// Automation 定义了自动化任务必须实现的接口
type Automation interface {
	Start(params AutomationParams) error
}

var registry []Automation

// Register 用于注册一个自动化任务
func Register(a Automation) {
	registry = append(registry, a)
}

// GetAll 返回所有已注册的自动化任务
func GetAll() []Automation {
	return registry
}
