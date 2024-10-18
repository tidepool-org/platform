package types

import (
	"sort"
	"time"

	"github.com/tidepool-org/platform/association"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/location"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/origin"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	timeZone "github.com/tidepool-org/platform/time/zone"
	"github.com/tidepool-org/platform/user"
)

const (
	ClockDriftOffsetMaximum = 24 * 60 * 60 * 1000  // TODO: Fix! Limit to reasonable values
	ClockDriftOffsetMinimum = -24 * 60 * 60 * 1000 // TODO: Fix! Limit to reasonable values
	DeviceTimeFormat        = "2006-01-02T15:04:05"
	NoteLengthMaximum       = 1000
	NotesLengthMaximum      = 100
	TagLengthMaximum        = 100
	TagsLengthMaximum       = 100
	TimeFormat              = time.RFC3339Nano
	TimeZoneOffsetMaximum   = 7 * 24 * 60  // TODO: Fix! Limit to reasonable values
	TimeZoneOffsetMinimum   = -7 * 24 * 60 // TODO: Fix! Limit to reasonable values
	VersionInternalMinimum  = 0
)

type Base struct {
	Active            bool                          `json:"-" bson:"_active"`
	Annotations       *metadata.MetadataArray       `json:"annotations,omitempty" bson:"annotations,omitempty"`
	ArchivedDataSetID *string                       `json:"archivedDatasetId,omitempty" bson:"archivedDatasetId,omitempty"`
	ArchivedTime      *time.Time                    `json:"archivedTime,omitempty" bson:"archivedTime,omitempty"`
	Associations      *association.AssociationArray `json:"associations,omitempty" bson:"associations,omitempty"`
	ClockDriftOffset  *int                          `json:"clockDriftOffset,omitempty" bson:"clockDriftOffset,omitempty"`
	ConversionOffset  *int                          `json:"conversionOffset,omitempty" bson:"conversionOffset,omitempty"`
	CreatedTime       *time.Time                    `json:"createdTime,omitempty" bson:"createdTime,omitempty"`
	CreatedUserID     *string                       `json:"createdUserId,omitempty" bson:"createdUserId,omitempty"`
	Deduplicator      *data.DeduplicatorDescriptor  `json:"deduplicator,omitempty" bson:"_deduplicator,omitempty"`
	DeletedTime       *time.Time                    `json:"deletedTime,omitempty" bson:"deletedTime,omitempty"`
	DeletedUserID     *string                       `json:"deletedUserId,omitempty" bson:"deletedUserId,omitempty"`
	DeviceID          *string                       `json:"deviceId,omitempty" bson:"deviceId,omitempty"`
	DeviceTime        *string                       `json:"deviceTime,omitempty" bson:"deviceTime,omitempty"`
	GUID              *string                       `json:"guid,omitempty" bson:"guid,omitempty"`
	ID                *string                       `json:"id,omitempty" bson:"id,omitempty"`
	Location          *location.Location            `json:"location,omitempty" bson:"location,omitempty"`
	ModifiedTime      *time.Time                    `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
	ModifiedUserID    *string                       `json:"modifiedUserId,omitempty" bson:"modifiedUserId,omitempty"`
	Notes             *[]string                     `json:"notes,omitempty" bson:"notes,omitempty"`
	Origin            *origin.Origin                `json:"origin,omitempty" bson:"origin,omitempty"`
	Payload           *metadata.Metadata            `json:"payload,omitempty" bson:"payload,omitempty"`
	Provenance        *data.Provenance              `json:"-" bson:"provenance,omitempty"`
	Source            *string                       `json:"source,omitempty" bson:"source,omitempty"`
	Tags              *[]string                     `json:"tags,omitempty" bson:"tags,omitempty"`
	Time              *time.Time                    `json:"time,omitempty" bson:"time,omitempty"`
	TimeZoneName      *string                       `json:"timezone,omitempty" bson:"timezone,omitempty"`             // TODO: Rename to timeZoneName
	TimeZoneOffset    *int                          `json:"timezoneOffset,omitempty" bson:"timezoneOffset,omitempty"` // TODO: Rename to timeZoneOffset
	Type              string                        `json:"type,omitempty" bson:"type,omitempty"`
	UploadID          *string                       `json:"uploadId,omitempty" bson:"uploadId,omitempty"`
	UserID            *string                       `json:"-" bson:"_userId,omitempty"`
	VersionInternal   int                           `json:"-" bson:"_version,omitempty"`
}

type Meta struct {
	Type string `json:"type,omitempty"`
}

func New(typ string) Base {
	return Base{
		Type: typ,
	}
}

func (b *Base) Meta() interface{} {
	return &Meta{
		Type: b.Type,
	}
}

func (b *Base) Parse(parser structure.ObjectParser) {
	b.Annotations = metadata.ParseMetadataArray(parser.WithReferenceArrayParser("annotations"))
	b.Associations = association.ParseAssociationArray(parser.WithReferenceArrayParser("associations"))
	b.ClockDriftOffset = parser.Int("clockDriftOffset")
	b.ConversionOffset = parser.Int("conversionOffset")
	b.DeviceID = parser.String("deviceId")
	b.DeviceTime = parser.String("deviceTime")
	b.ID = parser.String("id")
	b.Location = location.ParseLocation(parser.WithReferenceObjectParser("location"))
	b.Notes = parser.StringArray("notes")
	b.Origin = origin.ParseOrigin(parser.WithReferenceObjectParser("origin"))
	b.Payload = metadata.ParseMetadata(parser.WithReferenceObjectParser("payload"))
	b.Source = parser.String("source")
	b.Tags = parser.StringArray("tags")
	b.Time = parser.Time("time", TimeFormat)
	b.TimeZoneName = parser.String("timezone")
	b.TimeZoneOffset = parser.Int("timezoneOffset")
}

func (b *Base) Validate(validator structure.Validator) {
	// NOTE we copy these to default them if nil without writing to the originals
	// the logic below does not like null pointers
	var archivedTime time.Time
	var createdTime time.Time
	var modifiedTime time.Time

	if b.ArchivedTime != nil {
		archivedTime = *b.ArchivedTime
	}
	if b.CreatedTime != nil {
		createdTime = *b.CreatedTime
	}
	if b.ModifiedTime != nil {
		modifiedTime = *b.ModifiedTime
	}

	if b.Annotations != nil {
		b.Annotations.Validate(validator.WithReference("annotations"))
	}
	if b.Associations != nil {
		b.Associations.Validate(validator.WithReference("associations"))
	}

	if validator.Origin() <= structure.OriginInternal {
		if b.ArchivedTime != nil {
			validator.String("archivedDatasetId", b.ArchivedDataSetID).Exists().Using(data.SetIDValidator)
			validator.Time("archivedTime", b.ArchivedTime).After(createdTime).BeforeNow(time.Second)
		} else {
			validator.String("archivedDatasetId", b.ArchivedDataSetID).NotExists()
		}
	}

	validator.Int("clockDriftOffset", b.ClockDriftOffset).InRange(ClockDriftOffsetMinimum, ClockDriftOffsetMaximum)

	if validator.Origin() <= structure.OriginInternal {
		if b.CreatedTime != nil {
			validator.Time("createdTime", b.CreatedTime).BeforeNow(time.Second)
			validator.String("createdUserId", b.CreatedUserID).Using(user.IDValidator)
		} else {
			validator.Time("createdTime", b.CreatedTime).Exists()
			validator.String("createdUserId", b.CreatedUserID).NotExists()
		}

		if b.DeletedTime != nil {
			validator.Time("deletedTime", b.DeletedTime).After(latestTime(archivedTime, createdTime, modifiedTime)).BeforeNow(time.Second)
			validator.String("deletedUserId", b.DeletedUserID).Using(user.IDValidator)
		} else {
			validator.String("deletedUserId", b.DeletedUserID).NotExists()
		}
	}

	if b.Deduplicator != nil {
		b.Deduplicator.Validate(validator.WithReference("deduplicator"))
	}

	validator.String("deviceId", b.DeviceID).NotEmpty()
	validator.String("deviceTime", b.DeviceTime).AsTime(DeviceTimeFormat)

	validator.String("id", b.ID).Using(data.IDValidator)
	if validator.Origin() <= structure.OriginInternal {
		validator.String("id", b.ID).Exists()
	}

	if b.Location != nil {
		b.Location.Validate(validator.WithReference("location"))
	}

	if validator.Origin() <= structure.OriginInternal {
		if b.ModifiedTime != nil {
			validator.Time("modifiedTime", b.ModifiedTime).After(latestTime(archivedTime, createdTime)).BeforeNow(time.Second)
			validator.String("modifiedUserId", b.ModifiedUserID).Using(user.IDValidator)
		} else {
			if b.ArchivedTime != nil {
				validator.Time("modifiedTime", b.ModifiedTime).Exists()
			}
			validator.String("modifiedUserId", b.ModifiedUserID).NotExists()
		}
	}

	if b.Notes != nil {
		// notes from dexcom API fetch were set to be empty and would silently fail
		if len(*b.Notes) == 0 {
			b.Notes = nil
		}
	}

	validator.StringArray("notes", b.Notes).NotEmpty().LengthLessThanOrEqualTo(NotesLengthMaximum).Each(func(stringValidator structure.String) {
		stringValidator.Exists().NotEmpty().LengthLessThanOrEqualTo(NoteLengthMaximum)
	})

	if b.Origin != nil {
		b.Origin.Validate(validator.WithReference("origin"))
	}
	if b.Payload != nil {
		b.Payload.Validate(validator.WithReference("payload"))
	}

	validator.String("source", b.Source).EqualTo("carelink")
	validator.StringArray("tags", b.Tags).NotEmpty().LengthLessThanOrEqualTo(TagsLengthMaximum).Each(func(stringValidator structure.String) {
		stringValidator.Exists().NotEmpty().LengthLessThanOrEqualTo(TagLengthMaximum)
	}).EachUnique()

	timeValidator := validator.Time("time", b.Time)
	if b.Type != "upload" { // HACK: Need to replace upload.Upload with data.DataSet
		timeValidator.Exists().NotZero()
	}

	validator.String("timezone", b.TimeZoneName).Using(timeZone.NameValidator)
	validator.Int("timezoneOffset", b.TimeZoneOffset).InRange(TimeZoneOffsetMinimum, TimeZoneOffsetMaximum)
	validator.String("type", &b.Type).Exists().NotEmpty()

	if validator.Origin() <= structure.OriginInternal {
		validator.String("uploadId", b.UploadID).Exists().Using(data.SetIDValidator)
	}
	if validator.Origin() <= structure.OriginStore {
		validator.String("_userId", b.UserID).Exists().Using(user.IDValidator)
		validator.Int("_version", &b.VersionInternal).Exists().GreaterThanOrEqualTo(VersionInternalMinimum)
	}
}

func (b *Base) Normalize(normalizer data.Normalizer) {
	if b.Deduplicator != nil {
		b.Deduplicator.NormalizeDEPRECATED(normalizer.WithReference("deduplicator"))
	}

	if normalizer.Origin() == structure.OriginExternal {
		if b.ID == nil {
			b.ID = pointer.FromString(data.NewID())
		}
	}

	if b.Tags != nil {
		sort.Strings(*b.Tags)
	}
}

func (b *Base) IdentityFields() ([]string, error) {
	if b.UserID == nil {
		return nil, errors.New("user id is missing")
	}
	if *b.UserID == "" {
		return nil, errors.New("user id is empty")
	}
	if b.DeviceID == nil {
		return nil, errors.New("device id is missing")
	}
	if *b.DeviceID == "" {
		return nil, errors.New("device id is empty")
	}
	if b.Time == nil {
		return nil, errors.New("time is missing")
	}
	if (*b.Time).IsZero() {
		return nil, errors.New("time is empty")
	}
	if b.Type == "" {
		return nil, errors.New("type is empty")
	}

	return []string{*b.UserID, *b.DeviceID, (*b.Time).Format(TimeFormat), b.Type}, nil
}

func (b *Base) GetOrigin() *origin.Origin {
	return b.Origin
}

func (b *Base) GetPayload() *metadata.Metadata {
	return b.Payload
}

func (b *Base) GetTime() *time.Time {
	return b.Time
}

func (b *Base) GetTimeZoneOffset() *int {
	return b.TimeZoneOffset
}

func (b *Base) GetUploadID() *string {
	return b.UploadID
}

func (b *Base) GetType() string {
	return b.Type
}

func (b *Base) IsActive() bool {
	return b.Active
}

func (b *Base) SetType(typ string) {
	b.Type = typ
}

func (b *Base) SetUserID(userID *string) {
	b.UserID = userID
}

func (b *Base) SetDataSetID(dataSetID *string) {
	b.UploadID = dataSetID
}

func (b *Base) SetActive(active bool) {
	b.Active = active
}

func (b *Base) SetDeviceID(deviceID *string) {
	b.DeviceID = deviceID
}

func (b *Base) GetCreatedTime() *time.Time {
	return b.CreatedTime
}

func (b *Base) SetCreatedTime(createdTime *time.Time) {
	b.CreatedTime = createdTime
}

func (b *Base) SetCreatedUserID(createdUserID *string) {
	b.CreatedUserID = createdUserID
}

func (b *Base) SetModifiedTime(modifiedTime *time.Time) {
	b.ModifiedTime = modifiedTime
}

func (b *Base) SetModifiedUserID(modifiedUserID *string) {
	b.ModifiedUserID = modifiedUserID
}

func (b *Base) SetDeletedTime(deletedTime *time.Time) {
	b.DeletedTime = deletedTime
}

func (b *Base) SetDeletedUserID(deletedUserID *string) {
	b.DeletedUserID = deletedUserID
}

func (b *Base) DeduplicatorDescriptor() *data.DeduplicatorDescriptor {
	return b.Deduplicator
}

func (b *Base) SetDeduplicatorDescriptor(deduplicatorDescriptor *data.DeduplicatorDescriptor) {
	b.Deduplicator = deduplicatorDescriptor
}

func (b *Base) SetProvenance(provenance *data.Provenance) {
	b.Provenance = provenance
}

func latestTime(tms ...time.Time) time.Time {
	var latestTime time.Time
	for _, tm := range tms {
		if tm.After(latestTime) {
			latestTime = tm
		}
	}
	return latestTime
}
