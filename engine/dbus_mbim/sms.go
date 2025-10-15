package dbus_mbim

import (
	"fmt"
	"path"
	"strings"
	"tg_modem/engine"
	"time"

	"github.com/godbus/dbus/v5"
)

const messagingIface = "org.freedesktop.ModemManager1.Modem.Messaging"
const smsIface = "org.freedesktop.ModemManager1.Sms"

// ListSms 读取所有短信
func (e *DBusMBIMEngine) ListSms() (*engine.SmsListResult, error) {
	modemObj := e.Conn.Object(mmService, e.modemPath)

	var smsPaths []dbus.ObjectPath
	err := modemObj.Call(messagingIface+".List", 0).Store(&smsPaths)
	if err != nil {
		return nil, fmt.Errorf("无法列出短信: %w", err)
	}

	result := &engine.SmsListResult{
		Messages: make(map[string]dbus.ObjectPath),
	}

	if len(smsPaths) == 0 {
		result.DisplayText = "没有短信。"
		return result, nil
	}

	var builder strings.Builder
	for _, smsPath := range smsPaths {
		// 从路径中提取ID (e.g., /org/.../SMS/5 -> "5")
		id := path.Base(string(smsPath))
		result.Messages[id] = smsPath

		smsObj := e.Conn.Object(mmService, smsPath)

		numberVar, _ := smsObj.GetProperty(smsIface + ".Number")
		textVar, _ := smsObj.GetProperty(smsIface + ".Text")
		timestampVar, _ := smsObj.GetProperty(smsIface + ".Timestamp")

		ts, _ := time.Parse(time.RFC3339, timestampVar.Value().(string))

		// 在每条短信前加上ID
		builder.WriteString(fmt.Sprintf("✉️ *[ID: %s]* 来自: `%s` (%s)\n",
			id,
			numberVar.Value().(string),
			ts.Local().Format("2006-01-02 15:04"),
		))
		builder.WriteString(fmt.Sprintf("```\n%s\n```\n\n", textVar.Value().(string)))
	}
	result.DisplayText = builder.String()
	return result, nil
}

// SendSms 发送短信
func (e *DBusMBIMEngine) SendSms(recipient, text string) error {
	modemObj := e.Conn.Object(mmService, e.modemPath)

	props := map[string]dbus.Variant{
		"Text":   dbus.MakeVariant(text),
		"Number": dbus.MakeVariant(recipient),
	}

	var smsPath dbus.ObjectPath
	err := modemObj.Call(messagingIface+".Create", 0, props).Store(&smsPath)
	if err != nil {
		return fmt.Errorf("无法创建短信对象: %w", err)
	}

	smsObj := e.Conn.Object(mmService, smsPath)
	return smsObj.Call(smsIface+".Send", 0).Store()
}
