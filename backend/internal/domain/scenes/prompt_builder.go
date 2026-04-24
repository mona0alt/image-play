package scenes

import (
	"fmt"
	"sort"
	"strings"
)

type BuildInput struct {
	SceneKey    string
	TemplateKey string
	Fields      map[string]string
}

func BuildPrompt(input BuildInput) string {
	sceneTemplates, ok := sceneTemplates[input.SceneKey]
	if !ok {
		return buildFallbackPrompt(input)
	}

	template, ok := sceneTemplates[input.TemplateKey]
	if !ok {
		return buildFallbackPrompt(input)
	}

	var parts []string
	parts = append(parts, template.BasePrompt)

	if len(template.StyleWords) > 0 {
		parts = append(parts, "风格关键词："+strings.Join(template.StyleWords, ", "))
	}

	if len(input.Fields) > 0 {
		var fieldParts []string
		for k, v := range input.Fields {
			fieldParts = append(fieldParts, fmt.Sprintf("%s=%s", k, v))
		}
		sort.Strings(fieldParts)
		parts = append(parts, "自定义信息："+strings.Join(fieldParts, "; "))
	}

	return strings.Join(parts, "\n")
}

func buildFallbackPrompt(input BuildInput) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("场景：%s，模板：%s", input.SceneKey, input.TemplateKey))

	if len(input.Fields) > 0 {
		var fieldParts []string
		for k, v := range input.Fields {
			fieldParts = append(fieldParts, fmt.Sprintf("%s=%s", k, v))
		}
		sort.Strings(fieldParts)
		parts = append(parts, "自定义信息："+strings.Join(fieldParts, "; "))
	}

	return strings.Join(parts, "\n")
}
