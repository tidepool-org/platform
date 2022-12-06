# Clear all TIDEPOOL_* environment variables
unset `env | cut -d'=' -f1 | grep '^TIDEPOOL_' | xargs`

export TIDEPOOL_ENV="test"

export TIDEPOOL_LOGGER_LEVEL="error"

export TIDEPOOL_STORE_ADDRESSES="mongo4platform${RUN_ID}:27017"
#export TIDEPOOL_STORE_ADDRESSES="localhost:27018"
export TIDEPOOL_STORE_TLS="false"
export TIDEPOOL_STORE_DATABASE="tidepool_test"
export TIDEPOOL_STORE_MAX_CONNECTION_ATTEMPTS=5

export TIDEPOOL_CONFIRMATION_STORE_DATABASE="confirm_test"
export TIDEPOOL_DATA_STORE_DATABASE="data_test"
export TIDEPOOL_MESSAGE_STORE_DATABASE="messages_test"
export TIDEPOOL_PERMISSION_STORE_DATABASE="gatekeeper_test"
export TIDEPOOL_PROFILE_STORE_DATABASE="seagull_test"
export TIDEPOOL_SESSION_STORE_DATABASE="user_test"
export TIDEPOOL_SYNC_TASK_STORE_DATABASE="data_test"
export TIDEPOOL_USER_STORE_DATABASE="user_test"

export KEPT_IN_LEGACY_DATA_TYPES="basal"
export ARCHIVED_DATA_TYPES="cbg"
export BUCKETED_DATA_TYPES="cbg,basal"
