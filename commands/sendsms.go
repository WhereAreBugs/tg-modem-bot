package commands

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
	"tg_modem/engine"
)

func init() {
	Register(Command{
		Name:        "sendsms",
		Handler:     handleSendSms,
		AdminOnly:   true,
		Description: "<号码> <内容> - 发送短信",
	})
}

func handleSendSms(bot *tgbotapi.BotAPI, update tgbotapi.Update, eng engine.Engine) {
	args := strings.SplitN(update.Message.CommandArguments(), " ", 2)
	if len(args) < 2 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "格式错误. 请使用: /sendsms <号码> <内容>")
		bot.Send(msg)
		return
	}
	recipient, text := args[0], args[1]

	err := eng.SendSms(recipient, text)
	reply := "短信已发送成功。"
	if err != nil {
		log.Printf("发送短信失败: %v", err)
		reply = "发送短信失败: " + err.Error()
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
	bot.Send(msg)
}
