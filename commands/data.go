package commands

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
	"tg_modem/engine"
)

func init() {
	Register(Command{
		Name:        "data",
		Handler:     handleData,
		AdminOnly:   true,
		Description: "<on|off> - 开启或关闭移动数据",
	})
}

func handleData(bot *tgbotapi.BotAPI, update tgbotapi.Update, eng engine.Engine) {
	arg := strings.ToLower(update.Message.CommandArguments())
	var enable bool
	var action string

	switch arg {
	case "on":
		enable = true
		action = "开启"
	case "off":
		enable = false
		action = "关闭"
	default:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "无效的参数. 请使用 'on' 或 'off'.")
		bot.Send(msg)
		return
	}

	err := eng.SetData(enable)
	reply := "已尝试" + action + "移动数据。"
	if err != nil {
		log.Printf("%s移动数据失败: %v", action, err)
		reply = action + "移动数据失败: " + err.Error()
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
	bot.Send(msg)
}
