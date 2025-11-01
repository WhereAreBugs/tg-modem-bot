package commands

import (
	"fmt"
	"log"
	"strings"
	"tg_modem/engine"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func init() {
	Register(Command{
		Name:        "esim",
		Handler:     handleEsim,
		AdminOnly:   true,
		Description: "[AT] eSIM配置管理",
	})
}

func handleEsim(bot *tgbotapi.BotAPI, update tgbotapi.Update, eng engine.Engine) {
	args := strings.SplitN(update.Message.CommandArguments(), " ", 2)
	subcommand := args[0]

	// Check if the engine supports AT commands
	atEngine, ok := eng.(engine.ATEngine)
	if !ok {
		reply(bot, update, "错误: 当前引擎不支持eSIM (AT命令) 功能。")
		return
	}

	switch subcommand {
	case "info":
		handleEsimInfo(bot, update, atEngine)
	// Download is too complex for a simple AT command handler for now.
	// case "switch":
	// 	if len(args) < 2 {
	// 		reply(bot, update, "用法: /esim switch <ICCID>")
	// 		return
	// 	}
	// 	handleEsimSwitch(bot, update, atEngine, args[1])
	// case "delete":
	// 	if len(args) < 2 {
	// 		reply(bot, update, "用法: /esim delete <ICCID>")
	// 		return
	// 	}
	// 	handleEsimDelete(bot, update, atEngine, args[1])
	default:
		reply(bot, update, "未知的esim子命令")
	}
}

func handleEsimInfo(bot *tgbotapi.BotAPI, update tgbotapi.Update, eng engine.ATEngine) {
	msg, _ := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "⏳ 正在通过 AT 命令查询 ICCID..."))

	iccid, err := eng.GetEsimICCID()
	if err != nil {
		log.Printf("查询eSIM ICCID失败: %v", err)
		bot.Send(tgbotapi.NewEditMessageText(msg.Chat.ID, msg.MessageID, "❌ 查询失败: "+err.Error()))
		return
	}
	var builder strings.Builder
	builder.WriteString("📋 eSIM 基础信息:\n")
	builder.WriteString(fmt.Sprintf("ICCID: %s\n", iccid))

	status, err := eng.GetEsimStatus()
	if err != nil {
		log.Println("查询eSIM Status失败:%v", err)
	} else {
		builder.WriteString(fmt.Sprintf("eSIM状态: %s\n", status))
	}

	eid, err := eng.GetEsimEID()
	if err != nil {
		log.Println("查询eSIM EID失败:%v", err)
	} else {
		builder.WriteString(fmt.Sprintf("EID: %s\n", eid))
	}

	bot.Send(tgbotapi.NewEditMessageText(msg.Chat.ID, msg.MessageID, builder.String()))
}

func reply(bot *tgbotapi.BotAPI, update tgbotapi.Update, text string) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}
