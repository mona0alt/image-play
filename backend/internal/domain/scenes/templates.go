package scenes

import "strings"

type FormField struct {
	Name     string   `json:"name"`
	Label    string   `json:"label"`
	Type     string   `json:"type"`
	Required bool     `json:"required,omitempty"`
	Options  []string `json:"options,omitempty"`
}

type FormSchema []FormField

type PromptPreset struct {
	BasePrompt string   `json:"base_prompt"`
	StyleWords []string `json:"style_words,omitempty"`
}

func (p PromptPreset) IsUsable() bool {
	return strings.TrimSpace(p.BasePrompt) != ""
}

type Template struct {
	ID             int64        `json:"id"`
	SceneKey       string       `json:"scene_key"`
	TemplateKey    string       `json:"template_key"`
	Name           string       `json:"name"`
	FormSchema     FormSchema   `json:"form_schema"`
	PromptPreset   PromptPreset `json:"prompt_preset"`
	SampleImageURL string       `json:"sample_image_url,omitempty"`
	Active         bool         `json:"active"`
}
