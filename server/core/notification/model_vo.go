package notification

type MessageTemplateVO struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Type  string `json:"type"`
	Title string `json:"title"`
	Body  string `json:"body"`
	Lang  string `json:"lang"`
}

type MessageLogVO struct {
	ID          uint   `json:"id"`
	TemplateID  uint   `json:"templateID"`
	Destination string `json:"destination"`
	Channel     string `json:"channel"`
	Title       string `json:"title"`
	Body        string `json:"body"`
	Result      string `json:"result"`
}
