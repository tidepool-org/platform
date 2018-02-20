## HEAD

* Use consistent paradigm for creation of contained data objects
* Randomize most test input for data validate and normalize
* Fix meta reporting for embedded data (Status within Alarm, Bolus within Calculator)
* Distinguish between normalizing data from external origin versus internal or store origin
* Add Blob and BlobArray to encapsulate object and object array parse, validate, and normalize
* Update most all data object properties to use pointers
* Through validation of all data properties
* Improve all data validation and normalization tests to be complete and thorough
* Refactor all data validations to be accurate
* Use constants for all data validations
* Additional test helpers
* Add New and Clone functions for all test data objects
* Refactor global variables to make private or use functions to enforce as constant
* Log debug mongo connection configuration
* Add structure origin
* New structure validator validations
* Refactor error expectations using new test helpers
* Additional test helpers
* Add Object and Array structure validators
* Remove data.ValidateInterface...; does not work as expected
* Remove structure.Validating; does not work as expected
* Refactor data normalizer to use common structure normalizer
* Remove normalization from physical activity duration as necessary
* Expose ReportError function in structure parser, normalizer, and validator
* Update application package and tests
* Remove unnecessary verbose flag from tools; use environment variable instead
* Update confirmation package to use latest code and test style
* Additional test helpers
* Add OAuth client credentials provider
* Add golang.org/x/oauth2/clientcredentials dependency
* Refactor oauth package to allow alternate grant workflows
* Add Makefile generate target for code generation
* Update mechanism to find files with main()
* Add file embedding tool to Makefile
* Handle additional client response codes
* Strip trailing slashes from client base address
* Return error from config.Get rather than bool for compatibility
* Allow additional scopes on tools

## v1.23.0

* Fix start script source of environment file
* Add LifeScan Verio and LifeScan Verio Flex to supported devices
* Accomodate out-of-sync server time between Tidepool and Dexcom
* Add timezone data to release images
* Use standard `tidepool` user in Dockerfiles
* Add CA certificates to release images for external SSL requests
* Do not build Docker images without tag
* Fail Makefile when make lint fails
* Fix filename sorting issue on different file systems with make lint
* Automatic Docker build and push
* Use actual executables for migrations and tools
* Use /bin/sh in various start.sh scripts

## v1.22.0

* Remove legacy code to fixup unexpected data from Dexcom API
* Properly handle expired access tokens from Dexcom API
* Warn on excessive duration Dexcom API requests and overall task
* Add custom User-Agent header to client requests

## v1.21.0

* Dependency updates
* Use Go v1.9.2
* Update Copyright
* Minor Makefile updates
* Tool to compute metric hash from user id
* Add support for delta uploads
* Support private client data in upload
* Use environment variable for test database address

## v1.20.2

* Add fix to handle Dexcom API not correctly reporting daily G5 Mobile devices

## v1.20.1

* Fix alert settings snooze validation for Dexcom API
* Update Dockerfile to Go 1.9.1

## v1.20.0

* Use Go v1.9.1
* Update dependencies
* Add golang.org/x/oauth2 dependency
* Allow debugging from within VS Code
* Store request specific data in context
* Separate local authentication functionality from external (Shoreline) authentication
* Add provider sessions to support external service provider authentication
* Add external service provider OAuth require endpoints
* Add restricted tokens to support short-lived, restricted access authentication
* Implement uniform client functionality whether intra-service or inter-service
* Use responder uniformly across new services
* Refactor errors package
  * Remove package prefix string - Capture caller file, line, package, and function - Allow errors to be serialized/deserialized to/from JSON and BSON - Error can be one or multiple - Add pretty printing of all error info and multiple errors
* Add create/delete hooks for provider session create/delete to allow custom actions for specific providers
* Use gomega expectations rather than panics in test mocks
* Ensure mongo indexes for new services upon initialization
* Consistently use filter and pagination parameters for new service endpoints
* Refactor parsing, validation, normalization into common structure functionality for global use
* Improve test mocks and utility functions throughout
* Add HTTP test helpers
* Refactor generic client with custom client specific to each type of service provider (platform, OAuth, etc.)
* Add ability to mutate generic client request (to add request parameters, headers, etc.)
* Remove logger parameter from all new store session functions
* Add context as required parameter to all store functions (and functions that invoke store functions)
* Add data set as actual type (rather than just another data.Datum; code migration in progress)
* Add data source to represent a source of data, meta object that may represent multipe data sets
* Add variable threshold when validating time against now
* Add continuous data deduplicator; no deduplication actually performed
* Add activity/physical data type
* Add food data type
* Add insulin data type
* Add state/reported data type
* Update router path params to camelcase for consistency (not snakecase)
* Update data.Datum Annotations and Payload to be map[string]interface{} (was interface{})
* Add Dexcom API data types, including parsing, validation, and normalization
* Add Dexcom client as OAuth client
* Add Dexcom fetch task and runner
* Translate Dexcom API data types to Tidepool Data Model
* Periodically pull user data from Dexcom API, translate, and import into data store
* Allow id validation
* Logger uses refactor errors internally
* Add general OAuth provider implementation to support any external OAuth service
* Add global pagination mechanism
* Add service secrets to allow inter-service communication without authentication token
* Refactor common request functionality into separate package
  * Predefined errors
  * Parse, validate, and normalize requset parameters and response JSON body
  * Authentication details
  * Trace request/session
  * Responder
* Add authentication handler funcs to support service, user, and any authenticated session
* Revamp authentication middleware to support all known methods and record details into context details
* Common construct update functionality for mongo store
* Remove mongo agent as unnecessary
* Add task service that enables background task queue
* Allow structures/errors to be sanitized before returning in API (remove internal or private info)
* Make log package concurrency safe
* Replace Logger.SetLevel with Logger.WithLevel
* Rename log.Levels to log.LevelRanks
* Encapsulate service within api
* Use consistent route function naming
* Fix minor issues with use of defer
* Resolve general config reporter scopes issue with services
* Remove the 'services' suffix from all services
* Remove deprecated dataservices and userservices url path prefix
* Update url params to use underscores
* Refactor store sessions to allow multiple collections per store
* Add mongo config collection prefix to all collection name customization (particularly for tests)
* Add auth, notification, and task services
* Refactor and improve service package to streamline creation of new services
* Move authentication token logging to authenticate middleware
* Update service ports to align with deployed configurations
* Update Makefile to build service artifacts using common mechanism
* Remove broken `-notify` flag on Makefile watch target
* Add `format-write` Makefile target to write formatted code to source file
* Fix test package init check
* Add common Mock struct for test mocks
* Refactor auth middleware to make it the default on all routes
* Add auth client to handle user and server authentication and authorization for clients
* Refactor all client packages into common client package
* Update status and version APIs to use standard response functions
* Remove Clone function from all Config structs as unnecessary
* Refactor service Context into separate Responder struct
* Add various test mocks
* Remove deprecated git-hooks Makefile target
* Update Makefile .PHONY targets
* Refactor common functionality in migrations into migration package
* Update migrations to reflect new migration package
* Move migration executables into migrations folder
* Add new migration deploy artifact in Makefile
* Infer application name from executable name
* Update dependencies
* Move dataservices functionality into data package
* Move metricservices functionality into metric package
* Move userservices functionality into user package
* Move service executables into services folder
* Update Makefile to reflect service changes
* Move current task functionality to synctask package
* Move current notification functionality to confirmation package
* Refactor version package to align with new application package
* Refactor common functionality in service package into application package
* Update service package to reflect new application package
* Refactor common functionality in tools into tool package
* Update tools to reflect new tool package
* Update environment config to reflect common functionality in application package
* Introduce new runner mechanism to simplify main package for application package derived functionality
* Add package for version tests
* Introduce new log mechanism with dynamically configured levels and custom serializers
* Update null logger to reflect new log mechanism
* Remove log config as unnecessary for only one field
* Remove `github.com/sirupsen/logrus` dependency
* Add config Reporter functions `Set` and `Delete`
* Rename config Reporter functions `String` to `Get` and `StringOrDefault` to `GetWithDefault`
* Pull string functions out of app package into specific functions appropriate to including package
* Pull id functions out of app package into their own id package
* Pull pointer functions out of app package into their own pointer package
* Remove deprecated environment package
* Introduce new scoped config mechanism using only environment variables and not external files
* Update all config struct usage to reflect new scoped config mechanism
* Delete deprecated config files
* Remove `github.com/tidepool-org/configor` dependency
* Remove legacy group id from data replaced by user id
* Remove deprecated user services `Client.GetUserGroupID`

## v1.9.0

* Bump `hash_deactivate_old` data deduplicator to version 1.1.0
* Update `hash_deactivate_old` data deduplicator to use archived dataset id and time fields to accurately:
  * Deactivate deduplicate data on dataset addition
  * Activate undeduplicated data on dataset deletion
  * Record entire deduplication history
* Update mongo queries related to `hash_deactivate_old` data deduplicator
* Remove backwards-compatible legacy deduplicator name test in `DeduplicatorDescriptor.IsRegisteredWithNamedDeduplicator` (after `v1.8.0` required migration)
* Add archived dataset id and time fields to base data type
* Add MD5 hash of authentication token to request logger
* Add service middleware to extract select request headers and add as request logger fields
* Defer access to context store sessions and log until actually needed

## v1.8.0

* Add CHANGELOG.md
* **REQUIRED MIGRATION**: `migrate_data_deduplicator_descriptor` - data deduplicator descriptor name and version
* Force `precise` Ubuntu distribution for Travis (update to `trusty` later)
* Add deduplicator version
* Update deduplicator name scheme
* Add `github.com/blang/semver package` dependency
* Fix dependency import capitalization
* Update dependencies
* Remove unused data store functionality
* Remove unused data deduplicators

## v1.7.0

* See commit history for details on this and all previous releases
