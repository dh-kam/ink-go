package tuitest

type Transcript struct {
	SchemaVersion string            `json:"schemaVersion"`
	Scenario      string            `json:"scenario"`
	Runtime       string            `json:"runtime"`
	App           string            `json:"app"`
	Viewport      Viewport          `json:"viewport"`
	Environment   map[string]string `json:"environment,omitempty"`
	Frames        []TranscriptFrame `json:"frames"`
	Exit          *TranscriptExit   `json:"exit,omitempty"`
}

type TranscriptFrame struct {
	Index       int    `json:"index"`
	Step        string `json:"step"`
	Input       string `json:"input,omitempty"`
	Raw         string `json:"raw"`
	RawEscaped  string `json:"rawEscaped"`
	Plain       string `json:"plain"`
	ScreenPlain string `json:"screenPlain"`
}

type TranscriptExit struct {
	Step       string `json:"step,omitempty"`
	OK         bool   `json:"ok"`
	DurationMS int64  `json:"durationMs"`
	Error      string `json:"error,omitempty"`
}

const TranscriptSchemaVersion = "goink.tuitest.transcript/v1alpha1"
