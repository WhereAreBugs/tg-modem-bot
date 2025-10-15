package commands

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"tg_modem/engine"
)

func init() {
	Register(Command{
		Name:        "switchsim",
		Handler:     handleSwitchSim,
		AdminOnly:   true,
		Description: "<slot> - 切换SIM卡槽",
	})
}

func handleSwitchSim(bot *tgbotapi.BotAPI, update tgbotapi.Update, eng engine.Engine) {
	slotStr := update.Message.CommandArguments()
	slot, err := strconv.ParseUint(slotStr, 10, 32)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "无效的卡槽号. 请输入一个数字.")
		bot.Send(msg)
		return
	}

	err = eng.SwitchSim(uint32(slot))
	reply := fmt.Sprintf("已尝试切换到 SIM 卡槽 %d。", slot)
	if err != nil {
		log.Printf("切换SIM卡失败: %v", err)
		reply = "切换SIM卡失败: " + err.Error()
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
	bot.Send(msg)
}
