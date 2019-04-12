package image

import (
	"context"
	"fmt"
	"image/color"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tidepool-org/platform/association"
	"github.com/tidepool-org/platform/crypto"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/location"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/net"
	"github.com/tidepool-org/platform/origin"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"
)

const (
	ContentIntentAlternate      = "alternate"
	ContentIntentOriginal       = "original"
	HeightMaximum               = 10000
	HeightMinimum               = 1
	MediaTypeImageJPEG          = "image/jpeg"
	MediaTypeImagePNG           = "image/png"
	ModeDefault                 = ModeFit
	ModeFill                    = "fill"
	ModeFillDown                = "fillDown"
	ModeFit                     = "fit"
	ModeFitDown                 = "fitDown"
	ModePad                     = "pad"
	ModePadDown                 = "padDown"
	ModeScale                   = "scale"
	ModeScaleDown               = "scaleDown"
	NameLengthMaximum           = 100
	QualityDefault              = 95
	QualityMaximum              = 100
	QualityMinimum              = 1
	RenditionExtensionSeparator = "."
	RenditionFieldSeparator     = "_"
	RenditionKeyValueSeparator  = "="
	RenditionsLengthMaximum     = 10
	SizeMaximum                 = 100 * 1024 * 1024
	StatusAvailable             = "available"
	StatusCreated               = "created"
	WidthMaximum                = 10000
	WidthMinimum                = 1
)

func BackgroundDefault() *Color {
	return NewColor(0xFF, 0xFF, 0xFF, 0xFF)
}

func ContentIntents() []string {
	return []string{
		ContentIntentAlternate,
		ContentIntentOriginal,
	}
}

func MediaTypes() []string {
	return []string{
		MediaTypeImageJPEG,
		MediaTypeImagePNG,
	}
}

func Modes() []string {
	return []string{
		ModeFill,
		ModeFillDown,
		ModeFit,
		ModeFitDown,
		ModePad,
		ModePadDown,
		ModeScale,
		ModeScaleDown,
	}
}

func Statuses() []string {
	return []string{
		StatusAvailable,
		StatusCreated,
	}
}

type Client interface {
	List(ctx context.Context, userID string, filter *Filter, pagination *page.Pagination) (ImageArray, error)
	Create(ctx context.Context, userID string, metadata *Metadata, contentIntent string, content *Content) (*Image, error)
	CreateWithMetadata(ctx context.Context, userID string, metadata *Metadata) (*Image, error)
	CreateWithContent(ctx context.Context, userID string, contentIntent string, content *Content) (*Image, error)
	DeleteAll(ctx context.Context, userID string) error

	Get(ctx context.Context, id string) (*Image, error)
	GetMetadata(ctx context.Context, id string) (*Metadata, error)
	GetContent(ctx context.Context, id string, mediaType *string) (*Content, error)
	GetRenditionContent(ctx context.Context, id string, rendition *Rendition) (*Content, error)
	PutMetadata(ctx context.Context, id string, condition *request.Condition, metadata *Metadata) (*Image, error)
	PutContent(ctx context.Context, id string, condition *request.Condition, contentIntent string, content *Content) (*Image, error)
	Delete(ctx context.Context, id string, condition *request.Condition) (bool, error)
}

type Filter struct {
	Status        *[]string `json:"status,omitempty"`
	ContentIntent *[]string `json:"contentIntent,omitempty"`
}

func NewFilter() *Filter {
	return &Filter{}
}

func (f *Filter) Parse(parser structure.ObjectParser) {
	f.Status = parser.StringArray("status")
	f.ContentIntent = parser.StringArray("contentIntent")
}

func (f *Filter) Validate(validator structure.Validator) {
	validator.StringArray("status", f.Status).NotEmpty().EachOneOf(Statuses()...).EachUnique()
	validator.StringArray("contentIntent", f.ContentIntent).NotEmpty().EachOneOf(ContentIntents()...).EachUnique()
}

func (f *Filter) MutateRequest(req *http.Request) error {
	parameters := map[string][]string{}
	if f.Status != nil {
		parameters["status"] = *f.Status
	}
	if f.ContentIntent != nil {
		parameters["contentIntent"] = *f.ContentIntent
	}
	return request.NewArrayParametersMutator(parameters).MutateRequest(req)
}

type Metadata struct {
	Associations *association.AssociationArray `json:"associations,omitempty" bson:"associations,omitempty"`
	Location     *location.Location            `json:"location,omitempty" bson:"location,omitempty"`
	Metadata     *metadata.Metadata            `json:"metadata,omitempty" bson:"metadata,omitempty"`
	Name         *string                       `json:"name,omitempty" bson:"name,omitempty"`
	Origin       *origin.Origin                `json:"origin,omitempty" bson:"origin,omitempty"`
}

func ParseMetadata(parser structure.ObjectParser) *Metadata {
	if !parser.Exists() {
		return nil
	}
	datum := NewMetadata()
	parser.Parse(datum)
	return datum
}

func NewMetadata() *Metadata {
	return &Metadata{}
}

func (m *Metadata) Parse(parser structure.ObjectParser) {
	m.Associations = association.ParseAssociationArray(parser.WithReferenceArrayParser("associations"))
	m.Location = location.ParseLocation(parser.WithReferenceObjectParser("location"))
	m.Metadata = metadata.ParseMetadata(parser.WithReferenceObjectParser("metadata"))
	m.Name = parser.String("name")
	m.Origin = origin.ParseOrigin(parser.WithReferenceObjectParser("origin"))
}

func (m *Metadata) Validate(validator structure.Validator) {
	if m.Associations != nil {
		m.Associations.Validate(validator.WithReference("associations"))
	}
	if m.Location != nil {
		m.Location.Validate(validator.WithReference("location"))
	}
	if m.Metadata != nil {
		m.Metadata.Validate(validator.WithReference("metadata"))
	}
	validator.String("name", m.Name).NotEmpty().LengthLessThanOrEqualTo(NameLengthMaximum)
	if m.Origin != nil {
		m.Origin.Validate(validator.WithReference("origin"))
	}
}

func (m *Metadata) IsEmpty() bool {
	return m.Associations == nil && m.Location == nil && m.Metadata == nil && m.Name == nil && m.Origin == nil
}

type Content struct {
	Body      io.ReadCloser `json:"-"`
	DigestMD5 *string       `json:"digestMD5,omitempty"`
	MediaType *string       `json:"mediaType,omitempty"`
}

func NewContent() *Content {
	return &Content{}
}

func (c *Content) Validate(validator structure.Validator) {
	if c.Body == nil {
		validator.WithReference("body").ReportError(structureValidator.ErrorValueNotExists())
	}
	validator.String("digestMD5", c.DigestMD5).Using(crypto.Base64EncodedMD5HashValidator)
	validator.String("mediaType", c.MediaType).Exists().Using(MediaTypeValidator)
}

type Rendition struct {
	MediaType  *string `json:"mediaType,omitempty"`
	Width      *int    `json:"width,omitempty"`
	Height     *int    `json:"height,omitempty"`
	Mode       *string `json:"mode,omitempty"`
	Background *Color  `json:"background,omitempty"`
	Quality    *int    `json:"quality,omitempty"`
}

func ParseRenditionFromString(valueString string) (*Rendition, error) {
	values := url.Values{}

	switch extensionParts := strings.Split(valueString, RenditionExtensionSeparator); len(extensionParts) {
	case 2:
		mediaType, valid := MediaTypeFromExtension(extensionParts[1])
		if !valid {
			return nil, ErrorValueRenditionNotParsable(valueString)
		}
		values.Add("mediaType", mediaType)
		fallthrough
	case 1:
		for _, fieldParts := range strings.Split(extensionParts[0], RenditionFieldSeparator) {
			keyValueParts := strings.Split(fieldParts, RenditionKeyValueSeparator)
			if len(keyValueParts) != 2 {
				return nil, ErrorValueRenditionNotParsable(valueString)
			}

			switch keyValueParts[0] {
			case "width", "w":
				values.Add("width", keyValueParts[1])
			case "height", "h":
				values.Add("height", keyValueParts[1])
			case "mode", "m":
				values.Add("mode", keyValueParts[1])
			case "background", "b":
				values.Add("background", keyValueParts[1])
			case "quality", "q":
				values.Add("quality", keyValueParts[1])
			default:
				return nil, ErrorValueRenditionNotParsable(valueString)
			}
		}
	default:
		return nil, ErrorValueRenditionNotParsable(valueString)
	}

	rendition := NewRendition()
	if err := request.DecodeValues(values, rendition); err != nil {
		return nil, err
	}

	return rendition, nil
}

func NewRendition() *Rendition {
	return &Rendition{}
}

func (r *Rendition) Parse(parser structure.ObjectParser) {
	r.MediaType = parser.String("mediaType")
	r.Width = parser.Int("width")
	r.Height = parser.Int("height")
	r.Mode = parser.String("mode")
	if value := parser.String("background"); value != nil {
		if color, err := ParseColor(*value); err == nil {
			r.Background = color
		} else {
			parser.WithReferenceErrorReporter("background").ReportError(err)
		}
	}
	r.Quality = parser.Int("quality")
}

func (r *Rendition) Validate(validator structure.Validator) {
	validator.String("mediaType", r.MediaType).Exists().Using(MediaTypeValidator)
	validator.Int("width", r.Width).InRange(WidthMinimum, WidthMaximum)
	validator.Int("height", r.Height).InRange(HeightMinimum, HeightMaximum)
	if r.Width == nil && r.Height == nil {
		validator.ReportError(structureValidator.ErrorValuesNotExistForAny("width", "height"))
	}
	validator.String("mode", r.Mode).OneOf(Modes()...)
	if qualityValidator := validator.Int("quality", r.Quality); r.SupportsQuality() {
		qualityValidator.InRange(QualityMinimum, QualityMaximum)
	} else {
		qualityValidator.NotExists()
	}
}

func (r *Rendition) SupportsQuality() bool {
	return r.MediaType != nil && MediaTypeSupportsQuality(*r.MediaType)
}

func (r *Rendition) SupportsTransparency() bool {
	return r.MediaType != nil && MediaTypeSupportsTransparency(*r.MediaType)
}

func (r *Rendition) ConstrainWidth(aspectRatio float64) {
	if r.Height != nil {
		r.Width = pointer.FromInt(int(math.Round(float64(*r.Height) * aspectRatio)))
	}
}

func (r *Rendition) ConstrainHeight(aspectRatio float64) {
	if r.Width != nil {
		r.Height = pointer.FromInt(int(math.Round(float64(*r.Width) / aspectRatio)))
	}
}

func (r *Rendition) WithDefaults(aspectRatio float64) *Rendition {
	rendition := *r
	if rendition.Width == nil {
		rendition.ConstrainWidth(aspectRatio)
	}
	if rendition.Height == nil {
		rendition.ConstrainHeight(aspectRatio)
	}
	if rendition.Mode == nil {
		rendition.Mode = pointer.FromString(ModeDefault)
	}
	if rendition.Background == nil {
		rendition.Background = BackgroundDefault()
	}
	if rendition.Quality == nil && r.SupportsQuality() {
		rendition.Quality = pointer.FromInt(QualityDefault)
	}
	return &rendition
}

func (r *Rendition) String() string {
	var parts []string

	if r.Width != nil {
		parts = append(parts, fmt.Sprintf("%s%s%d", "w", RenditionKeyValueSeparator, *r.Width))
	}
	if r.Height != nil {
		parts = append(parts, fmt.Sprintf("%s%s%d", "h", RenditionKeyValueSeparator, *r.Height))
	}
	if r.Mode != nil {
		parts = append(parts, fmt.Sprintf("%s%s%s", "m", RenditionKeyValueSeparator, *r.Mode))
	}
	if r.Background != nil {
		parts = append(parts, fmt.Sprintf("%s%s%s", "b", RenditionKeyValueSeparator, r.Background.String()))
	}
	if r.Quality != nil {
		parts = append(parts, fmt.Sprintf("%s%s%d", "q", RenditionKeyValueSeparator, *r.Quality))
	}

	valueString := strings.Join(parts, RenditionFieldSeparator)

	if r.MediaType != nil {
		if extension, valid := ExtensionFromMediaType(*r.MediaType); valid {
			valueString = fmt.Sprintf("%s%s%s", valueString, RenditionExtensionSeparator, extension)
		}
	}

	return valueString
}

type ContentAttributes struct {
	DigestMD5    *string    `json:"digestMD5,omitempty" bson:"digestMD5,omitempty"`
	MediaType    *string    `json:"mediaType,omitempty" bson:"mediaType,omitempty"`
	Width        *int       `json:"width,omitempty" bson:"width,omitempty"`
	Height       *int       `json:"height,omitempty" bson:"height,omitempty"`
	Size         *int       `json:"size,omitempty" bson:"size,omitempty"`
	CreatedTime  *time.Time `json:"createdTime,omitempty" bson:"createdTime,omitempty"`
	ModifiedTime *time.Time `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
}

func ParseContentAttributes(parser structure.ObjectParser) *ContentAttributes {
	if !parser.Exists() {
		return nil
	}
	datum := NewContentAttributes()
	parser.Parse(datum)
	return datum
}

func NewContentAttributes() *ContentAttributes {
	return &ContentAttributes{}
}

func (c *ContentAttributes) Parse(parser structure.ObjectParser) {
	c.DigestMD5 = parser.String("digestMD5")
	c.MediaType = parser.String("mediaType")
	c.Width = parser.Int("width")
	c.Height = parser.Int("height")
	c.Size = parser.Int("size")
	c.CreatedTime = parser.Time("createdTime", time.RFC3339Nano)
	c.ModifiedTime = parser.Time("modifiedTime", time.RFC3339Nano)
}

func (c *ContentAttributes) Validate(validator structure.Validator) {
	validator.String("digestMD5", c.DigestMD5).Exists().Using(crypto.Base64EncodedMD5HashValidator)
	validator.String("mediaType", c.MediaType).Exists().Using(MediaTypeValidator)
	validator.Int("width", c.Width).Exists().GreaterThan(0)
	validator.Int("height", c.Height).Exists().GreaterThan(0)
	validator.Int("size", c.Size).Exists().GreaterThan(0)
	validator.Time("createdTime", c.CreatedTime).Exists().NotZero().BeforeNow(time.Second)
	validator.Time("modifiedTime", c.ModifiedTime).NotZero().After(pointer.ToTime(c.CreatedTime)).BeforeNow(time.Second)
}

func (c *ContentAttributes) SupportsQuality() bool {
	return c.MediaType != nil && MediaTypeSupportsQuality(*c.MediaType)
}

func (c *ContentAttributes) SupportsTransparency() bool {
	return c.MediaType != nil && MediaTypeSupportsTransparency(*c.MediaType)
}

type Image struct {
	ID                *string            `json:"id,omitempty" bson:"id,omitempty"`
	UserID            *string            `json:"userId,omitempty" bson:"userId,omitempty"`
	Status            *string            `json:"status,omitempty" bson:"status,omitempty"`
	Metadata          *Metadata          `json:"metadata,omitempty" bson:"metadata,omitempty"`
	ContentID         *string            `json:"contentId,omitempty" bson:"contentId,omitempty"`
	ContentIntent     *string            `json:"contentIntent,omitempty" bson:"contentIntent,omitempty"`
	ContentAttributes *ContentAttributes `json:"contentAttributes,omitempty" bson:"contentAttributes,omitempty"`
	RenditionsID      *string            `json:"renditionsId,omitempty" bson:"renditionsId,omitempty"`
	Renditions        *[]string          `json:"renditions,omitempty" bson:"renditions,omitempty"`
	CreatedTime       *time.Time         `json:"createdTime,omitempty" bson:"createdTime,omitempty"`
	ModifiedTime      *time.Time         `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
	DeletedTime       *time.Time         `json:"deletedTime,omitempty" bson:"deletedTime,omitempty"`
	Revision          *int               `json:"revision,omitempty" bson:"revision,omitempty"`
}

func (i *Image) Parse(parser structure.ObjectParser) {
	i.ID = parser.String("id")
	i.UserID = parser.String("userId")
	i.Status = parser.String("status")
	i.Metadata = ParseMetadata(parser.WithReferenceObjectParser("metadata"))
	i.ContentID = parser.String("contentId")
	i.ContentIntent = parser.String("contentIntent")
	i.ContentAttributes = ParseContentAttributes(parser.WithReferenceObjectParser("contentAttributes"))
	i.RenditionsID = parser.String("renditionsId")
	i.Renditions = parser.StringArray("renditions")
	i.CreatedTime = parser.Time("createdTime", time.RFC3339Nano)
	i.ModifiedTime = parser.Time("modifiedTime", time.RFC3339Nano)
	i.DeletedTime = parser.Time("deletedTime", time.RFC3339Nano)
	i.Revision = parser.Int("revision")
}

func (i *Image) Validate(validator structure.Validator) {
	contentIDValidator := validator.String("contentId", i.ContentID)
	contentIntentValidator := validator.String("contentIntent", i.ContentIntent)
	contentAttributesValidator := validator.WithReference("contentAttributes")
	renditionsIDValidator := validator.String("renditionsId", i.RenditionsID)
	renditionsValidator := validator.StringArray("renditions", i.Renditions)

	validator.String("id", i.ID).Exists().Using(IDValidator)
	validator.String("userId", i.UserID).Exists().Using(user.IDValidator)
	validator.String("status", i.Status).Exists().OneOf(Statuses()...)
	if i.Metadata != nil {
		i.Metadata.Validate(validator.WithReference("metadata"))
	}
	if i.Status != nil && *i.Status == StatusCreated {
		contentIDValidator.NotExists()
		contentIntentValidator.NotExists()
		if i.ContentAttributes != nil {
			contentAttributesValidator.ReportError(structureValidator.ErrorValueExists())
		}
		renditionsIDValidator.NotExists()
		renditionsValidator.NotExists()
	} else {
		if i.Status != nil && *i.Status == StatusAvailable {
			contentIntentValidator.Exists()
			if i.ContentAttributes == nil {
				contentAttributesValidator.ReportError(structureValidator.ErrorValueNotExists())
			}
			if i.RenditionsID != nil {
				renditionsValidator.Exists()
			} else {
				renditionsValidator.NotExists()
			}
		}
		contentIDValidator.Using(ContentIDValidator)
		contentIntentValidator.OneOf(ContentIntents()...)
		if i.ContentAttributes != nil {
			i.ContentAttributes.Validate(contentAttributesValidator)
		}
		renditionsIDValidator.Using(RenditionsIDValidator)
		renditionsValidator.EachNotEmpty().EachUnique()
	}
	validator.Time("createdTime", i.CreatedTime).Exists().NotZero().BeforeNow(time.Second)
	validator.Time("modifiedTime", i.ModifiedTime).NotZero().After(pointer.ToTime(i.CreatedTime)).BeforeNow(time.Second)
	validator.Time("deletedTime", i.DeletedTime).NotZero().After(pointer.ToTime(i.CreatedTime)).BeforeNow(time.Second)
	validator.Int("revision", i.Revision).Exists().GreaterThanOrEqualTo(0)
}

func (i *Image) HasContent() bool {
	return i.Status != nil && *i.Status == StatusAvailable && i.ContentID != nil && i.ContentIntent != nil && i.ContentAttributes != nil
}

func (i *Image) HasRendition(rendition Rendition) bool {
	if i.Renditions != nil {
		renditionString := rendition.String()
		for _, r := range *i.Renditions {
			if r == renditionString {
				return true
			}
		}
	}
	return false
}

func (i *Image) Sanitize(details request.Details) error {
	if details == nil || !details.IsService() {
		i.ContentID = nil
		i.RenditionsID = nil
		i.Renditions = nil
	}
	return nil
}

type ImageArray []*Image

func (i ImageArray) Sanitize(details request.Details) error {
	for _, datum := range i {
		if err := datum.Sanitize(details); err != nil {
			return err
		}
	}
	return nil
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

var idExpression = regexp.MustCompile("^[0-9a-z]{32}$")

func NewContentID() string {
	return id.Must(id.New(8))
}

func IsValidContentID(value string) bool {
	return ValidateContentID(value) == nil
}

func ContentIDValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateContentID(value))
}

func ValidateContentID(value string) error {
	if value == "" {
		return structureValidator.ErrorValueEmpty()
	} else if !contentIDExpression.MatchString(value) {
		return ErrorValueStringAsContentIDNotValid(value)
	}
	return nil
}

var contentIDExpression = regexp.MustCompile("^[0-9a-z]{16}$")

func NewRenditionsID() string {
	return id.Must(id.New(8))
}

func IsValidRenditionsID(value string) bool {
	return ValidateRenditionsID(value) == nil
}

func RenditionsIDValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateRenditionsID(value))
}

func ValidateRenditionsID(value string) error {
	if value == "" {
		return structureValidator.ErrorValueEmpty()
	} else if !renditionsIDExpression.MatchString(value) {
		return ErrorValueStringAsRenditionsIDNotValid(value)
	}
	return nil
}

var renditionsIDExpression = regexp.MustCompile("^[0-9a-z]{16}$")

func IsValidContentIntent(value string) bool {
	return ValidateContentIntent(value) == nil
}

func ContentIntentValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateContentIntent(value))
}

func ValidateContentIntent(value string) error {
	if value == "" {
		return structureValidator.ErrorValueEmpty()
	} else if value != ContentIntentAlternate && value != ContentIntentOriginal {
		return ErrorValueStringAsContentIntentNotValid(value)
	}
	return nil
}

func IsValidMediaType(value string) bool {
	return ValidateMediaType(value) == nil
}

func MediaTypeValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateMediaType(value))
}

func ValidateMediaType(value string) error {
	if err := net.ValidateMediaType(value); err != nil {
		return err
	} else if _, ok := ExtensionFromMediaType(value); !ok {
		return request.ErrorMediaTypeNotSupported(value)
	}
	return nil
}

func MediaTypeSupportsQuality(mediaType string) bool {
	switch mediaType {
	case MediaTypeImageJPEG:
		return true
	}
	return false
}

func MediaTypeSupportsTransparency(mediaType string) bool {
	switch mediaType {
	case MediaTypeImagePNG:
		return true
	}
	return false
}

func MediaTypeFromExtension(extension string) (string, bool) {
	switch strings.ToLower(extension) {
	case "jpeg", "jpg":
		return MediaTypeImageJPEG, true
	case "png":
		return MediaTypeImagePNG, true
	}
	return "", false
}

func IsValidExtension(value string) bool {
	return ValidateExtension(value) == nil
}

func ExtensionValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateExtension(value))
}

func ValidateExtension(value string) error {
	if value == "" {
		return structureValidator.ErrorValueEmpty()
	} else if _, ok := MediaTypeFromExtension(value); !ok {
		return request.ErrorExtensionNotSupported(value)
	}
	return nil
}

func ExtensionFromMediaType(mediaType string) (string, bool) {
	switch mediaType {
	case MediaTypeImageJPEG:
		return "jpeg", true
	case MediaTypeImagePNG:
		return "png", true
	}
	return "", false
}

func NormalizeMode(mode string) string {
	switch mode {
	case ModeFillDown:
		return ModeFill
	case ModeFitDown:
		return ModeFit
	case ModePadDown:
		return ModePad
	case ModeScaleDown:
		return ModeScale
	}
	return mode
}

type Color struct {
	color.NRGBA
}

func ParseColor(value string) (*Color, error) {
	normalized := strings.TrimPrefix(value, "0x")
	if parsed, err := strconv.ParseUint(normalized, 16, 32); err == nil {
		switch len(normalized) {
		case 6:
			return NewColor(uint8((parsed>>16)&0xFF), uint8((parsed>>8)&0xFF), uint8(parsed&0xFF), 0xFF), nil
		case 8:
			return NewColor(uint8((parsed>>24)&0xFF), uint8((parsed>>16)&0xFF), uint8((parsed>>8)&0xFF), uint8(parsed&0xFF)), nil
		}
	}
	return nil, ErrorValueStringAsColorNotValid(value)
}

func NewColor(r uint8, g uint8, b uint8, a uint8) *Color {
	return &Color{NRGBA: color.NRGBA{R: r, G: g, B: b, A: a}}
}

func (c *Color) String() string {
	return fmt.Sprintf("%02x%02x%02x%02x", c.R, c.G, c.B, c.A)
}
