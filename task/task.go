package task

import (
	"context"
	"net/http"
	"regexp"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

//go:generate mockgen -source=task.go -destination=test/task_mocks.go -package=test -typed

type Client interface {
	ListTasks(ctx context.Context, filter *TaskFilter, pagination *page.Pagination) (Tasks, error)
	CreateTask(ctx context.Context, create *TaskCreate) (*Task, error)
	GetTask(ctx context.Context, id string, condition *request.Condition) (*Task, error)
	// UpdateTask applies the update and returns the resulting task. A runner that updates its
	// own running task (for example, to persist progress in Data) must replace its in-memory
	// task with the returned task: the update changes fields such as the revision, and the
	// queue writes the in-memory task's fields back when it completes the task, so a stale
	// in-memory copy would overwrite the update.
	UpdateTask(ctx context.Context, id string, condition *request.Condition, update *TaskUpdate) (*Task, error)
	DeleteTask(ctx context.Context, id string, condition *request.Condition) error
}

const (
	TaskStatePending   = "pending"
	TaskStateRunning   = "running"
	TaskStateFailed    = "failed"
	TaskStateCompleted = "completed"
)

func TaskStates() []string {
	return []string{
		TaskStatePending,
		TaskStateRunning,
		TaskStateFailed,
		TaskStateCompleted,
	}
}

type TaskFilter struct {
	Name  *string `json:"name,omitempty"`
	Type  *string `json:"type,omitempty"`
	State *string `json:"state,omitempty"`
}

func NewTaskFilter() *TaskFilter {
	return &TaskFilter{}
}

func (t *TaskFilter) Parse(parser structure.ObjectParser) {
	t.Name = parser.String("name")
	t.Type = parser.String("type")
	t.State = parser.String("state")
}

func (t *TaskFilter) Validate(validator structure.Validator) {
	validator.String("name", t.Name).NotEmpty()
	validator.String("type", t.Type).NotEmpty()
	validator.String("state", t.State).OneOf(TaskStates()...)
}

func (t *TaskFilter) MutateRequest(req *http.Request) error {
	parameters := map[string]string{}
	if t.Name != nil {
		parameters["name"] = *t.Name
	}
	if t.Type != nil {
		parameters["type"] = *t.Type
	}
	if t.State != nil {
		parameters["state"] = *t.State
	}
	return request.NewParametersMutator(parameters).MutateRequest(req)
}

type TaskCreate struct {
	Name          *string        `json:"name,omitempty"`
	Type          string         `json:"type,omitempty"`
	Data          map[string]any `json:"data,omitempty"`
	AvailableTime *time.Time     `json:"availableTime,omitempty"`
}

func NewTaskCreate() *TaskCreate {
	return &TaskCreate{}
}

func (t *TaskCreate) Parse(parser structure.ObjectParser) {
	t.Name = parser.String("name")
	if ptr := parser.String("type"); ptr != nil {
		t.Type = *ptr
	}
	if ptr := parser.Object("data"); ptr != nil {
		t.Data = *ptr
	}
	t.AvailableTime = parser.Time("availableTime", time.RFC3339Nano)
}

func (t *TaskCreate) Validate(validator structure.Validator) {
	validator.String("name", t.Name).NotEmpty()
	validator.String("type", &t.Type).NotEmpty()
}

type TaskUpdate struct {
	Data          *map[string]any      `json:"data,omitempty" bson:"data,omitempty"`
	AvailableTime *time.Time           `json:"availableTime,omitempty" bson:"availableTime,omitempty"`
	Error         *errors.Serializable `json:"error,omitempty" bson:"error,omitempty"`
}

func NewTaskUpdate() *TaskUpdate {
	return &TaskUpdate{}
}

func (t *TaskUpdate) Parse(parser structure.ObjectParser) {
	t.Data = parser.Object("data")
	t.AvailableTime = parser.Time("availableTime", time.RFC3339Nano)
	if parser.ReferenceExists("error") {
		t.Error = &errors.Serializable{}
		t.Error.Parse("error", parser)
	}
}

func (t *TaskUpdate) Validate(validator structure.Validator) {
	if t.Error != nil {
		t.Error.Validate(validator.WithReference("error"))
	}
}

func (t *TaskUpdate) Normalize(normalizer structure.Normalizer) {
	if t.Error != nil {
		t.Error.Normalize(normalizer.WithReference("error"))
	}
}

func (t *TaskUpdate) IsEmpty() bool {
	return t.Data == nil && t.AvailableTime == nil && t.Error == nil
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
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as task id", value)
}

var idExpression = regexp.MustCompile("^[0-9a-f]{32}$")

type Task struct {
	ID            string               `json:"id" bson:"id"`
	Name          *string              `json:"name,omitempty" bson:"name,omitempty"`
	Type          string               `json:"type" bson:"type"`
	Data          map[string]any       `json:"data,omitempty" bson:"data,omitempty"`
	AvailableTime *time.Time           `json:"availableTime,omitempty" bson:"availableTime,omitempty"`
	State         string               `json:"state" bson:"state"`
	Error         *errors.Serializable `json:"error,omitempty" bson:"error,omitempty"`
	RunTime       *time.Time           `json:"runTime,omitempty" bson:"runTime,omitempty"`
	Duration      *float64             `json:"duration,omitempty" bson:"duration,omitempty"`
	CreatedTime   time.Time            `json:"createdTime" bson:"createdTime"`
	ModifiedTime  *time.Time           `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
	Revision      int                  `json:"revision" bson:"revision"`

	// Database only

	// Use to enforce only one state transition at a time. This is a unique value that changes on every update.
	StateLock    *string    `json:"-" bson:"stateLock,omitempty"`
	DeadlineTime *time.Time `json:"-" bson:"deadlineTime,omitempty"`
}

func NewTask(ctx context.Context, create *TaskCreate) (*Task, error) {
	if create == nil {
		return nil, errors.New("create is missing")
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(create); err != nil {
		return nil, errors.Wrap(err, "create is invalid")
	}

	now := time.Now().UTC()

	availableTime := create.AvailableTime
	if availableTime == nil || availableTime.Before(now) {
		availableTime = pointer.From(now)
	}

	return &Task{
		ID:            NewID(),
		Name:          create.Name,
		Type:          create.Type,
		Data:          create.Data,
		AvailableTime: availableTime,
		State:         TaskStatePending,
		CreatedTime:   now,
		Revision:      1,
	}, nil
}

func (t *Task) Parse(parser structure.ObjectParser) {
	if ptr := parser.String("id"); ptr != nil {
		t.ID = *ptr
	}
	t.Name = parser.String("name")
	if ptr := parser.String("type"); ptr != nil {
		t.Type = *ptr
	}
	if ptr := parser.Object("data"); ptr != nil {
		t.Data = *ptr
	}
	t.AvailableTime = parser.Time("availableTime", time.RFC3339Nano)
	if ptr := parser.String("state"); ptr != nil {
		t.State = *ptr
	}
	if parser.ReferenceExists("error") {
		t.Error = &errors.Serializable{}
		t.Error.Parse("error", parser)
	}
	t.RunTime = parser.Time("runTime", time.RFC3339Nano)
	t.Duration = parser.Float64("duration")
	if ptr := parser.Time("createdTime", time.RFC3339Nano); ptr != nil {
		t.CreatedTime = *ptr
	}
	t.ModifiedTime = parser.Time("modifiedTime", time.RFC3339Nano)
	if ptr := parser.Int("revision"); ptr != nil {
		t.Revision = *ptr
	}
}

func (t *Task) Validate(validator structure.Validator) {
	validator.String("id", &t.ID).Using(IDValidator)
	validator.String("name", t.Name).NotEmpty()
	validator.String("type", &t.Type).NotEmpty()
	validator.String("state", &t.State).OneOf(TaskStates()...)
	if t.Error != nil {
		t.Error.Validate(validator.WithReference("error"))
	}
	validator.Time("runTime", t.RunTime).After(t.CreatedTime).BeforeNow(time.Second)
	validator.Float64("duration", t.Duration).GreaterThanOrEqualTo(0)
	validator.Time("createdTime", &t.CreatedTime).NotZero().BeforeNow(time.Second)
	validator.Time("modifiedTime", t.ModifiedTime).After(t.CreatedTime).BeforeNow(time.Second)
	validator.Int("revision", &t.Revision).GreaterThanOrEqualTo(0)
}

func (t *Task) Normalize(normalizer structure.Normalizer) {
	if t.Error != nil {
		t.Error.Normalize(normalizer.WithReference("error"))
	}
}

func (t *Task) Sanitize(details request.AuthDetails) error {
	if details != nil && details.IsService() {
		return nil
	}
	return errors.New("unable to sanitize")
}

func (t *Task) RepeatAvailableAt(availableTime time.Time) {
	t.State = TaskStatePending
	t.AvailableTime = pointer.FromTime(availableTime)
}

func (t *Task) RepeatAvailableAfter(availableDuration time.Duration) {
	t.RepeatAvailableAt(time.Now().Add(availableDuration))
}

func (t *Task) IsFailed() bool {
	return t.State == TaskStateFailed
}

func (t *Task) SetFailed() {
	t.State = TaskStateFailed
}

func (t *Task) SetFailedWithError(err error) {
	t.State = TaskStateFailed
	t.AppendError(err)
}

func (t *Task) IsCompleted() bool {
	return t.State == TaskStateCompleted
}

func (t *Task) SetCompleted() {
	t.State = TaskStateCompleted
}

func (t *Task) HasError() bool {
	return t.Error != nil && t.Error.Error != nil
}

func (t *Task) GetError() error {
	if t.Error != nil {
		return t.Error.Error
	}
	return nil
}

func (t *Task) AppendError(err error) {
	if err != nil {
		if t.Error == nil {
			t.Error = &errors.Serializable{}
		}
		t.Error.Error = errors.Append(t.Error.Error, err)
	}
}

func (t *Task) ClearError() {
	t.Error = nil
}

func (t *Task) LogFields() log.Fields {
	return log.Fields{
		"id":        t.ID,
		"type":      t.Type,
		"state":     t.State,
		"revision":  t.Revision,
		"stateLock": t.StateLock,
	}
}

type Tasks []*Task

func (t Tasks) Sanitize(details request.AuthDetails) error {
	for _, tsk := range t {
		if err := tsk.Sanitize(details); err != nil {
			return err
		}
	}
	return nil
}
