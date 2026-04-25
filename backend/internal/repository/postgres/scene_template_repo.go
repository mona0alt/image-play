package postgres

import (
	"context"
	"database/sql"
	"encoding/json"

	"image-play/internal/domain/scenes"
)

type SceneTemplateRepo struct {
	db *sql.DB
}

func NewSceneTemplateRepo(db *sql.DB) *SceneTemplateRepo {
	return &SceneTemplateRepo{db: db}
}

func (r *SceneTemplateRepo) ListActiveByScene(ctx context.Context, sceneKey string) ([]scenes.Template, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, scene_key, template_key, name, form_schema, prompt_preset, sample_image_url, is_active
		FROM scene_templates
		WHERE scene_key = $1 AND is_active = TRUE
		ORDER BY id ASC
	`, sceneKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]scenes.Template, 0)
	for rows.Next() {
		template, err := scanSceneTemplateRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *template)
	}
	return items, rows.Err()
}

func (r *SceneTemplateRepo) GetActiveTemplate(ctx context.Context, sceneKey, templateKey string) (*scenes.Template, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, scene_key, template_key, name, form_schema, prompt_preset, sample_image_url, is_active
		FROM scene_templates
		WHERE scene_key = $1 AND template_key = $2 AND is_active = TRUE
		LIMIT 1
	`, sceneKey, templateKey)

	template, err := scanSceneTemplate(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return template, nil
}

func scanSceneTemplateRow(rows *sql.Rows) (*scenes.Template, error) {
	var (
		template        scenes.Template
		formSchemaRaw   []byte
		promptPresetRaw []byte
		sampleImageURL  sql.NullString
	)

	err := rows.Scan(
		&template.ID,
		&template.SceneKey,
		&template.TemplateKey,
		&template.Name,
		&formSchemaRaw,
		&promptPresetRaw,
		&sampleImageURL,
		&template.Active,
	)
	if err != nil {
		return nil, err
	}

	if err := hydrateTemplate(&template, formSchemaRaw, promptPresetRaw, sampleImageURL); err != nil {
		return nil, err
	}
	return &template, nil
}

func scanSceneTemplate(row *sql.Row) (*scenes.Template, error) {
	var (
		template        scenes.Template
		formSchemaRaw   []byte
		promptPresetRaw []byte
		sampleImageURL  sql.NullString
	)

	err := row.Scan(
		&template.ID,
		&template.SceneKey,
		&template.TemplateKey,
		&template.Name,
		&formSchemaRaw,
		&promptPresetRaw,
		&sampleImageURL,
		&template.Active,
	)
	if err != nil {
		return nil, err
	}

	if err := hydrateTemplate(&template, formSchemaRaw, promptPresetRaw, sampleImageURL); err != nil {
		return nil, err
	}
	return &template, nil
}

func hydrateTemplate(template *scenes.Template, formSchemaRaw, promptPresetRaw []byte, sampleImageURL sql.NullString) error {
	if len(formSchemaRaw) > 0 {
		if err := json.Unmarshal(formSchemaRaw, &template.FormSchema); err != nil {
			return err
		}
	}
	if len(promptPresetRaw) > 0 {
		if err := json.Unmarshal(promptPresetRaw, &template.PromptPreset); err != nil {
			return err
		}
	}
	if sampleImageURL.Valid {
		template.SampleImageURL = sampleImageURL.String
	}
	return nil
}
