package v1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "user/service/api/v1")
}

// type TestFlags struct {
// 	flags map[string]bool
// }

// func NewTestFlags() *TestFlags {
// 	return &TestFlags{
// 		flags: map[string]bool{},
// 	}
// }

// func (t *TestFlags) Set(flagsToSet ...string) *TestFlags {
// 	flags := map[string]bool{}
// 	for key, value := range t.flags {
// 		flags[key] = value
// 	}

// 	for _, flagToSet := range flagsToSet {
// 		flags[flagToSet] = true
// 	}

// 	return &TestFlags{
// 		flags: flags,
// 	}
// }

// func (t *TestFlags) IsSet(flag string) bool {
// 	set, ok := t.flags[flag]
// 	return ok && set
// }

// type RespondWithInternalServerFailureInput struct {
// 	message string
// 	failure []interface{}
// }

// type RespondWithStatusAndErrorsInput struct {
// 	statusCode int
// 	errors     []*service.Error
// }

// type RespondWithStatusAndDataInput struct {
// 	statusCode int
// 	data       interface{}
// }

// type RecordMetricInput struct {
// 	context context.Context
// 	metric  string
// 	data    []map[string]string
// }

// type TestMetricClient struct {
// 	RecordMetricInputs  []RecordMetricInput
// 	RecordMetricOutputs []error
// }

// func (t *TestMetricClient) RecordMetric(ctx context.Context, metric string, data ...map[string]string) error {
// 	t.RecordMetricInputs = append(t.RecordMetricInputs, RecordMetricInput{ctx, metric, data})
// 	output := t.RecordMetricOutputs[0]
// 	t.RecordMetricOutputs = t.RecordMetricOutputs[1:]
// 	return output
// }

// func (t *TestMetricClient) ValidateTest() bool {
// 	return len(t.RecordMetricOutputs) == 0
// }

// type GetUserPermissionsInput struct {
// 	context       context.Context
// 	requestUserID string
// 	targetUserID  string
// }

// type GetUserPermissionsOutput struct {
// 	permissions user.Permissions
// 	err         error
// }

// type TestUserClient struct {
// 	GetUserPermissionsInputs  []GetUserPermissionsInput
// 	GetUserPermissionsOutputs []GetUserPermissionsOutput
// }

// func (t *TestUserClient) GetUserPermissions(ctx context.Context, requestUserID string, targetUserID string) (user.Permissions, error) {
// 	t.GetUserPermissionsInputs = append(t.GetUserPermissionsInputs, GetUserPermissionsInput{ctx, requestUserID, targetUserID})
// 	output := t.GetUserPermissionsOutputs[0]
// 	t.GetUserPermissionsOutputs = t.GetUserPermissionsOutputs[1:]
// 	return output.permissions, output.err
// }

// func (t *TestUserClient) ValidateTest() bool {
// 	return len(t.GetUserPermissionsOutputs) == 0
// }

// type DestroyDataForUserByIDInput struct {
// 	context context.Context
// 	userID  string
// }

// type TestDataClient struct {
// 	DestroyDataForUserByIDInputs  []DestroyDataForUserByIDInput
// 	DestroyDataForUserByIDOutputs []error
// }

// func (t *TestDataClient) DestroyDataForUserByID(ctx context.Context, userID string) error {
// 	t.DestroyDataForUserByIDInputs = append(t.DestroyDataForUserByIDInputs, DestroyDataForUserByIDInput{ctx, userID})
// 	output := t.DestroyDataForUserByIDOutputs[0]
// 	t.DestroyDataForUserByIDOutputs = t.DestroyDataForUserByIDOutputs[1:]
// 	return output
// }

// func (t *TestDataClient) ValidateTest() bool {
// 	return len(t.DestroyDataForUserByIDOutputs) == 0
// }

// type TestConfirmationSession struct {
// 	DeleteUserConfirmationsInputs  []string
// 	DeleteUserConfirmationsOutputs []error
// }

// func (t *TestConfirmationSession) IsClosed() bool {
// 	panic("Unexpected invocation of IsClosed on TestConfirmationSession")
// }

// func (t *TestConfirmationSession) Close() {
// 	panic("Unexpected invocation of Close on TestConfirmationSession")
// }

// func (t *TestConfirmationSession) Logger() log.Logger {
// 	panic("Unexpected invocation of Logger on TestConfirmationSession")
// }

// func (t *TestConfirmationSession) DeleteUserConfirmations(userID string) error {
// 	t.DeleteUserConfirmationsInputs = append(t.DeleteUserConfirmationsInputs, userID)
// 	output := t.DeleteUserConfirmationsOutputs[0]
// 	t.DeleteUserConfirmationsOutputs = t.DeleteUserConfirmationsOutputs[1:]
// 	return output
// }

// func (t *TestConfirmationSession) ValidateTest() bool {
// 	return len(t.DeleteUserConfirmationsOutputs) == 0
// }

// type TestMessagesSession struct {
// 	DeleteMessagesFromUserInputs      []*messageStore.User
// 	DeleteMessagesFromUserOutputs     []error
// 	DestroyMessagesForUserByIDInputs  []string
// 	DestroyMessagesForUserByIDOutputs []error
// }

// func (t *TestMessagesSession) IsClosed() bool {
// 	panic("Unexpected invocation of IsClosed on TestMessagesSession")
// }

// func (t *TestMessagesSession) Close() {
// 	panic("Unexpected invocation of Close on TestMessagesSession")
// }

// func (t *TestMessagesSession) Logger() log.Logger {
// 	panic("Unexpected invocation of Logger on TestMessagesSession")
// }

// func (t *TestMessagesSession) DeleteMessagesFromUser(deleteUser *messageStore.User) error {
// 	t.DeleteMessagesFromUserInputs = append(t.DeleteMessagesFromUserInputs, deleteUser)
// 	output := t.DeleteMessagesFromUserOutputs[0]
// 	t.DeleteMessagesFromUserOutputs = t.DeleteMessagesFromUserOutputs[1:]
// 	return output
// }

// func (t *TestMessagesSession) DestroyMessagesForUserByID(userID string) error {
// 	t.DestroyMessagesForUserByIDInputs = append(t.DestroyMessagesForUserByIDInputs, userID)
// 	output := t.DestroyMessagesForUserByIDOutputs[0]
// 	t.DestroyMessagesForUserByIDOutputs = t.DestroyMessagesForUserByIDOutputs[1:]
// 	return output
// }

// func (t *TestMessagesSession) ValidateTest() bool {
// 	return len(t.DeleteMessagesFromUserOutputs) == 0 &&
// 		len(t.DestroyMessagesForUserByIDOutputs) == 0
// }

// type TestPermissionsSession struct {
// 	DestroyPermissionsForUserByIDInputs  []string
// 	DestroyPermissionsForUserByIDOutputs []error
// }

// func (t *TestPermissionsSession) IsClosed() bool {
// 	panic("Unexpected invocation of IsClosed on TestPermissionsSession")
// }

// func (t *TestPermissionsSession) Close() {
// 	panic("Unexpected invocation of Close on TestPermissionsSession")
// }

// func (t *TestPermissionsSession) Logger() log.Logger {
// 	panic("Unexpected invocation of Logger on TestPermissionsSession")
// }

// func (t *TestPermissionsSession) DestroyPermissionsForUserByID(userID string) error {
// 	t.DestroyPermissionsForUserByIDInputs = append(t.DestroyPermissionsForUserByIDInputs, userID)
// 	output := t.DestroyPermissionsForUserByIDOutputs[0]
// 	t.DestroyPermissionsForUserByIDOutputs = t.DestroyPermissionsForUserByIDOutputs[1:]
// 	return output
// }

// func (t *TestPermissionsSession) ValidateTest() bool {
// 	return len(t.DestroyPermissionsForUserByIDOutputs) == 0
// }

// type GetProfileByIDOutput struct {
// 	*profile.Profile
// 	err error
// }

// type TestProfilesSession struct {
// 	GetProfileByIDInputs      []string
// 	GetProfileByIDOutputs     []GetProfileByIDOutput
// 	DestroyProfileByIDInputs  []string
// 	DestroyProfileByIDOutputs []error
// }

// func (t *TestProfilesSession) IsClosed() bool {
// 	panic("Unexpected invocation of IsClosed on TestProfilesSession")
// }

// func (t *TestProfilesSession) Close() {
// 	panic("Unexpected invocation of Close on TestProfilesSession")
// }

// func (t *TestProfilesSession) Logger() log.Logger {
// 	panic("Unexpected invocation of Logger on TestProfilesSession")
// }

// func (t *TestProfilesSession) GetProfileByID(profileID string) (*profile.Profile, error) {
// 	t.GetProfileByIDInputs = append(t.GetProfileByIDInputs, profileID)
// 	output := t.GetProfileByIDOutputs[0]
// 	t.GetProfileByIDOutputs = t.GetProfileByIDOutputs[1:]
// 	return output.Profile, output.err
// }

// func (t *TestProfilesSession) DestroyProfileByID(profileID string) error {
// 	t.DestroyProfileByIDInputs = append(t.DestroyProfileByIDInputs, profileID)
// 	output := t.DestroyProfileByIDOutputs[0]
// 	t.DestroyProfileByIDOutputs = t.DestroyProfileByIDOutputs[1:]
// 	return output
// }

// func (t *TestProfilesSession) ValidateTest() bool {
// 	return len(t.GetProfileByIDOutputs) == 0 &&
// 		len(t.DestroyProfileByIDOutputs) == 0
// }

// type TestSessionsSession struct {
// 	DestroySessionsForUserByIDInputs  []string
// 	DestroySessionsForUserByIDOutputs []error
// }

// func (t *TestSessionsSession) IsClosed() bool {
// 	panic("Unexpected invocation of IsClosed on TestSessionsSession")
// }

// func (t *TestSessionsSession) Close() {
// 	panic("Unexpected invocation of Close on TestSessionsSession")
// }

// func (t *TestSessionsSession) Logger() log.Logger {
// 	panic("Unexpected invocation of Logger on TestSessionsSession")
// }

// func (t *TestSessionsSession) DestroySessionsForUserByID(userID string) error {
// 	t.DestroySessionsForUserByIDInputs = append(t.DestroySessionsForUserByIDInputs, userID)
// 	output := t.DestroySessionsForUserByIDOutputs[0]
// 	t.DestroySessionsForUserByIDOutputs = t.DestroySessionsForUserByIDOutputs[1:]
// 	return output
// }

// func (t *TestSessionsSession) ValidateTest() bool {
// 	return len(t.DestroySessionsForUserByIDOutputs) == 0
// }

// type GetUserByIDOutput struct {
// 	*user.User
// 	err error
// }

// type PasswordMatchesInput struct {
// 	*user.User
// 	password string
// }

// type TestUsersSession struct {
// 	GetUserByIDInputs      []string
// 	GetUserByIDOutputs     []GetUserByIDOutput
// 	DeleteUserInputs       []*user.User
// 	DeleteUserOutputs      []error
// 	DestroyUserByIDInputs  []string
// 	DestroyUserByIDOutputs []error
// 	PasswordMatchesInputs  []PasswordMatchesInput
// 	PasswordMatchesOutputs []bool
// }

// func (t *TestUsersSession) IsClosed() bool {
// 	panic("Unexpected invocation of IsClosed on TestUsersSession")
// }

// func (t *TestUsersSession) Close() {
// 	panic("Unexpected invocation of Close on TestUsersSession")
// }

// func (t *TestUsersSession) Logger() log.Logger {
// 	panic("Unexpected invocation of Logger on TestUsersSession")
// }

// func (t *TestUsersSession) GetUserByID(profileID string) (*user.User, error) {
// 	t.GetUserByIDInputs = append(t.GetUserByIDInputs, profileID)
// 	output := t.GetUserByIDOutputs[0]
// 	t.GetUserByIDOutputs = t.GetUserByIDOutputs[1:]
// 	return output.User, output.err
// }

// func (t *TestUsersSession) DeleteUser(deleteUser *user.User) error {
// 	t.DeleteUserInputs = append(t.DeleteUserInputs, deleteUser)
// 	output := t.DeleteUserOutputs[0]
// 	t.DeleteUserOutputs = t.DeleteUserOutputs[1:]
// 	return output
// }

// func (t *TestUsersSession) DestroyUserByID(userID string) error {
// 	t.DestroyUserByIDInputs = append(t.DestroyUserByIDInputs, userID)
// 	output := t.DestroyUserByIDOutputs[0]
// 	t.DestroyUserByIDOutputs = t.DestroyUserByIDOutputs[1:]
// 	return output
// }

// func (t *TestUsersSession) PasswordMatches(matchUser *user.User, password string) bool {
// 	t.PasswordMatchesInputs = append(t.PasswordMatchesInputs, PasswordMatchesInput{matchUser, password})
// 	output := t.PasswordMatchesOutputs[0]
// 	t.PasswordMatchesOutputs = t.PasswordMatchesOutputs[1:]
// 	return output
// }

// func (t *TestUsersSession) ValidateTest() bool {
// 	return len(t.GetUserByIDOutputs) == 0 &&
// 		len(t.DeleteUserOutputs) == 0 &&
// 		len(t.DestroyUserByIDOutputs) == 0 &&
// 		len(t.PasswordMatchesOutputs) == 0
// }

// type TestContext struct {
// 	RespondWithErrorInputs                 []*service.Error
// 	RespondWithInternalServerFailureInputs []RespondWithInternalServerFailureInput
// 	RespondWithStatusAndErrorsInputs       []RespondWithStatusAndErrorsInput
// 	RespondWithStatusAndDataInputs         []RespondWithStatusAndDataInput
// 	MetricClientImpl                       *TestMetricClient
// 	UserClientImpl                         *TestUserClient
// 	DataClientImpl                         *TestDataClient
// 	ConfirmationSessionImpl               *TestConfirmationSession
// 	MessagesSessionImpl                    *TestMessagesSession
// 	PermissionsSessionImpl                 *TestPermissionsSession
// 	ProfilesSessionImpl                    *TestProfilesSession
// 	SessionsSessionImpl                    *TestSessionsSession
// 	UsersSessionImpl                       *TestUsersSession
// }

// func NewTestContext() *TestContext {
// 	return &TestContext{
// 		MetricClientImpl:         &TestMetricClient{},
// 		UserClientImpl:           &TestUserClient{},
// 		DataClientImpl:           &TestDataClient{},
// 		ConfirmationSessionImpl: &TestConfirmationSession{},
// 		MessagesSessionImpl:      &TestMessagesSession{},
// 		PermissionsSessionImpl:   &TestPermissionsSession{},
// 		ProfilesSessionImpl:      &TestProfilesSession{},
// 		SessionsSessionImpl:      &TestSessionsSession{},
// 		UsersSessionImpl:         &TestUsersSession{},
// 	}
// }

// func (t *TestContext) Response() rest.ResponseWriter {
// 	panic("Unexpected invocation of Response on TestContext")
// }

// func (t *TestContext) RespondWithError(err *service.Error) {
// 	t.RespondWithErrorInputs = append(t.RespondWithErrorInputs, err)
// }

// func (t *TestContext) RespondWithInternalServerFailure(message string, failure ...interface{}) {
// 	t.RespondWithInternalServerFailureInputs = append(t.RespondWithInternalServerFailureInputs, RespondWithInternalServerFailureInput{message, failure})
// }

// func (t *TestContext) RespondWithStatusAndErrors(statusCode int, errors []*service.Error) {
// 	t.RespondWithStatusAndErrorsInputs = append(t.RespondWithStatusAndErrorsInputs, RespondWithStatusAndErrorsInput{statusCode, errors})
// }

// func (t *TestContext) RespondWithStatusAndData(statusCode int, data interface{}) {
// 	t.RespondWithStatusAndDataInputs = append(t.RespondWithStatusAndDataInputs, RespondWithStatusAndDataInput{statusCode, data})
// }

// func (t *TestContext) MetricClient() metric.Client {
// 	return t.MetricClientImpl
// }

// func (t *TestContext) UserClient() user.Client {
// 	return t.UserClientImpl
// }

// func (t *TestContext) DataClient() dataClient.Client {
// 	return t.DataClientImpl
// }

// func (t *TestContext) ConfirmationSession() confirmationStore.ConfirmationSession {
// 	return t.ConfirmationSessionImpl
// }

// func (t *TestContext) MessagesSession() messageStore.MessagesSession {
// 	return t.MessagesSessionImpl
// }

// func (t *TestContext) PermissionsSession() permissionStore.PermissionsSession {
// 	return t.PermissionsSessionImpl
// }

// func (t *TestContext) ProfilesSession() profileStore.ProfilesSession {
// 	return t.ProfilesSessionImpl
// }

// func (t *TestContext) SessionsSession() sessionStore.SessionsSession {
// 	return t.SessionsSessionImpl
// }

// func (t *TestContext) UsersSession() userStore.UsersSession {
// 	return t.UsersSessionImpl
// }

// func (t *TestContext) ValidateTest() bool {
// 	return (t.MetricClientImpl == nil || t.MetricClientImpl.ValidateTest()) &&
// 		(t.UserClientImpl == nil || t.UserClientImpl.ValidateTest()) &&
// 		(t.DataClientImpl == nil || t.DataClientImpl.ValidateTest()) &&
// 		(t.ConfirmationSessionImpl == nil || t.ConfirmationSessionImpl.ValidateTest()) &&
// 		(t.MessagesSessionImpl == nil || t.MessagesSessionImpl.ValidateTest()) &&
// 		(t.PermissionsSessionImpl == nil || t.PermissionsSessionImpl.ValidateTest()) &&
// 		(t.ProfilesSessionImpl == nil || t.ProfilesSessionImpl.ValidateTest()) &&
// 		(t.SessionsSessionImpl == nil || t.SessionsSessionImpl.ValidateTest()) &&
// 		(t.UsersSessionImpl == nil || t.UsersSessionImpl.ValidateTest())
// }
