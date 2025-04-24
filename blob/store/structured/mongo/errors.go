package mongo

import "github.com/tidepool-org/platform/errors"

var errorBlobIDNotValid = errors.New("id is invalid")
var errorBlobIDMissing = errors.New("id is missing")
var errorUserIDNotValid = errors.New("user id is invalid")
var errorUserIDMissing = errors.New("user id is missing")
