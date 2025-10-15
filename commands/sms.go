package commands

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"tg_modem/engine"
)

func init() {
	Register(Command{
		Name:        "sms",
		Handler:     handleSms,
		AdminOnly:   true,
		Description: "读取所有短信",
	})
}

func handleSms(bot *tgbotapi.BotAPI, update tgbotapi.Update, eng engine.Engine) {
	smsResult, err := eng.ListSms()
	var replyText string

	if err != nil {
		log.Printf("读取短信失败: %v", err)
		replyText = "读取短信失败: " + err.Error()
	} else {
		replyText = smsResult.DisplayText

		// 缓存结果以供 /deletesms 使用
		cacheMutex.Lock()
		userSmsCache[update.Message.Chat.ID] = smsResult.Messages
		cacheMutex.Unlock()
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, replyText)
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}
