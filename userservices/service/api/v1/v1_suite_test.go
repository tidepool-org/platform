package v1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"net/http"
	"net/url"
	"testing"

	"github.com/ant0ine/go-json-rest/rest"

	dataservicesClient "github.com/tidepool-org/platform/dataservices/client"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/null"
	messageStore "github.com/tidepool-org/platform/message/store"
	metricservicesClient "github.com/tidepool-org/platform/metricservices/client"
	notificationStore "github.com/tidepool-org/platform/notification/store"
	permissionStore "github.com/tidepool-org/platform/permission/store"
	"github.com/tidepool-org/platform/profile"
	profileStore "github.com/tidepool-org/platform/profile/store"
	"github.com/tidepool-org/platform/service"
	sessionStore "github.com/tidepool-org/platform/session/store"
	commonStore "github.com/tidepool-org/platform/store"
	"github.com/tidepool-org/platform/user"
	userStore "github.com/tidepool-org/platform/user/store"
	userservicesClient "github.com/tidepool-org/platform/userservices/client"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "userservices/service/api/v1")
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
	context metricservicesClient.Context
	metric  string
	data    []map[string]string
}

type TestMetricServicesClient struct {
	RecordMetricInputs  []RecordMetricInput
	RecordMetricOutputs []error
}

func (t *TestMetricServicesClient) RecordMetric(context metricservicesClient.Context, metric string, data ...map[string]string) error {
	t.RecordMetricInputs = append(t.RecordMetricInputs, RecordMetricInput{context, metric, data})
	output := t.RecordMetricOutputs[0]
	t.RecordMetricOutputs = t.RecordMetricOutputs[1:]
	return output
}

func (t *TestMetricServicesClient) ValidateTest() bool {
	return len(t.RecordMetricOutputs) == 0
}

type GetUserPermissionsInput struct {
	context       service.Context
	requestUserID string
	targetUserID  string
}

type GetUserPermissionsOutput struct {
	permissions userservicesClient.Permissions
	err         error
}

type TestUserServicesClient struct {
	GetUserPermissionsInputs  []GetUserPermissionsInput
	GetUserPermissionsOutputs []GetUserPermissionsOutput
}

func (t *TestUserServicesClient) Start() error {
	panic("Unexpected invocation of Start on TestUserServicesClient")
}

func (t *TestUserServicesClient) Close() {
	panic("Unexpected invocation of Close on TestUserServicesClient")
}

func (t *TestUserServicesClient) ValidateAuthenticationToken(context service.Context, authenticationToken string) (userservicesClient.AuthenticationDetails, error) {
	panic("Unexpected invocation of ValidateAuthenticationToken on TestUserServicesClient")
}

func (t *TestUserServicesClient) GetUserPermissions(context service.Context, requestUserID string, targetUserID string) (userservicesClient.Permissions, error) {
	t.GetUserPermissionsInputs = append(t.GetUserPermissionsInputs, GetUserPermissionsInput{context, requestUserID, targetUserID})
	output := t.GetUserPermissionsOutputs[0]
	t.GetUserPermissionsOutputs = t.GetUserPermissionsOutputs[1:]
	return output.permissions, output.err
}

func (t *TestUserServicesClient) ServerToken() (string, error) {
	panic("Unexpected invocation of ServerToken on TestUserServicesClient")
}

func (t *TestUserServicesClient) ValidateTest() bool {
	return len(t.GetUserPermissionsOutputs) == 0
}

type DestroyDataForUserByIDInput struct {
	context dataservicesClient.Context
	userID  string
}

type TestDataServicesClient struct {
	DestroyDataForUserByIDInputs  []DestroyDataForUserByIDInput
	DestroyDataForUserByIDOutputs []error
}

func (t *TestDataServicesClient) DestroyDataForUserByID(context dataservicesClient.Context, userID string) error {
	t.DestroyDataForUserByIDInputs = append(t.DestroyDataForUserByIDInputs, DestroyDataForUserByIDInput{context, userID})
	output := t.DestroyDataForUserByIDOutputs[0]
	t.DestroyDataForUserByIDOutputs = t.DestroyDataForUserByIDOutputs[1:]
	return output
}

func (t *TestDataServicesClient) ValidateTest() bool {
	return len(t.DestroyDataForUserByIDOutputs) == 0
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

func (t *TestMessageStoreSession) SetAgent(agent commonStore.Agent) {
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

type TestNotificationStoreSession struct {
	DestroyNotificationsForUserByIDInputs  []string
	DestroyNotificationsForUserByIDOutputs []error
}

func (t *TestNotificationStoreSession) IsClosed() bool {
	panic("Unexpected invocation of IsClosed on TestNotificationStoreSession")
}

func (t *TestNotificationStoreSession) Close() {
	panic("Unexpected invocation of Close on TestNotificationStoreSession")
}

func (t *TestNotificationStoreSession) Logger() log.Logger {
	panic("Unexpected invocation of Logger on TestNotificationStoreSession")
}

func (t *TestNotificationStoreSession) SetAgent(agent commonStore.Agent) {
	panic("Unexpected invocation of SetAgent on TestNotificationStoreSession")
}

func (t *TestNotificationStoreSession) DestroyNotificationsForUserByID(userID string) error {
	t.DestroyNotificationsForUserByIDInputs = append(t.DestroyNotificationsForUserByIDInputs, userID)
	output := t.DestroyNotificationsForUserByIDOutputs[0]
	t.DestroyNotificationsForUserByIDOutputs = t.DestroyNotificationsForUserByIDOutputs[1:]
	return output
}

func (t *TestNotificationStoreSession) ValidateTest() bool {
	return len(t.DestroyNotificationsForUserByIDOutputs) == 0
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

func (t *TestPermissionStoreSession) SetAgent(agent commonStore.Agent) {
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

func (t *TestProfileStoreSession) SetAgent(agent commonStore.Agent) {
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

func (t *TestSessionStoreSession) SetAgent(agent commonStore.Agent) {
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

func (t *TestUserStoreSession) SetAgent(agent commonStore.Agent) {
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

type TestAuthenticationDetails struct {
	IsServerOutputs []bool
	UserIDOutputs   []string
}

func (t *TestAuthenticationDetails) Token() string {
	panic("Unexpected invocation of Token on TestAuthenticationDetails")
}

func (t *TestAuthenticationDetails) IsServer() bool {
	output := t.IsServerOutputs[0]
	t.IsServerOutputs = t.IsServerOutputs[1:]
	return output
}

func (t *TestAuthenticationDetails) UserID() string {
	output := t.UserIDOutputs[0]
	t.UserIDOutputs = t.UserIDOutputs[1:]
	return output
}

func (t *TestAuthenticationDetails) ValidateTest() bool {
	return len(t.IsServerOutputs) == 0 &&
		len(t.UserIDOutputs) == 0
}

type TestContext struct {
	LoggerImpl                             log.Logger
	RequestImpl                            *rest.Request
	RespondWithErrorInputs                 []*service.Error
	RespondWithInternalServerFailureInputs []RespondWithInternalServerFailureInput
	RespondWithStatusAndErrorsInputs       []RespondWithStatusAndErrorsInput
	RespondWithStatusAndDataInputs         []RespondWithStatusAndDataInput
	MetricServicesClientImpl               *TestMetricServicesClient
	UserServicesClientImpl                 *TestUserServicesClient
	DataServicesClientImpl                 *TestDataServicesClient
	MessageStoreSessionImpl                *TestMessageStoreSession
	NotificationStoreSessionImpl           *TestNotificationStoreSession
	PermissionStoreSessionImpl             *TestPermissionStoreSession
	ProfileStoreSessionImpl                *TestProfileStoreSession
	SessionStoreSessionImpl                *TestSessionStoreSession
	UserStoreSessionImpl                   *TestUserStoreSession
	AuthenticationDetailsImpl              *TestAuthenticationDetails
}

func NewTestContext() *TestContext {
	return &TestContext{
		LoggerImpl: null.NewLogger(),
		RequestImpl: &rest.Request{
			Request: &http.Request{
				URL: &url.URL{},
			},
			PathParams: map[string]string{},
		},
		MetricServicesClientImpl:     &TestMetricServicesClient{},
		UserServicesClientImpl:       &TestUserServicesClient{},
		DataServicesClientImpl:       &TestDataServicesClient{},
		MessageStoreSessionImpl:      &TestMessageStoreSession{},
		NotificationStoreSessionImpl: &TestNotificationStoreSession{},
		PermissionStoreSessionImpl:   &TestPermissionStoreSession{},
		ProfileStoreSessionImpl:      &TestProfileStoreSession{},
		SessionStoreSessionImpl:      &TestSessionStoreSession{},
		UserStoreSessionImpl:         &TestUserStoreSession{},
		AuthenticationDetailsImpl:    &TestAuthenticationDetails{},
	}
}

func (t *TestContext) Logger() log.Logger {
	return t.LoggerImpl
}

func (t *TestContext) Request() *rest.Request {
	return t.RequestImpl
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

func (t *TestContext) MetricServicesClient() metricservicesClient.Client {
	return t.MetricServicesClientImpl
}

func (t *TestContext) UserServicesClient() userservicesClient.Client {
	return t.UserServicesClientImpl
}

func (t *TestContext) DataServicesClient() dataservicesClient.Client {
	return t.DataServicesClientImpl
}

func (t *TestContext) MessageStoreSession() messageStore.Session {
	return t.MessageStoreSessionImpl
}

func (t *TestContext) NotificationStoreSession() notificationStore.Session {
	return t.NotificationStoreSessionImpl
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

func (t *TestContext) AuthenticationDetails() userservicesClient.AuthenticationDetails {
	return t.AuthenticationDetailsImpl
}

func (t *TestContext) SetAuthenticationDetails(authenticationDetails userservicesClient.AuthenticationDetails) {
	panic("Unexpected invocation of SetAuthenticationDetails on TestContext")
}

func (t *TestContext) ValidateTest() bool {
	return (t.MetricServicesClientImpl == nil || t.MetricServicesClientImpl.ValidateTest()) &&
		(t.UserServicesClientImpl == nil || t.UserServicesClientImpl.ValidateTest()) &&
		(t.DataServicesClientImpl == nil || t.DataServicesClientImpl.ValidateTest()) &&
		(t.MessageStoreSessionImpl == nil || t.MessageStoreSessionImpl.ValidateTest()) &&
		(t.NotificationStoreSessionImpl == nil || t.NotificationStoreSessionImpl.ValidateTest()) &&
		(t.PermissionStoreSessionImpl == nil || t.PermissionStoreSessionImpl.ValidateTest()) &&
		(t.ProfileStoreSessionImpl == nil || t.ProfileStoreSessionImpl.ValidateTest()) &&
		(t.SessionStoreSessionImpl == nil || t.SessionStoreSessionImpl.ValidateTest()) &&
		(t.UserStoreSessionImpl == nil || t.UserStoreSessionImpl.ValidateTest()) &&
		(t.AuthenticationDetailsImpl == nil || t.AuthenticationDetailsImpl.ValidateTest())
}
