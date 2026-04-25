package scenes

import (
	"fmt"
	"sort"
	"strings"
)

type BuildInput struct {
	SceneKey    string
	TemplateKey string
	Preset      PromptPreset
	Fields      map[string]string
}

func BuildPrompt(input BuildInput) string {
	if input.Preset.BasePrompt == "" && len(input.Preset.StyleWords) == 0 {
		return buildFallbackPrompt(input)
	}

	var parts []string
	parts = append(parts, input.Preset.BasePrompt)

	if len(input.Preset.StyleWords) > 0 {
		parts = append(parts, "风格关键词："+strings.Join(input.Preset.StyleWords, ", "))
	}

	if fieldText := joinFieldParts(input.Fields); fieldText != "" {
		parts = append(parts, "自定义信息："+fieldText)
	}

	return strings.Join(parts, "\n")
}

func buildFallbackPrompt(input BuildInput) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("场景：%s，模板：%s", input.SceneKey, input.TemplateKey))

	if fieldText := joinFieldParts(input.Fields); fieldText != "" {
		parts = append(parts, "自定义信息："+fieldText)
	}

	return strings.Join(parts, "\n")
}

func joinFieldParts(fields map[string]string) string {
	if len(fields) == 0 {
		return ""
	}

	fieldParts := make([]string, 0, len(fields))
	for k, v := range fields {
		fieldParts = append(fieldParts, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(fieldParts)
	return strings.Join(fieldParts, "; ")
}
