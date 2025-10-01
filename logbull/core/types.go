package core

type LogLevel string

const (
	DEBUG    LogLevel = "DEBUG"
	INFO     LogLevel = "INFO"
	WARNING  LogLevel = "WARNING"
	ERROR    LogLevel = "ERROR"
	CRITICAL LogLevel = "CRITICAL"
)

type LogEntry struct {
	Level     string         `json:"level"`
	Message   string         `json:"message"`
	Timestamp string         `json:"timestamp"`
	Fields    map[string]any `json:"fields"`
}

type LogBatch struct {
	Logs []LogEntry `json:"logs"`
}

type LogBullResponse struct {
	Accepted int           `json:"accepted"`
	Rejected int           `json:"rejected"`
	Message  string        `json:"message,omitempty"`
	Errors   []RejectedLog `json:"errors,omitempty"`
}

type RejectedLog struct {
	Index   int    `json:"index"`
	Message string `json:"message"`
}

type Config struct {
	ProjectID string
	Host      string
	APIKey    string
	LogLevel  LogLevel
}

var levelPriority = map[LogLevel]int{
	DEBUG:    10,
	INFO:     20,
	WARNING:  30,
	ERROR:    40,
	CRITICAL: 50,
}

func (l LogLevel) Priority() int {
	return levelPriority[l]
}

func (l LogLevel) String() string {
	return string(l)
}
