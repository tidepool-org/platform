package push

import (
	"context"
	"os"

	"github.com/tidepool-org/platform/devicetokens"
	"github.com/tidepool-org/platform/log"
	logjson "github.com/tidepool-org/platform/log/json"
	lognull "github.com/tidepool-org/platform/log/null"
)

// LogPusher simply logs notifications instead of sending push notifications.
//
// Useful for dev or testing situations.
type LogPusher struct {
	log.Logger
}

func NewLogPusher(l log.Logger) *LogPusher {
	if l == nil {
		var err error
		l, err = logjson.NewLogger(os.Stderr, log.DefaultLevelRanks(), log.DefaultLevel())
		if err != nil {
			l = lognull.NewLogger()
		}
	}
	return &LogPusher{Logger: l}
}

func (p *LogPusher) Push(ctx context.Context, deviceToken *devicetokens.DeviceToken, note *Notification) error {
	p.Logger.WithFields(log.Fields{
		"deviceToken": deviceToken,
		"message":     note.Message,
	}).Info("logging push notification")
	return nil
}
