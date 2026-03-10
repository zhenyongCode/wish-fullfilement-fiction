package task

// Task represents a unit of work that can be executed by an agent.
type Task struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Status      string `json:"status"` // e.g., "pending", "in_progress", "completed"
	Type        string `json:"type"`
	Result      string `json:"result,omitempty"`
}
