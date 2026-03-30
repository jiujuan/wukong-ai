package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jiujuan/wukong-ai/internal/db/repository"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// DagHandler DAG 处理器
type DagHandler struct{}

// NewDagHandler 创建 DAG 处理器
func NewDagHandler() *DagHandler {
	return &DagHandler{}
}

// DagResponse DAG 响应
type DagResponse struct {
	TaskID string     `json:"task_id"`
	Mode   string     `json:"mode"`
	Nodes  []DagNode  `json:"nodes"`
	Edges  []DagEdge  `json:"edges"`
}

// DagNode DAG 节点
type DagNode struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Status string `json:"status"`
}

// DagEdge DAG 边
type DagEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

// Handle 处理 DAG 请求
func (h *DagHandler) Handle(c *gin.Context) {
	taskID := c.Query("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task_id is required"})
		return
	}

	task, err := repository.GetTaskByID(taskID)
	if err != nil {
		logger.Error("failed to get task", "task_id", taskID, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get task"})
		return
	}

	if task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	// 根据模式确定节点和边
	nodes, edges := h.buildDagForMode(task.Mode, task.Status, task.LastNode.String)

	c.JSON(http.StatusOK, DagResponse{
		TaskID: taskID,
		Mode:   task.Mode,
		Nodes:  nodes,
		Edges:  edges,
	})
}

// buildDagForMode 根据模式构建 DAG
func (h *DagHandler) buildDagForMode(mode, taskStatus, lastNode string) ([]DagNode, []DagEdge) {
	var nodes []DagNode
	var edges []DagEdge

	switch mode {
	case "flash":
		nodes = []DagNode{
			{h.nodeID("coordinator"), "Coordinator", h.getNodeStatus("coordinator", taskStatus, lastNode)},
		}
		edges = []DagEdge{}

	case "standard":
		nodes = []DagNode{
			{h.nodeID("coordinator"), "Coordinator", h.getNodeStatus("coordinator", taskStatus, lastNode)},
			{h.nodeID("background"), "Background", h.getNodeStatus("background", taskStatus, lastNode)},
			{h.nodeID("researcher"), "Researcher", h.getNodeStatus("researcher", taskStatus, lastNode)},
			{h.nodeID("reporter"), "Reporter", h.getNodeStatus("reporter", taskStatus, lastNode)},
		}
		edges = []DagEdge{
			{"coordinator", "background"},
			{"background", "researcher"},
			{"researcher", "reporter"},
		}

	case "pro":
		nodes = []DagNode{
			{h.nodeID("coordinator"), "Coordinator", h.getNodeStatus("coordinator", taskStatus, lastNode)},
			{h.nodeID("planner"), "Planner", h.getNodeStatus("planner", taskStatus, lastNode)},
			{h.nodeID("researcher"), "Researcher", h.getNodeStatus("researcher", taskStatus, lastNode)},
			{h.nodeID("reporter"), "Reporter", h.getNodeStatus("reporter", taskStatus, lastNode)},
		}
		edges = []DagEdge{
			{"coordinator", "planner"},
			{"planner", "researcher"},
			{"researcher", "reporter"},
		}

	case "ultra":
		nodes = []DagNode{
			{h.nodeID("coordinator"), "Coordinator", h.getNodeStatus("coordinator", taskStatus, lastNode)},
			{h.nodeID("planner"), "Planner", h.getNodeStatus("planner", taskStatus, lastNode)},
			{h.nodeID("subagentmanager"), "SubAgentManager", h.getNodeStatus("subagentmanager", taskStatus, lastNode)},
			{h.nodeID("reporter"), "Reporter", h.getNodeStatus("reporter", taskStatus, lastNode)},
		}
		edges = []DagEdge{
			{"coordinator", "planner"},
			{"planner", "subagentmanager"},
			{"subagentmanager", "reporter"},
		}

	default:
		nodes = []DagNode{
			{h.nodeID("coordinator"), "Coordinator", h.getNodeStatus("coordinator", taskStatus, lastNode)},
			{h.nodeID("reporter"), "Reporter", h.getNodeStatus("reporter", taskStatus, lastNode)},
		}
		edges = []DagEdge{
			{"coordinator", "reporter"},
		}
	}

	return nodes, edges
}

func (h *DagHandler) nodeID(name string) string {
	return name
}

func (h *DagHandler) getNodeStatus(nodeName, taskStatus, lastNode string) string {
	if taskStatus == "pending" || taskStatus == "queued" {
		return "pending"
	}

	if taskStatus == "success" {
		return "success"
	}

	if taskStatus == "failed" {
		return "failed"
	}

	// running 状态
	nodeOrder := []string{"coordinator", "background", "planner", "researcher", "subagentmanager", "reporter"}

	lastIdx := -1
	for i, n := range nodeOrder {
		if n == lastNode {
			lastIdx = i
			break
		}
	}

	for i, n := range nodeOrder {
		if n == nodeName {
			if i < lastIdx {
				return "success"
			}
			if i == lastIdx {
				return "running"
			}
			return "pending"
		}
	}

	return "pending"
}
