package commands

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"tg_modem/engine"
)

func init() {
	Register(Command{
		Name:        "deletesms",
		Handler:     handleDeleteSms,
		AdminOnly:   true,
		Description: "<ID> - 删除指定ID的短信",
	})
}

func handleDeleteSms(bot *tgbotapi.BotAPI, update tgbotapi.Update, eng engine.Engine) {
	smsID := update.Message.CommandArguments()
	if smsID == "" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "请提供要删除的短信ID. \n用法: /deletesms <ID>")
		bot.Send(msg)
		return
	}

	chatID := update.Message.Chat.ID

	// 从缓存中查找短信路径
	cacheMutex.Lock()
	userMessages, ok := userSmsCache[chatID]
	if !ok || len(userMessages) == 0 {
		cacheMutex.Unlock()
		msg := tgbotapi.NewMessage(chatID, "未找到短信列表缓存. 请先运行 /sms 命令。")
		bot.Send(msg)
		return
	}

	smsPath, found := userMessages[smsID]
	cacheMutex.Unlock() // 尽快释放锁

	if !found {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("无效的短信ID: %s. 请运行 /sms 查看可用ID。", smsID))
		bot.Send(msg)
		return
	}

	// 执行删除操作
	err := eng.DeleteSms(smsPath)
	var replyText string
	if err != nil {
		log.Printf("删除短信 %s (ID: %s) 失败: %v", smsPath, smsID, err)
		replyText = fmt.Sprintf("删除短信 ID %s 失败: %s", smsID, err.Error())
	} else {
		log.Printf("成功删除短信 %s (ID: %s)", smsPath, smsID)
		replyText = fmt.Sprintf("✅ 短信 ID %s 已成功删除。", smsID)

		// 从缓存中移除已删除的短信，防止重复删除
		cacheMutex.Lock()
		if userMessages, ok := userSmsCache[chatID]; ok {
			delete(userMessages, smsID)
		}
		cacheMutex.Unlock()
	}

	msg := tgbotapi.NewMessage(chatID, replyText)
	bot.Send(msg)
}
