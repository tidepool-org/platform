module github.com/tidepool-org/platform

go 1.11

require (
	github.com/Shopify/sarama v1.27.0
	github.com/ant0ine/go-json-rest v3.3.2+incompatible
	github.com/aws/aws-sdk-go v1.29.23
	github.com/blang/semver v3.5.1+incompatible
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/disintegration/imaging v1.5.0
	github.com/fatih/color v1.7.0 // indirect
	github.com/githubnemo/CompileDaemon v1.0.0
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8
	github.com/google/uuid v1.1.1
	github.com/howeyc/fsnotify v0.9.0 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kr/pty v1.1.1 // indirect
	github.com/mattn/go-colorable v0.0.9 // indirect
	github.com/mattn/go-isatty v0.0.4 // indirect
	github.com/mitchellh/go-homedir v1.0.0
	github.com/mjibson/esc v0.1.0
	github.com/onsi/ginkgo v1.10.2
	github.com/onsi/gomega v1.9.0
	github.com/tidepool-org/devices v0.0.0-20200709151222-3d1edbd4a58e
	github.com/tidepool-org/go-common v0.6.3-0.20201001200420-0bf25851089b
	github.com/urfave/cli v1.20.0
	go.mongodb.org/mongo-driver v1.3.2
	go.uber.org/fx v1.12.0
	golang.org/x/crypto v0.0.0-20200510223506-06a226fb4e37
	golang.org/x/image v0.0.0-20181116024801-cd38e8056d9b // indirect
	golang.org/x/lint v0.0.0-20190930215403-16217165b5de
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/tools v0.0.0-20191114200427-caa0b0f7d508
	google.golang.org/grpc v1.30.0
	gopkg.in/tylerb/graceful.v1 v1.2.15
	gopkg.in/yaml.v2 v2.3.0
	syreclabs.com/go/faker v1.2.2
)

replace gopkg.in/fsnotify.v1 v1.4.7 => gopkg.in/fsnotify/fsnotify.v1 v1.4.7
