package audit

// import (
// 	"context"

// 	"github.com/tidepool-org/platform/errors"
// 	"github.com/tidepool-org/platform/log"
// )

// type Auditor interface {
// 	Audit(resourceOwnerID string, )

// 	WithAuthenticatedUserID(authenticatedUserID string) Auditor

// 	WithContext(ctx context.Context) Auditor

// 	WithError(err error) Auditor

// 	WithField(key string, value interface{}) Auditor
// 	WithFields(fields Fields) Auditor
// }

// type LoggerAuditor struct {
// 	logger log.Logger
// }

// func NewLoggerAuditor(logger log.Logger) (*Audit, error) {
// 	if logger == nil {
// 		return nil, errors.New("logger is missing")
// 	}

// 	return &LoggerAuditor{
// 		logger: logger,
// 	}, nil
// }

// func (l *LoggerAuditor) Audit(ctx context.Context) {

// }

// // authenticated user id
// // action
// // target user id
// // resources

// // all APIs forward incoming AuthDetails
// // so even if API called with server secret, originating authenticated user is captured?
// // or capture it via requesttrace? Only capture authenticated user id on top-level API audit?

// // audit should capture request trace and session trace
// // we get this from context

// // or reverse this an create auditor
// // works like logger?
// // WithField
// // WithFields
// // WithError
