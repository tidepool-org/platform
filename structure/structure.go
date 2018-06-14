package structure

import "strings"

type Origin int

const (
	OriginStore Origin = iota
	OriginInternal
	OriginExternal
)

func Origins() []Origin {
	return []Origin{
		OriginStore,
		OriginInternal,
		OriginExternal,
	}
}

type OriginReporter interface {
	Origin() Origin
}

type SourceReporter interface {
	HasSource() bool
	Source() Source
}

type MetaReporter interface {
	HasMeta() bool
	Meta() interface{}
}

type ErrorReporter interface {
	HasError() bool
	Error() error
	ReportError(err error)
}

type Source interface {
	Parameter() string
	Pointer() string

	WithReference(reference string) Source
}

type ParameterSource struct {
	parameter string
}

func NewParameterSource() *ParameterSource {
	return &ParameterSource{}
}

func (p *ParameterSource) Parameter() string {
	return p.parameter
}

func (p *ParameterSource) Pointer() string {
	return ""
}

func (p *ParameterSource) WithReference(reference string) Source {
	if p.parameter != "" {
		return p
	}

	return &ParameterSource{
		parameter: reference,
	}
}

type PointerSource struct {
	pointer string
}

func NewPointerSource() *PointerSource {
	return &PointerSource{}
}

func (p *PointerSource) Parameter() string {
	return ""
}

func (p *PointerSource) Pointer() string {
	return p.pointer
}

func (p *PointerSource) WithReference(reference string) Source {
	return &PointerSource{
		pointer: p.pointer + "/" + EncodePointerReference(reference),
	}
}

func EncodePointerReference(reference string) string {
	return strings.Replace(strings.Replace(reference, "~", "~0", -1), "/", "~1", -1)
}
