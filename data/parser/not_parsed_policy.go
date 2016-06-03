package parser

type NotParsedPolicy int

const (
	IgnoreNotParsed NotParsedPolicy = iota
	WarnLoggerNotParsed
	AppendErrorNotParsed
)
