package work

import (
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/user"
	"github.com/tidepool-org/platform/work"
)

//go:generate mockgen -source=mixin.go -destination=test/mixin_mocks.go -package=test -typed

const MetadataKeyUserID = "userId"

type Metadata struct {
	UserID *string `json:"userId,omitempty" bson:"userId,omitempty"`
}

func (m *Metadata) Parse(parser structure.ObjectParser) {
	m.UserID = parser.String(MetadataKeyUserID)
}

func (m *Metadata) Validate(validator structure.Validator) {
	validator.String(MetadataKeyUserID, m.UserID).Using(user.IDValidator)
}

type Mixin interface {
	UserClient() user.Client

	HasUser() bool
	User() *user.User
	SetUser(user *user.User) *work.ProcessResult

	FetchUser(userID string) *work.ProcessResult

	AddUserToContext()
}

type MixinFromWork interface {
	Mixin

	HasWorkMetadata() bool

	FetchUserFromWorkMetadata() *work.ProcessResult
	UpdateWorkMetadataFromUser() *work.ProcessResult
}

func NewMixin(provider work.Provider, userClient user.Client) (Mixin, error) {
	if provider == nil {
		return nil, errors.New("provider is missing")
	}
	if userClient == nil {
		return nil, errors.New("user client is missing")
	}
	return &mixin{
		Provider:   provider,
		userClient: userClient,
	}, nil
}

func NewMixinFromWork(provider work.Provider, userClient user.Client, workMetadata *Metadata) (MixinFromWork, error) {
	if provider == nil {
		return nil, errors.New("provider is missing")
	}
	if userClient == nil {
		return nil, errors.New("user client is missing")
	}
	if workMetadata == nil {
		return nil, errors.New("work metadata is missing")
	}
	return &mixin{
		Provider:     provider,
		userClient:   userClient,
		workMetadata: workMetadata,
	}, nil
}

type mixin struct {
	work.Provider
	userClient   user.Client
	user         *user.User
	workMetadata *Metadata
}

func (m *mixin) UserClient() user.Client {
	return m.userClient
}

func (m *mixin) HasUser() bool {
	return m.user != nil
}

func (m *mixin) User() *user.User {
	return m.user
}

func (m *mixin) SetUser(user *user.User) *work.ProcessResult {
	m.user = user
	m.AddUserToContext()
	return nil
}

func (m *mixin) FetchUser(userID string) *work.ProcessResult {
	if user, err := m.userClient.Get(m.Context(), userID); err != nil {
		return m.Failing(errors.Wrap(err, "unable to get user"))
	} else if user == nil {
		return m.Failed(errors.New("user is missing"))
	} else {
		return m.SetUser(user)
	}
}

func (m *mixin) HasWorkMetadata() bool {
	return m.workMetadata != nil
}

func (m *mixin) FetchUserFromWorkMetadata() *work.ProcessResult {
	if m.workMetadata == nil {
		return m.Failed(errors.New("work metadata is missing"))
	} else if m.workMetadata.UserID == nil {
		return m.Failed(errors.New("work metadata user id is missing"))
	} else {
		return m.FetchUser(*m.workMetadata.UserID)
	}
}

func (m *mixin) UpdateWorkMetadataFromUser() *work.ProcessResult {
	if m.user == nil {
		return m.Failed(errors.New("user is missing"))
	} else if m.user.UserID == nil {
		return m.Failed(errors.New("user id is missing"))
	} else if m.workMetadata == nil {
		return m.Failed(errors.New("work metadata is missing"))
	}
	m.workMetadata.UserID = pointer.Clone(m.user.UserID)
	return nil
}

func (m *mixin) AddUserToContext() {
	m.AddFieldToContext("user", userToFields(m.user))
}

func userToFields(user *user.User) log.Fields {
	if user == nil {
		return nil
	}
	return log.Fields{
		"id": user.UserID,
	}
}
