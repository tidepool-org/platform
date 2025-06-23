# Clear all TIDEPOOL_* environment variables
unset $(env | cut -d'=' -f1 | grep '^TIDEPOOL_' | xargs)

export TIDEPOOL_ENV="test"

export TIDEPOOL_LOGGER_LEVEL="error"

export TIDEPOOL_STORE_ADDRESSES="localhost"
export TIDEPOOL_STORE_TLS="false"
export TIDEPOOL_STORE_DATABASE="tidepool_test"

export TIDEPOOL_CONFIRMATION_STORE_DATABASE="confirm_test"
export TIDEPOOL_DEPRECATED_DATA_STORE_DATABASE="data_test"
export TIDEPOOL_MESSAGE_STORE_DATABASE="messages_test"
export TIDEPOOL_PERMISSION_STORE_DATABASE="gatekeeper_test"
export TIDEPOOL_PROFILE_STORE_DATABASE="seagull_test"
export TIDEPOOL_SESSION_STORE_DATABASE="user_test"
export TIDEPOOL_SYNC_TASK_STORE_DATABASE="data_test"
export TIDEPOOL_USER_STORE_DATABASE="user_test"

export TIDEPOOL_KEYCLOAK_CLIENT_ID="client_id"
export TIDEPOOL_KEYCLOAK_CLIENT_SECRET="client_secret"
export TIDEPOOL_KEYCLOAK_LONG_LIVED_CLIENT_ID="long_lived_client_id"
export TIDEPOOL_KEYCLOAK_LONG_LIVED_CLIENT_SECRET="long_lived_client_secret"
export TIDEPOOL_KEYCLOAK_BACKEND_CLIENT_ID="backend_client_id"
export TIDEPOOL_KEYCLOAK_BACKEND_CLIENT_SECRET="backend_client_secret"
export TIDEPOOL_KEYCLOAK_BASE_URL="http://localhost:8080"
export TIDEPOOL_KEYCLOAK_REALM="realm"
export TIDEPOOL_KEYCLOAK_ADMIN_USERNAME="admin_username"
export TIDEPOOL_KEYCLOAK_ADMIN_PASSWORD="admin_password"

