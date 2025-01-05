package work

import (
	"regexp"
	"sort"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/net"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	GroupIDLengthMaximum         = 1000
	DeduplicationIDLengthMaximum = 1000
	SerialIDLengthMaximum        = 1000
	ProcessingTimeoutMaximum     = 24 * 60 * 60 // seconds

	TypeQuantitiesLengthMaximum = 100

	StatePending    = "pending"
	StateProcessing = "processing"
)

func States() []string {
	return []string{
		StatePending,
		StateProcessing,
	}
}

type Poll struct {
	TypeQuantities TypeQuantities `json:"typeQuantities,omitempty"`
}

func NewPoll() *Poll {
	return &Poll{}
}

func (p *Poll) Parse(parser structure.ObjectParser) {
	p.TypeQuantities = ParseTypeQuantities(parser.WithReferenceObjectParser("typeQuantities"))
}

func (p *Poll) Validate(validator structure.Validator) {
	p.TypeQuantities.Validate(validator.WithReference("typeQuantities"))
}

type TypeQuantities map[string]int

func ParseTypeQuantities(parser structure.ObjectParser) TypeQuantities {
	datum := NewTypeQuantities()
	if parser.Exists() {
		parser.Parse(&datum)
	}
	return datum
}

func NewTypeQuantities() TypeQuantities {
	return TypeQuantities{}
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
	return len(t) > 0
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
	nonZero := NewTypeQuantities()
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

type Filter struct {
	Types   *[]string `json:"types,omitempty"`
	GroupID *string   `json:"groupId,omitempty"`
}

func NewFilter() *Filter {
	return &Filter{}
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
	Type                    string             `json:"type,omitempty"`
	GroupID                 *string            `json:"groupId,omitempty"`
	DeduplicationID         *string            `json:"deduplicationId,omitempty"`
	SerialID                *string            `json:"serialId,omitempty"`
	ProcessingAvailableTime time.Time          `json:"processingAvailableTime,omitempty"`
	ProcessingPriority      int                `json:"processingPriority,omitempty"`
	ProcessingTimeout       int                `json:"processingTimeout,omitempty"` // seconds
	Metadata                *metadata.Metadata `json:"metadata,omitempty"`
}

func ParseCreate(parser structure.ObjectParser) *Create {
	if !parser.Exists() {
		return nil
	}
	datum := NewCreate()
	parser.Parse(datum)
	return datum
}

func NewCreate() *Create {
	return &Create{}
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
	c.Metadata = metadata.ParseMetadata(parser.WithReferenceObjectParser("metadata"))
}

func (c *Create) Validate(validator structure.Validator) {
	validator.String("type", &c.Type).Using(net.ReverseDomainValidator)
	validator.String("groupId", c.GroupID).NotEmpty().LengthLessThanOrEqualTo(GroupIDLengthMaximum)
	validator.String("deduplicationId", c.DeduplicationID).NotEmpty().LengthLessThanOrEqualTo(DeduplicationIDLengthMaximum)
	validator.String("serialId", c.SerialID).NotEmpty().LengthLessThanOrEqualTo(SerialIDLengthMaximum)
	validator.Int("processingTimeout", &c.ProcessingTimeout).GreaterThan(0).LessThanOrEqualTo(ProcessingTimeoutMaximum)
	if c.Metadata != nil {
		c.Metadata.Validate(validator.WithReference("metadata"))
	}
}

type PendingUpdate struct {
	ProcessingAvailableTime time.Time          `json:"processingAvailableTime,omitempty"`
	ProcessingPriority      int                `json:"processingPriority,omitempty"`
	ProcessingTimeout       int                `json:"processingTimeout,omitempty"`
	Metadata                *metadata.Metadata `json:"metadata,omitempty"`
}

func ParsePendingUpdate(parser structure.ObjectParser) *PendingUpdate {
	if !parser.Exists() {
		return nil
	}
	datum := NewPendingUpdate()
	parser.Parse(datum)
	return datum
}

func NewPendingUpdate() *PendingUpdate {
	return &PendingUpdate{}
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
	p.Metadata = metadata.ParseMetadata(parser.WithReferenceObjectParser("metadata"))
}

func (p *PendingUpdate) Validate(validator structure.Validator) {
	validator.Int("processingTimeout", &p.ProcessingTimeout).GreaterThan(0).LessThanOrEqualTo(ProcessingTimeoutMaximum)
	if p.Metadata != nil {
		p.Metadata.Validate(validator.WithReference("metadata"))
	}
}

type ProcessingUpdate struct {
	Metadata *metadata.Metadata `json:"metadata,omitempty"`
}

func ParseProcessingUpdate(parser structure.ObjectParser) *ProcessingUpdate {
	if !parser.Exists() {
		return nil
	}
	datum := NewProcessingUpdate()
	parser.Parse(datum)
	return datum
}

func NewProcessingUpdate() *ProcessingUpdate {
	return &ProcessingUpdate{}
}

func (p *ProcessingUpdate) Parse(parser structure.ObjectParser) {
	p.Metadata = metadata.ParseMetadata(parser.WithReferenceObjectParser("metadata"))
}

func (p *ProcessingUpdate) Validate(validator structure.Validator) {
	if p.Metadata != nil {
		p.Metadata.Validate(validator.WithReference("metadata"))
	}
}

type StateUpdate struct {
	State string `json:"state,omitempty"`
}

func ParseStateUpdate(parser structure.ObjectParser) *StateUpdate {
	if !parser.Exists() {
		return nil
	}
	datum := NewStateUpdate()
	parser.Parse(datum)
	return datum
}

func NewStateUpdate() *StateUpdate {
	return &StateUpdate{}
}

func (s *StateUpdate) Parse(parser structure.ObjectParser) {
	if ptr := parser.String("state"); ptr != nil {
		s.State = *ptr
	}
}

func (s *StateUpdate) Validate(validator structure.Validator) {
	validator.String("state", &s.State).OneOf(States()...)
}

type Update struct {
	*PendingUpdate    `json:"pendingUpdate,omitempty"`
	*ProcessingUpdate `json:"processingUpdate,omitempty"`
	*StateUpdate      `json:"stateUpdate,omitempty"`
}

func NewUpdate() *Update {
	return &Update{}
}

func (u *Update) Parse(parser structure.ObjectParser) {
	u.PendingUpdate = ParsePendingUpdate(parser.WithReferenceObjectParser("pendingUpdate"))
	u.ProcessingUpdate = ParseProcessingUpdate(parser.WithReferenceObjectParser("processingUpdate"))
	u.StateUpdate = ParseStateUpdate(parser.WithReferenceObjectParser("stateUpdate"))
}

func (u *Update) Validate(validator structure.Validator) {
	if u.PendingUpdate != nil {
		u.PendingUpdate.Validate(validator.WithReference("pendingUpdate"))
		u.ProcessingUpdate.Validate(validator.WithReference("processingUpdate"))
		u.StateUpdate.Validate(validator.WithReference("stateUpdate"))
	}
}

func (u *Update) IsEmpty() bool {
	return u.PendingUpdate == nil && u.ProcessingUpdate == nil && u.StateUpdate == nil
}

type Work struct {
	ID                      string             `json:"id,omitempty"`
	Type                    string             `json:"type,omitempty"`
	GroupID                 *string            `json:"groupId,omitempty"`
	DeduplicationID         *string            `json:"deduplicationId,omitempty"`
	SerialID                *string            `json:"serialId,omitempty"`
	ProcessingAvailableTime time.Time          `json:"processingAvailableTime,omitempty"`
	ProcessingPriority      int                `json:"processingPriority,omitempty"`
	ProcessingTimeout       int                `json:"processingTimeout,omitempty"`
	Metadata                *metadata.Metadata `json:"metadata,omitempty"`
	PendingTime             time.Time          `json:"pendingTime,omitempty"`
	ProcessingTime          *time.Time         `json:"processingTime,omitempty"`
	ProcessingTimeoutTime   *time.Time         `json:"processingTimeoutTime,omitempty"`
	ProcessingDuration      *float64           `json:"processingDuration,omitempty"` // seconds
	State                   string             `json:"state,omitempty"`
	CreatedTime             time.Time          `json:"createdTime,omitempty"`
	ModifiedTime            *time.Time         `json:"modifiedTime,omitempty"`
	Revision                int                `json:"revision,omitempty"`
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
	w.Metadata = metadata.ParseMetadata(parser.WithReferenceObjectParser("metadata"))
	if ptr := parser.Time("pendingTime", time.RFC3339Nano); ptr != nil {
		w.PendingTime = *ptr
	}
	w.ProcessingTime = parser.Time("processingTime", time.RFC3339Nano)
	w.ProcessingTimeoutTime = parser.Time("processingTimeoutTime", time.RFC3339Nano)
	w.ProcessingDuration = parser.Float64("processingDuration")
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
	if w.Metadata != nil {
		w.Metadata.Validate(validator.WithReference("metadata"))
	}
	validator.Time("pendingTime", &w.PendingTime).After(w.CreatedTime).BeforeNow(time.Second)
	processingTimeValidator := validator.Time("processingTime", w.ProcessingTime)
	processingTimeoutTimeValidator := validator.Time("processingTimeoutTime", w.ProcessingTimeoutTime)
	processingDurationValidator := validator.Float64("processingDuration", w.ProcessingDuration)
	switch w.State {
	case StatePending:
		processingTimeValidator.After(w.CreatedTime).BeforeNow(time.Second)
		processingTimeoutTimeValidator.NotExists()
		if w.ProcessingTime != nil {
			processingDurationValidator.Exists().GreaterThanOrEqualTo(0)
		} else {
			processingDurationValidator.NotExists()
		}
	case StateProcessing:
		processingTimeValidator.Exists().After(w.CreatedTime).BeforeNow(time.Second)
		processingTimeoutTimeValidator.Exists()
		if w.ProcessingTime != nil {
			processingTimeoutTimeValidator.After(*w.ProcessingTime)
		}
		processingDurationValidator.NotExists()
	}
	validator.String("state", &w.State).OneOf(States()...)
	validator.Time("createdTime", &w.CreatedTime).NotZero().BeforeNow(time.Second)
	validator.Time("modifiedTime", w.ModifiedTime).After(w.CreatedTime).BeforeNow(time.Second)
	validator.Int("revision", &w.Revision).GreaterThanOrEqualTo(0)
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
