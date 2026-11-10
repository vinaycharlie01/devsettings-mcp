package domain

type DocEntry struct {
	Title    string       `json:"title"`
	URL      string       `json:"url"`
	Editor   EditorID     `json:"editor,omitempty"`
	Topic    string       `json:"topic"`
	Sections []DocSection `json:"sections,omitempty"`
}

type DocSection struct {
	Heading string `json:"heading"`
	Content string `json:"content"`
	URL     string `json:"url,omitempty"`
}

type DocSearchResult struct {
	Entries []DocEntry `json:"entries"`
	Query   string     `json:"query"`
}
