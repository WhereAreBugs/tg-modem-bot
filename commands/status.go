package commands

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"tg_modem/engine"
)

func init() {
	Register(Command{
		Name:        "status",
		Handler:     handleStatus,
		AdminOnly:   true,
		Description: "查询调制解调器状态",
	})
}

func handleStatus(bot *tgbotapi.BotAPI, update tgbotapi.Update, eng engine.Engine) {
	status, err := eng.GetStatus()
	if err != nil {
		log.Printf("获取状态失败: %v", err)
		status = "获取状态失败: " + err.Error()
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, status)
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}
