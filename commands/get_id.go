package commands

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"tg_modem/engine"
)

func init() {
	Register(Command{
		Name:        "getid",
		Handler:     handleGetID,
		AdminOnly:   false,
		Description: "获取你当前的 Chat ID",
	})
}

func handleGetID(bot *tgbotapi.BotAPI, update tgbotapi.Update, eng engine.Engine) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("你的 Chat ID 是: `%d`", update.Message.Chat.ID))
	msg.ParseMode = "MarkdownV2"
	bot.Send(msg)
}
