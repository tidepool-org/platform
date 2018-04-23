package zone

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"time"
)

func Names() []string {
	return _Names
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
