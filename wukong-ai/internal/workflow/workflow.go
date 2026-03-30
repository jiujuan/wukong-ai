package workflow

// Node 接口定义
type Node interface {
	// Name 返回节点名称
	Name() string
	// Run 执行节点逻辑
	Run(ctx *WukongContext) error
}

// Workflow 工作流结构体
type Workflow struct {
	Start     Node              // 起始节点
	Nodes     map[string]Node   // 所有节点
	Edges     map[string][]Node // 节点名 → 后继节点列表
	TotalSteps int              // 总步骤数
}

// NewWorkflow 创建新的工作流
func NewWorkflow() *Workflow {
	return &Workflow{
		Nodes: make(map[string]Node),
		Edges: make(map[string][]Node),
	}
}

// AddNode 添加节点
func (wf *Workflow) AddNode(node Node) {
	wf.Nodes[node.Name()] = node
}

// AddEdge 添加边（从前驱节点指向后继节点）
func (wf *Workflow) AddEdge(from, to string) {
	wf.Edges[from] = append(wf.Edges[from], wf.Nodes[to])
}

// SetStart 设置起始节点
func (wf *Workflow) SetStart(node Node) {
	wf.Start = node
}

// GetNode 获取节点
func (wf *Workflow) GetNode(name string) (Node, bool) {
	node, ok := wf.Nodes[name]
	return node, ok
}

// GetSuccessors 获取节点的后继节点
func (wf *Workflow) GetSuccessors(nodeName string) []Node {
	return wf.Edges[nodeName]
}

// CalculateTotalSteps 计算总步骤数
func (wf *Workflow) CalculateTotalSteps() int {
	if wf.Start == nil {
		return 0
	}

	visited := make(map[string]bool)
	var count int
	wf.countSteps(wf.Start, visited, &count)
	wf.TotalSteps = count
	return count
}

func (wf *Workflow) countSteps(node Node, visited map[string]bool, count *int) {
	if node == nil || visited[node.Name()] {
		return
	}
	visited[node.Name()] = true
	*count++

	for _, successor := range wf.Edges[node.Name()] {
		wf.countSteps(successor, visited, count)
	}
}
