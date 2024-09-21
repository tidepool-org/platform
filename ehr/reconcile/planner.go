package reconcile

import (
	"context"

	duration "github.com/xhit/go-str2duration/v2"

	"github.com/tidepool-org/platform/clinics"
	"github.com/tidepool-org/platform/ehr/sync"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/task"
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
		parsed, err := duration.ParseDuration(string(settings.ScheduledReports.Cadence))
		if err != nil {
			p.logger.WithField("clinicId", clinicId).WithError(err).Error("unable to parse scheduled report cadence")
			continue
		}
		cadenceFromSettings = parsed

		tsk, exists := syncTasks[clinicId]
		if exists {

			delete(syncTasks, clinicId)
			if cadenceFromSettings == 0 {
				toDelete = append(toDelete, tsk)
				continue
			}

			cadenceFromTask := sync.GetCadence(tsk.Data)
			if cadenceFromTask == nil || *cadenceFromTask != cadenceFromSettings {
				sync.SetCadence(tsk.Data, cadenceFromSettings)
				update := task.NewTaskUpdate()
				update.Data = &tsk.Data
				toUpdate[tsk.ID] = update
			}
		} else if cadenceFromSettings != 0 {
			// The task doesn't exist yet and scheduled reports are not disabled
			create := sync.NewTaskCreate(clinicId, cadenceFromSettings)
			toCreate = append(toCreate, *create)
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
