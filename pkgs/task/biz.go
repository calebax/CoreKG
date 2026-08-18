package task

type CommonPayload struct {
	TaskType string `json:"task_type"`
	Timeout  int64  `json:"timeout"`
}
