## HEAD

- Infer application name from executable name
- Update dependencies
- Move dataservices functionality into data package
- Move metricservices functionality into metric package
- Move userservices functionality into user package
- Move service executables into services folder
- Update Makefile to reflect service changes
- Move current task functionality to synctask package
- Move current notification functionality to confirmation package
- Refactor version package to align with new application package
- Refactor common functionality in service package into application package
- Update service package to reflect new application package
- Refactor common functionality in tools into tool package
- Update tools to reflect new tool package
- Update environment config to reflect common functionality in application package
- Introduce new runner mechanism to simplify main package for application package derived functionality
- Add package for version tests
- Introduce new log mechanism with dynamically configured levels and custom serializers
- Update null logger to reflect new log mechanism
- Remove log config as unnecessary for only one field
- Remove `github.com/sirupsen/logrus` dependency
- Add config Reporter functions `Set` and `Delete`
- Rename config Reporter functions `String` to `Get` and `StringOrDefault` to `GetWithDefault`
- Pull string functions out of app package into specific functions appropriate to including package
- Pull id functions out of app package into their own id package
- Pull pointer functions out of app package into their own pointer package
- Remove deprecated environment package
- Introduce new scoped config mechanism using only environment variables and not external files
- Update all config struct usage to reflect new scoped config mechanism
- Delete deprecated config files
- Remove `github.com/tidepool-org/configor` dependency
- Remove legacy group id from data replaced by user id
- Remove deprecated user services `Client.GetUserGroupID`

## v1.9.0 (2017-08-10)

- Bump `hash_deactivate_old` data deduplicator to version 1.1.0
- Update `hash_deactivate_old` data deduplicator to use archived dataset id and time fields to accurately:
  - Deactivate deduplicate data on dataset addition
  - Activate undeduplicated data on dataset deletion
  - Record entire deduplication history
- Update mongo queries related to `hash_deactivate_old` data deduplicator
- Remove backwards-compatible legacy deduplicator name test in `DeduplicatorDescriptor.IsRegisteredWithNamedDeduplicator` (after `v1.8.0` required migration)
- Add archived dataset id and time fields to base data type
- Add MD5 hash of authentication token to request logger
- Add service middleware to extract select request headers and add as request logger fields
- Defer access to context store sessions and log until actually needed

## v1.8.0 (2017-08-09)

- Add CHANGELOG.md
- **REQUIRED MIGRATION**: `migrate_data_deduplicator_descriptor` - data deduplicator descriptor name and version
- Force `precise` Ubuntu distribution for Travis (update to `trusty` later)
- Add deduplicator version
- Update deduplicator name scheme
- Add `github.com/blang/semver package` dependency
- Fix dependency import capitalization
- Update dependencies
- Remove unused data store functionality
- Remove unused data deduplicators

## v1.7.0 (2017-06-22)

- See commit history for details on this and all previous releases
