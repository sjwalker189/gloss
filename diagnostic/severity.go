package diagnostic

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
