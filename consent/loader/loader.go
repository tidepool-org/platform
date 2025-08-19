package loader

import (
	"context"
	"embed"
	_ "embed"
	"io/fs"
	"path"
	"regexp"
	"strconv"

	"github.com/tidepool-org/platform/log"

	"github.com/tidepool-org/platform/consent"
	"github.com/tidepool-org/platform/errors"
)

//go:embed content/*.md
var content embed.FS
var directory = "content"

// Allow injecting FS in tests
var contentFS fs.FS = content

func SetContentFS(testFS fs.FS, dir string) {
	contentFS = testFS
	directory = dir
}

func ResetContentFS() {
	contentFS = content
	directory = "content"
}

// Matches (name).v(version).md
var markdownContent = regexp.MustCompile(`(?P<name>[a-zA-Z0-9_-]+)\.v(?P<version>[0-9]+)\.md$`)

func SeedConsents(ctx context.Context, logger log.Logger, service consent.Service) error {
	entries, err := fs.ReadDir(contentFS, directory)
	if err != nil {
		return errors.Wrap(err, "unable to read consent content directory")
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !markdownContent.MatchString(entry.Name()) {
			continue
		}

		matches := markdownContent.FindStringSubmatch(entry.Name())
		nameIndex := markdownContent.SubexpIndex("name")
		versionIndex := markdownContent.SubexpIndex("version")
		if nameIndex == -1 || versionIndex == -1 {
			return errors.Newf("invalid content file name %s", entry.Name())
		}

		version, err := strconv.Atoi(matches[versionIndex])
		if err != nil {
			return errors.Newf("invalid version %s for consent %s", matches[versionIndex], entry.Name())
		}

		consentContent, err := fs.ReadFile(contentFS, path.Join(directory, entry.Name()))
		if err != nil {
			return errors.Wrapf(err, "unable to read consent content from %s", entry.Name())
		}

		cons := consent.NewConsent()
		cons.ContentType = consent.ContentTypeMarkdown
		cons.Type = matches[nameIndex]
		cons.Version = version
		cons.Content = string(consentContent)

		err = service.EnsureConsent(log.NewContextWithLogger(ctx, logger), cons)
		if err != nil {
			return errors.Wrapf(err, "unable to ensure consent %s exists", entry.Name())
		}
	}

	return nil
}
