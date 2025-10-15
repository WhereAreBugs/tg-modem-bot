package commands

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"tg_modem/engine"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func init() {
	Register(Command{
		Name:        "help",
		Handler:     handleHelp,
		AdminOnly:   false,
		Description: "显示此帮助信息",
	})
}

func handleHelp(bot *tgbotapi.BotAPI, update tgbotapi.Update, eng engine.Engine) {
	// 为了判断用户身份，我们需要获取 ADMIN_CHAT_ID
	// 这是一种简便方法，避免了修改所有命令处理器的签名
	adminChatIDStr := os.Getenv("ADMIN_CHAT_ID")
	adminChatID, err := strconv.ParseInt(adminChatIDStr, 10, 64)
	if err != nil {
		log.Printf("无法解析 ADMIN_CHAT_ID 以生成帮助信息: %v", err)
		// 即使出错，也只显示公开命令
		adminChatID = -1 // 设置一个无效ID
	}

	callerIsAdmin := (update.Message.Chat.ID == adminChatID)
	allCommands := GetAll()

	// 对命令进行排序，以便输出是确定的
	var cmdNames []string
	for name := range allCommands {
		cmdNames = append(cmdNames, name)
	}
	sort.Strings(cmdNames)

	var builder strings.Builder
	builder.WriteString("你好！我是一个调制解调器控制机器人。\n\n")

	// 公开命令部分
	builder.WriteString("*公开命令:*\n")
	for _, name := range cmdNames {
		cmd := allCommands[name]
		if !cmd.AdminOnly {
			builder.WriteString(fmt.Sprintf("/%s - %s\n", cmd.Name, cmd.Description))
		}
	}

	// 如果是管理员，额外显示管理员命令
	if callerIsAdmin {
		builder.WriteString("\n*管理员命令:*\n")
		for _, name := range cmdNames {
			cmd := allCommands[name]
			if cmd.AdminOnly {
				builder.WriteString(fmt.Sprintf("/%s - %s\n", cmd.Name, cmd.Description))
			}
		}
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, builder.String())
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}
