---
AIGC:
    ContentProducer: Minimax Agent AI
    ContentPropagator: Minimax Agent AI
    Label: AIGC
    ProduceID: "00000000000000000000000000000000"
    PropagateID: "00000000000000000000000000000000"
    ReservedCode1: 30450221009b56ec0083a93d554c9d162b2292c1b1e73f192b301fb62151d30146d71e1cf50220584bcaf9c947545bbd4071680b96a8919134340fb4df827dc0f304e3d6bc6836
    ReservedCode2: 3044022025b6e31454bc280c466084898803a53e50dfc88b8a9e85b3f6459610d1573bb6022025a1e2eed1a77d2009b752e8d79984904417d5b7c9267aa1684009c5e20315b7
---

# 自定义技能说明

## 概述

用户可以在此目录下添加自定义技能。

## 创建自定义技能

1. 创建一个新的 Go 文件，例如 `custom_skill.go`
2. 实现 `skills.Skill` 接口：

```go
package custom

import (
    "context"
    "github.com/jiujuan/wukong-ai/internal/skills"
)

type CustomSkill struct {
    // 添加依赖
}

func NewCustomSkill() *CustomSkill {
    return &CustomSkill{}
}

func (s *CustomSkill) Name() string {
    return "custom_skill"
}

func (s *CustomSkill) Description() string {
    return "Description of your custom skill"
}

func (s *CustomSkill) Execute(ctx context.Context, input string) (string, error) {
    // 实现技能逻辑
    return "result", nil
}

func (s *CustomSkill) GetPrompt() string {
    return "System prompt for this skill"
}

// 确保实现接口
var _ skills.Skill = (*CustomSkill)(nil)
```

3. 在 `skill_registry.go` 中注册技能

## 示例技能

参考 `../basic/` 目录下的基础技能实现。
