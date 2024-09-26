package reconcile

import (
	"context"
	"regexp"
	"strconv"
	"strings"
	"time"

	api "github.com/tidepool-org/clinic/client"

	"github.com/tidepool-org/platform/errors"

	"github.com/tidepool-org/platform/clinics"
	"github.com/tidepool-org/platform/ehr/sync"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/task"
)

var (
	cadenceRegexp = regexp.MustCompile("(\\d{1,3})d")
)

type Planner struct {
	clinicsClient clinics.Client
	logger        log.Logger
}

func NewPlanner(clinicsClient clinics.Client, logger log.Logger) *Planner {
	return &Planner{
		clinicsClient: clinicsClient,
		logger:        logger,
	}
}

func (p *Planner) GetReconciliationPlan(ctx context.Context, syncTasks map[string]task.Task) (*ReconciliationPlan, error) {
	toCreate := make([]task.TaskCreate, 0)
	toDelete := make([]task.Task, 0)
	toUpdate := make(map[string]*task.TaskUpdate)

	// Get the list of all EHR enabled clinics
	clinicsList, err := p.clinicsClient.ListEHREnabledClinics(ctx)
	if err != nil {
		return nil, err
	}

	// At the end of the loop syncTasks will contain only the tasks that need to be deleted,
	// and toCreate will contain tasks for new clinics that need to be synced.
	for _, clinic := range clinicsList {
		clinicId := *clinic.Id
		settings, err := p.clinicsClient.GetEHRSettings(ctx, clinicId)
		if err != nil {
			return nil, err
		} else if settings == nil || !settings.Enabled {
			continue
		}

		// Use the default value for all clinics which don't have a cadence
		cadenceFromSettings := sync.DefaultCadence
		scheduledReportsDisabled := false

		if settings.ScheduledReports.Cadence == api.DISABLED {
			scheduledReportsDisabled = true
		}

		if !scheduledReportsDisabled {
			cadence := string(settings.ScheduledReports.Cadence)
			if !cadenceRegexp.MatchString(string(settings.ScheduledReports.Cadence)) {
				err = errors.New("invalid scheduled report cadence format")
			} else {
				var days int
				cadence = strings.TrimSuffix(cadence, "d")
				days, err = strconv.Atoi(cadence)
				cadenceFromSettings = time.Duration(days) * time.Hour * 24
				if cadenceFromSettings == 0 {
					scheduledReportsDisabled = true
				}
			}
			if err != nil {
				p.logger.WithField("clinicId", clinicId).WithError(err).Error("unable to parse scheduled report cadence")
				continue
			}
		}

		tsk, exists := syncTasks[clinicId]
		if !exists && !scheduledReportsDisabled {
			// Create the tasks for the clinic if reports are not disabled and we don't have a task already
			create := sync.NewTaskCreate(clinicId, cadenceFromSettings)
			toCreate = append(toCreate, *create)
		} else if exists && scheduledReportsDisabled {
			// Delete the task if it tasks exists but reports are now disabled
			delete(syncTasks, clinicId)
			toDelete = append(toDelete, tsk)
		} else if exists && !scheduledReportsDisabled {
			delete(syncTasks, clinicId)
			// Update the task if the configured cadence has been changed
			cadenceFromTask := sync.GetCadence(tsk.Data)
			if cadenceFromTask == nil || *cadenceFromTask != cadenceFromSettings {
				sync.SetCadence(tsk.Data, cadenceFromSettings)
				update := task.NewTaskUpdate()
				update.Data = &tsk.Data
				toUpdate[tsk.ID] = update
			}
		}
	}

	for _, tsk := range syncTasks {
		toDelete = append(toDelete, tsk)
	}
	return &ReconciliationPlan{
		ToCreate: toCreate,
		ToDelete: toDelete,
		ToUpdate: toUpdate,
	}, nil
}
