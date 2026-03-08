package validation

type Level string

const (
	LevelError   Level = "ERROR"
	LevelWarning Level = "WARNING"
	LevelInfo    Level = "INFO"
)

type Issue struct {
	Level   Level  `json:"level"`
	Path    string `json:"path"`
	Message string `json:"message"`
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
}

type Report struct {
	Valid   bool    `json:"valid"`
	Issues  []Issue `json:"issues"`
	Summary struct {
		Errors   int `json:"errors"`
		Warnings int `json:"warnings"`
		Info     int `json:"info"`
	} `json:"summary"`
}
