package port

import "github.com/slack-go/slack"

const (
	PlainText      = "plain_text"
	PlainTextInput = "plain_text_input"
)

type Block struct {
	ID             string `json:"block_id,omitempty"`
	Type           string
	DispatchAction bool `json:",omitempty"`

	// Sub-structs
	Label *Element `json:",omitempty"`

	Element  *Element   `json:",omitempty"`
	Elements []*Element `json:",omitempty"`
}

func (b Block) BlockType() slack.MessageBlockType {
	return slack.MessageBlockType(b.Type)
}

type Element struct {
	ActionID string `json:",omitempty"`
	Type     string

	// Element properties for TextInput
	Text         string
	Emoji        bool   `json:",omitempty"`
	InitialValue string `json:",omitempty"`
	Multiline    bool
	MinLength    int `json:",omitempty"`
	MaxLength    int `json:",omitempty"`
}

func (e Element) ElementType() slack.MessageElementType {
	return slack.MessageElementType(e.Type)
}
