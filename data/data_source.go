package data

import (
	"context"
	"net/http"
	"regexp"
	"time"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"
)

type DataSourceAccessor interface {
	ListUserDataSources(ctx context.Context, userID string, filter *DataSourceFilter, pagination *page.Pagination) (DataSources, error)
	CreateUserDataSource(ctx context.Context, userID string, create *DataSourceCreate) (*DataSource, error)
	GetDataSource(ctx context.Context, id string) (*DataSource, error)
	UpdateDataSource(ctx context.Context, id string, update *DataSourceUpdate) (*DataSource, error)
	DeleteDataSource(ctx context.Context, id string) error
}

const (
	DataSourceStateConnected    = "connected"
	DataSourceStateDisconnected = "disconnected"
	DataSourceStateError        = "error"
)

func DataSourceStates() []string {
	return []string{
		DataSourceStateConnected,
		DataSourceStateDisconnected,
		DataSourceStateError,
	}
}

type DataSourceFilter struct {
	ProviderType      *string `json:"providerType,omitempty" bson:"providerType,omitempty"`
	ProviderName      *string `json:"providerName,omitempty" bson:"providerName,omitempty"`
	ProviderSessionID *string `json:"providerSessionId,omitempty" bson:"providerSessionId,omitempty"`
	State             *string `json:"state,omitempty" bson:"state,omitempty"`
}

func NewDataSourceFilter() *DataSourceFilter {
	return &DataSourceFilter{}
}

func (d *DataSourceFilter) Parse(parser structure.ObjectParser) {
	d.ProviderType = parser.String("providerType")
	d.ProviderName = parser.String("providerName")
	d.ProviderSessionID = parser.String("providerSessionId")
	d.State = parser.String("state")
}

func (d *DataSourceFilter) Validate(validator structure.Validator) {
	validator.String("providerType", d.ProviderType).OneOf(auth.ProviderTypes()...)
	validator.String("providerName", d.ProviderName).NotEmpty()
	validator.String("providerSessionId", d.ProviderSessionID).Using(auth.ProviderSessionIDValidator)
	validator.String("state", d.State).OneOf(DataSourceStates()...)
}

func (d *DataSourceFilter) MutateRequest(req *http.Request) error {
	parameters := map[string]string{}
	if d.ProviderType != nil {
		parameters["providerType"] = *d.ProviderType
	}
	if d.ProviderName != nil {
		parameters["providerName"] = *d.ProviderName
	}
	if d.ProviderSessionID != nil {
		parameters["providerSessionId"] = *d.ProviderSessionID
	}
	if d.State != nil {
		parameters["state"] = *d.State
	}
	return request.NewParametersMutator(parameters).MutateRequest(req)
}

type DataSourceCreate struct {
	ProviderType      string `json:"providerType" bson:"providerType"`
	ProviderName      string `json:"providerName" bson:"providerName"`
	ProviderSessionID string `json:"providerSessionId" bson:"providerSessionId"`
	State             string `json:"state,omitempty" bson:"state,omitempty"`
}

func NewDataSourceCreate() *DataSourceCreate {
	return &DataSourceCreate{}
}

func (d *DataSourceCreate) Parse(parser structure.ObjectParser) {
	if ptr := parser.String("providerType"); ptr != nil {
		d.ProviderType = *ptr
	}
	if ptr := parser.String("providerName"); ptr != nil {
		d.ProviderName = *ptr
	}
	if ptr := parser.String("providerSessionId"); ptr != nil {
		d.ProviderSessionID = *ptr
	}
	if ptr := parser.String("state"); ptr != nil {
		d.State = *ptr
	}
}

func (d *DataSourceCreate) Validate(validator structure.Validator) {
	validator.String("providerType", &d.ProviderType).OneOf(auth.ProviderTypes()...)
	validator.String("providerName", &d.ProviderName).NotEmpty()
	validator.String("providerSessionId", &d.ProviderSessionID).Using(auth.ProviderSessionIDValidator)
	validator.String("state", &d.State).OneOf(DataSourceStates()...)
}

type DataSourceUpdate struct {
	State            *string              `json:"state,omitempty" bson:"state,omitempty"`
	Error            *errors.Serializable `json:"error,omitempty" bson:"error,omitempty"`
	DataSetIDs       *[]string            `json:"dataSetIds,omitempty" bson:"dataSetIds,omitempty"`
	EarliestDataTime *time.Time           `json:"earliestDataTime,omitempty" bson:"earliestDataTime,omitempty"`
	LatestDataTime   *time.Time           `json:"latestDataTime,omitempty" bson:"latestDataTime,omitempty"`
	LastImportTime   *time.Time           `json:"lastImportTime,omitempty" bson:"lastImportTime,omitempty"`
}

func NewDataSourceUpdate() *DataSourceUpdate {
	return &DataSourceUpdate{}
}

func (d *DataSourceUpdate) HasUpdates() bool {
	return d.State != nil || d.Error != nil || d.DataSetIDs != nil || d.EarliestDataTime != nil || d.LatestDataTime != nil || d.LastImportTime != nil
}

func (d *DataSourceUpdate) Parse(parser structure.ObjectParser) {
	d.State = parser.String("state")
	if parser.ReferenceExists("error") {
		d.Error = &errors.Serializable{}
		d.Error.Parse("error", parser)
	}
	d.DataSetIDs = parser.StringArray("dataSetIds")
	d.EarliestDataTime = parser.Time("earliestDataTime", time.RFC3339)
	d.LatestDataTime = parser.Time("latestDataTime", time.RFC3339)
	d.LastImportTime = parser.Time("lastImportTime", time.RFC3339)
}

func (d *DataSourceUpdate) Validate(validator structure.Validator) {
	validator.String("state", d.State).OneOf(DataSourceStates()...)
	if d.Error != nil {
		d.Error.Validate(validator.WithReference("error"))
	}
	validator.StringArray("dataSetIds", d.DataSetIDs).NotEmpty()
	validator.Time("earliestDataTime", d.EarliestDataTime).NotZero().BeforeNow(time.Second)
	if d.EarliestDataTime != nil {
		validator.Time("latestDataTime", d.LatestDataTime).After(*d.EarliestDataTime).BeforeNow(time.Second)
	} else {
		validator.Time("latestDataTime", d.LatestDataTime).NotZero().BeforeNow(time.Second)
	}
	validator.Time("lastImportTime", d.LastImportTime).NotZero().BeforeNow(time.Second)
}

func (d *DataSourceUpdate) Normalize(normalizer structure.Normalizer) {
	if d.Error != nil {
		d.Error.Normalize(normalizer.WithReference("error"))
	}
	if d.EarliestDataTime != nil {
		d.EarliestDataTime = pointer.FromTime((*d.EarliestDataTime).Truncate(time.Second))
	}
	if d.LatestDataTime != nil {
		d.LatestDataTime = pointer.FromTime((*d.LatestDataTime).Truncate(time.Second))
	}
	if d.LastImportTime != nil {
		d.LastImportTime = pointer.FromTime((*d.LastImportTime).Truncate(time.Second))
	}
}

func NewSourceID() string {
	return id.Must(id.New(16))
}

func IsValidSourceID(value string) bool {
	return ValidateSourceID(value) == nil
}

func SourceIDValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateSourceID(value))
}

func ValidateSourceID(value string) error {
	if value == "" {
		return structureValidator.ErrorValueEmpty()
	} else if !setIDExpression.MatchString(value) {
		return ErrorValueStringAsSourceIDNotValid(value)
	}
	return nil
}

func ErrorValueStringAsSourceIDNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as data source id", value)
}

var sourceIDExpression = regexp.MustCompile("^[0-9a-z]{32}$")

type DataSource struct {
	ID                string               `json:"id" bson:"id"`
	UserID            string               `json:"userId" bson:"userId"`
	ProviderType      string               `json:"providerType" bson:"providerType"`
	ProviderName      string               `json:"providerName" bson:"providerName"`
	ProviderSessionID *string              `json:"providerSessionId,omitempty" bson:"providerSessionId,omitempty"`
	State             string               `json:"state,omitempty" bson:"state,omitempty"`
	Error             *errors.Serializable `json:"error,omitempty" bson:"error,omitempty"`
	DataSetIDs        []string             `json:"dataSetIds,omitempty" bson:"dataSetIds,omitempty"`
	EarliestDataTime  *time.Time           `json:"earliestDataTime,omitempty" bson:"earliestDataTime,omitempty"`
	LatestDataTime    *time.Time           `json:"latestDataTime,omitempty" bson:"latestDataTime,omitempty"`
	LastImportTime    *time.Time           `json:"lastImportTime,omitempty" bson:"lastImportTime,omitempty"`
	CreatedTime       time.Time            `json:"createdTime" bson:"createdTime"`
	ModifiedTime      *time.Time           `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
}

func NewDataSource(userID string, create *DataSourceCreate) (*DataSource, error) {
	if userID == "" {
		return nil, errors.New("user id is missing")
	}
	if create == nil {
		return nil, errors.New("create is missing")
	} else if err := structureValidator.New().Validate(create); err != nil {
		return nil, errors.Wrap(err, "create is invalid")
	}

	return &DataSource{
		ID:                NewSourceID(),
		UserID:            userID,
		ProviderType:      create.ProviderType,
		ProviderName:      create.ProviderName,
		ProviderSessionID: pointer.FromString(create.ProviderSessionID),
		State:             create.State,
		CreatedTime:       time.Now().Truncate(time.Second),
	}, nil
}

func (d *DataSource) Parse(parser structure.ObjectParser) {
	if ptr := parser.String("id"); ptr != nil {
		d.ID = *ptr
	}
	if ptr := parser.String("userId"); ptr != nil {
		d.UserID = *ptr
	}
	if ptr := parser.String("providerType"); ptr != nil {
		d.ProviderType = *ptr
	}
	if ptr := parser.String("providerName"); ptr != nil {
		d.ProviderName = *ptr
	}
	d.ProviderSessionID = parser.String("providerSessionId")
	if ptr := parser.String("state"); ptr != nil {
		d.State = *ptr
	}
	if parser.ReferenceExists("error") {
		d.Error = &errors.Serializable{}
		d.Error.Parse("error", parser)
	}
	if ptr := parser.StringArray("dataSetIds"); ptr != nil {
		d.DataSetIDs = *ptr
	}
	d.EarliestDataTime = parser.Time("earliestDataTime", time.RFC3339)
	d.LatestDataTime = parser.Time("latestDataTime", time.RFC3339)
	d.LastImportTime = parser.Time("lastImportTime", time.RFC3339)
	if ptr := parser.Time("createdTime", time.RFC3339); ptr != nil {
		d.CreatedTime = *ptr
	}
	d.ModifiedTime = parser.Time("modifiedTime", time.RFC3339)
}

func (d *DataSource) Validate(validator structure.Validator) {
	validator.String("id", &d.ID).Using(SourceIDValidator)
	validator.String("userId", &d.UserID).Using(user.IDValidator)
	validator.String("providerType", &d.ProviderType).OneOf(auth.ProviderTypes()...)
	validator.String("providerName", &d.ProviderName).NotEmpty()
	validator.String("providerSessionId", d.ProviderSessionID).Using(auth.ProviderSessionIDValidator)
	validator.String("state", &d.State).OneOf(DataSourceStates()...)
	if d.Error != nil {
		d.Error.Validate(validator.WithReference("error"))
	}
	validator.StringArray("dataSetIds", &d.DataSetIDs)
	validator.Time("earliestDataTime", d.EarliestDataTime).NotZero().BeforeNow(time.Second)
	if d.EarliestDataTime != nil {
		validator.Time("latestDataTime", d.LatestDataTime).After(*d.EarliestDataTime).BeforeNow(time.Second)
	} else {
		validator.Time("latestDataTime", d.LatestDataTime).NotZero().BeforeNow(time.Second)
	}
	validator.Time("lastImportTime", d.LastImportTime).NotZero().BeforeNow(time.Second)
	validator.Time("createdTime", &d.CreatedTime).NotZero().BeforeNow(time.Second)
	validator.Time("modifiedTime", d.ModifiedTime).After(d.CreatedTime).BeforeNow(time.Second)
}

func (d *DataSource) Normalize(normalizer structure.Normalizer) {
	if d.Error != nil {
		d.Error.Normalize(normalizer.WithReference("error"))
	}
}

func (d *DataSource) Sanitize(details request.Details) error {
	if details == nil {
		return errors.New("unable to sanitize")
	}

	if details.IsUser() {
		d.ProviderSessionID = nil
		if d.Error != nil && d.Error.Error != nil {
			if cause := errors.Cause(d.Error.Error); errors.Code(cause) == request.ErrorCodeUnauthenticated {
				d.Error.Error = cause
			}
			d.Error.Error = errors.Sanitize(d.Error.Error)
		}
	}

	return nil
}

type DataSources []*DataSource

func (d DataSources) Sanitize(details request.Details) error {
	for _, dataSource := range d {
		if err := dataSource.Sanitize(details); err != nil {
			return err
		}
	}
	return nil
}
