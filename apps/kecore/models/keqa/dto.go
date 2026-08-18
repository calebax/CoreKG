package keqa

import "encoding/json"

type FlagAnswer string

const (
	FlagThinking  FlagAnswer = "thinking"
	FlagSearching FlagAnswer = "searching"
	FlagFound     FlagAnswer = "found"
	FlagAnswering FlagAnswer = "answering"
)

type WriteResult struct {
	ReasoningContent string     `json:"reasoning_content"`
	Content          string     `json:"content"`
	ReasoningSeconds float64    `json:"reasoning_seconds"`
	Flag             FlagAnswer `json:"flag"`
	Reference        Reference  `json:"reference"`
}

func (w WriteResult) String() string {
	jsonData, err := json.Marshal(w)
	if err != nil {
		return "\n"
	}
	return string(jsonData) + "\n"
}

type Reference struct {
	ForestID uint   `json:"forest_id"`
	FileID   uint   `json:"file_id"`
	Name     string `json:"name"`
}
