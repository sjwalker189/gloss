package parser

import (
	"gloss/token"
)

type Severity int

const (
	SeverityError Severity = iota
	SeverityWarn
	SeverityInfo
)

var severityName = map[Severity]string{
	SeverityError: "Error",
	SeverityWarn:  "Warn",
	SeverityInfo:  "Info",
}

func (s Severity) String() string {
	return severityName[s]
}

type Diagnostic struct {
	Line     int
	Column   int
	Text     string
	Severiry Severity
}

type DiagnosticList struct {
	Items []Diagnostic
}

func (dl *DiagnosticList) Error(t token.Token, msg string) {
	dl.Items = append(dl.Items, Diagnostic{
		Line:     t.Line,
		Column:   t.Column,
		Text:     msg,
		Severiry: SeverityError,
	})
}

func (dl *DiagnosticList) Warn(t token.Token, msg string) {
	dl.Items = append(dl.Items, Diagnostic{
		Line:     t.Line,
		Column:   t.Column,
		Text:     msg,
		Severiry: SeverityWarn,
	})
}

func (dl *DiagnosticList) Raise(t token.Token, msg string) {
	dl.Items = append(dl.Items, Diagnostic{
		Line:     t.Line,
		Column:   t.Column,
		Text:     msg,
		Severiry: SeverityInfo,
	})
}
