module github.com/tidepool-org/platform

go 1.11

require (
	contrib.go.opencensus.io/exporter/ocagent v0.6.0
	github.com/ant0ine/go-json-rest v3.3.2+incompatible
	github.com/aws/aws-sdk-go v1.16.20
	github.com/blang/semver v3.5.1+incompatible
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/disintegration/imaging v1.5.0
	github.com/fatih/color v1.7.0 // indirect
	github.com/githubnemo/CompileDaemon v1.0.0
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8
	github.com/howeyc/fsnotify v0.9.0 // indirect
	github.com/mattn/go-colorable v0.0.9 // indirect
	github.com/mattn/go-isatty v0.0.4 // indirect
	github.com/mitchellh/go-homedir v1.0.0
	github.com/mjibson/esc v0.1.0
	github.com/onsi/ginkgo v1.7.0
	github.com/onsi/gomega v1.4.3
	github.com/urfave/cli v1.20.0
	go.opencensus.io v0.22.2
	golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2
	golang.org/x/image v0.0.0-20181116024801-cd38e8056d9b // indirect
	golang.org/x/lint v0.0.0-20190409202823-959b441ac422
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/tools v0.0.0-20190524140312-2c0ae7006135
	gopkg.in/tylerb/graceful.v1 v1.2.15
	gopkg.in/yaml.v2 v2.2.2
)

replace gopkg.in/fsnotify.v1 v1.4.7 => gopkg.in/fsnotify/fsnotify.v1 v1.4.7
