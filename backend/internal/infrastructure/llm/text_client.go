package llm

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type textClient struct {
	chatModel model.ChatModel
}

func NewTextClient(cfg TextConfig) (TextClient, error) {
	ctx := context.Background()
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: cfg.BaseURL,
		Model:   cfg.Model,
		APIKey:  cfg.APIKey,
		Timeout: cfg.Timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("llm: create text client: %w", err)
	}
	return &textClient{chatModel: chatModel}, nil
}

func (c *textClient) Chat(ctx context.Context, messages []Message) (string, error) {
	inputs, err := toSchemaMessages(messages)
	if err != nil {
		return "", err
	}
	resp, err := c.chatModel.Generate(ctx, inputs)
	if err != nil {
		return "", fmt.Errorf("llm: chat generate failed: %w", err)
	}
	return resp.Content, nil
}

func (c *textClient) ChatStream(ctx context.Context, messages []Message) (StreamReader, error) {
	inputs, err := toSchemaMessages(messages)
	if err != nil {
		return nil, err
	}
	stream, err := c.chatModel.Stream(ctx, inputs)
	if err != nil {
		return nil, fmt.Errorf("llm: chat stream failed: %w", err)
	}
	return &textStreamReader{stream: stream}, nil
}

type textStreamReader struct {
	stream *schema.StreamReader[*schema.Message]
}

func (r *textStreamReader) Recv() (Chunk, error) {
	msg, err := r.stream.Recv()
	if err != nil {
		return Chunk{}, err
	}
	return Chunk{Content: msg.Content}, nil
}

func (r *textStreamReader) Close() {
	r.stream.Close()
}
