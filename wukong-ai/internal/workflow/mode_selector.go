package workflow

import (
	"fmt"
)

// Mode 执行模式
type Mode int

const (
	ModeFlash    Mode = iota // Coordinator
	ModeStandard             // Coordinator → Background → Researcher → Reporter
	ModePro                  // Coordinator → Planner → Researcher → Reporter
	ModeUltra                // Coordinator → Planner → SubAgentManager → Reporter
)

// ModeName 返回模式名称
func (m Mode) String() string {
	switch m {
	case ModeFlash:
		return "flash"
	case ModeStandard:
		return "standard"
	case ModePro:
		return "pro"
	case ModeUltra:
		return "ultra"
	default:
		return "unknown"
	}
}

// ParseMode 解析模式字符串
func ParseMode(modeStr string) (Mode, error) {
	switch modeStr {
	case "flash":
		return ModeFlash, nil
	case "standard":
		return ModeStandard, nil
	case "pro":
		return ModePro, nil
	case "ultra":
		return ModeUltra, nil
	default:
		return ModeFlash, fmt.Errorf("unknown mode: %s", modeStr)
	}
}

// AutoSelectMode 根据配置自动选择模式
func AutoSelectMode(cfg *RunConfig) Mode {
	switch {
	case !cfg.ThinkingEnabled && !cfg.PlanEnabled && !cfg.SubAgentEnabled:
		return ModeFlash
	case cfg.ThinkingEnabled && !cfg.PlanEnabled && !cfg.SubAgentEnabled:
		return ModeStandard
	case cfg.ThinkingEnabled && cfg.PlanEnabled && !cfg.SubAgentEnabled:
		return ModePro
	case cfg.ThinkingEnabled && cfg.PlanEnabled && cfg.SubAgentEnabled:
		return ModeUltra
	default:
		return ModeFlash
	}
}

// NodeSet 节点集合
type NodeSet struct {
	Coordinator     Node
	Background      Node
	Planner         Node
	Researcher      Node
	SubAgentManager Node
	Reporter        Node
}

// BuildWorkflow 根据模式构建对应 DAG
func BuildWorkflow(mode Mode, nodes *NodeSet) *Workflow {
	wf := NewWorkflow()

	// 添加所有节点
	wf.AddNode(nodes.Coordinator)
	if nodes.Background != nil {
		wf.AddNode(nodes.Background)
	}
	if nodes.Planner != nil {
		wf.AddNode(nodes.Planner)
	}
	if nodes.Researcher != nil {
		wf.AddNode(nodes.Researcher)
	}
	if nodes.SubAgentManager != nil {
		wf.AddNode(nodes.SubAgentManager)
	}
	if nodes.Reporter != nil {
		wf.AddNode(nodes.Reporter)
	}

	// 设置起始节点
	wf.SetStart(nodes.Coordinator)

	// 根据模式构建边
	switch mode {
	case ModeFlash:
	case ModeStandard:
		wf.AddEdge("coordinator", "background")
		wf.AddEdge("background", "researcher")
		wf.AddEdge("researcher", "reporter")

	case ModePro:
		wf.AddEdge("coordinator", "planner")
		wf.AddEdge("planner", "researcher")
		wf.AddEdge("researcher", "reporter")

	case ModeUltra:
		wf.AddEdge("coordinator", "planner")
		wf.AddEdge("planner", "subagentmanager")
		wf.AddEdge("subagentmanager", "reporter")
	}

	// 计算总步骤数
	wf.CalculateTotalSteps()

	return wf
}
