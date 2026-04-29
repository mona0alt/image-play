package llm

import (
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToSchemaMessages_TextOnly(t *testing.T) {
	msgs := []Message{
		{Role: "user", Parts: []Part{{Type: PartTypeText, Content: "hello"}}},
	}
	result, err := toSchemaMessages(msgs)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, schema.User, result[0].Role)
	assert.Equal(t, "hello", result[0].Content)
	assert.Empty(t, result[0].UserInputMultiContent)
}

func TestToSchemaMessages_Multimodal(t *testing.T) {
	msgs := []Message{
		{
			Role: "user",
			Parts: []Part{
				{Type: PartTypeText, Content: "analyze this"},
				{Type: PartTypeImage, Content: "data:image/png;base64,abc"},
			},
		},
	}
	result, err := toSchemaMessages(msgs)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, schema.User, result[0].Role)
	assert.Empty(t, result[0].Content)
	require.Len(t, result[0].UserInputMultiContent, 2)
	assert.Equal(t, schema.ChatMessagePartTypeText, result[0].UserInputMultiContent[0].Type)
	assert.Equal(t, "analyze this", result[0].UserInputMultiContent[0].Text)
	assert.Equal(t, schema.ChatMessagePartTypeImageURL, result[0].UserInputMultiContent[1].Type)
	assert.Equal(t, "data:image/png;base64,abc", *result[0].UserInputMultiContent[1].Image.URL)
}

func TestToSchemaMessages_UnsupportedPartType(t *testing.T) {
	msgs := []Message{
		{Role: "user", Parts: []Part{{Type: PartType("audio"), Content: "x"}}},
	}
	_, err := toSchemaMessages(msgs)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported part type")
}
