#!/bin/bash
API_ENV=qa3.development
USER_ID_ONE=$1
SERVER_TOKEN=$2

check_val() {
    if [[ -z "$1" ]]; then
        echo "missing $2 value"
        exit 2
    fi
}

check_val $SERVER_SECRET "SERVER_SECRET"
check_val $USER_ID_ONE "USER_ID_ONE"

if [[ -z "$SERVER_TOKEN" ]]; then

    SERVER_TOKEN="$(curl -s -I -X POST -H "X-Tidepool-Server-Secret: $SERVER_SECRET" -H "X-Tidepool-Server-Name: devops" "https://${API_ENV}.tidepool.org/auth/serverlogin" | grep 'x-tidepool-session-token' | sed 's/[^:]*: //')"
fi

check_val $SERVER_TOKEN "SERVER_TOKEN"

http_response=$(curl -s -w "%{response_code}" --request DELETE \
    --url https://${API_ENV}.tidepool.org/v1/users/${USER_ID_ONE}/data \
    --header 'Accept: */*' \
    --header "X-Tidepool-Session-Token: $SERVER_TOKEN")

echo "status $http_response deleting data for user $USER_ID_ONE"
