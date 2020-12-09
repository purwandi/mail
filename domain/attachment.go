package domain

// Attachment ...
type Attachment struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	Filepath string `json:"filepath"`
	Type     string `json:"type"`
}
