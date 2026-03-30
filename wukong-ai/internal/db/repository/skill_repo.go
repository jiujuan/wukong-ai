package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/jiujuan/wukong-ai/internal/db"
)

// Skill 技能结构体
type Skill struct {
	ID          int64           `json:"id"`
	Name        string          `json:"name"`
	Description sql.NullString  `json:"description"`
	SkillType   string          `json:"skill_type"`
	Config      json.RawMessage `json:"config"`
	Enabled     bool            `json:"enabled"`
	CreateTime  sql.NullString  `json:"create_time"`
}

// CreateSkill 创建技能
func CreateSkill(skill *Skill) (int64, error) {
	db := db.Get()
	query := `
		INSERT INTO skills (name, description, skill_type, config, enabled, create_time)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id
	`
	var id int64
	err := db.QueryRow(query,
		skill.Name, skill.Description, skill.SkillType, skill.Config, skill.Enabled,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create skill: %w", err)
	}
	return id, nil
}

// GetAllSkills 获取所有启用的技能
func GetAllSkills() ([]*Skill, error) {
	db := db.Get()
	query := `
		SELECT id, name, description, skill_type, config, enabled, create_time
		FROM skills
		WHERE enabled = true
		ORDER BY skill_type ASC, name ASC
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get skills: %w", err)
	}
	defer rows.Close()

	var skills []*Skill
	for rows.Next() {
		var skill Skill
		err := rows.Scan(
			&skill.ID, &skill.Name, &skill.Description, &skill.SkillType,
			&skill.Config, &skill.Enabled, &skill.CreateTime,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan skill: %w", err)
		}
		skills = append(skills, &skill)
	}

	return skills, nil
}

// GetSkillByName 根据名称获取技能
func GetSkillByName(name string) (*Skill, error) {
	db := db.Get()
	query := `
		SELECT id, name, description, skill_type, config, enabled, create_time
		FROM skills
		WHERE name = $1
	`
	var skill Skill
	err := db.QueryRow(query, name).Scan(
		&skill.ID, &skill.Name, &skill.Description, &skill.SkillType,
		&skill.Config, &skill.Enabled, &skill.CreateTime,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get skill: %w", err)
	}
	return &skill, nil
}

// UpdateSkill 更新技能
func UpdateSkill(skill *Skill) error {
	db := db.Get()
	query := `
		UPDATE skills SET
			description = $1,
			config = $2,
			enabled = $3
		WHERE name = $4
	`
	_, err := db.Exec(query, skill.Description, skill.Config, skill.Enabled, skill.Name)
	if err != nil {
		return fmt.Errorf("failed to update skill: %w", err)
	}
	return nil
}

// DeleteSkill 删除技能
func DeleteSkill(name string) error {
	db := db.Get()
	query := `DELETE FROM skills WHERE name = $1`
	_, err := db.Exec(query, name)
	if err != nil {
		return fmt.Errorf("failed to delete skill: %w", err)
	}
	return nil
}
