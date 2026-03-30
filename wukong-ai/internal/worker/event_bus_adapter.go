package worker

import (
	"github.com/jiujuan/wukong-ai/internal/event"
	"github.com/jiujuan/wukong-ai/internal/workflow"
)

type workflowEventBusAdapter struct {
	bus *event.EventBus
}

func NewWorkflowEventBusAdapter(bus *event.EventBus) workflow.EventBus {
	if bus == nil {
		return nil
	}
	return &workflowEventBusAdapter{bus: bus}
}

func (a *workflowEventBusAdapter) Publish(taskID string, evt workflow.ProgressEvent) {
	if a == nil || a.bus == nil {
		return
	}
	a.bus.Publish(taskID, event.ProgressEvent{
		Type:      evt.Type,
		Node:      evt.Node,
		Status:    evt.Status,
		Progress:  evt.Progress,
		Total:     evt.Total,
		Done:      evt.Done,
		Latest:    evt.Latest,
		Output:    evt.Output,
		Timestamp: evt.Timestamp,
	})
}
