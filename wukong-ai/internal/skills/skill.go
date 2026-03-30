package skills

import (
	"context"
)

// Skill 技能接口
type Skill interface {
	// Name 返回技能名称
	Name() string
	// Description 返回技能描述
	Description() string
	// Execute 执行技能
	Execute(ctx context.Context, input string) (string, error)
	// GetPrompt 获取系统提示词
	GetPrompt() string
}

// SkillRegistry 技能注册表
type SkillRegistry struct {
	skills map[string]Skill
}

// NewSkillRegistry 创建技能注册表
func NewSkillRegistry() *SkillRegistry {
	return &SkillRegistry{
		skills: make(map[string]Skill),
	}
}

// Register 注册技能
func (r *SkillRegistry) Register(s Skill) {
	r.skills[s.Name()] = s
}

// Get 获取技能
func (r *SkillRegistry) Get(name string) (Skill, bool) {
	s, ok := r.skills[name]
	return s, ok
}

// List 获取所有技能
func (r *SkillRegistry) List() []Skill {
	skills := make([]Skill, 0, len(r.skills))
	for _, s := range r.skills {
		skills = append(skills, s)
	}
	return skills
}

// GetNames 获取所有技能名称
func (r *SkillRegistry) GetNames() []string {
	names := make([]string, 0, len(r.skills))
	for name := range r.skills {
		names = append(names, name)
	}
	return names
}

// LoadSkillsFromDB 从数据库加载技能
func (r *SkillRegistry) LoadSkillsFromDB() error {
	// 从数据库加载技能并注册
	// 简化实现
	return nil
}
