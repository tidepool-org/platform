package errors

import (
	"path"
	"runtime"
	"strings"

	"github.com/tidepool-org/platform/structure"
)

type Caller struct {
	Package  string `json:"package,omitempty" bson:"package,omitempty"`
	Function string `json:"function,omitempty" bson:"function,omitempty"`
	File     string `json:"file,omitempty" bson:"file,omitempty"`
	Line     int    `json:"line,omitempty" bson:"line,omitempty"`
}

func GetCaller(frame int) *Caller {
	if pc, file, line, ok := runtime.Caller(frame + 1); ok {
		var pkg string
		var fnctn string

		parts := strings.SplitN(file, "/", ignoredFileSegments)
		length := len(parts)
		file = parts[length-1]

		parts = strings.Split(runtime.FuncForPC(pc).Name(), "/")
		length = len(parts)
		pkg = strings.Join(append(parts[:length-1], strings.Split(parts[length-1], ".")[0]), "/")

		return &Caller{
			Package:  pkg,
			Function: fnctn,
			File:     file,
			Line:     line,
		}
	}
	return nil
}

func (c *Caller) PackageName() string {
	parts := strings.Split(c.Package, "/")
	return parts[len(parts)-1]
}

func (c *Caller) FileName() string {
	_, fileName := path.Split(c.File)
	return fileName
}

func (c *Caller) Parse(parser structure.ObjectParser) {
	if ptr := parser.String("package"); ptr != nil {
		c.Package = *ptr
	}
	if ptr := parser.String("function"); ptr != nil {
		c.Function = *ptr
	}
	if ptr := parser.String("file"); ptr != nil {
		c.File = *ptr
	}
	if ptr := parser.Int("line"); ptr != nil {
		c.Line = *ptr
	}
}

func (c *Caller) Validate(validator structure.Validator) {
	validator.String("package", &c.Package).NotEmpty()
	validator.String("function", &c.Function).NotEmpty()
	validator.String("file", &c.File).NotEmpty()
	validator.Int("line", &c.Line).GreaterThanOrEqualTo(0)
}

func (c *Caller) Normalize(normalizer structure.Normalizer) {}

func init() {
	if _, file, _, ok := runtime.Caller(0); ok {
		ignoredFileSegments = len(strings.Split(file, "/")) - 1
	}
}

var ignoredFileSegments int
