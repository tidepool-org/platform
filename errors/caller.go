package errors

import (
	"path"
	"runtime"
	"strings"

	"github.com/tidepool-org/platform/structure"
)

type Caller struct {
	Package  string `json:"dbl_package,omitempty" bson:"package,omitempty"`
	Function string `json:"dbl_function,omitempty" bson:"function,omitempty"`
	File     string `json:"dbl_file,omitempty" bson:"file,omitempty"`
	Line     int    `json:"dbl_line,omitempty" bson:"line,omitempty"`
}

func GetCaller(frame int) *Caller {
	if pc, file, line, ok := runtime.Caller(frame + 1); ok {
		fileParts := strings.SplitN(file, "/", ignoredFileSegments)
		fileLength := len(fileParts)
		pkgParts := strings.Split(runtime.FuncForPC(pc).Name(), "/")
		pkgLength := len(pkgParts)
		fnctnParts := strings.Split(pkgParts[pkgLength-1], ".")
		fnctnLength := len(fnctnParts)

		return &Caller{
			Package:  strings.Join(append(pkgParts[:pkgLength-1], fnctnParts[0]), "/"),
			Function: fnctnParts[fnctnLength-1],
			File:     fileParts[fileLength-1],
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
	if ptr := parser.String("dbl_package"); ptr != nil {
		c.Package = *ptr
	}
	if ptr := parser.String("dbl_function"); ptr != nil {
		c.Function = *ptr
	}
	if ptr := parser.String("dbl_file"); ptr != nil {
		c.File = *ptr
	}
	if ptr := parser.Int("dbl_line"); ptr != nil {
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
