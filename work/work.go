package work

import (
	"regexp"
	"sort"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/net"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

// TODO: BACK-3612 - Add unit tests for work package for reasonable coverage

const (
	GroupIDLengthMaximum         = 1000
	DeduplicationIDLengthMaximum = 1000
	SerialIDLengthMaximum        = 1000
	ProcessingTimeoutMaximum     = 24 * 60 * 60 // seconds
	MetadataLengthMaximum        = 4 * 1024

	TypeQuantitiesLengthMaximum = 100

	StatePending    = "pending"
	StateProcessing = "processing"
	StateFailing    = "failing"
	StateFailed     = "failed"
	StateSuccess    = "success"
)

func States() []string {
	return []string{
		StatePending,
		StateProcessing,
		StateFailing,
		StateFailed,
		StateSuccess,
	}
}

type TypeQuantities map[string]int

func ParseTypeQuantities(parser structure.ObjectParser) TypeQuantities {
	datum := TypeQuantities{}
	if parser.Exists() {
		parser.Parse(&datum)
	}
	return datum
}

func (t TypeQuantities) Parse(parser structure.ObjectParser) {
	for _, reference := range parser.References() {
		if quantity := parser.Int(reference); quantity != nil {
			t[reference] = *quantity
		}
	}
}

func (t TypeQuantities) Validate(validator structure.Validator) {
	if length := len(t); length == 0 {
		validator.ReportError(structureValidator.ErrorValueEmpty())
	} else if length > TypeQuantitiesLengthMaximum {
		validator.ReportError(structureValidator.ErrorLengthNotLessThanOrEqualTo(length, TypeQuantitiesLengthMaximum))
	}
	for _, typ := range t.sortedTypes() {
		quantity := t[typ]
		datumValidator := validator.WithReference(typ)
		datumValidator.String(typ, &typ).Using(net.ReverseDomainValidator)
		datumValidator.Int(typ, &quantity).GreaterThan(0)
	}
}

func (t TypeQuantities) IsEmpty() bool {
	return len(t) == 0
}

func (t TypeQuantities) Get(typ string) int {
	quantity, ok := t[typ]
	if !ok {
		return 0
	}
	return quantity
}

func (t TypeQuantities) Set(typ string, quantity int) {
	if quantity < 0 {
		quantity = 0
	}
	t[typ] = quantity
}

func (t TypeQuantities) Increment(typ string) {
	if quantity, ok := t[typ]; ok {
		t[typ] = quantity + 1
	}
}

func (t TypeQuantities) Decrement(typ string) {
	if quantity, ok := t[typ]; ok && quantity > 0 {
		t[typ] = quantity - 1
	}
}

func (t TypeQuantities) Total() int {
	var total int
	for _, quantity := range t {
		total += quantity
	}
	return total
}

func (t TypeQuantities) NonZero() TypeQuantities {
	nonZero := TypeQuantities{}
	for typ, quantity := range t {
		if quantity > 0 {
			nonZero[typ] = quantity
		}
	}
	return nonZero
}

func (t TypeQuantities) sortedTypes() []string {
	var typs []string
	for typ := range t {
		typs = append(typs, typ)
	}
	sort.Strings(typs)
	return typs
}

type Poll struct {
	TypeQuantities TypeQuantities `json:"typeQuantities,omitempty"`
}

func (p *Poll) Parse(parser structure.ObjectParser) {
	p.TypeQuantities = ParseTypeQuantities(parser.WithReferenceObjectParser("typeQuantities"))
}

func (p *Poll) Validate(validator structure.Validator) {
	p.TypeQuantities.Validate(validator.WithReference("typeQuantities"))
}

type Filter struct {
	Types   *[]string `json:"types,omitempty"`
	GroupID *string   `json:"groupId,omitempty"`
}

func (f *Filter) Parse(parser structure.ObjectParser) {
	f.Types = parser.StringArray("types")
	f.GroupID = parser.String("groupId")
}

func (f *Filter) Validate(validator structure.Validator) {
	validator.StringArray("type", f.Types).NotEmpty().EachUsing(net.ReverseDomainValidator).EachUnique()
	validator.String("groupId", f.GroupID).NotEmpty().LengthLessThanOrEqualTo(GroupIDLengthMaximum)
}

type Create struct {
	Type                    string         `json:"type,omitempty"`
	GroupID                 *string        `json:"groupId,omitempty"`
	DeduplicationID         *string        `json:"deduplicationId,omitempty"`
	SerialID                *string        `json:"serialId,omitempty"`
	ProcessingAvailableTime time.Time      `json:"processingAvailableTime,omitempty"`
	ProcessingPriority      int            `json:"processingPriority,omitempty"`
	ProcessingTimeout       int            `json:"processingTimeout,omitempty"` // seconds
	Metadata                map[string]any `json:"metadata,omitempty"`
}

func ParseCreate(parser structure.ObjectParser) *Create {
	if !parser.Exists() {
		return nil
	}
	datum := &Create{}
	parser.Parse(datum)
	return datum
}

func (c *Create) Parse(parser structure.ObjectParser) {
	if ptr := parser.String("type"); ptr != nil {
		c.Type = *ptr
	}
	c.GroupID = parser.String("groupId")
	c.DeduplicationID = parser.String("deduplicationId")
	c.SerialID = parser.String("serialId")
	if ptr := parser.Time("processingAvailableTime", time.RFC3339Nano); ptr != nil {
		c.ProcessingAvailableTime = *ptr
	}
	if ptr := parser.Int("processingPriority"); ptr != nil {
		c.ProcessingPriority = *ptr
	}
	if ptr := parser.Int("processingTimeout"); ptr != nil {
		c.ProcessingTimeout = *ptr
	}
	if ptr := parser.Object("metadata"); ptr != nil {
		c.Metadata = *ptr
	}
}

func (c *Create) Validate(validator structure.Validator) {
	validator.String("type", &c.Type).Using(net.ReverseDomainValidator)
	validator.String("groupId", c.GroupID).NotEmpty().LengthLessThanOrEqualTo(GroupIDLengthMaximum)
	validator.String("deduplicationId", c.DeduplicationID).NotEmpty().LengthLessThanOrEqualTo(DeduplicationIDLengthMaximum)
	validator.String("serialId", c.SerialID).NotEmpty().LengthLessThanOrEqualTo(SerialIDLengthMaximum)
	validator.Int("processingTimeout", &c.ProcessingTimeout).GreaterThan(0).LessThanOrEqualTo(ProcessingTimeoutMaximum)
	validator.Object("metadata", &c.Metadata).SizeLessThanOrEqualTo(MetadataLengthMaximum)
}

type PendingUpdate struct {
	ProcessingAvailableTime time.Time      `json:"processingAvailableTime,omitempty" bson:"processingAvailableTime,omitempty"`
	ProcessingPriority      int            `json:"processingPriority,omitempty" bson:"processingPriority,omitempty"`
	ProcessingTimeout       int            `json:"processingTimeout,omitempty" bson:"processingTimeout,omitempty"`
	Metadata                map[string]any `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

func ParsePendingUpdate(parser structure.ObjectParser) *PendingUpdate {
	if !parser.Exists() {
		return nil
	}
	datum := &PendingUpdate{}
	parser.Parse(datum)
	return datum
}

func (p *PendingUpdate) Parse(parser structure.ObjectParser) {
	if ptr := parser.Time("processingAvailableTime", time.RFC3339Nano); ptr != nil {
		p.ProcessingAvailableTime = *ptr
	}
	if ptr := parser.Int("processingPriority"); ptr != nil {
		p.ProcessingPriority = *ptr
	}
	if ptr := parser.Int("processingTimeout"); ptr != nil {
		p.ProcessingTimeout = *ptr
	}
	if ptr := parser.Object("metadata"); ptr != nil {
		p.Metadata = *ptr
	}
}

func (p *PendingUpdate) Validate(validator structure.Validator) {
	validator.Int("processingTimeout", &p.ProcessingTimeout).GreaterThan(0).LessThanOrEqualTo(ProcessingTimeoutMaximum)
	validator.Object("metadata", &p.Metadata).SizeLessThanOrEqualTo(MetadataLengthMaximum)
}

type ProcessingUpdate struct {
	Metadata map[string]any `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

func ParseProcessingUpdate(parser structure.ObjectParser) *ProcessingUpdate {
	if !parser.Exists() {
		return nil
	}
	datum := &ProcessingUpdate{}
	parser.Parse(datum)
	return datum
}

func (p *ProcessingUpdate) Parse(parser structure.ObjectParser) {
	if ptr := parser.Object("metadata"); ptr != nil {
		p.Metadata = *ptr
	}
}

func (p *ProcessingUpdate) Validate(validator structure.Validator) {
	validator.Object("metadata", &p.Metadata).SizeLessThanOrEqualTo(MetadataLengthMaximum)
}

type FailingUpdate struct {
	FailingError      errors.Serializable `json:"failingError,omitempty" bson:"failingError,omitempty"`
	FailingRetryCount int                 `json:"failingRetryCount,omitempty" bson:"failingRetryCount,omitempty"`
	FailingRetryTime  time.Time           `json:"failingRetryTime,omitempty" bson:"failingRetryTime,omitempty"`
	Metadata          map[string]any      `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

func ParseFailingUpdate(parser structure.ObjectParser) *FailingUpdate {
	if !parser.Exists() {
		return nil
	}
	datum := &FailingUpdate{}
	parser.Parse(datum)
	return datum
}

func (f *FailingUpdate) Parse(parser structure.ObjectParser) {
	if parser.ReferenceExists("failingError") {
		f.FailingError = errors.Serializable{}
		f.FailingError.Parse("failingError", parser)
	}
	if ptr := parser.Int("failingRetryCount"); ptr != nil {
		f.FailingRetryCount = *ptr
	}
	if ptr := parser.Time("failingRetryTime", time.RFC3339Nano); ptr != nil {
		f.FailingRetryTime = *ptr
	}
	if ptr := parser.Object("metadata"); ptr != nil {
		f.Metadata = *ptr
	}
}

func (f *FailingUpdate) Validate(validator structure.Validator) {
	f.FailingError.Validate(validator.WithReference("failingError"))
	validator.Int("failingRetryCount", &f.FailingRetryCount).GreaterThanOrEqualTo(0)
	validator.Object("metadata", &f.Metadata).SizeLessThanOrEqualTo(MetadataLengthMaximum)
}

type FailedUpdate struct {
	FailedError errors.Serializable `json:"failedError,omitempty" bson:"failedError,omitempty"`
	Metadata    map[string]any      `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

func ParseFailedUpdate(parser structure.ObjectParser) *FailedUpdate {
	if !parser.Exists() {
		return nil
	}
	datum := &FailedUpdate{}
	parser.Parse(datum)
	return datum
}

func (f *FailedUpdate) Parse(parser structure.ObjectParser) {
	if parser.ReferenceExists("failedError") {
		f.FailedError = errors.Serializable{}
		f.FailedError.Parse("failedError", parser)
	}
	if ptr := parser.Object("metadata"); ptr != nil {
		f.Metadata = *ptr
	}
}

func (f *FailedUpdate) Validate(validator structure.Validator) {
	f.FailedError.Validate(validator.WithReference("failedError"))
	validator.Object("metadata", &f.Metadata).SizeLessThanOrEqualTo(MetadataLengthMaximum)
}

type SuccessUpdate struct {
	Metadata map[string]any `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

func ParseSuccessUpdate(parser structure.ObjectParser) *SuccessUpdate {
	if !parser.Exists() {
		return nil
	}
	datum := &SuccessUpdate{}
	parser.Parse(datum)
	return datum
}

func (s *SuccessUpdate) Parse(parser structure.ObjectParser) {
	if ptr := parser.Object("metadata"); ptr != nil {
		s.Metadata = *ptr
	}
}

func (s *SuccessUpdate) Validate(validator structure.Validator) {
	validator.Object("metadata", &s.Metadata).SizeLessThanOrEqualTo(MetadataLengthMaximum)
}

type Update struct {
	State            string            `json:"state,omitempty"`
	PendingUpdate    *PendingUpdate    `json:"pendingUpdate,omitempty"`
	ProcessingUpdate *ProcessingUpdate `json:"processingUpdate,omitempty"`
	FailingUpdate    *FailingUpdate    `json:"failingUpdate,omitempty"`
	FailedUpdate     *FailedUpdate     `json:"failedUpdate,omitempty"`
	SuccessUpdate    *SuccessUpdate    `json:"successUpdate,omitempty"`
}

func (u *Update) Parse(parser structure.ObjectParser) {
	if ptr := parser.String("state"); ptr != nil {
		u.State = *ptr
	}
	u.PendingUpdate = ParsePendingUpdate(parser.WithReferenceObjectParser("pendingUpdate"))
	u.ProcessingUpdate = ParseProcessingUpdate(parser.WithReferenceObjectParser("processingUpdate"))
	u.FailingUpdate = ParseFailingUpdate(parser.WithReferenceObjectParser("failingUpdate"))
	u.FailedUpdate = ParseFailedUpdate(parser.WithReferenceObjectParser("failedUpdate"))
	u.SuccessUpdate = ParseSuccessUpdate(parser.WithReferenceObjectParser("successUpdate"))
}

func (u *Update) Validate(validator structure.Validator) {
	validator.String("state", &u.State).OneOf(States()...)
	if pendingUpdateValidator := validator.WithReference("pendingUpdate"); u.PendingUpdate != nil {
		if u.State == StatePending {
			u.PendingUpdate.Validate(pendingUpdateValidator)
		} else {
			pendingUpdateValidator.ReportError(structureValidator.ErrorValueExists())
		}
	} else if u.State == StatePending {
		pendingUpdateValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
	if processingUpdateValidator := validator.WithReference("processingUpdate"); u.ProcessingUpdate != nil {
		if u.State == StateProcessing {
			u.ProcessingUpdate.Validate(processingUpdateValidator)
		} else {
			processingUpdateValidator.ReportError(structureValidator.ErrorValueExists())
		}
	} else if u.State == StateProcessing {
		processingUpdateValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
	if failingUpdateValidator := validator.WithReference("failingUpdate"); u.FailingUpdate != nil {
		if u.State == StateFailing {
			u.FailingUpdate.Validate(failingUpdateValidator)
		} else {
			failingUpdateValidator.ReportError(structureValidator.ErrorValueExists())
		}
	} else if u.State == StateFailing {
		failingUpdateValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
	if failedUpdateValidator := validator.WithReference("failedUpdate"); u.FailedUpdate != nil {
		if u.State == StateFailed {
			u.FailedUpdate.Validate(failedUpdateValidator)
		} else {
			failedUpdateValidator.ReportError(structureValidator.ErrorValueExists())
		}
	} else if u.State == StateFailed {
		failedUpdateValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
	if successUpdateValidator := validator.WithReference("successUpdate"); u.SuccessUpdate != nil {
		if u.State == StateSuccess {
			u.SuccessUpdate.Validate(successUpdateValidator)
		} else {
			successUpdateValidator.ReportError(structureValidator.ErrorValueExists())
		}
	} else if u.State == StateSuccess {
		successUpdateValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
}

type Work struct {
	ID                      string               `json:"id,omitempty"`
	Type                    string               `json:"type,omitempty"`
	GroupID                 *string              `json:"groupId,omitempty"`
	DeduplicationID         *string              `json:"deduplicationId,omitempty"`
	SerialID                *string              `json:"serialId,omitempty"`
	ProcessingAvailableTime time.Time            `json:"processingAvailableTime,omitempty"`
	ProcessingPriority      int                  `json:"processingPriority,omitempty"`
	ProcessingTimeout       int                  `json:"processingTimeout,omitempty"`
	Metadata                map[string]any       `json:"metadata,omitempty"`
	PendingTime             time.Time            `json:"pendingTime,omitempty"`
	ProcessingTime          *time.Time           `json:"processingTime,omitempty"`
	ProcessingTimeoutTime   *time.Time           `json:"processingTimeoutTime,omitempty"`
	ProcessingDuration      *float64             `json:"processingDuration,omitempty"` // seconds
	FailingTime             *time.Time           `json:"failingTime,omitempty"`
	FailingError            *errors.Serializable `json:"failingError,omitempty"`
	FailingRetryCount       *int                 `json:"failingRetryCount,omitempty"`
	FailingRetryTime        *time.Time           `json:"failingRetryTime,omitempty"`
	FailedTime              *time.Time           `json:"failedTime,omitempty"`
	FailedError             *errors.Serializable `json:"failedError,omitempty"`
	SuccessTime             *time.Time           `json:"successTime,omitempty"`
	State                   string               `json:"state,omitempty"`
	CreatedTime             time.Time            `json:"createdTime,omitempty"`
	ModifiedTime            *time.Time           `json:"modifiedTime,omitempty"`
	Revision                int                  `json:"revision,omitempty"`
}

func (w *Work) Parse(parser structure.ObjectParser) {
	if ptr := parser.String("id"); ptr != nil {
		w.ID = *ptr
	}
	if ptr := parser.String("type"); ptr != nil {
		w.Type = *ptr
	}
	w.GroupID = parser.String("groupId")
	w.DeduplicationID = parser.String("deduplicationId")
	w.SerialID = parser.String("serialId")
	if ptr := parser.Time("processingAvailableTime", time.RFC3339Nano); ptr != nil {
		w.ProcessingAvailableTime = *ptr
	}
	if ptr := parser.Int("processingPriority"); ptr != nil {
		w.ProcessingPriority = *ptr
	}
	if ptr := parser.Int("processingTimeout"); ptr != nil {
		w.ProcessingTimeout = *ptr
	}
	if ptr := parser.Object("metadata"); ptr != nil {
		w.Metadata = *ptr
	}
	if ptr := parser.Time("pendingTime", time.RFC3339Nano); ptr != nil {
		w.PendingTime = *ptr
	}
	w.ProcessingTime = parser.Time("processingTime", time.RFC3339Nano)
	w.ProcessingTimeoutTime = parser.Time("processingTimeoutTime", time.RFC3339Nano)
	w.ProcessingDuration = parser.Float64("processingDuration")
	w.FailingTime = parser.Time("failingTime", time.RFC3339Nano)
	if parser.ReferenceExists("failingError") {
		w.FailingError = &errors.Serializable{}
		w.FailingError.Parse("failingError", parser)
	}
	w.FailingRetryCount = parser.Int("failingRetryCount")
	w.FailingRetryTime = parser.Time("failingRetryTime", time.RFC3339Nano)
	w.FailedTime = parser.Time("failedTime", time.RFC3339Nano)
	if parser.ReferenceExists("failedError") {
		w.FailedError = &errors.Serializable{}
		w.FailedError.Parse("failedError", parser)
	}
	w.SuccessTime = parser.Time("successTime", time.RFC3339Nano)
	if ptr := parser.String("state"); ptr != nil {
		w.State = *ptr
	}
	if ptr := parser.Time("createdTime", time.RFC3339Nano); ptr != nil {
		w.CreatedTime = *ptr
	}
	w.ModifiedTime = parser.Time("modifiedTime", time.RFC3339Nano)
	if ptr := parser.Int("revision"); ptr != nil {
		w.Revision = *ptr
	}
}

func (w *Work) Validate(validator structure.Validator) {
	validator.String("id", &w.ID).Using(IDValidator)
	validator.String("type", &w.Type).Using(net.ReverseDomainValidator)
	validator.String("groupId", w.GroupID).NotEmpty().LengthLessThanOrEqualTo(GroupIDLengthMaximum)
	validator.String("deduplicationId", w.DeduplicationID).NotEmpty().LengthLessThanOrEqualTo(DeduplicationIDLengthMaximum)
	validator.String("serialId", w.SerialID).NotEmpty().LengthLessThanOrEqualTo(SerialIDLengthMaximum)
	validator.Int("processingTimeout", &w.ProcessingTimeout).GreaterThan(0).LessThanOrEqualTo(ProcessingTimeoutMaximum)
	validator.Object("metadata", &w.Metadata).SizeLessThanOrEqualTo(MetadataLengthMaximum)
	validator.Time("pendingTime", &w.PendingTime).After(w.CreatedTime).BeforeNow(time.Second)

	processingTimeValidator := validator.Time("processingTime", w.ProcessingTime)
	processingTimeoutTimeValidator := validator.Time("processingTimeoutTime", w.ProcessingTimeoutTime)
	processingDurationValidator := validator.Float64("processingDuration", w.ProcessingDuration)
	failingTimeValidator := validator.Time("failingTime", w.FailingTime)
	failingErrorValidator := validator.WithReference("failingError")
	failingRetryCountValidator := validator.Int("failingRetryCount", w.FailingRetryCount)
	failingRetryTimeValidator := validator.Time("failingRetryTime", w.FailingRetryTime)
	failedTimeValidator := validator.Time("failedTime", w.FailedTime)
	failedErrorValidator := validator.WithReference("failedError")
	successTimeValidator := validator.Time("successTime", w.FailedTime)

	switch w.State {
	case StateProcessing:
		processingTimeValidator.Exists().After(w.CreatedTime).BeforeNow(time.Second)
		processingTimeoutTimeValidator.Exists().After(w.CreatedTime)
		if w.ProcessingTime != nil {
			processingTimeoutTimeValidator.After(*w.ProcessingTime)
		}
		processingDurationValidator.NotExists()
	case StatePending, StateFailing, StateFailed, StateSuccess:
		processingTimeValidator.After(w.CreatedTime).BeforeNow(time.Second)
		processingTimeoutTimeValidator.NotExists()
		if w.ProcessingTime != nil {
			processingDurationValidator.Exists().GreaterThanOrEqualTo(0)
		} else {
			processingDurationValidator.NotExists()
		}
	}

	switch w.State {
	case StatePending, StateSuccess:
		failingTimeValidator.NotExists()
		if w.FailingError != nil {
			failingErrorValidator.ReportError(structureValidator.ErrorValueExists())
		}
		failingRetryCountValidator.NotExists()
		failingRetryTimeValidator.NotExists()
	case StateFailing:
		failingTimeValidator.Exists().After(w.CreatedTime)
		if w.FailingError != nil {
			w.FailingError.Validate(failingErrorValidator)
		} else {
			failingErrorValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
		failingRetryCountValidator.Exists().GreaterThanOrEqualTo(0)
		failingRetryTimeValidator.Exists().After(w.CreatedTime)
	case StateProcessing, StateFailed:
		failingTimeValidator.After(w.CreatedTime)
		if w.FailingError != nil {
			w.FailingError.Validate(failingErrorValidator)
		}
		failingRetryCountValidator.GreaterThanOrEqualTo(0)
		failingRetryTimeValidator.After(w.CreatedTime)
	}

	switch w.State {
	case StateFailed:
		failedTimeValidator.Exists().After(w.CreatedTime)
		if w.FailedError != nil {
			w.FailedError.Validate(failedErrorValidator)
		} else {
			failedErrorValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	case StatePending, StateProcessing, StateFailing, StateSuccess:
		failedTimeValidator.NotExists()
		if w.FailedError != nil {
			failedErrorValidator.ReportError(structureValidator.ErrorValueExists())
		}
	}

	switch w.State {
	case StateSuccess:
		successTimeValidator.Exists().After(w.CreatedTime)
	case StatePending, StateProcessing, StateFailing, StateFailed:
		successTimeValidator.NotExists()
	}

	validator.String("state", &w.State).OneOf(States()...)
	validator.Time("createdTime", &w.CreatedTime).NotZero().BeforeNow(time.Second)
	validator.Time("modifiedTime", w.ModifiedTime).After(w.CreatedTime).BeforeNow(time.Second)
	validator.Int("revision", &w.Revision).GreaterThanOrEqualTo(0)
}

func (w *Work) EnsureMetadata() {
	if w.Metadata == nil {
		w.Metadata = map[string]any{}
	}
}

func (w *Work) ProcessingTimeoutDuration() time.Duration {
	return time.Duration(w.ProcessingTimeout) * time.Second
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

var idExpression = regexp.MustCompile("^[0-9a-fA-F]{24}$")
