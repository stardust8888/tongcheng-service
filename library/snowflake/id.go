package snowflake

func MsgID() string {
	return "MSG" + GenerateString()
}

func ChatID() string {
	return "CHAT" + GenerateString()
}
