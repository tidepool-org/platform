package work

import (
	"context"
	"regexp"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/net"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	StatePending    = "pending"
	StateProcessing = "processing"
)

func States() []string {
	return []string{
		StatePending,
		StateProcessing,
	}
}

type Create struct {
	Type                 string             `json:"type,omitempty"`
	Priority             int                `json:"priority,omitempty"`
	DeduplicationId      *string            `json:"deduplicationId,omitempty"`
	GroupId              *string            `json:"groupId,omitempty"`
	Metadata             *metadata.Metadata `json:"metadata,omitempty"`
	PendingAvailableTime *time.Time         `json:"pendingAvailableTime,omitempty"`
}

func NewCreate() *Create {
	return &Create{}
}

func (c *Create) Parse(parser structure.ObjectParser) {
	if ptr := parser.String("type"); ptr != nil {
		c.Type = *ptr
	}
	if ptr := parser.Int("priority"); ptr != nil {
		c.Priority = *ptr
	}
	c.DeduplicationId = parser.String("deduplicationId")
	c.GroupId = parser.String("groupId")
	c.Metadata = metadata.ParseMetadata(parser.WithReferenceObjectParser("metadata"))
	c.PendingAvailableTime = parser.Time("pendingAvailableTime", time.RFC3339Nano)
}

func (c *Create) Validate(validator structure.Validator) {
	validator.String("type", &c.Type).Using(net.ReverseDomainValidator)
	validator.String("deduplicationId", c.DeduplicationId).NotEmpty()
	validator.String("groupId", c.GroupId).NotEmpty()
	if c.Metadata != nil {
		c.Metadata.Validate(validator.WithReference("metadata"))
	}
}

type Process struct {
	Types []string `json:"types,omitempty"`
}

func NewProcess() *Process {
	return &Process{}
}

func (p *Process) Parse(parser structure.ObjectParser) {
	if ptr := parser.StringArray("types"); ptr != nil {
		p.Types = *ptr
	}
}

func (p *Process) Validate(validator structure.Validator) {
	validator.StringArray("types", &p.Types).NotEmpty().EachUsing(net.ReverseDomainValidator)
}

type Repeat struct {
	PendingAvailableTime *time.Time `json:"pendingAvailableTime,omitempty"`
}

func NewRepeat() *Repeat {
	return &Repeat{}
}

func (r *Repeat) Parse(parser structure.ObjectParser) {
	r.PendingAvailableTime = parser.Time("pendingAvailableTime", time.RFC3339Nano)
}

type Work struct {
	ID                     string             `json:"id,omitempty" bson:"id,omitempty"`
	Type                   string             `json:"type,omitempty" bson:"type,omitempty"`
	Priority               int                `json:"priority,omitempty" bson:"priority,omitempty"`
	DeduplicationId        *string            `json:"deduplicationId,omitempty" bson:"deduplicationId,omitempty"`
	GroupId                *string            `json:"groupId,omitempty" bson:"groupId,omitempty"`
	Metadata               *metadata.Metadata `json:"metadata,omitempty" bson:"metadata,omitempty"`
	State                  string             `json:"state,omitempty" bson:"state,omitempty"`
	PendingAvailableTime   *time.Time         `json:"pendingAvailableTime,omitempty" bson:"pendingAvailableTime,omitempty"`
	ProcessingTime         *time.Time         `json:"processingTime,omitempty" bson:"processingTime,omitempty"`
	ProcessingDeadlineTime *time.Time         `json:"processingDeadlineTime,omitempty" bson:"processingDeadlineTime,omitempty"`
	ProcessingDuration     *float64           `json:"processingDuration,omitempty" bson:"processingDuration,omitempty"`
	CreatedTime            time.Time          `json:"createdTime,omitempty" bson:"createdTime,omitempty"`
	ModifiedTime           *time.Time         `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
}

func NewWork(ctx context.Context, create *Create) (*Work, error) {
	if create == nil {
		return nil, errors.New("create is missing")
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(create); err != nil {
		return nil, errors.Wrap(err, "create is invalid")
	}

	return &Work{
		ID:                   NewID(),
		Type:                 create.Type,
		Priority:             create.Priority,
		DeduplicationId:      create.DeduplicationId,
		GroupId:              create.GroupId,
		Metadata:             create.Metadata,
		State:                StatePending,
		PendingAvailableTime: create.PendingAvailableTime,
		CreatedTime:          time.Now(),
	}, nil
}

func (w *Work) Parse(parser structure.ObjectParser) {
	if ptr := parser.String("id"); ptr != nil {
		w.ID = *ptr
	}
	if ptr := parser.String("type"); ptr != nil {
		w.Type = *ptr
	}
	if ptr := parser.Int("priority"); ptr != nil {
		w.Priority = *ptr
	}
	w.DeduplicationId = parser.String("deduplicationId")
	w.GroupId = parser.String("groupId")
	w.Metadata = metadata.ParseMetadata(parser.WithReferenceObjectParser("metadata"))
	if ptr := parser.String("state"); ptr != nil {
		w.State = *ptr
	}
	w.PendingAvailableTime = parser.Time("pendingAvailableTime", time.RFC3339Nano)
	w.ProcessingTime = parser.Time("processingTime", time.RFC3339Nano)
	w.ProcessingDeadlineTime = parser.Time("processingDeadlineTime", time.RFC3339Nano)
	w.ProcessingDuration = parser.Float64("processingDuration")
	if ptr := parser.Time("createdTime", time.RFC3339Nano); ptr != nil {
		w.CreatedTime = *ptr
	}
	w.ModifiedTime = parser.Time("modifiedTime", time.RFC3339Nano)
}

func (w *Work) Validate(validator structure.Validator) {
	validator.String("id", &w.ID).Using(IDValidator)
	validator.String("type", &w.Type).Using(net.ReverseDomainValidator)
	validator.String("deduplicationId", w.DeduplicationId).NotEmpty()
	validator.String("groupId", w.GroupId).NotEmpty()
	if w.Metadata != nil {
		w.Metadata.Validate(validator.WithReference("metadata"))
	}
	validator.String("state", &w.State).OneOf(States()...)
	switch w.State {
	case StatePending:
		validator.Time("processingDeadlingTime", w.ProcessingDeadlineTime).NotExists()
	case StateProcessing:
		validator.Time("processingTime", w.ProcessingTime).Exists().After(w.CreatedTime).BeforeNow(time.Second)
		processingDeadlingTimeValidator := validator.Time("processingDeadlingTime", w.ProcessingDeadlineTime)
		processingDeadlingTimeValidator.Exists()
		if w.ProcessingTime != nil {
			processingDeadlingTimeValidator.After(*w.ProcessingTime)
		}
	}
	validator.Float64("processingDuration", w.ProcessingDuration).GreaterThanOrEqualTo(0)
	validator.Time("createdTime", &w.CreatedTime).NotZero().BeforeNow(time.Second)
	validator.Time("modifiedTime", w.ModifiedTime).After(w.CreatedTime).BeforeNow(time.Second)
}

func NewID() string {
	return id.Must(id.New(16))
}

func IsValidID(value string) bool {
	return ValidateID(value) == nil
}

func IDValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateID(value))
}

func ValidateID(value string) error {
	if value == "" {
		return structureValidator.ErrorValueEmpty()
	} else if !idExpression.MatchString(value) {
		return ErrorValueStringAsIDNotValid(value)
	}
	return nil
}

func ErrorValueStringAsIDNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as work id", value)
}

var idExpression = regexp.MustCompile("^[0-9a-f]{32}$")
