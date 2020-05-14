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
	ArchivedTimeFormat      = time.RFC3339Nano
	ClockDriftOffsetMaximum = 24 * 60 * 60 * 1000  // TODO: Fix! Limit to reasonable values
	ClockDriftOffsetMinimum = -24 * 60 * 60 * 1000 // TODO: Fix! Limit to reasonable values
	CreatedTimeFormat       = time.RFC3339Nano
	DeletedTimeFormat       = time.RFC3339Nano
	DeviceTimeFormat        = "2006-01-02T15:04:05"
	ModifiedTimeFormat      = time.RFC3339Nano
	NoteLengthMaximum       = 1000
	NotesLengthMaximum      = 100
	SchemaVersionCurrent    = SchemaVersionMaximum
	SchemaVersionMaximum    = 3
	SchemaVersionMinimum    = 1
	TagLengthMaximum        = 100
	TagsLengthMaximum       = 100
	TimeFormat              = time.RFC3339Nano
	TimeZoneOffsetMaximum   = 7 * 24 * 60  // TODO: Fix! Limit to reasonable values
	TimeZoneOffsetMinimum   = -7 * 24 * 60 // TODO: Fix! Limit to reasonable values
	VersionMinimum          = 0
	parsingTimeFormat       = "2006-01-02T15:04:05.999-0700"
)

type Base struct {
	Active            bool                          `json:"-" bson:"_active"`
	Annotations       *metadata.MetadataArray       `json:"annotations,omitempty" bson:"annotations,omitempty"`
	ArchivedDataSetID *string                       `json:"archivedDatasetId,omitempty" bson:"archivedDatasetId,omitempty"`
	ArchivedTime      *string                       `json:"archivedTime,omitempty" bson:"archivedTime,omitempty"`
	Associations      *association.AssociationArray `json:"associations,omitempty" bson:"associations,omitempty"`
	ClockDriftOffset  *int                          `json:"clockDriftOffset,omitempty" bson:"clockDriftOffset,omitempty"`
	ConversionOffset  *int                          `json:"conversionOffset,omitempty" bson:"conversionOffset,omitempty"`
	CreatedTime       *string                       `json:"createdTime,omitempty" bson:"createdTime,omitempty"`
	CreatedUserID     *string                       `json:"createdUserId,omitempty" bson:"createdUserId,omitempty"`
	Deduplicator      *data.DeduplicatorDescriptor  `json:"deduplicator,omitempty" bson:"_deduplicator,omitempty"`
	DeletedTime       *string                       `json:"deletedTime,omitempty" bson:"deletedTime,omitempty"`
	DeletedUserID     *string                       `json:"deletedUserId,omitempty" bson:"deletedUserId,omitempty"`
	DeviceID          *string                       `json:"deviceId,omitempty" bson:"deviceId,omitempty"`
	DeviceTime        *string                       `json:"deviceTime,omitempty" bson:"deviceTime,omitempty"`
	GUID              *string                       `json:"guid,omitempty" bson:"guid,omitempty"`
	ID                *string                       `json:"id,omitempty" bson:"id,omitempty"`
	Location          *location.Location            `json:"location,omitempty" bson:"location,omitempty"`
	ModifiedTime      *string                       `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
	ModifiedUserID    *string                       `json:"modifiedUserId,omitempty" bson:"modifiedUserId,omitempty"`
	Notes             *[]string                     `json:"notes,omitempty" bson:"notes,omitempty"`
	Origin            *origin.Origin                `json:"origin,omitempty" bson:"origin,omitempty"`
	Payload           *metadata.Metadata            `json:"payload,omitempty" bson:"payload,omitempty"`
	SchemaVersion     int                           `json:"-" bson:"_schemaVersion,omitempty"`
	Source            *string                       `json:"source,omitempty" bson:"source,omitempty"`
	Tags              *[]string                     `json:"tags,omitempty" bson:"tags,omitempty"`
	Time              *string                       `json:"time,omitempty" bson:"time,omitempty"`
	TimeZoneName      *string                       `json:"timezone,omitempty" bson:"timezone,omitempty"`             // TODO: Rename to timeZoneName
	TimeZoneOffset    *int                          `json:"timezoneOffset,omitempty" bson:"timezoneOffset,omitempty"` // TODO: Rename to timeZoneOffset
	Type              string                        `json:"type,omitempty" bson:"type,omitempty"`
	UploadID          *string                       `json:"uploadId,omitempty" bson:"uploadId,omitempty"`
	UserID            *string                       `json:"-" bson:"_userId,omitempty"`
	Version           int                           `json:"-" bson:"_version,omitempty"`
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
	b.Time = parser.String("time")
	b.TimeZoneName = parser.String("timezone")
	b.TimeZoneOffset = parser.Int("timezoneOffset")
}

func (b *Base) Validate(validator structure.Validator) {
	var archivedTime time.Time
	var createdTime time.Time
	var modifiedTime time.Time

	if b.ArchivedTime != nil {
		archivedTime, _ = time.Parse(ArchivedTimeFormat, *b.ArchivedTime)
	}
	if b.CreatedTime != nil {
		createdTime, _ = time.Parse(CreatedTimeFormat, *b.CreatedTime)
	}
	if b.ModifiedTime != nil {
		modifiedTime, _ = time.Parse(ModifiedTimeFormat, *b.ModifiedTime)
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
			validator.String("archivedTime", b.ArchivedTime).AsTime(ArchivedTimeFormat).After(createdTime).BeforeNow(time.Second)
		} else {
			validator.String("archivedDatasetId", b.ArchivedDataSetID).NotExists()
		}
	}

	validator.Int("clockDriftOffset", b.ClockDriftOffset).InRange(ClockDriftOffsetMinimum, ClockDriftOffsetMaximum)

	if validator.Origin() <= structure.OriginInternal {
		if b.CreatedTime != nil {
			validator.String("createdTime", b.CreatedTime).AsTime(CreatedTimeFormat).BeforeNow(time.Second)
			validator.String("createdUserId", b.CreatedUserID).Using(user.IDValidator)
		} else {
			validator.String("createdTime", b.CreatedTime).Exists()
			validator.String("createdUserId", b.CreatedUserID).NotExists()
		}

		if b.DeletedTime != nil {
			validator.String("deletedTime", b.DeletedTime).AsTime(DeletedTimeFormat).After(latestTime(archivedTime, createdTime, modifiedTime)).BeforeNow(time.Second)
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
			validator.String("modifiedTime", b.ModifiedTime).AsTime(ModifiedTimeFormat).After(latestTime(archivedTime, createdTime)).BeforeNow(time.Second)
			validator.String("modifiedUserId", b.ModifiedUserID).Using(user.IDValidator)
		} else {
			if b.ArchivedTime != nil {
				validator.String("modifiedTime", b.ModifiedTime).Exists()
			}
			validator.String("modifiedUserId", b.ModifiedUserID).NotExists()
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

	if validator.Origin() <= structure.OriginStore {
		validator.Int("_schemaVersion", &b.SchemaVersion).Exists().InRange(SchemaVersionMinimum, SchemaVersionMaximum)
	}

	validator.String("source", b.Source).EqualTo("carelink")
	validator.StringArray("tags", b.Tags).NotEmpty().LengthLessThanOrEqualTo(TagsLengthMaximum).Each(func(stringValidator structure.String) {
		stringValidator.Exists().NotEmpty().LengthLessThanOrEqualTo(TagLengthMaximum)
	}).EachUnique()

	timeValidator := validator.String("time", b.Time)
	if b.Type != "upload" { // HACK: Need to replace upload.Upload with data.DataSet
		timeValidator.Exists()
	}
	timeValidator.AsTime(TimeFormat)

	validator.String("timezone", b.TimeZoneName).Using(timeZone.NameValidator)
	validator.Int("timezoneOffset", b.TimeZoneOffset).InRange(TimeZoneOffsetMinimum, TimeZoneOffsetMaximum)
	validator.String("type", &b.Type).Exists().NotEmpty()

	if validator.Origin() <= structure.OriginInternal {
		validator.String("uploadId", b.UploadID).Exists().Using(data.SetIDValidator)
	}
	if validator.Origin() <= structure.OriginStore {
		validator.String("_userId", b.UserID).Exists().Using(user.IDValidator)
		validator.Int("_version", &b.Version).Exists().GreaterThanOrEqualTo(VersionMinimum)
	}
}

// IsValid returns true if there is no error and no warning in the validator
func (b *Base) IsValid(validator structure.Validator) bool {
	return !(validator.HasError())
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

	if normalizer.Origin() == structure.OriginExternal {
		if b.SchemaVersion == 0 {
			b.SchemaVersion = SchemaVersionCurrent
		}
	}

	if b.Tags != nil {
		sort.Strings(*b.Tags)
	}
	if b.Time != nil && *b.Time != "" {
		parsedTime, err := time.Parse(TimeFormat, *b.Time)
		if err != nil {
			parsedTime, err = time.Parse(parsingTimeFormat, *b.Time)
		}
		if err == nil {
			utcTimeString := parsedTime.UTC().Format(TimeFormat)
			_, offset := parsedTime.Zone()
			// Time field is not well formatted in UTC
			if utcTimeString != *b.Time {
				b.Time = pointer.FromString(utcTimeString)
				// Time field was not set to UTC timezone
				// we update zone name / zone offset
				if utcTimeString != parsedTime.Format(TimeFormat) {
					b.TimeZoneOffset = pointer.FromInt(offset / 60)
				}
			}
			if b.TimeZoneOffset == nil {
				if b.TimeZoneName != nil && *b.TimeZoneName != "" {
					zoneLoc, err := time.LoadLocation(*b.TimeZoneName)
					if err == nil {
						_, offset := parsedTime.UTC().In(zoneLoc).Zone()
						b.TimeZoneOffset = pointer.FromInt(offset / 60)
					}
				}
				if b.TimeZoneOffset == nil {
					b.TimeZoneOffset = pointer.FromInt(0)
				}
			}
			offsetZone := time.FixedZone("offsetZone", *b.TimeZoneOffset*60)
			// For TimeZoneName we only check that :
			// the current timezone offset is valid for the TimeZoneName passed in
			// if not TimeZoneName is reset to nil
			if b.TimeZoneName != nil {
				tzCompareFormat := "15:04:05 -0700"
				currentZoneName := *b.TimeZoneName
				zoneLoc, err := time.LoadLocation(currentZoneName)
				if err == nil {
					localZoneTime := parsedTime.In(zoneLoc)
					localOffsetTime := parsedTime.In(offsetZone)
					if localZoneTime.Format(tzCompareFormat) != localOffsetTime.Format(tzCompareFormat) {
						b.TimeZoneName = nil
					}
				} else {
					b.TimeZoneName = nil
				}
			}
			// Setting DeviceTime to Time in offset zone (with the correct format)
			b.DeviceTime = pointer.FromString(parsedTime.UTC().In(offsetZone).Format(DeviceTimeFormat))
		}

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
	if *b.Time == "" {
		return nil, errors.New("time is empty")
	}
	if b.Type == "" {
		return nil, errors.New("type is empty")
	}

	return []string{*b.UserID, *b.DeviceID, *b.Time, b.Type}, nil
}

func (b *Base) GetOrigin() *origin.Origin {
	return b.Origin
}

func (b *Base) GetPayload() *metadata.Metadata {
	return b.Payload
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

func (b *Base) SetCreatedTime(createdTime *string) {
	b.CreatedTime = createdTime
}

func (b *Base) SetCreatedUserID(createdUserID *string) {
	b.CreatedUserID = createdUserID
}

func (b *Base) SetModifiedTime(modifiedTime *string) {
	b.ModifiedTime = modifiedTime
}

func (b *Base) SetModifiedUserID(modifiedUserID *string) {
	b.ModifiedUserID = modifiedUserID
}

func (b *Base) SetDeletedTime(deletedTime *string) {
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

func latestTime(tms ...time.Time) time.Time {
	var latestTime time.Time
	for _, tm := range tms {
		if tm.After(latestTime) {
			latestTime = tm
		}
	}
	return latestTime
}
