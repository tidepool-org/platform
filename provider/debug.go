package provider

import (
	"context"
	"maps"
	"slices"

	"github.com/tidepool-org/platform/auth"
	providerSession "github.com/tidepool-org/platform/auth/providersession"
	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type DebugSubjectInfo struct {
	ID               string                 `json:"id,omitempty"`
	DataSources      dataSource.SourceArray `json:"dataSources,omitempty"`
	ProviderSessions auth.ProviderSessions  `json:"providerSessions,omitempty"`
}

type DebugInfo struct {
	Config   DebuggerConfig      `json:"config,omitempty"`
	Errors   errors.Serializable `json:"errors,omitzero"`
	User     DebugSubjectInfo    `json:"user,omitempty"`
	External *DebugSubjectInfo   `json:"external,omitempty"`
}

func (d *DebugInfo) AppendError(err error) {
	d.Errors.Error = errors.Append(d.Errors.Error, err)
}

func (d *DebugInfo) ProviderSessions() auth.ProviderSessions {
	providerSessionsMap := map[string]*auth.ProviderSession{}
	for _, providerSession := range d.User.ProviderSessions {
		providerSessionsMap[providerSession.ID] = providerSession
	}
	if d.External != nil {
		for _, providerSession := range d.External.ProviderSessions {
			providerSessionsMap[providerSession.ID] = providerSession
		}
	}
	return slices.SortedFunc(maps.Values(providerSessionsMap), func(left *auth.ProviderSession, right *auth.ProviderSession) int {
		return left.CreatedTime.Compare(right.CreatedTime)
	})
}

func (d *DebugInfo) DataSources() dataSource.SourceArray {
	dataSrcsMap := map[string]*dataSource.Source{}
	for _, dataSrc := range d.User.DataSources {
		dataSrcsMap[dataSrc.ID] = dataSrc
	}
	if d.External != nil {
		for _, dataSrc := range d.External.DataSources {
			dataSrcsMap[dataSrc.ID] = dataSrc
		}
	}
	return slices.SortedFunc(maps.Values(dataSrcsMap), func(left *dataSource.Source, right *dataSource.Source) int {
		return left.CreatedTime.Compare(right.CreatedTime)
	})
}

type DebuggerConfig struct {
	Type *string `json:"type,omitempty"`
	Name *string `json:"name,omitempty"`
}

type DebuggerDependencies struct {
	Config                DebuggerConfig
	ProviderSessionClient providerSession.Client
	DataSourceClient      dataSource.Client
}

type Debugger struct {
	DebuggerDependencies
}

func NewDebugger(dependencies DebuggerDependencies) Debugger {
	return Debugger{
		DebuggerDependencies: dependencies,
	}
}

func (d Debugger) GetDebugInfo(ctx context.Context, userID string, externalID *string) DebugInfo {
	debugInfo := DebugInfo{Config: d.Config}

	dataSrcFilter := &dataSource.Filter{
		ProviderType: d.Config.Type,
		ProviderName: d.Config.Name,
	}
	if dataSrcs, err := d.DataSourceClient.List(ctx, userID, dataSrcFilter, nil); err != nil {
		debugInfo.AppendError(errors.Wrapf(err, "unable to list user data sources"))
	} else {
		debugInfo.User.DataSources = dataSrcs
	}

	providerSessionFilter := &auth.ProviderSessionFilter{
		UserID: &userID,
		Type:   d.Config.Type,
		Name:   d.Config.Name,
	}
	if providerSessions, err := d.ProviderSessionClient.ListProviderSessions(ctx, providerSessionFilter, nil); err != nil {
		debugInfo.AppendError(errors.Wrapf(err, "unable to list user provider sessions"))
	} else {
		debugInfo.User.ProviderSessions = providerSessions.Redacted()
	}

	if externalID != nil {
		debugInfo.External = &DebugSubjectInfo{ID: *externalID}

		providerSessionFilter := &auth.ProviderSessionFilter{
			Type:       d.Config.Type,
			Name:       d.Config.Name,
			ExternalID: externalID,
		}
		if providerSessions, err := d.ProviderSessionClient.ListProviderSessions(ctx, providerSessionFilter, nil); err != nil {
			debugInfo.AppendError(errors.Wrapf(err, "unable to list external provider sessions"))
		} else {
			debugInfo.External.ProviderSessions = providerSessions.Redacted()

			for _, providerSession := range debugInfo.External.ProviderSessions {
				if dataSrc, err := d.DataSourceClient.GetFromProviderSession(ctx, providerSession.ID); err != nil {
					debugInfo.AppendError(errors.Wrapf(err, "unable to get data source for provider session with id %q", providerSession.ID))
				} else if dataSrc != nil {
					debugInfo.External.DataSources = append(debugInfo.External.DataSources, dataSrc)
				}
			}
		}
	}

	// General sanity checks
	lgr := log.LoggerFromContext(ctx)
	for _, dataSrc := range debugInfo.DataSources() {
		if err := structureValidator.New(lgr).Validate(dataSrc); err != nil {
			debugInfo.AppendError(errors.Wrapf(err, "data source %q is invalid", dataSrc.ID))
		}
	}
	for _, providerSession := range debugInfo.ProviderSessions() {
		if err := structureValidator.New(lgr).Validate(providerSession); err != nil {
			debugInfo.AppendError(errors.Wrapf(err, "provider session %q is invalid", providerSession.ID))
		}
	}

	// User-specific sanity checks
	for _, dataSrc := range debugInfo.User.DataSources {
		if !slices.ContainsFunc(debugInfo.User.ProviderSessions, func(providerSession *auth.ProviderSession) bool {
			return dataSrc.ProviderSessionID != nil && *dataSrc.ProviderSessionID == providerSession.ID
		}) {
			debugInfo.AppendError(errors.Newf("user data source %q missing user provider session", dataSrc.ID))
		}
	}
	for _, providerSession := range debugInfo.User.ProviderSessions {
		if !slices.ContainsFunc(debugInfo.User.DataSources, func(dataSrc *dataSource.Source) bool {
			return dataSrc.ProviderSessionID != nil && *dataSrc.ProviderSessionID == providerSession.ID
		}) {
			debugInfo.AppendError(errors.Newf("user provider session %q missing user data source", providerSession.ID))
		}
	}

	return debugInfo
}
