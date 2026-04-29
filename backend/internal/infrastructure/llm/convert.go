package llm

import (
	"fmt"

	"github.com/cloudwego/eino/schema"
)

func toSchemaMessages(messages []Message) ([]*schema.Message, error) {
	result := make([]*schema.Message, 0, len(messages))
	for _, msg := range messages {
		m := &schema.Message{
			Role: schema.RoleType(msg.Role),
		}
		if len(msg.Parts) == 0 {
			result = append(result, m)
			continue
		}
		if len(msg.Parts) == 1 && msg.Parts[0].Type == PartTypeText {
			m.Content = msg.Parts[0].Content
			result = append(result, m)
			continue
		}
		parts := make([]schema.MessageInputPart, 0, len(msg.Parts))
		for _, part := range msg.Parts {
			switch part.Type {
			case PartTypeText:
				parts = append(parts, schema.MessageInputPart{
					Type: schema.ChatMessagePartTypeText,
					Text: part.Content,
				})
			case PartTypeImage:
				parts = append(parts, schema.MessageInputPart{
					Type: schema.ChatMessagePartTypeImageURL,
					Image: &schema.MessageInputImage{
						MessagePartCommon: schema.MessagePartCommon{
							URL: &part.Content,
						},
					},
				})
			default:
				return nil, fmt.Errorf("llm: unsupported part type %q", part.Type)
			}
		}
		m.UserInputMultiContent = parts
		result = append(result, m)
	}
	return result, nil
}
