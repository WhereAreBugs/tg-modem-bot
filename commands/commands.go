package commands

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/godbus/dbus/v5"
	"sync"
	"tg_modem/engine"
)

// CommandHandler 定义了命令处理函数的签名
type CommandHandler func(bot *tgbotapi.BotAPI, update tgbotapi.Update, eng engine.Engine)

// Command 定义了一个命令的结构
type Command struct {
	Name        string
	Handler     CommandHandler
	AdminOnly   bool
	Description string
}

var (
	userSmsCache = make(map[int64]map[string]dbus.ObjectPath)
	cacheMutex   = &sync.Mutex{}
)
var commandRegistry = make(map[string]Command)

// Register 用于注册一个命令
func Register(cmd Command) {
	if _, exists := commandRegistry[cmd.Name]; exists {
		return
	}
	commandRegistry[cmd.Name] = cmd
}

// Get 返回一个已注册的命令
func Get(name string) (Command, bool) {
	cmd, ok := commandRegistry[name]
	return cmd, ok
}

// GetAll 返回所有已注册的命令
func GetAll() map[string]Command {
	return commandRegistry
}
