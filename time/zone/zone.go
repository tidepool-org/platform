package zone

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func Names() []string {
	return _Names
}

func IsValidName(value string) bool {
	return ValidateName(value) == nil
}

func NameValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateName(value))
}

func ValidateName(value string) error {
	if value == "" {
		return structureValidator.ErrorValueEmpty()
	} else if index := sort.SearchStrings(_Names, value); index == len(_Names) || _Names[index] != value {
		return ErrorValueStringAsNameNotValid(value)
	}
	return nil
}

func ErrorValueStringAsNameNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid time zone name", value)
}

var _Names []string

var _Directories = []string{
	"/usr/share/zoneinfo",
	"/usr/share/lib/zoneinfo",
	"/usr/lib/locale/TZ",
}

var _PartRegexp = regexp.MustCompile("^[A-Z][A-Za-z0-9_+-]*$")

func init() {
	directoryNames, err := scanDirectoriesForNames(_Directories)
	if err != nil {
		panic(err)
	}
	_Names = directoryNames
}

func scanDirectoriesForNames(directories []string) ([]string, error) {
	var names []string
	for _, directory := range directories {
		directoryNames, err := scanDirectoryForNames(directory, "")
		if err != nil {
			return nil, err
		}
		names = append(names, directoryNames...)
	}
	sort.Strings(names)
	return names, nil
}

func scanDirectoryForNames(directory string, prefix string) ([]string, error) {
	if directory == "" {
		return nil, errors.New("directory is missing")
	}

	fileInfos, err := ioutil.ReadDir(filepath.Join(directory, prefix))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var names []string
	for _, fileInfo := range fileInfos {
		part := fileInfo.Name()
		if !_PartRegexp.MatchString(part) {
			continue
		}

		if fileInfo.IsDir() {
			var directoryNames []string
			directoryNames, err = scanDirectoryForNames(directory, filepath.Join(prefix, part))
			if err != nil {
				return nil, err
			}
			names = append(names, directoryNames...)
		} else {
			name := path.Join(prefix, part)
			if _, err = time.LoadLocation(name); err == nil {
				names = append(names, name)
			}
		}
	}

	sort.Strings(names)
	return names, nil
}
