package provider

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/config"
	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/dexcom"
	dexcomFetch "github.com/tidepool-org/platform/dexcom/fetch"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/oauth"
	oauthProvider "github.com/tidepool-org/platform/oauth/provider"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/task"
)

type Provider struct {
	*oauthProvider.Provider
	dataSourceClient dataSource.Client
	taskClient       task.Client
}

// Compile time check for making sure Provider is a valid oauth.Provider
var _ oauth.Provider = &Provider{}

func New(configReporter config.Reporter, dataSourceClient dataSource.Client, taskClient task.Client) (*Provider, error) {
	if configReporter == nil {
		return nil, errors.New("config reporter is missing")
	}
	if dataSourceClient == nil {
		return nil, errors.New("data source client is missing")
	}
	if taskClient == nil {
		return nil, errors.New("task client is missing")
	}

	cfg, err := oauthProvider.NewConfigWithConfigReporter(configReporter.WithScopes(dexcom.ProviderName))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create provider config")
	}

	// Attach prometheus round tripper to default client transport
	prometheusRequestMetricsRoundTripper.WithRoundTripper(http.DefaultClient.Transport)

	// Create http client
	httpClient := &http.Client{
		Transport:     prometheusRequestMetricsRoundTripper,
		CheckRedirect: http.DefaultClient.CheckRedirect,
		Jar:           http.DefaultClient.Jar,
		Timeout:       2 * time.Minute,
	}

	prvdr, err := oauthProvider.New(dexcom.ProviderName, cfg, httpClient, nil)
	if err != nil {
		return nil, err
	}

	return &Provider{
		Provider:         prvdr,
		dataSourceClient: dataSourceClient,
		taskClient:       taskClient,
	}, nil
}

func (p *Provider) OnCreate(ctx context.Context, providerSession *auth.ProviderSession) error {
	if providerSession == nil {
		return errors.New("provider session is missing")
	}

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"type": p.Type(), "name": p.Name()})

	filter := dataSource.NewFilter()
	filter.ProviderType = pointer.FromString(p.Type())
	filter.ProviderName = pointer.FromString(p.Name())
	sources, err := p.dataSourceClient.List(ctx, providerSession.UserID, filter, nil)
	if err != nil {
		return errors.Wrap(err, "unable to fetch data sources")
	}

	var source *dataSource.Source
	if count := len(sources); count > 0 {
		if count > 1 {
			logger.WithField("count", count).Warn("unexpected number of data sources found")
		}

		for _, source := range sources {
			if source.State != dataSource.StateDisconnected {
				logger.WithFields(log.Fields{"id": source.ID, "state": source.State}).Warn("data source in unexpected state")

				update := dataSource.NewUpdate()
				update.State = pointer.FromString(dataSource.StateDisconnected)

				_, err = p.dataSourceClient.Update(ctx, source.ID, nil, update)
				if err != nil {
					return errors.Wrap(err, "unable to update data source")
				}
			}
		}

		source = sources[0]
	} else {
		create := dataSource.NewCreate()
		create.ProviderType = p.Type()
		create.ProviderName = p.Name()

		source, err = p.dataSourceClient.Create(ctx, providerSession.UserID, create)
		if err != nil {
			return errors.Wrap(err, "unable to create data source")
		}
	}

	taskCreate, err := dexcomFetch.NewTaskCreate(providerSession.ID, source.ID)
	if err != nil {
		return errors.Wrap(err, "unable to create task create")
	}

	task, err := p.taskClient.CreateTask(ctx, taskCreate)
	if err != nil {
		return errors.Wrap(err, "unable to create task")
	}

	// Update data source to connected after task successfully created
	update := dataSource.NewUpdate()
	update.ProviderSessionID = pointer.FromString(providerSession.ID)
	update.State = pointer.FromString(dataSource.StateConnected)
	if _, err = p.dataSourceClient.Update(ctx, source.ID, nil, update); err != nil {

		// Attempt to delete task if data source not marked as connected
		if taskErr := p.taskClient.DeleteTask(context.WithoutCancel(ctx), task.ID, nil); taskErr != nil {
			logger.WithError(taskErr).Error("Failure deleting task after failed data source update")
		}

		return errors.Wrap(err, "unable to update data source")
	}

	return nil
}

func (p *Provider) OnDelete(ctx context.Context, providerSession *auth.ProviderSession) error {
	if providerSession == nil {
		return errors.New("provider session is missing")
	}

	logger := log.LoggerFromContext(ctx)

	taskFilter := task.NewTaskFilter()
	taskFilter.Name = pointer.FromString(dexcomFetch.TaskName(providerSession.ID))
	tasks, err := p.taskClient.ListTasks(ctx, taskFilter, nil)
	if err != nil {
		logger.WithError(err).Error("unable to list tasks while deleting provider session")
		return nil
	}

	for _, task := range tasks {
		if dataSourceID, ok := task.Data[dexcom.DataKeyDataSourceID].(string); ok && dataSourceID != "" {
			update := dataSource.NewUpdate()
			update.State = pointer.FromString(dataSource.StateDisconnected)
			_, err = p.dataSourceClient.Update(ctx, dataSourceID, nil, update)
			if err != nil {
				logger.WithError(err).WithField(dexcom.DataKeyDataSourceID, dataSourceID).Error("Unable to update data source while deleting provider session")
			}
		}
		if err = p.taskClient.DeleteTask(ctx, task.ID, nil); err != nil {
			logger.WithError(err).WithField("taskId", task.ID).Error("unable to delete task while deleting provider session")
		}
	}
	return nil
}

// Some Dexcom API responses include a "request-time" header with, supposedly, the internal duration of the request.
// This could be useful for debugging connection issues if compared against the calculated request duration.
// See client/prometheus.go for details on how the request duration is calculated and recorded.

const RequestTimeHeaderName = "request-time"

type PrometheusRequestMetricsRoundTripper struct {
	*client.PrometheusRequestMetricsRoundTripper
	requestTimeHistogramVec *prometheus.HistogramVec
}

func NewPrometheusRequestMetricsRoundTripper(name string, help string) *PrometheusRequestMetricsRoundTripper {
	return &PrometheusRequestMetricsRoundTripper{
		PrometheusRequestMetricsRoundTripper: client.NewPrometheusRequestMetricsRoundTripper(name, help),
		requestTimeHistogramVec: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    fmt.Sprintf("%s_request_time_seconds", name),
				Help:    fmt.Sprintf("%s request time (seconds)", help),
				Buckets: client.DurationBucketsDefault,
			},
			client.PrometheusLabelNames(),
		),
	}
}

func (p *PrometheusRequestMetricsRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	res, err := p.ResolvedRoundTripper().RoundTrip(req)

	if res != nil {
		if labels := p.Labels(req, res); labels != nil {
			if requestTime, parseErr := time.ParseDuration(res.Header.Get(RequestTimeHeaderName)); parseErr == nil {
				p.requestTimeHistogramVec.With(*labels).Observe(requestTime.Seconds())
			}
		}
	}

	return res, err
}

var prometheusRequestMetricsRoundTripper = NewPrometheusRequestMetricsRoundTripper("tidepool_dexcom_api", "Tidepool Dexcom API")
