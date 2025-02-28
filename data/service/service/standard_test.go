package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tidepool-org/platform/application"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/test"
)

func TestAuthClientIsInitialized(t *testing.T) {
	tt := newStandardTest(t)
	tt.Setenv("SECRET", "something secret")
	tt.Setenv("SERVER_ADDRESS", "somehost")
	tt.SetenvToFilePath("SERVER_TLS_CERTIFICATE_FILE", "contents")
	tt.SetenvToFilePath("SERVER_TLS_KEY_FILE", "contents")
	tt.Setenv("AUTH_CLIENT_ADDRESS", "something secret")
	tt.Setenv("AUTH_CLIENT_EXTERNAL_ADDRESS", "something secret")
	tt.Setenv("AUTH_CLIENT_SERVICE_SECRET", "something secret")
	tt.Setenv("AUTH_CLIENT_EXTERNAL_SERVICE_SECRET", "something secret")
	tt.Setenv("AUTH_CLIENT_SERVER_SESSION_TOKEN_SECRET", "something secret")
	tt.Setenv("AUTH_CLIENT_EXTERNAL_SERVER_SESSION_TOKEN_SECRET", "something secret")
	tt.Setenv("METRIC_CLIENT_ADDRESS", "something secret")
	tt.Setenv("PERMISSION_CLIENT_ADDRESS", "something secret")
	tt.Setenv("DEPRECATED_DATA_STORE_DATABASE", "test_data_database")
	tt.Setenv("SYNC_TASK_STORE_DATABASE", "test_sync_task_database")
	// These have no prefixes
	t.Setenv("KAFKA_BROKERS", "somehost")
	t.Setenv("KAFKA_TOPIC_PREFIX", "somehost")
	t.Setenv("KAFKA_REQUIRE_SSL", "false")
	t.Setenv("KAFKA_VERSION", "2.5.0")

	std := NewStandard()
	err := std.Initialize(tt.Provider)
	if err != nil {
		t.Errorf("expected successful initialization, got %s", err)
	}
}

type standardTest struct {
	Prefix   string
	Name     string
	Scopes   []string
	Provider application.Provider

	t       testing.TB
	tempDir string
}

func newStandardTest(t *testing.T) *standardTest {
	prefix := test.RandomStringFromRangeAndCharset(4, 8, test.CharsetUppercase)
	name := test.RandomStringFromRangeAndCharset(4, 8, test.CharsetAlphaNumeric)
	scopes := test.RandomStringArrayFromRangeAndCharset(0, 2, test.CharsetAlphaNumeric)

	t.Setenv(fmt.Sprintf("%s_LOGGER_LEVEL", prefix), "error")
	oldName := os.Args[0]
	os.Args[0] = name
	t.Cleanup(func() { os.Args[0] = oldName })

	application.VersionBase = netTest.RandomSemanticVersion()
	application.VersionFullCommit = test.RandomStringFromRangeAndCharset(40, 40, test.CharsetHexidecimalLowercase)
	application.VersionShortCommit = application.VersionFullCommit[0:8]

	provider, err := application.NewProvider(prefix, scopes...)
	if err != nil {
		t.Fatalf("unable to initialize provider")
	}

	return &standardTest{
		Prefix:   prefix,
		Name:     name,
		Scopes:   scopes,
		Provider: provider,

		t:       t,
		tempDir: t.TempDir(),
	}
}

func (t *standardTest) Setenv(keySuffix, value string) {
	baseKey := strings.Join(append([]string{t.Prefix, t.Name}, t.Scopes...), "_")
	t.t.Setenv(strings.ToUpper(baseKey+"_"+keySuffix), value)
}

func (t *standardTest) SetenvToFilePath(keySuffix, contents string) {
	filename := filepath.Join(t.tempDir, keySuffix)
	if err := os.WriteFile(filename, []byte(contents), 0600); err != nil {
		t.t.Fatalf("opening tempfile %q: %s", filename, err)
	}
	t.Setenv(keySuffix, filename)
}
