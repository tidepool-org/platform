package v1_test

import (
	"github.com/ant0ine/go-json-rest/rest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"

	"github.com/tidepool-org/platform/auth"
	testAuth "github.com/tidepool-org/platform/auth/test"
	confirmationStore "github.com/tidepool-org/platform/confirmation/store"
	dataClient "github.com/tidepool-org/platform/data/client"
	"github.com/tidepool-org/platform/log"
	messageStore "github.com/tidepool-org/platform/message/store"
	metricClient "github.com/tidepool-org/platform/metric/client"
	permissionStore "github.com/tidepool-org/platform/permission/store"
	"github.com/tidepool-org/platform/profile"
	profileStore "github.com/tidepool-org/platform/profile/store"
	"github.com/tidepool-org/platform/service"
	sessionStore "github.com/tidepool-org/platform/session/store"
	"github.com/tidepool-org/platform/store"
	"github.com/tidepool-org/platform/user"
	userClient "github.com/tidepool-org/platform/user/client"
	userStore "github.com/tidepool-org/platform/user/store"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "user/service/api/v1")
}

type TestFlags struct {
	flags map[string]bool
}

func NewTestFlags() *TestFlags {
	return &TestFlags{
		flags: map[string]bool{},
	}
}

func (t *TestFlags) Set(flagsToSet ...string) *TestFlags {
	flags := map[string]bool{}
	for key, value := range t.flags {
		flags[key] = value
	}

	for _, flagToSet := range flagsToSet {
		flags[flagToSet] = true
	}

	return &TestFlags{
		flags: flags,
	}
}

func (t *TestFlags) IsSet(flag string) bool {
	set, ok := t.flags[flag]
	return ok && set
}

type RespondWithInternalServerFailureInput struct {
	message string
	failure []interface{}
}

type RespondWithStatusAndErrorsInput struct {
	statusCode int
	errors     []*service.Error
}

type RespondWithStatusAndDataInput struct {
	statusCode int
	data       interface{}
}

type RecordMetricInput struct {
	context auth.Context
	metric  string
	data    []map[string]string
}

type TestMetricClient struct {
	RecordMetricInputs  []RecordMetricInput
	RecordMetricOutputs []error
}

func (t *TestMetricClient) RecordMetric(context auth.Context, metric string, data ...map[string]string) error {
	t.RecordMetricInputs = append(t.RecordMetricInputs, RecordMetricInput{context, metric, data})
	output := t.RecordMetricOutputs[0]
	t.RecordMetricOutputs = t.RecordMetricOutputs[1:]
	return output
}

func (t *TestMetricClient) ValidateTest() bool {
	return len(t.RecordMetricOutputs) == 0
}

type GetUserPermissionsInput struct {
	context       auth.Context
	requestUserID string
	targetUserID  string
}

type GetUserPermissionsOutput struct {
	permissions userClient.Permissions
	err         error
}

type TestUserClient struct {
	GetUserPermissionsInputs  []GetUserPermissionsInput
	GetUserPermissionsOutputs []GetUserPermissionsOutput
}

func (t *TestUserClient) GetUserPermissions(context auth.Context, requestUserID string, targetUserID string) (userClient.Permissions, error) {
	t.GetUserPermissionsInputs = append(t.GetUserPermissionsInputs, GetUserPermissionsInput{context, requestUserID, targetUserID})
	output := t.GetUserPermissionsOutputs[0]
	t.GetUserPermissionsOutputs = t.GetUserPermissionsOutputs[1:]
	return output.permissions, output.err
}

func (t *TestUserClient) ValidateTest() bool {
	return len(t.GetUserPermissionsOutputs) == 0
}

type DestroyDataForUserByIDInput struct {
	context auth.Context
	userID  string
}

type TestDataClient struct {
	DestroyDataForUserByIDInputs  []DestroyDataForUserByIDInput
	DestroyDataForUserByIDOutputs []error
}

func (t *TestDataClient) DestroyDataForUserByID(context auth.Context, userID string) error {
	t.DestroyDataForUserByIDInputs = append(t.DestroyDataForUserByIDInputs, DestroyDataForUserByIDInput{context, userID})
	output := t.DestroyDataForUserByIDOutputs[0]
	t.DestroyDataForUserByIDOutputs = t.DestroyDataForUserByIDOutputs[1:]
	return output
}

func (t *TestDataClient) ValidateTest() bool {
	return len(t.DestroyDataForUserByIDOutputs) == 0
}

type TestConfirmationStoreSession struct {
	DestroyConfirmationsForUserByIDInputs  []string
	DestroyConfirmationsForUserByIDOutputs []error
}

func (t *TestConfirmationStoreSession) IsClosed() bool {
	panic("Unexpected invocation of IsClosed on TestConfirmationStoreSession")
}

func (t *TestConfirmationStoreSession) Close() {
	panic("Unexpected invocation of Close on TestConfirmationStoreSession")
}

func (t *TestConfirmationStoreSession) Logger() log.Logger {
	panic("Unexpected invocation of Logger on TestConfirmationStoreSession")
}

func (t *TestConfirmationStoreSession) SetAgent(agent store.Agent) {
	panic("Unexpected invocation of SetAgent on TestConfirmationStoreSession")
}

func (t *TestConfirmationStoreSession) DestroyConfirmationsForUserByID(userID string) error {
	t.DestroyConfirmationsForUserByIDInputs = append(t.DestroyConfirmationsForUserByIDInputs, userID)
	output := t.DestroyConfirmationsForUserByIDOutputs[0]
	t.DestroyConfirmationsForUserByIDOutputs = t.DestroyConfirmationsForUserByIDOutputs[1:]
	return output
}

func (t *TestConfirmationStoreSession) ValidateTest() bool {
	return len(t.DestroyConfirmationsForUserByIDOutputs) == 0
}

type TestMessageStoreSession struct {
	DeleteMessagesFromUserInputs      []*messageStore.User
	DeleteMessagesFromUserOutputs     []error
	DestroyMessagesForUserByIDInputs  []string
	DestroyMessagesForUserByIDOutputs []error
}

func (t *TestMessageStoreSession) IsClosed() bool {
	panic("Unexpected invocation of IsClosed on TestMessageStoreSession")
}

func (t *TestMessageStoreSession) Close() {
	panic("Unexpected invocation of Close on TestMessageStoreSession")
}

func (t *TestMessageStoreSession) Logger() log.Logger {
	panic("Unexpected invocation of Logger on TestMessageStoreSession")
}

func (t *TestMessageStoreSession) SetAgent(agent store.Agent) {
	panic("Unexpected invocation of SetAgent on TestMessageStoreSession")
}

func (t *TestMessageStoreSession) DeleteMessagesFromUser(deleteUser *messageStore.User) error {
	t.DeleteMessagesFromUserInputs = append(t.DeleteMessagesFromUserInputs, deleteUser)
	output := t.DeleteMessagesFromUserOutputs[0]
	t.DeleteMessagesFromUserOutputs = t.DeleteMessagesFromUserOutputs[1:]
	return output
}

func (t *TestMessageStoreSession) DestroyMessagesForUserByID(userID string) error {
	t.DestroyMessagesForUserByIDInputs = append(t.DestroyMessagesForUserByIDInputs, userID)
	output := t.DestroyMessagesForUserByIDOutputs[0]
	t.DestroyMessagesForUserByIDOutputs = t.DestroyMessagesForUserByIDOutputs[1:]
	return output
}

func (t *TestMessageStoreSession) ValidateTest() bool {
	return len(t.DeleteMessagesFromUserOutputs) == 0 &&
		len(t.DestroyMessagesForUserByIDOutputs) == 0
}

type TestPermissionStoreSession struct {
	DestroyPermissionsForUserByIDInputs  []string
	DestroyPermissionsForUserByIDOutputs []error
}

func (t *TestPermissionStoreSession) IsClosed() bool {
	panic("Unexpected invocation of IsClosed on TestPermissionStoreSession")
}

func (t *TestPermissionStoreSession) Close() {
	panic("Unexpected invocation of Close on TestPermissionStoreSession")
}

func (t *TestPermissionStoreSession) Logger() log.Logger {
	panic("Unexpected invocation of Logger on TestPermissionStoreSession")
}

func (t *TestPermissionStoreSession) SetAgent(agent store.Agent) {
	panic("Unexpected invocation of SetAgent on TestPermissionStoreSession")
}

func (t *TestPermissionStoreSession) DestroyPermissionsForUserByID(userID string) error {
	t.DestroyPermissionsForUserByIDInputs = append(t.DestroyPermissionsForUserByIDInputs, userID)
	output := t.DestroyPermissionsForUserByIDOutputs[0]
	t.DestroyPermissionsForUserByIDOutputs = t.DestroyPermissionsForUserByIDOutputs[1:]
	return output
}

func (t *TestPermissionStoreSession) ValidateTest() bool {
	return len(t.DestroyPermissionsForUserByIDOutputs) == 0
}

type GetProfileByIDOutput struct {
	*profile.Profile
	err error
}

type TestProfileStoreSession struct {
	GetProfileByIDInputs      []string
	GetProfileByIDOutputs     []GetProfileByIDOutput
	DestroyProfileByIDInputs  []string
	DestroyProfileByIDOutputs []error
}

func (t *TestProfileStoreSession) IsClosed() bool {
	panic("Unexpected invocation of IsClosed on TestProfileStoreSession")
}

func (t *TestProfileStoreSession) Close() {
	panic("Unexpected invocation of Close on TestProfileStoreSession")
}

func (t *TestProfileStoreSession) Logger() log.Logger {
	panic("Unexpected invocation of Logger on TestProfileStoreSession")
}

func (t *TestProfileStoreSession) SetAgent(agent store.Agent) {
	panic("Unexpected invocation of SetAgent on TestProfileStoreSession")
}

func (t *TestProfileStoreSession) GetProfileByID(profileID string) (*profile.Profile, error) {
	t.GetProfileByIDInputs = append(t.GetProfileByIDInputs, profileID)
	output := t.GetProfileByIDOutputs[0]
	t.GetProfileByIDOutputs = t.GetProfileByIDOutputs[1:]
	return output.Profile, output.err
}

func (t *TestProfileStoreSession) DestroyProfileByID(profileID string) error {
	t.DestroyProfileByIDInputs = append(t.DestroyProfileByIDInputs, profileID)
	output := t.DestroyProfileByIDOutputs[0]
	t.DestroyProfileByIDOutputs = t.DestroyProfileByIDOutputs[1:]
	return output
}

func (t *TestProfileStoreSession) ValidateTest() bool {
	return len(t.GetProfileByIDOutputs) == 0 &&
		len(t.DestroyProfileByIDOutputs) == 0
}

type TestSessionStoreSession struct {
	DestroySessionsForUserByIDInputs  []string
	DestroySessionsForUserByIDOutputs []error
}

func (t *TestSessionStoreSession) IsClosed() bool {
	panic("Unexpected invocation of IsClosed on TestSessionStoreSession")
}

func (t *TestSessionStoreSession) Close() {
	panic("Unexpected invocation of Close on TestSessionStoreSession")
}

func (t *TestSessionStoreSession) Logger() log.Logger {
	panic("Unexpected invocation of Logger on TestSessionStoreSession")
}

func (t *TestSessionStoreSession) SetAgent(agent store.Agent) {
	panic("Unexpected invocation of SetAgent on TestSessionStoreSession")
}

func (t *TestSessionStoreSession) DestroySessionsForUserByID(userID string) error {
	t.DestroySessionsForUserByIDInputs = append(t.DestroySessionsForUserByIDInputs, userID)
	output := t.DestroySessionsForUserByIDOutputs[0]
	t.DestroySessionsForUserByIDOutputs = t.DestroySessionsForUserByIDOutputs[1:]
	return output
}

func (t *TestSessionStoreSession) ValidateTest() bool {
	return len(t.DestroySessionsForUserByIDOutputs) == 0
}

type GetUserByIDOutput struct {
	*user.User
	err error
}

type PasswordMatchesInput struct {
	*user.User
	password string
}

type TestUserStoreSession struct {
	GetUserByIDInputs      []string
	GetUserByIDOutputs     []GetUserByIDOutput
	DeleteUserInputs       []*user.User
	DeleteUserOutputs      []error
	DestroyUserByIDInputs  []string
	DestroyUserByIDOutputs []error
	PasswordMatchesInputs  []PasswordMatchesInput
	PasswordMatchesOutputs []bool
}

func (t *TestUserStoreSession) IsClosed() bool {
	panic("Unexpected invocation of IsClosed on TestUserStoreSession")
}

func (t *TestUserStoreSession) Close() {
	panic("Unexpected invocation of Close on TestUserStoreSession")
}

func (t *TestUserStoreSession) Logger() log.Logger {
	panic("Unexpected invocation of Logger on TestUserStoreSession")
}

func (t *TestUserStoreSession) SetAgent(agent store.Agent) {
	panic("Unexpected invocation of SetAgent on TestUserStoreSession")
}

func (t *TestUserStoreSession) GetUserByID(profileID string) (*user.User, error) {
	t.GetUserByIDInputs = append(t.GetUserByIDInputs, profileID)
	output := t.GetUserByIDOutputs[0]
	t.GetUserByIDOutputs = t.GetUserByIDOutputs[1:]
	return output.User, output.err
}

func (t *TestUserStoreSession) DeleteUser(deleteUser *user.User) error {
	t.DeleteUserInputs = append(t.DeleteUserInputs, deleteUser)
	output := t.DeleteUserOutputs[0]
	t.DeleteUserOutputs = t.DeleteUserOutputs[1:]
	return output
}

func (t *TestUserStoreSession) DestroyUserByID(userID string) error {
	t.DestroyUserByIDInputs = append(t.DestroyUserByIDInputs, userID)
	output := t.DestroyUserByIDOutputs[0]
	t.DestroyUserByIDOutputs = t.DestroyUserByIDOutputs[1:]
	return output
}

func (t *TestUserStoreSession) PasswordMatches(matchUser *user.User, password string) bool {
	t.PasswordMatchesInputs = append(t.PasswordMatchesInputs, PasswordMatchesInput{matchUser, password})
	output := t.PasswordMatchesOutputs[0]
	t.PasswordMatchesOutputs = t.PasswordMatchesOutputs[1:]
	return output
}

func (t *TestUserStoreSession) ValidateTest() bool {
	return len(t.GetUserByIDOutputs) == 0 &&
		len(t.DeleteUserOutputs) == 0 &&
		len(t.DestroyUserByIDOutputs) == 0 &&
		len(t.PasswordMatchesOutputs) == 0
}

type TestContext struct {
	*testAuth.Context
	RespondWithErrorInputs                 []*service.Error
	RespondWithInternalServerFailureInputs []RespondWithInternalServerFailureInput
	RespondWithStatusAndErrorsInputs       []RespondWithStatusAndErrorsInput
	RespondWithStatusAndDataInputs         []RespondWithStatusAndDataInput
	MetricClientImpl                       *TestMetricClient
	UserClientImpl                         *TestUserClient
	DataClientImpl                         *TestDataClient
	ConfirmationStoreSessionImpl           *TestConfirmationStoreSession
	MessageStoreSessionImpl                *TestMessageStoreSession
	PermissionStoreSessionImpl             *TestPermissionStoreSession
	ProfileStoreSessionImpl                *TestProfileStoreSession
	SessionStoreSessionImpl                *TestSessionStoreSession
	UserStoreSessionImpl                   *TestUserStoreSession
}

func NewTestContext() *TestContext {
	return &TestContext{
		Context:                      testAuth.NewContext(),
		MetricClientImpl:             &TestMetricClient{},
		UserClientImpl:               &TestUserClient{},
		DataClientImpl:               &TestDataClient{},
		ConfirmationStoreSessionImpl: &TestConfirmationStoreSession{},
		MessageStoreSessionImpl:      &TestMessageStoreSession{},
		PermissionStoreSessionImpl:   &TestPermissionStoreSession{},
		ProfileStoreSessionImpl:      &TestProfileStoreSession{},
		SessionStoreSessionImpl:      &TestSessionStoreSession{},
		UserStoreSessionImpl:         &TestUserStoreSession{},
	}
}

func (t *TestContext) Response() rest.ResponseWriter {
	panic("Unexpected invocation of Response on TestContext")
}

func (t *TestContext) RespondWithError(err *service.Error) {
	t.RespondWithErrorInputs = append(t.RespondWithErrorInputs, err)
}

func (t *TestContext) RespondWithInternalServerFailure(message string, failure ...interface{}) {
	t.RespondWithInternalServerFailureInputs = append(t.RespondWithInternalServerFailureInputs, RespondWithInternalServerFailureInput{message, failure})
}

func (t *TestContext) RespondWithStatusAndErrors(statusCode int, errors []*service.Error) {
	t.RespondWithStatusAndErrorsInputs = append(t.RespondWithStatusAndErrorsInputs, RespondWithStatusAndErrorsInput{statusCode, errors})
}

func (t *TestContext) RespondWithStatusAndData(statusCode int, data interface{}) {
	t.RespondWithStatusAndDataInputs = append(t.RespondWithStatusAndDataInputs, RespondWithStatusAndDataInput{statusCode, data})
}

func (t *TestContext) MetricClient() metricClient.Client {
	return t.MetricClientImpl
}

func (t *TestContext) UserClient() userClient.Client {
	return t.UserClientImpl
}

func (t *TestContext) DataClient() dataClient.Client {
	return t.DataClientImpl
}

func (t *TestContext) ConfirmationStoreSession() confirmationStore.Session {
	return t.ConfirmationStoreSessionImpl
}

func (t *TestContext) MessageStoreSession() messageStore.Session {
	return t.MessageStoreSessionImpl
}

func (t *TestContext) PermissionStoreSession() permissionStore.Session {
	return t.PermissionStoreSessionImpl
}

func (t *TestContext) ProfileStoreSession() profileStore.Session {
	return t.ProfileStoreSessionImpl
}

func (t *TestContext) SessionStoreSession() sessionStore.Session {
	return t.SessionStoreSessionImpl
}

func (t *TestContext) UserStoreSession() userStore.Session {
	return t.UserStoreSessionImpl
}

func (t *TestContext) ValidateTest() bool {
	return (t.Context == nil) || (t.Context.UnusedOutputsCount() == 0) &&
		(t.MetricClientImpl == nil || t.MetricClientImpl.ValidateTest()) &&
		(t.UserClientImpl == nil || t.UserClientImpl.ValidateTest()) &&
		(t.DataClientImpl == nil || t.DataClientImpl.ValidateTest()) &&
		(t.ConfirmationStoreSessionImpl == nil || t.ConfirmationStoreSessionImpl.ValidateTest()) &&
		(t.MessageStoreSessionImpl == nil || t.MessageStoreSessionImpl.ValidateTest()) &&
		(t.PermissionStoreSessionImpl == nil || t.PermissionStoreSessionImpl.ValidateTest()) &&
		(t.ProfileStoreSessionImpl == nil || t.ProfileStoreSessionImpl.ValidateTest()) &&
		(t.SessionStoreSessionImpl == nil || t.SessionStoreSessionImpl.ValidateTest()) &&
		(t.UserStoreSessionImpl == nil || t.UserStoreSessionImpl.ValidateTest())
}
