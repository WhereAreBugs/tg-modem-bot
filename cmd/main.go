package main

import (
	"log"
	"os"
	"sort"
	"strconv"
	"tg_modem/automation"
	_ "tg_modem/automation"
	"tg_modem/commands"
	_ "tg_modem/commands" // 空导入以执行 commands 包中的 init() 函数
	"tg_modem/engine"
	"tg_modem/engine/at"
	"tg_modem/engine/dbus_mbim"
	_ "tg_modem/engine/dbus_mbim"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// 1. 加载配置
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("环境变量 TELEGRAM_BOT_TOKEN 未设置")
	}

	adminChatIDStr := os.Getenv("ADMIN_CHAT_ID")
	if adminChatIDStr == "" {
		log.Fatal("环境变量 ADMIN_CHAT_ID 未设置")
	}
	atPortStr := os.Getenv("AT_PORT")
	if atPortStr == "" {
		log.Println("WARN:环境变量AT_PORT未设置，使用/dev/wwan0at0.")
		atPortStr = "/dev/wwan0at0"
	}
	adminChatID, err := strconv.ParseInt(adminChatIDStr, 10, 64)
	if err != nil {
		log.Fatalf("无效的 ADMIN_CHAT_ID: %v", err)
	}

	// 2. 初始化引擎
	eng := engine.Get("dbus_mbim")
	if eng == nil {
		log.Fatal("无法找到 'dbus_mbim' 引擎")
	}

	if err := eng.Init(); err != nil {
		log.Fatalf("引擎初始化失败: %v", err)
	}
	if s, ok := eng.(engine.ATSetter); ok {
		log.Printf("Setting AT Handler:%s", atPortStr)
		s.SetATHandler(at.NewHandler(atPortStr))
	}
	log.Println("Modem 引擎初始化成功")

	// 3. 初始化 Telegram Bot
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = false
	log.Printf("已授权为机器人: %s", bot.Self.UserName)
	setupTelegramCommands(bot, adminChatID)
	// Add
	dbusEngine, ok := eng.(*dbus_mbim.DBusMBIMEngine)
	if !ok {
		log.Fatal("自动化任务需要 'dbus_mbim' 引擎, 但加载了其他类型")
	}

	autoParams := automation.AutomationParams{
		Bot:         bot,
		AdminChatID: adminChatID,
		Conn:        dbusEngine.Conn,
		ModemPath:   dbusEngine.GetModemPath(), // 需要为 DBusMBIMEngine 添加一个 Getter
	}

	for _, task := range automation.GetAll() {
		if err := task.Start(autoParams); err != nil {
			log.Printf("启动自动化任务失败: %v", err)
		}
	}

	// 4. 设置更新通道并开始监听
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	log.Println("开始监听 Telegram 更新...")

	for update := range updates {
		if update.Message == nil || !update.Message.IsCommand() {
			continue
		}

		cmdName := update.Message.Command()
		cmd, ok := commands.Get(cmdName)
		if !ok {
			continue
		}

		if cmd.AdminOnly && update.Message.Chat.ID != adminChatID {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "无权执行此命令。")
			bot.Send(msg)
			log.Printf("拒绝来自非管理员 (%d) 的命令: /%s", update.Message.Chat.ID, cmdName)
			continue
		}

		log.Printf("收到来自 %d 的命令: /%s, 参数: %s", update.Message.Chat.ID, cmdName, update.Message.CommandArguments())
		go cmd.Handler(bot, update, eng)
	}
}

func setupTelegramCommands(bot *tgbotapi.BotAPI, adminChatID int64) {
	publicCmdsAPI := []tgbotapi.BotCommand{}
	adminCmdsAPI := []tgbotapi.BotCommand{}

	// 从命令注册表中构建列表
	allCommands := commands.GetAll()

	// 为了排序，先提取 key
	var cmdNames []string
	for name := range allCommands {
		cmdNames = append(cmdNames, name)
	}
	sort.Strings(cmdNames)

	for _, name := range cmdNames {
		cmd := allCommands[name]
		botCmd := tgbotapi.BotCommand{Command: name, Description: cmd.Description}

		if !cmd.AdminOnly {
			publicCmdsAPI = append(publicCmdsAPI, botCmd)
		}
		// 管理员应该看到所有命令
		adminCmdsAPI = append(adminCmdsAPI, botCmd)
	}

	// 1. 为所有普通用户设置公开命令 (默认作用域)
	defaultScope := tgbotapi.NewSetMyCommandsWithScope(tgbotapi.NewBotCommandScopeDefault(), publicCmdsAPI...)
	if _, err := bot.Request(defaultScope); err != nil {
		log.Printf("设置默认命令列表失败: %v", err)
	} else {
		log.Println("已成功为普通用户设置命令列表。")
	}

	// 2. 为管理员设置所有命令 (特定聊天作用域)
	adminScope := tgbotapi.NewSetMyCommandsWithScope(tgbotapi.NewBotCommandScopeChat(adminChatID), adminCmdsAPI...)
	if _, err := bot.Request(adminScope); err != nil {
		log.Printf("设置管理员命令列表失败: %v", err)
	} else {
		log.Println("已成功为管理员设置命令列表。")
	}
}
