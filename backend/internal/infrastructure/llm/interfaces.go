package llm

import (
	"context"
	"time"
)

type PartType string

const (
	PartTypeText  PartType = "text"
	PartTypeImage PartType = "image"
)

type Part struct {
	Type    PartType
	Content string
}

type Message struct {
	Role  string
	Parts []Part
}

type Chunk struct {
	Content string
}

// StreamReader 流式读取模型输出
// Recv 返回 io.EOF 表示流正常结束
// 调用方必须在使用完成后调用 Close
type StreamReader interface {
	Recv() (Chunk, error)
	Close()
}

// TextClient 文本/多模态大模型客户端
type TextClient interface {
	Chat(ctx context.Context, messages []Message) (string, error)
	ChatStream(ctx context.Context, messages []Message) (StreamReader, error)
}

// ImageClient 文生图大模型客户端
type ImageClient interface {
	Generate(ctx context.Context, prompt string) (imageURL string, err error)
}

type TextConfig struct {
	APIKey  string
	BaseURL string
	Model   string
	Timeout time.Duration
}

type ImageConfig struct {
	APIKey  string
	BaseURL string
	Model   string
	Timeout time.Duration
}
