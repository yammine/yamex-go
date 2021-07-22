package app

type Application struct {
}

type BotMention struct {
	Platform string
	BotID    string
	UserID   string
	Text     string
}

type BotResponse struct {
	Text string
}

func (a Application) ProcessAppMention(m *BotMention) BotResponse {
	return BotResponse{Text: "I love you too"}
}
