module github.com/tidepool-org/platform

go 1.15

require (
	github.com/Shopify/sarama v1.27.0
	github.com/ant0ine/go-json-rest v3.3.2+incompatible
	github.com/aws/aws-sdk-go v1.35.3
	github.com/blang/semver v3.5.1+incompatible
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/disintegration/imaging v1.6.2
	github.com/githubnemo/CompileDaemon v1.4.0
	github.com/golang/mock v1.6.0
	github.com/google/uuid v1.3.0
	github.com/gowebpki/jcs v0.0.0-20210215032300-680d9436c864
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.14.0 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mjibson/esc v0.2.0
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.3
	github.com/prometheus/client_golang v1.11.1
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/rinchsan/device-check-go v1.2.3
	github.com/tidepool-org/clinic/client v0.0.0-20211118205743-020bf46ac989
	github.com/tidepool-org/devices/api v0.0.0-20220914225528-c7373eb1babc
	github.com/tidepool-org/go-common v0.9.0
	github.com/tidepool-org/hydrophone/client v0.0.0-20221219223301-92bd47a8a11c
	github.com/urfave/cli v1.22.4
	go.mongodb.org/mongo-driver v1.8.2
	go.uber.org/fx v1.13.1
	go.uber.org/zap v1.13.0 // indirect
	golang.org/x/crypto v0.0.0-20220411220226-7b82a4e95df4
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616
	golang.org/x/net v0.4.0 // indirect
	golang.org/x/oauth2 v0.2.0
	golang.org/x/sync v0.1.0
	golang.org/x/tools v0.1.12
	google.golang.org/genproto v0.0.0-20221207170731-23e4bf6bdc37 // indirect
	google.golang.org/grpc v1.51.0
	gopkg.in/yaml.v2 v2.4.0
	syreclabs.com/go/faker v1.2.2
)

replace gopkg.in/fsnotify.v1 v1.4.7 => gopkg.in/fsnotify/fsnotify.v1 v1.4.7
