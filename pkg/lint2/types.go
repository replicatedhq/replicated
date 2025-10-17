package lint2

// LintResult represents the outcome of linting a chart
type LintResult struct {
	Success  bool
	Messages []LintMessage
}

// LintMessage represents a single finding from helm lint
type LintMessage struct {
	Severity string // "ERROR", "WARNING", "INFO"
	Path     string // File path (if provided by helm)
	Message  string // The lint message
}
