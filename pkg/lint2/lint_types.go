package lint2

// LintIssue is the common interface for lint issues across all linting tools.
type LintIssue interface {
	GetLine() int
	GetColumn() int
	GetMessage() string
	GetField() string
}

// FileLintResult is the per-file result structure shared across all linting tools.
// Info uses json:"info" — the troubleshoot.sh tools do not emit an info field,
// so it will always be empty for preflight/support-bundle output.
type FileLintResult[T LintIssue] struct {
	FilePath string `json:"filePath"`
	Errors   []T    `json:"errors"`
	Warnings []T    `json:"warnings"`
	Info     []T    `json:"info"`
}

// LintOutput is the top-level JSON structure shared across all linting tools.
type LintOutput[T LintIssue] struct {
	Results []FileLintResult[T] `json:"results"`
}
