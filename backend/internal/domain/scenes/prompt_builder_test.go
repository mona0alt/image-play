package scenes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildPromptForInvitation(t *testing.T) {
	input := BuildInput{
		SceneKey:    "invitation",
		TemplateKey: "wedding-classic",
		Fields: map[string]string{
			"host_name":   "林然与苏晴",
			"event_time":  "2026-10-01 18:00",
			"event_place": "杭州西湖国宾馆",
		},
	}

	prompt := BuildPrompt(input)
	require.Contains(t, prompt, "婚礼请柬")
	require.Contains(t, prompt, "林然与苏晴")
	require.Contains(t, prompt, "elegant")
}

func TestBuildPromptForPortrait(t *testing.T) {
	input := BuildInput{
		SceneKey:    "portrait",
		TemplateKey: "office-pro",
		Fields: map[string]string{
			"subject_name": "张三",
			"position":     "高级产品经理",
		},
	}

	prompt := BuildPrompt(input)
	require.Contains(t, prompt, "职业形象照")
	require.Contains(t, prompt, "张三")
	require.Contains(t, prompt, "高级产品经理")
	require.Contains(t, prompt, "professional")
}

func TestBuildPromptEmptyFields(t *testing.T) {
	input := BuildInput{
		SceneKey:    "festival",
		TemplateKey: "spring-festival",
		Fields:      map[string]string{},
	}

	prompt := BuildPrompt(input)
	require.Contains(t, prompt, "春节贺卡")
	require.NotContains(t, prompt, "自定义信息")
}

func TestBuildPromptCombinesAllParts(t *testing.T) {
	input := BuildInput{
		SceneKey:    SceneInvitation,
		TemplateKey: "wedding-classic",
		Fields: map[string]string{
			"host_name": "Alice",
		},
	}

	prompt := BuildPrompt(input)
	require.Contains(t, prompt, "婚礼请柬")
	require.Contains(t, prompt, "elegant")
	require.Contains(t, prompt, "Alice")
}

func TestBuildPromptUnknownTemplate(t *testing.T) {
	input := BuildInput{
		SceneKey:    "invitation",
		TemplateKey: "nonexistent-template",
		Fields: map[string]string{
			"host_name": "Alice",
		},
	}

	prompt := BuildPrompt(input)
	require.Contains(t, prompt, "场景：invitation")
	require.Contains(t, prompt, "模板：nonexistent-template")
	require.Contains(t, prompt, "host_name=Alice")
}

func TestBuildPromptNilFields(t *testing.T) {
	input := BuildInput{
		SceneKey:    "invitation",
		TemplateKey: "wedding-classic",
		Fields:      nil,
	}

	prompt := BuildPrompt(input)
	require.Contains(t, prompt, "婚礼请柬")
	require.NotContains(t, prompt, "自定义信息")
}

func TestBuildPromptUnknownScene(t *testing.T) {
	input := BuildInput{
		SceneKey:    "unknown-scene",
		TemplateKey: "unknown-template",
		Fields: map[string]string{
			"foo": "bar",
		},
	}

	prompt := BuildPrompt(input)
	require.Contains(t, prompt, "unknown-scene")
	require.Contains(t, prompt, "unknown-template")
	require.Contains(t, prompt, "foo=bar")
}
