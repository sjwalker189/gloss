package diagnostic

import "gloss/token"

type Message struct {
	Line     int
	Column   int
	Text     string
	Severiry Severity
}

type MessageList struct {
	list []Message
}

func (dl *MessageList) Any() bool {
	return len(dl.list) > 0
}

func (dl *MessageList) Messages() []Message {
	return dl.list
}

func (dl *MessageList) Error(t token.Token, msg string) {
	dl.list = append(dl.list, Message{
		Line:     t.Line,
		Column:   t.Column,
		Text:     msg,
		Severiry: SeverityError,
	})
}

func (dl *MessageList) Warn(t token.Token, msg string) {
	dl.list = append(dl.list, Message{
		Line:     t.Line,
		Column:   t.Column,
		Text:     msg,
		Severiry: SeverityWarn,
	})
}

func (dl *MessageList) Raise(t token.Token, msg string) {
	dl.list = append(dl.list, Message{
		Line:     t.Line,
		Column:   t.Column,
		Text:     msg,
		Severiry: SeverityInfo,
	})
}
