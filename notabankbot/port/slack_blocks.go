package port

import "github.com/slack-go/slack"

const (
	PlainText      = "plain_text"
	PlainTextInput = "plain_text_input"
)

type Block struct {
	ID             string `json:"block_id,omitempty"`
	Type           string `json:"type"`
	DispatchAction bool   `json:"dispatch_action,omitempty"`

	// Sub-structs
	Label *Element `json:"label,omitempty"`

	Element  *Element   `json:"element,omitempty"`
	Elements []*Element `json:"elements,omitempty"`
}

func (b Block) BlockType() slack.MessageBlockType {
	return slack.MessageBlockType(b.Type)
}

type Element struct {
	ActionID string `json:"action_id,omitempty"`
	Type     string `json:"type"`

	// Element properties for TextInput
	Text         string `json:"text"`
	Emoji        bool   `json:"emoji,omitempty"`
	InitialValue string `json:"initial_value,omitempty"`
	Multiline    bool   `json:"multiline"`
	MinLength    int    `json:"min_length,omitempty"`
	MaxLength    int    `json:"max_length,omitempty"`
}

func (e Element) ElementType() slack.MessageElementType {
	return slack.MessageElementType(e.Type)
}
