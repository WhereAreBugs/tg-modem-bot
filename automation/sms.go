package automation

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/godbus/dbus/v5"
)

const (
	mmService      = "org.freedesktop.ModemManager1"
	messagingIface = "org.freedesktop.ModemManager1.Modem.Messaging"
	smsIface       = "org.freedesktop.ModemManager1.Sms"
)

func init() {
	Register(&SmsListener{})
}

// SmsListener 实现了监听新短信的自动化任务
type SmsListener struct{}

// Start 开始监听 D-Bus 上的短信 "Added" 信号
func (s *SmsListener) Start(params AutomationParams) error {
	err := params.Conn.AddMatchSignal(
		dbus.WithMatchObjectPath(params.ModemPath),
		dbus.WithMatchInterface(messagingIface),
	)
	if err != nil {
		return fmt.Errorf("无法添加 D-Bus 信号匹配规则 (SMS): %w", err)
	}

	sigChan := make(chan *dbus.Signal, 10)
	params.Conn.Signal(sigChan)

	log.Println("自动化任务：短信监听器已启动")

	go func() {
		for sig := range sigChan {
			// 过滤出我们关心的 "Added" 信号
			if sig.Name != messagingIface+".Added" {
				continue
			}

			// 信号体包含短信路径和一个布尔值
			if len(sig.Body) < 1 {
				continue
			}
			smsPath, ok := sig.Body[0].(dbus.ObjectPath)
			if !ok {
				continue
			}

			log.Printf("检测到新短信: %s", smsPath)
			s.processSms(params, smsPath)
		}
	}()

	return nil
}

// processSms 处理单条新短信：推送并删除
func (s *SmsListener) processSms(params AutomationParams, smsPath dbus.ObjectPath) {
	smsObj := params.Conn.Object(mmService, smsPath)

	// 获取短信内容
	numberVar, err := smsObj.GetProperty(smsIface + ".Number")
	if err != nil {
		log.Printf("无法获取短信发件人: %v", err)
		return
	}
	textVar, err := smsObj.GetProperty(smsIface + ".Text")
	if err != nil {
		log.Printf("无法获取短信内容: %v", err)
		return
	}
	timeS := ""
	Time, err := smsObj.GetProperty(smsIface + ".Timestamp")
	if err != nil {
		log.Printf("无法获取短信时间戳")
	} else {
		timeS = Time.String()
	}
	Ref, err := smsObj.GetProperty(smsIface + ".MessageReference")
	RefStr := ""
	if err != nil {
		log.Printf("无法获取Ref")
	} else {
		RefStr = Ref.String()
	}
	// 推送通知
	notificationText := fmt.Sprintf("*新短信*\n*来自:* `%s`\n*内容:*\n%s\n*时间:* %s\n*Ref*: %s",
		numberVar.Value().(string),
		textVar.Value().(string),
		timeS,
		RefStr,
	)
	msg := tgbotapi.NewMessage(params.AdminChatID, notificationText)
	msg.ParseMode = "Markdown"
	_, err = params.Bot.Send(msg)
	if err != nil {
		log.Printf("已成功处理并删除短信: %s", smsPath)
		return
	}
	modemObj := params.Conn.Object(mmService, params.ModemPath)
	//    调用 Delete 方法，并把短信的路径作为参数传入
	if err := modemObj.Call(messagingIface+".Delete", 0, smsPath).Store(); err != nil {
		log.Printf("删除短信 %s 失败: %v", smsPath, err)
	} else {
		log.Printf("已成功处理并删除短信: %s", smsPath)
	}
}
