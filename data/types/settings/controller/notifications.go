package controller

import (
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	AlertStyleAlert            = "alert"
	AlertStyleBanner           = "banner"
	AlertStyleNone             = "none"
	AuthorizationAuthorized    = "authorized"
	AuthorizationDenied        = "denied"
	AuthorizationEphemeral     = "ephemeral"
	AuthorizationNotDetermined = "notDetermined"
	AuthorizationProvisional   = "provisional"
)

func AlertStyles() []string {
	return []string{
		AlertStyleAlert,
		AlertStyleBanner,
		AlertStyleNone,
	}
}

func Authorizations() []string {
	return []string{
		AuthorizationAuthorized,
		AuthorizationDenied,
		AuthorizationEphemeral,
		AuthorizationNotDetermined,
		AuthorizationProvisional,
	}
}

type Notifications struct {
	Authorization      *string `json:"authorization,omitempty" bson:"authorization,omitempty"`
	Alert              *bool   `json:"alert,omitempty" bson:"alert,omitempty"`
	CriticalAlert      *bool   `json:"criticalAlert,omitempty" bson:"criticalAlert,omitempty"`
	Badge              *bool   `json:"badge,omitempty" bson:"badge,omitempty"`
	Sound              *bool   `json:"sound,omitempty" bson:"sound,omitempty"`
	Announcement       *bool   `json:"announcement,omitempty" bson:"announcement,omitempty"`
	NotificationCenter *bool   `json:"notificationCenter,omitempty" bson:"notificationCenter,omitempty"`
	LockScreen         *bool   `json:"lockScreen,omitempty" bson:"lockScreen,omitempty"`
	AlertStyle         *string `json:"alertStyle,omitempty" bson:"alertStyle,omitempty"`
}

func ParseNotifications(parser structure.ObjectParser) *Notifications {
	if !parser.Exists() {
		return nil
	}
	datum := NewNotifications()
	parser.Parse(datum)
	return datum
}

func NewNotifications() *Notifications {
	return &Notifications{}
}

func (n *Notifications) Parse(parser structure.ObjectParser) {
	n.Authorization = parser.String("authorization")
	n.Alert = parser.Bool("alert")
	n.CriticalAlert = parser.Bool("criticalAlert")
	n.Badge = parser.Bool("badge")
	n.Sound = parser.Bool("sound")
	n.Announcement = parser.Bool("announcement")
	n.NotificationCenter = parser.Bool("notificationCenter")
	n.LockScreen = parser.Bool("lockScreen")
	n.AlertStyle = parser.String("alertStyle")
}

func (n *Notifications) Validate(validator structure.Validator) {
	validator.String("authorization", n.Authorization).OneOf(Authorizations()...)
	validator.String("alertStyle", n.AlertStyle).OneOf(AlertStyles()...)

	if n.Authorization == nil && n.Alert == nil && n.CriticalAlert == nil && n.Badge == nil && n.Sound == nil && n.Announcement == nil && n.NotificationCenter == nil && n.LockScreen == nil && n.AlertStyle == nil {
		validator.ReportError(structureValidator.ErrorValuesNotExistForAny("authorization", "alert", "criticalAlert", "badge", "sound", "announcement", "notificationCenter", "lockScreen", "alertStyle"))
	}
}
