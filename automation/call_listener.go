package automation

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/godbus/dbus/v5"
)

const (
	voiceIface = "org.freedesktop.ModemManager1.Modem.Voice"
	callIface  = "org.freedesktop.ModemManager1.Call"
)

func init() {
	Register(&CallListener{})
}

// CallListener 实现了监听来电的自动化任务
type CallListener struct{}

// Start 开始监听 D-Bus 上的来电 "CallAdded" 信号
func (c *CallListener) Start(params AutomationParams) error {
	err := params.Conn.AddMatchSignal(
		dbus.WithMatchObjectPath(params.ModemPath),
		dbus.WithMatchInterface(voiceIface),
	)
	if err != nil {
		return fmt.Errorf("无法添加 D-Bus 信号匹配规则 (Call): %w", err)
	}

	sigChan := make(chan *dbus.Signal, 10)
	params.Conn.Signal(sigChan)

	log.Println("自动化任务：来电监听器已启动")

	go func() {
		for sig := range sigChan {
			if sig.Name != voiceIface+".CallAdded" {
				continue
			}

			if len(sig.Body) < 1 {
				continue
			}
			callPath, ok := sig.Body[0].(dbus.ObjectPath)
			if !ok {
				continue
			}

			log.Printf("检测到新来电: %s", callPath)
			c.processCall(params, callPath)
		}
	}()

	return nil
}

func (c *CallListener) processCall(params AutomationParams, callPath dbus.ObjectPath) {
	callObj := params.Conn.Object(mmService, callPath)
	numberVar, err := callObj.GetProperty(callIface + ".Number")
	if err != nil {
		log.Printf("无法获取来电号码: %v", err)
		return
	}

	number := numberVar.Value().(string)
	if number == "" {
		number = "未知号码"
	}

	notificationText := fmt.Sprintf("📞 *来电提醒*\n*来自:* `%s`", number)
	msg := tgbotapi.NewMessage(params.AdminChatID, notificationText)
	msg.ParseMode = "Markdown"
	params.Bot.Send(msg)
}
