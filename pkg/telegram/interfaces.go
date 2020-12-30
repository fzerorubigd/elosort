package telegram

import (
	"context"
)

type (
	button struct {
		keys []string
		text string
		inl  bool
	}
	message struct {
		hideButton bool
		text       string
	}

	multi struct {
		resp []Response
	}
)

func (m *multi) buttonText() string {
	return ""
}

func (m *multi) inline() bool {
	return false
}

func (m *multi) SetText(s string) {

}

func (m *message) inline() bool {
	return false
}

func (b *button) inline() bool {
	return b.inl
}

func (m *message) buttonText() string {
	return m.text
}

func (m *message) SetText(s string) {
	m.text = s
}

func (b *button) buttonText() string {
	return b.text
}

func (b *button) SetText(s string) {
	b.text = s
}

// Response is a telegram response
type Response interface {
	buttonText() string
	inline() bool
	SetText(string)
}

// Menu is the menu to handle the all menus in bot
type Menu interface {
	Reset(ctx context.Context) Response
	Process(ctx context.Context, message string) Response
}

// NewButtonResponse create new button list
func NewButtonResponse(text string, items ...string) Response {
	return &button{
		keys: items,
		text: text,
	}
}

// NewTextResponse create a new text message
func NewTextResponse(text string, hideButton bool) Response {
	return &message{
		hideButton: hideButton,
		text:       text,
	}
}

func NewMultiResponse(resp ...Response) Response {
	return &multi{
		resp: resp,
	}
}

func NewInlineResponse(text string, items ...string) Response {
	return &button{
		keys: items,
		text: text,
		inl:  true,
	}
}
