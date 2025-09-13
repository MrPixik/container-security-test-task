package models

type EnqueueReq struct {
	Id         string `json:"id,omitempty"`
	Payload    string `json:"payload,omitempty"`
	MaxRetries int    `json:"max-retries,omitempty"`
}

type EnqueueRes struct {
	TaskId int `json:"task_id"`
}
