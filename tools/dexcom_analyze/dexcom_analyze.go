package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli"
	"golang.org/x/exp/maps"

	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/auth"
	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/dexcom"
	dexcomFetch "github.com/tidepool-org/platform/dexcom/fetch"
	dexcomProvider "github.com/tidepool-org/platform/dexcom/provider"
	"github.com/tidepool-org/platform/errors"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/task"
	"github.com/tidepool-org/platform/tool"
)

const (
	DataSourcesFileFlag      = "data-sources-file-flag"
	ProviderSessionsFileFlag = "provider-sessions-file-flag"
	TasksFileFlag            = "tasks-file-flag"

	Issue_DataSource_ProviderSessionID_DoesNotExist                      = "data source provider session id does not exist"
	Issue_DataSource_ProviderSessionID_Empty                             = "data source provider session id empty"
	Issue_DataSource_UserID_Empty                                        = "data source user id empty"
	Issue_DataSource_UserID_Missing                                      = "data source user id missing"
	Issue_DataSource_With_DataSetIDs_DataSetIDs_Length_Invalid           = "data source with data set ids data set ids length invalid"
	Issue_DataSource_With_DataSetIDs_EarliestDataTime_Missing            = "data source with data set ids earliest data time missing"
	Issue_DataSource_With_DataSetIDs_LastImportTime_Missing              = "data source with data set ids last import time missing"
	Issue_DataSource_With_DataSetIDs_LatestDataTime_Missing              = "data source with data set ids latest data time missing"
	Issue_DataSource_With_State_Connected_Error_Present                  = "data source with state connected error present"
	Issue_DataSource_With_State_Connected_ProviderSession_Missing        = "data source with state connected provider session missing"
	Issue_DataSource_With_State_Connected_ProviderSessionID_Missing      = "data source with state connected provider session id missing"
	Issue_DataSource_With_State_Connected_Task_Missing                   = "data source with state connected task missing"
	Issue_DataSource_With_State_Disconnected_Error_Present               = "data source with state disconnected error present"
	Issue_DataSource_With_State_Disconnected_ProviderSession_Present     = "data source with state disconnected provider session present"
	Issue_DataSource_With_State_Disconnected_ProviderSessionID_Present   = "data source with state disconnected provider session id present"
	Issue_DataSource_With_State_Disconnected_Task_Present                = "data source with state disconnected task present"
	Issue_DataSource_With_State_Error_Error_Missing                      = "data source with state error error missing"
	Issue_DataSource_With_State_Error_ProviderSession_Missing            = "data source with state error provider session missing"
	Issue_DataSource_With_State_Error_ProviderSessionID_Missing          = "data source with state error provider session id missing"
	Issue_DataSource_With_State_Error_Task_Missing                       = "data source with state error task missing"
	Issue_DataSource_Without_DataSetIDs_EarliestDataTime_Present         = "data source without data set ids earliest data time present"
	Issue_DataSource_Without_DataSetIDs_LatestDataTime_Present           = "data source without data set ids latest data time present"
	Issue_ProviderSession_Task_Missing                                   = "provider session task missing"
	Issue_ProviderSession_UserID_Empty                                   = "provider session user id empty"
	Issue_Task_Data_Missing                                              = "task data missing"
	Issue_Task_DataSourceID_DoesNotExist                                 = "task data source id does not exist"
	Issue_Task_DataSourceID_Empty                                        = "task data source id empty"
	Issue_Task_DataSourceID_Missing                                      = "task data source id missing"
	Issue_Task_DeviceHash_Invalid                                        = "task device hash invalid"
	Issue_Task_DeviceHashes_Invalid                                      = "task device hashes invalid"
	Issue_Task_Name_Invalid                                              = "task name invalid"
	Issue_Task_Name_Missing                                              = "task name missing"
	Issue_Task_ProviderSession_Missing                                   = "task provider session missing"
	Issue_Task_ProviderSessionID_DoesNotExist                            = "task provider session id does not exist"
	Issue_Task_ProviderSessionID_Empty                                   = "task provider session id empty"
	Issue_Task_ProviderSessionID_Missing                                 = "task provider session id missing"
	Issue_Task_With_DeviceHashes_And_DataSource_DataSetIDs_Missing       = "task with device hashes and data source data set ids missing"
	Issue_Task_With_DeviceHashes_And_DataSource_EarliestDataTime_Missing = "task with device hashes and data source earliest data time missing"
	Issue_Task_With_DeviceHashes_And_DataSource_LastImportTime_Missing   = "task with device hashes and data source last import time missing"
	Issue_Task_With_DeviceHashes_And_DataSource_LatestDataTime_Missing   = "task with device hashes and data source latest data time missing"
	Issue_Task_With_State_Failed_AvailableTime_Present                   = "task with state failed available time present"
	Issue_Task_With_State_Failed_DeadlineTime_Present                    = "task with state failed deadline time present"
	Issue_Task_With_State_Failed_Error_Missing                           = "task with state failed error missing"
	Issue_Task_With_State_Failed_ExpirationTime_Present                  = "task with state failed expiration time present"
	Issue_Task_With_State_Pending_DeadlineTime_Present                   = "task with state pending deadline time present"
	Issue_Task_With_State_Pending_ExpirationTime_Present                 = "task with state pending expiration time present"
	Issue_Task_With_State_Running_AvailableTime_Present                  = "task with state running available time present"
	Issue_Task_With_State_Running_DeadlineTime_Missing                   = "task with state running deadline time missing"
	Issue_Task_With_State_Running_Error_Present                          = "task with state running error present"
	Issue_Task_With_State_Running_ExpirationTime_Present                 = "task with state running expiration time present"

	IssueFormat_DataSource_Invalid                          = "data source invalid ('%s', '%s')"
	IssueFormat_DataSource_ProviderSession_Mismatch         = "data source provider session mismatch ('%s', '%s')"
	IssueFormat_DataSource_Task_Mismatch                    = "data source task mismatch ('%s', '%s')"
	IssueFormat_DataSource_User_Mismatch                    = "data source user mismatch ('%s', '%s')"
	IssueFormat_ProviderSession_DataSource_Mismatch         = "provider session data source mismatch ('%s', '%s')"
	IssueFormat_ProviderSession_Invalid                     = "provider session invalid ('%s', '%s')"
	IssueFormat_ProviderSession_Task_Mismatch               = "provider session task mismatch ('%s', '%s')"
	IssueFormat_ProviderSession_User_Mismatch               = "provider session user mismatch ('%s', '%s')"
	IssueFormat_Task_DataSource_Mismatch                    = "task data source mismatch ('%s', '%s')"
	IssueFormat_Task_Invalid                                = "task invalid ('%s', '%s')"
	IssueFormat_Task_ProviderSession_And_DataSource_Invalid = "task provider session and data source invalid ('%s')"
	IssueFormat_Task_ProviderSession_Mismatch               = "task provider session mismatch ('%s', '%s')"
	IssueFormat_Task_State_Invalid                          = "task state invalid ('%s')"
	IssueFormat_Task_User_Mismatch                          = "task user mismatch ('%s', '%s')"
	IssueFormat_User_DataSource_Mismatch                    = "user data source mismatch ('%s', '%s')"
	IssueFormat_User_ProviderSession_Mismatch               = "user provider session mismatch ('%s', '%s')"
	IssueFormat_User_Task_Mismatch                          = "user task mismatch ('%s', '%s')"
)

func Issues() []string {
	return []string{
		Issue_DataSource_ProviderSessionID_DoesNotExist,
		Issue_DataSource_ProviderSessionID_Empty,
		Issue_DataSource_UserID_Empty,
		Issue_DataSource_UserID_Missing,
		Issue_DataSource_With_DataSetIDs_DataSetIDs_Length_Invalid,
		Issue_DataSource_With_DataSetIDs_EarliestDataTime_Missing,
		Issue_DataSource_With_DataSetIDs_LastImportTime_Missing,
		Issue_DataSource_With_DataSetIDs_LatestDataTime_Missing,
		Issue_DataSource_With_State_Connected_Error_Present,
		Issue_DataSource_With_State_Connected_ProviderSessionID_Missing,
		Issue_DataSource_With_State_Connected_Task_Missing,
		Issue_DataSource_With_State_Disconnected_Error_Present,
		Issue_DataSource_With_State_Disconnected_ProviderSessionID_Present,
		Issue_DataSource_With_State_Disconnected_Task_Present,
		Issue_DataSource_With_State_Error_Error_Missing,
		Issue_DataSource_With_State_Error_ProviderSessionID_Missing,
		Issue_DataSource_With_State_Error_Task_Missing,
		Issue_DataSource_Without_DataSetIDs_EarliestDataTime_Present,
		Issue_DataSource_Without_DataSetIDs_LatestDataTime_Present,
		Issue_ProviderSession_Task_Missing,
		Issue_ProviderSession_UserID_Empty,
		Issue_Task_Data_Missing,
		Issue_Task_DataSourceID_DoesNotExist,
		Issue_Task_DataSourceID_Empty,
		Issue_Task_DataSourceID_Missing,
		Issue_Task_DeviceHash_Invalid,
		Issue_Task_DeviceHashes_Invalid,
		Issue_Task_Name_Invalid,
		Issue_Task_Name_Missing,
		Issue_Task_ProviderSessionID_DoesNotExist,
		Issue_Task_ProviderSessionID_Empty,
		Issue_Task_ProviderSessionID_Missing,
		Issue_Task_With_DeviceHashes_And_DataSource_DataSetIDs_Missing,
		Issue_Task_With_DeviceHashes_And_DataSource_EarliestDataTime_Missing,
		Issue_Task_With_DeviceHashes_And_DataSource_LastImportTime_Missing,
		Issue_Task_With_DeviceHashes_And_DataSource_LatestDataTime_Missing,
		Issue_Task_With_State_Failed_AvailableTime_Present,
		Issue_Task_With_State_Failed_DeadlineTime_Present,
		Issue_Task_With_State_Failed_Error_Missing,
		Issue_Task_With_State_Failed_ExpirationTime_Present,
		Issue_Task_With_State_Pending_DeadlineTime_Present,
		Issue_Task_With_State_Pending_ExpirationTime_Present,
		Issue_Task_With_State_Running_AvailableTime_Present,
		Issue_Task_With_State_Running_DeadlineTime_Missing,
		Issue_Task_With_State_Running_Error_Present,
		Issue_Task_With_State_Running_ExpirationTime_Present,
	}
}

func IssueFormats() []string {
	return []string{
		IssueFormat_DataSource_Invalid,
		IssueFormat_DataSource_ProviderSession_Mismatch,
		IssueFormat_DataSource_Task_Mismatch,
		IssueFormat_DataSource_User_Mismatch,
		IssueFormat_ProviderSession_DataSource_Mismatch,
		IssueFormat_ProviderSession_Invalid,
		IssueFormat_ProviderSession_Task_Mismatch,
		IssueFormat_ProviderSession_User_Mismatch,
		IssueFormat_Task_DataSource_Mismatch,
		IssueFormat_Task_Invalid,
		IssueFormat_Task_ProviderSession_And_DataSource_Invalid,
		IssueFormat_Task_ProviderSession_Mismatch,
		IssueFormat_Task_State_Invalid,
		IssueFormat_Task_User_Mismatch,
		IssueFormat_User_DataSource_Mismatch,
		IssueFormat_User_ProviderSession_Mismatch,
		IssueFormat_User_Task_Mismatch,
	}
}

type IDs []string

func (i IDs) Add(others ...IDs) IDs {
	clone := slices.Clone(i)
	if len(others) > 0 {
		for _, other := range others {
			clone = append(clone, other...)
		}
	}
	return clone
}

func (i IDs) Subtract(others ...IDs) IDs {
	if len(i) == 0 || len(others) == 0 {
		return slices.Clone(i)
	}

	other := others[0].Add(others[1:]...).Sort().Compact()
	if len(other) == 0 {
		return slices.Clone(i)
	}

	difference := make(IDs, 0, len(i))
	for _, id := range i {
		if _, ok := slices.BinarySearch(other, id); !ok {
			difference = append(difference, id)
		}
	}
	return difference
}

func (i IDs) Intersection(others ...IDs) IDs {
	clone := slices.Clone(i)
	for _, other := range others {
		var intersection IDs

		other = other.Sort()
		for _, id := range clone {
			if _, ok := slices.BinarySearch(other, id); ok {
				intersection = append(intersection, id)
			}
		}
		if len(intersection) == 0 {
			return nil
		}

		clone = intersection
	}
	return clone
}

func (i IDs) Sort() IDs {
	clone := slices.Clone(i)
	slices.Sort(clone)
	return clone
}

func (i IDs) Compact() IDs {
	return slices.Compact(i)
}

type IssueReporter interface {
	ReportIssues() []string
}

type WithIssues []string

func (w *WithIssues) AppendIssue(issue string) {
	if issue != "" {
		*w = append(*w, issue)
	}
}

func (w *WithIssues) AppendIssuef(format string, a ...interface{}) {
	w.AppendIssue(fmt.Sprintf(format, a...))
}

func (w *WithIssues) ReportIssues() []string {
	return *w
}

type Marshalable struct {
	DataSource      *dataSource.Source    `json:"data_source,omitempty"`
	ProviderSession *auth.ProviderSession `json:"provider_session,omitempty"`
	Task            *task.Task            `json:"task,omitempty"`
	User            *UserData             `json:"user,omitempty"`
}

type Marshalables []*Marshalable

func (m Marshalables) DataSources() DataSources {
	var dataSources DataSources
	for _, marshable := range m {
		dataSources = append(dataSources, marshable.DataSource)
	}
	return dataSources
}

func (m Marshalables) ProviderSessions() ProviderSessions {
	var providerSession ProviderSessions
	for _, marshable := range m {
		providerSession = append(providerSession, marshable.ProviderSession)
	}
	return providerSession
}

func (m Marshalables) Tasks() Tasks {
	var tasks Tasks
	for _, marshable := range m {
		tasks = append(tasks, marshable.Task)
	}
	return tasks
}

type AsMarshalable interface {
	AsMarshalable() *Marshalable
}

type IssueReporterAsMarshalable interface {
	IssueReporter
	AsMarshalable
}

type IssueMarshalableMap map[string]Marshalables

func (i IssueMarshalableMap) AppendIssueReporterAsMarshalable(reportedErrorsAsMarshalable IssueReporterAsMarshalable) {
	marshalable := reportedErrorsAsMarshalable.AsMarshalable()
	for _, issue := range reportedErrorsAsMarshalable.ReportIssues() {
		i[issue] = append(i[issue], marshalable)
	}
}

type IssueFormatRE struct {
	IssueFormat string
	RE          *regexp.Regexp
}

type DataSources []*dataSource.Source

func (d DataSources) IDs() IDs {
	var ids IDs
	for _, dataSource := range d {
		ids = append(ids, *dataSource.ID)
	}
	slices.Sort(ids)
	return ids
}

func (d DataSources) ProviderSessionIDs() IDs {
	var ids IDs
	for _, dataSource := range d {
		ids = append(ids, *dataSource.ProviderSessionID)
	}
	slices.Sort(ids)
	return ids
}

type DataSource struct {
	dataSource.Source
	WithIssues

	providerSession *ProviderSession
	task            *Task
	user            *User
}

func (d *DataSource) SetProviderSession(providerSession *ProviderSession) {
	if providerSession == nil {
		return
	}

	if d.providerSession == nil {
		d.providerSession = providerSession
		providerSession.SetDataSource(d)
		if d.task != nil {
			providerSession.SetTask(d.task)
			d.task.SetProviderSession(providerSession)
		}
		if d.user != nil {
			providerSession.SetUser(d.user)
			d.user.SetProviderSession(providerSession)
		}
	} else if d.providerSession != providerSession {
		d.AppendIssuef(IssueFormat_DataSource_ProviderSession_Mismatch, d.providerSession.ID, providerSession.ID)
	}
}

func (d *DataSource) SetTask(task *Task) {
	if task == nil {
		return
	}

	if d.task == nil {
		d.task = task
		task.SetDataSource(d)
		if d.providerSession != nil {
			task.SetProviderSession(d.providerSession)
			d.providerSession.SetTask(task)
		}
		if d.user != nil {
			task.SetUser(d.user)
			d.user.SetTask(task)
		}
	} else if d.task != task {
		d.AppendIssuef(IssueFormat_DataSource_Task_Mismatch, d.task.ID, task.ID)
	}
}

func (d *DataSource) SetUser(user *User) {
	if user == nil {
		return
	}

	if d.user == nil {
		d.user = user
		user.SetDataSource(d)
		if d.providerSession != nil {
			user.SetProviderSession(d.providerSession)
			d.providerSession.SetUser(user)
		}
		if d.task != nil {
			user.SetTask(d.task)
			d.task.SetUser(user)
		}
	} else if d.user != user {
		d.AppendIssuef(IssueFormat_DataSource_User_Mismatch, d.user.ID, user.ID)
	}
}

func (d *DataSource) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.AsMarshalable())
}

func (d *DataSource) AsMarshalable() *Marshalable {
	marshable := &Marshalable{
		DataSource: &d.Source,
	}
	if d.providerSession != nil {
		marshable.ProviderSession = &d.providerSession.ProviderSession
	}
	if d.task != nil {
		marshable.Task = &d.task.Task
	}
	if d.user != nil {
		marshable.User = &d.user.UserData
	}
	return marshable
}

type ProviderSessions []*auth.ProviderSession

func (p ProviderSessions) IDs() IDs {
	var ids IDs
	for _, providerSession := range p {
		ids = append(ids, providerSession.ID)
	}
	slices.Sort(ids)
	return ids
}

type ProviderSession struct {
	auth.ProviderSession
	WithIssues

	dataSource *DataSource
	task       *Task
	user       *User
}

func (p *ProviderSession) SetDataSource(dataSource *DataSource) {
	if dataSource == nil {
		return
	}

	if p.dataSource == nil {
		p.dataSource = dataSource
		dataSource.SetProviderSession(p)
		if p.task != nil {
			dataSource.SetTask(p.task)
			p.task.SetDataSource(dataSource)
		}
		if p.user != nil {
			dataSource.SetUser(p.user)
			p.user.SetDataSource(dataSource)
		}
	} else if p.dataSource != dataSource {
		p.AppendIssuef(IssueFormat_ProviderSession_DataSource_Mismatch, *p.dataSource.ID, *dataSource.ID)
	}
}

func (p *ProviderSession) SetTask(task *Task) {
	if task == nil {
		return
	}

	if p.task == nil {
		p.task = task
		task.SetProviderSession(p)
		if p.dataSource != nil {
			task.SetDataSource(p.dataSource)
			p.dataSource.SetTask(task)
		}
		if p.user != nil {
			task.SetUser(p.user)
			p.user.SetTask(task)
		}
	} else if p.task != task {
		p.AppendIssuef(IssueFormat_ProviderSession_Task_Mismatch, p.task.ID, task.ID)
	}
}

func (p *ProviderSession) SetUser(user *User) {
	if user == nil {
		return
	}

	if p.user == nil {
		p.user = user
		user.SetProviderSession(p)
		if p.dataSource != nil {
			user.SetDataSource(p.dataSource)
			p.dataSource.SetUser(user)
		}
		if p.task != nil {
			user.SetTask(p.task)
			p.task.SetUser(user)
		}
	} else if p.user != user {
		p.AppendIssuef(IssueFormat_ProviderSession_User_Mismatch, p.user.ID, user.ID)
	}
}

func (p *ProviderSession) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.AsMarshalable())
}

func (p *ProviderSession) AsMarshalable() *Marshalable {
	marshable := &Marshalable{
		ProviderSession: &p.ProviderSession,
	}
	if p.dataSource != nil {
		marshable.DataSource = &p.dataSource.Source
	}
	if p.task != nil {
		marshable.Task = &p.task.Task
	}
	if p.user != nil {
		marshable.User = &p.user.UserData
	}
	return marshable
}

type Tasks []*task.Task

func (t Tasks) IDs() IDs {
	var ids IDs
	for _, task := range t {
		ids = append(ids, task.ID)
	}
	slices.Sort(ids)
	return ids
}

type Task struct {
	task.Task
	WithIssues

	dataSource      *DataSource
	providerSession *ProviderSession
	user            *User
}

func (t *Task) SetDataSource(dataSource *DataSource) {
	if dataSource == nil {
		return
	}

	if t.dataSource == nil {
		t.dataSource = dataSource
		dataSource.SetTask(t)
		if t.providerSession != nil {
			dataSource.SetProviderSession(t.providerSession)
			t.providerSession.SetDataSource(dataSource)
		}
		if t.user != nil {
			dataSource.SetUser(t.user)
			t.user.SetDataSource(dataSource)
		}
	} else if t.dataSource != dataSource {
		t.AppendIssuef(IssueFormat_Task_DataSource_Mismatch, *t.dataSource.ID, *dataSource.ID)
	}
}
func (t *Task) SetProviderSession(providerSession *ProviderSession) {
	if providerSession == nil {
		return
	}

	if t.providerSession == nil {
		t.providerSession = providerSession
		providerSession.SetTask(t)
		if t.dataSource != nil {
			providerSession.SetDataSource(t.dataSource)
			t.dataSource.SetProviderSession(providerSession)
		}
		if t.user != nil {
			providerSession.SetUser(t.user)
			t.user.SetProviderSession(providerSession)
		}
	} else if t.providerSession != providerSession {
		t.AppendIssuef(IssueFormat_Task_ProviderSession_Mismatch, t.providerSession.ID, providerSession.ID)
	}
}

func (t *Task) SetUser(user *User) {
	if user == nil {
		return
	}

	if t.user == nil {
		t.user = user
		user.SetTask(t)
		if t.dataSource != nil {
			user.SetDataSource(t.dataSource)
			t.dataSource.SetUser(user)
		}
		if t.providerSession != nil {
			user.SetProviderSession(t.providerSession)
			t.providerSession.SetUser(user)
		}
	} else if t.user != user {
		t.AppendIssuef(IssueFormat_Task_User_Mismatch, t.user.ID, user.ID)
	}
}

func (t *Task) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.AsMarshalable())
}

func (t *Task) AsMarshalable() *Marshalable {
	marshable := &Marshalable{
		Task: &t.Task,
	}
	if t.dataSource != nil {
		marshable.DataSource = &t.dataSource.Source
	}
	if t.providerSession != nil {
		marshable.ProviderSession = &t.providerSession.ProviderSession
	}
	if t.user != nil {
		marshable.User = &t.user.UserData
	}
	return marshable
}

type UserData struct {
	ID string `json:"id,omitempty"`
}

type User struct {
	UserData
	WithIssues

	dataSource      *DataSource
	providerSession *ProviderSession
	task            *Task
}

func (u *User) SetDataSource(dataSource *DataSource) {
	if dataSource == nil {
		return
	}

	if u.dataSource == nil {
		u.dataSource = dataSource
		dataSource.SetUser(u)
		if u.providerSession != nil {
			dataSource.SetProviderSession(u.providerSession)
			u.providerSession.SetDataSource(dataSource)
		}
		if u.task != nil {
			dataSource.SetTask(u.task)
			u.task.SetDataSource(dataSource)
		}
	} else if u.dataSource != dataSource {
		u.AppendIssuef(IssueFormat_User_DataSource_Mismatch, *u.dataSource.ID, *dataSource.ID)
	}
}

func (u *User) SetProviderSession(providerSession *ProviderSession) {
	if providerSession == nil {
		return
	}

	if u.providerSession == nil {
		u.providerSession = providerSession
		providerSession.SetUser(u)
		if u.dataSource != nil {
			providerSession.SetDataSource(u.dataSource)
			u.dataSource.SetProviderSession(providerSession)
		}
		if u.task != nil {
			providerSession.SetTask(u.task)
			u.task.SetProviderSession(providerSession)
		}
	} else if u.providerSession != providerSession {
		u.AppendIssuef(IssueFormat_User_ProviderSession_Mismatch, u.providerSession.ID, providerSession.ID)
	}
}

func (u *User) SetTask(task *Task) {
	if task == nil {
		return
	}

	if u.task == nil {
		u.task = task
		task.SetUser(u)
		if u.dataSource != nil {
			task.SetDataSource(u.dataSource)
			u.dataSource.SetTask(task)
		}
		if u.providerSession != nil {
			task.SetProviderSession(u.providerSession)
			u.providerSession.SetTask(task)
		}
	} else if u.task != task {
		u.AppendIssuef(IssueFormat_User_Task_Mismatch, u.task.ID, task.ID)
	}
}

func (u *User) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.AsMarshalable())
}

func (u *User) AsMarshalable() *Marshalable {
	marshable := &Marshalable{
		User: &u.UserData,
	}
	if u.dataSource != nil {
		marshable.DataSource = &u.dataSource.Source
	}
	if u.providerSession != nil {
		marshable.ProviderSession = &u.providerSession.ProviderSession
	}
	if u.task != nil {
		marshable.Task = &u.task.Task
	}
	return marshable
}

func main() {
	application.RunAndExit(NewTool())
}

type Tool struct {
	*tool.Tool
	issueFormatREs       []IssueFormatRE
	dataSourcesFile      string
	dataSourcesMap       map[string]*DataSource
	providerSessionsFile string
	providerSessionsMap  map[string]*ProviderSession
	tasksFile            string
	tasksMap             map[string]*Task
	usersMap             map[string]*User
	output               io.Writer
}

func NewTool() *Tool {
	return &Tool{
		Tool:   tool.New(),
		output: os.Stdout,
	}
}

func (t *Tool) Initialize(provider application.Provider) error {
	if err := t.Tool.Initialize(provider); err != nil {
		return err
	}

	for _, issueFormat := range IssueFormats() {
		expression := fmt.Sprintf("^%s$", strings.ReplaceAll(regexp.QuoteMeta(issueFormat), "%s", "(.*)"))
		if re, err := regexp.Compile(expression); err != nil {
			return err
		} else {
			issueFormatRE := IssueFormatRE{
				IssueFormat: issueFormat,
				RE:          re,
			}
			t.issueFormatREs = append(t.issueFormatREs, issueFormatRE)
		}
	}

	t.CLI().Usage = "Analyze Dexcom"
	t.CLI().Authors = []cli.Author{
		{
			Name:  "Darin Krauss",
			Email: "darin.krauss@tidepool.org",
		},
	}
	t.CLI().Flags = append(t.CLI().Flags,
		cli.StringFlag{
			Name:  fmt.Sprintf("%s,%s", DataSourcesFileFlag, "d"),
			Usage: "data sources file",
		},
		cli.StringFlag{
			Name:  fmt.Sprintf("%s,%s", ProviderSessionsFileFlag, "p"),
			Usage: "provider sessions file",
		},
		cli.StringFlag{
			Name:  fmt.Sprintf("%s,%s", TasksFileFlag, "t"),
			Usage: "tasks file",
		},
	)
	t.CLI().Action = func(ctx *cli.Context) error {
		if !t.ParseContext(ctx) {
			return nil
		}
		return t.execute()
	}

	return nil
}

func (t *Tool) ParseContext(ctx *cli.Context) bool {
	if parsed := t.Tool.ParseContext(ctx); !parsed {
		return parsed
	}

	t.dataSourcesFile = ctx.String(DataSourcesFileFlag)
	t.providerSessionsFile = ctx.String(ProviderSessionsFileFlag)
	t.tasksFile = ctx.String(TasksFileFlag)

	return true
}

func (t *Tool) execute() error {
	if err := t.load(); err != nil {
		return errors.Wrap(err, "unable to load")
	}

	t.analyze()
	return nil
}

func (t *Tool) load() error {
	dataSources, err := loadFile[DataSource](t.dataSourcesFile)
	if err != nil {
		return errors.Wrap(err, "unable to load data sources file")
	}
	dataSourcesMap := map[string]*DataSource{}
	for _, dataSource := range dataSources {
		if dataSource.ProviderType != nil && *dataSource.ProviderType == auth.ProviderTypeOAuth && dataSource.ProviderName != nil && *dataSource.ProviderName == dexcomProvider.ProviderName {
			dataSourcesMap[*dataSource.ID] = dataSource
		}
	}

	providerSessions, err := loadFile[ProviderSession](t.providerSessionsFile)
	if err != nil {
		return errors.Wrap(err, "unable to load provider sessions file")
	}
	providerSessionsMap := map[string]*ProviderSession{}
	for _, providerSession := range providerSessions {
		if providerSession.Type == auth.ProviderTypeOAuth && providerSession.Name == dexcomProvider.ProviderName {
			providerSessionsMap[providerSession.ID] = providerSession
		}
	}

	tasks, err := loadFile[Task](t.tasksFile)
	if err != nil {
		return errors.Wrap(err, "unable to load data sources file")
	}
	tasksMap := map[string]*Task{}
	for _, task := range tasks {
		if task.Type == dexcomFetch.Type {
			tasksMap[task.ID] = task
		}
	}

	usersMap := map[string]*User{}

	t.dataSourcesMap = dataSourcesMap
	t.providerSessionsMap = providerSessionsMap
	t.tasksMap = tasksMap
	t.usersMap = usersMap
	return nil
}

func (t *Tool) analyze() {
	t.analyzeAssociations()
	t.analyzeDataSources()
	t.analyzeProviderSessions()
	t.analyzeTasks()
	t.analyzeIssues()
}

func (t *Tool) analyzeAssociations() {
	for _, record := range t.providerSessionsMap {
		if userID := record.UserID; userID != "" {
			user, ok := t.usersMap[userID]
			if !ok {
				user = &User{UserData: UserData{ID: userID}}
				t.usersMap[userID] = user
			}
			record.SetUser(user)
		} else {
			record.AppendIssue(Issue_ProviderSession_UserID_Empty)
		}
	}

	for _, record := range t.dataSourcesMap {
		if userID := record.UserID; userID != nil {
			if *userID != "" {
				user, ok := t.usersMap[*userID]
				if !ok {
					user = &User{UserData: UserData{ID: *userID}}
					t.usersMap[*userID] = user
				}
				record.SetUser(user)
			} else {
				record.AppendIssue(Issue_DataSource_UserID_Empty)
			}
		} else {
			record.AppendIssue(Issue_DataSource_UserID_Missing)
		}

		if providerSessionID := record.ProviderSessionID; providerSessionID != nil {
			if *providerSessionID != "" {
				if providerSession, ok := t.providerSessionsMap[*providerSessionID]; ok {
					record.SetProviderSession(providerSession)
				} else {
					record.AppendIssue(Issue_DataSource_ProviderSessionID_DoesNotExist)
				}
			} else {
				record.AppendIssue(Issue_DataSource_ProviderSessionID_Empty)
			}
		} // Missing provider session id is valid when disconnected (will be verified below)
	}

	for _, record := range t.tasksMap {
		var providerSessionIssue string
		if providerSessionID, ok := record.Data[dexcom.DataKeyProviderSessionID].(string); ok {
			if providerSessionID != "" {
				if providerSession, ok := t.providerSessionsMap[providerSessionID]; ok {
					record.SetProviderSession(providerSession)
				} else {
					providerSessionIssue = Issue_Task_ProviderSessionID_DoesNotExist
				}
			} else {
				providerSessionIssue = Issue_Task_ProviderSessionID_Empty
			}
		} else {
			providerSessionIssue = Issue_Task_ProviderSessionID_Missing
		}

		var dataSourceIssue string
		if dataSourceID, ok := record.Data[dexcom.DataKeyDataSourceID].(string); ok {
			if dataSourceID != "" {
				if dataSource, ok := t.dataSourcesMap[dataSourceID]; ok {
					record.SetDataSource(dataSource)
				} else {
					dataSourceIssue = Issue_Task_DataSourceID_DoesNotExist
				}
			} else {
				dataSourceIssue = Issue_Task_DataSourceID_Empty
			}
		} else {
			dataSourceIssue = Issue_Task_DataSourceID_Missing
		}

		if record.dataSource == nil && record.providerSession == nil {
			record.AppendIssuef(IssueFormat_Task_ProviderSession_And_DataSource_Invalid, record.State)
		} else {
			record.AppendIssue(providerSessionIssue)
			record.AppendIssue(dataSourceIssue)
		}
	}
}

func (t *Tool) analyzeDataSources() {
	for _, record := range t.dataSourcesMap {
		if err := structureValidator.New(t.Logger()).Validate(record); err != nil {
			for _, err = range errors.ToArray(err) {
				record.AppendIssuef(IssueFormat_DataSource_Invalid, err.Error(), errors.AsSource(err).Pointer())
			}
		}

		switch *record.State {
		case dataSource.StateConnected:
			if record.ProviderSessionID == nil {
				record.AppendIssue(Issue_DataSource_With_State_Connected_ProviderSessionID_Missing)
			}
			if record.Error != nil {
				record.AppendIssue(Issue_DataSource_With_State_Connected_Error_Present)
			}
			if record.providerSession == nil {
				record.AppendIssue(Issue_DataSource_With_State_Connected_ProviderSession_Missing)
			}
			if record.task == nil {
				record.AppendIssue(Issue_DataSource_With_State_Connected_Task_Missing)
			}
		case dataSource.StateDisconnected:
			if record.ProviderSessionID != nil {
				record.AppendIssue(Issue_DataSource_With_State_Disconnected_ProviderSessionID_Present)
			}
			if record.Error != nil {
				record.AppendIssue(Issue_DataSource_With_State_Disconnected_Error_Present)
			}
			if record.providerSession != nil {
				record.AppendIssue(Issue_DataSource_With_State_Disconnected_ProviderSession_Present)
			}
			if record.task != nil {
				record.AppendIssue(Issue_DataSource_With_State_Disconnected_Task_Present)
			}
		case dataSource.StateError:
			if record.ProviderSessionID == nil {
				record.AppendIssue(Issue_DataSource_With_State_Error_ProviderSessionID_Missing)
			}
			if record.Error == nil {
				record.AppendIssue(Issue_DataSource_With_State_Error_Error_Missing)
			}
			if record.providerSession == nil {
				record.AppendIssue(Issue_DataSource_With_State_Error_ProviderSession_Missing)
			}
			if record.task == nil {
				record.AppendIssue(Issue_DataSource_With_State_Error_Task_Missing)
			}
		}

		if record.DataSetIDs != nil {
			if len(*record.DataSetIDs) != 1 {
				record.AppendIssue(Issue_DataSource_With_DataSetIDs_DataSetIDs_Length_Invalid)
			}
			if record.LastImportTime == nil {
				record.AppendIssue(Issue_DataSource_With_DataSetIDs_LastImportTime_Missing)
			}
			if record.LatestDataTime == nil {
				record.AppendIssue(Issue_DataSource_With_DataSetIDs_LatestDataTime_Missing)
			}
			if record.EarliestDataTime == nil {
				record.AppendIssue(Issue_DataSource_With_DataSetIDs_EarliestDataTime_Missing)
			}
		} else {
			if record.LatestDataTime != nil {
				record.AppendIssue(Issue_DataSource_Without_DataSetIDs_LatestDataTime_Present)
			}
			if record.EarliestDataTime != nil {
				record.AppendIssue(Issue_DataSource_Without_DataSetIDs_EarliestDataTime_Present)
			}
		}
	}
}

func (t *Tool) analyzeProviderSessions() {
	for _, record := range t.providerSessionsMap {
		if err := structureValidator.New(t.Logger()).Validate(record); err != nil {
			for _, err = range errors.ToArray(err) {
				record.AppendIssuef(IssueFormat_ProviderSession_Invalid, err.Error(), errors.AsSource(err).Pointer())
			}
		}

		if record.task == nil {
			record.AppendIssue(Issue_ProviderSession_Task_Missing)
		}
	}
}

func (t *Tool) analyzeTasks() {
	for _, record := range t.tasksMap {
		if err := structureValidator.New(t.Logger()).Validate(record); err != nil {
			for _, err = range errors.ToArray(err) {
				record.AppendIssuef(IssueFormat_Task_Invalid, err.Error(), errors.AsSource(err).Pointer())
			}
		}

		if record.Name != nil {
			if providerSessionID, ok := record.Data[dexcom.DataKeyProviderSessionID].(string); ok {
				if *record.Name != fmt.Sprintf("%s:%s", dexcomFetch.Type, providerSessionID) {
					record.AppendIssue(Issue_Task_Name_Invalid)
				}
			}
		} else {
			record.AppendIssue(Issue_Task_Name_Missing)
		}

		if record.Data != nil {
			if deviceHashesRaw, ok := record.Data[dexcom.DataKeyDeviceHashes]; ok {
				if deviceHashesRawMap, ok := deviceHashesRaw.(map[string]interface{}); ok {
					deviceHashes := map[string]string{}
					for key, value := range deviceHashesRawMap {
						if valueString, valueStringOK := value.(string); valueStringOK {
							deviceHashes[key] = valueString
						} else {
							record.AppendIssue(Issue_Task_DeviceHash_Invalid)
						}
					}
				} else {
					record.AppendIssue(Issue_Task_DeviceHashes_Invalid)
				}

				if record.dataSource != nil {
					if len(*record.dataSource.DataSetIDs) == 0 {
						record.AppendIssue(Issue_Task_With_DeviceHashes_And_DataSource_DataSetIDs_Missing)
					}
					if record.dataSource.LastImportTime == nil {
						record.AppendIssue(Issue_Task_With_DeviceHashes_And_DataSource_LastImportTime_Missing)
					} else if record.dataSource.LatestDataTime == nil {
						record.AppendIssue(Issue_Task_With_DeviceHashes_And_DataSource_LatestDataTime_Missing)
					} else if record.dataSource.EarliestDataTime == nil {
						record.AppendIssue(Issue_Task_With_DeviceHashes_And_DataSource_EarliestDataTime_Missing)
					}
				}
			}
		} else {
			record.AppendIssue(Issue_Task_Data_Missing)
		}

		switch record.State {
		case task.TaskStatePending:
			if record.DeadlineTime != nil {
				record.AppendIssue(Issue_Task_With_State_Pending_DeadlineTime_Present)
			}
			if record.ExpirationTime != nil {
				record.AppendIssue(Issue_Task_With_State_Pending_ExpirationTime_Present)
			}
		case task.TaskStateRunning:
			if record.AvailableTime != nil {
				record.AppendIssue(Issue_Task_With_State_Running_AvailableTime_Present)
			}
			if record.DeadlineTime == nil {
				record.AppendIssue(Issue_Task_With_State_Running_DeadlineTime_Missing)
			}
			if record.Error != nil {
				record.AppendIssue(Issue_Task_With_State_Running_Error_Present)
			}
			if record.ExpirationTime != nil {
				record.AppendIssue(Issue_Task_With_State_Running_ExpirationTime_Present)
			}
		case task.TaskStateFailed:
			if record.AvailableTime != nil {
				record.AppendIssue(Issue_Task_With_State_Failed_AvailableTime_Present)
			}
			if record.DeadlineTime != nil {
				record.AppendIssue(Issue_Task_With_State_Failed_DeadlineTime_Present)
			}
			if record.Error == nil {
				record.AppendIssue(Issue_Task_With_State_Failed_Error_Missing)
			}
			if record.ExpirationTime != nil {
				record.AppendIssue(Issue_Task_With_State_Failed_ExpirationTime_Present)
			}
		case task.TaskStateCompleted:
			record.AppendIssuef(IssueFormat_Task_State_Invalid, record.State)
		}

		if record.providerSession == nil {
			record.AppendIssue(Issue_Task_ProviderSession_Missing)
		}
	}
}

func (t *Tool) analyzeIssues() {
	issueMarshalableMap := IssueMarshalableMap{}
	for _, record := range t.providerSessionsMap {
		issueMarshalableMap.AppendIssueReporterAsMarshalable(record)
	}
	for _, record := range t.dataSourcesMap {
		issueMarshalableMap.AppendIssueReporterAsMarshalable(record)
	}
	for _, record := range t.tasksMap {
		issueMarshalableMap.AppendIssueReporterAsMarshalable(record)
	}
	for _, record := range t.usersMap {
		issueMarshalableMap.AppendIssueReporterAsMarshalable(record)
	}

	t.outputIssues(issueMarshalableMap)
}

func (t *Tool) outputIssues(issueMarshalableMap IssueMarshalableMap) {
	t.outputIssuesHeader(issueMarshalableMap)

	issues := maps.Keys(issueMarshalableMap)
	sort.StringSlice(issues).Sort()
	for _, issue := range issues {
		t.outputIssue(issue, issueMarshalableMap[issue], issueMarshalableMap)
	}

	t.outputIssuesFooter()
}

func (t *Tool) outputIssue(issue string, marshalables Marshalables, issueMarshalableMap IssueMarshalableMap) {
	t.outputIssueHeader(issue, marshalables)

	switch issue {
	case Issue_DataSource_ProviderSessionID_DoesNotExist:
		t.outputResolutionHeader("Examine each to determine why it does not exist.")
		t.outputMongoReadOperationsHeader()
		t.outputMongoDataSourcesAggregation(marshalables.DataSources().IDs())
		t.outputMongoWriteOperationsHeader()
		t.outputMongoOperationf("db.data_sources.updateMany({id: {$in: [%s]}}, {$set: {state: 'disconnected'}, $unset: {providerSessionId: true}})", mongoIDs(marshalables.DataSources().IDs()))
		return
	case Issue_DataSource_ProviderSessionID_Empty:
	case Issue_DataSource_UserID_Empty:
	case Issue_DataSource_UserID_Missing:
	case Issue_DataSource_With_DataSetIDs_DataSetIDs_Length_Invalid:
	case Issue_DataSource_With_DataSetIDs_EarliestDataTime_Missing, Issue_DataSource_With_DataSetIDs_LastImportTime_Missing, Issue_DataSource_With_DataSetIDs_LatestDataTime_Missing:
		t.outputResolutionHeader("FIXED with BACK-3113. Will keep occurring until deployed. Will need to manually fix.")
		// t.outputMongoReadOperationsHeader()

		// var earliestDataTimeIDs IDs
		// if otherMarshable, ok := issueMarshalableMap[Issue_DataSource_With_DataSetIDs_EarliestDataTime_Missing]; ok {
		// 	earliestDataTimeIDs = otherMarshable.DataSources().IDs()
		// }

		// var lastImportTimeIDs IDs
		// if otherMarshable, ok := issueMarshalableMap[Issue_DataSource_With_DataSetIDs_LastImportTime_Missing]; ok {
		// 	lastImportTimeIDs = otherMarshable.DataSources().IDs()
		// }

		// var latestDataTimeIDs IDs
		// if otherMarshable, ok := issueMarshalableMap[Issue_DataSource_With_DataSetIDs_LatestDataTime_Missing]; ok {
		// 	latestDataTimeIDs = otherMarshable.DataSources().IDs()
		// }

		// switch issue {
		// case Issue_DataSource_With_DataSetIDs_EarliestDataTime_Missing:
		// 	t.outputDescription("ALL:")
		// 	t.outputMongoDataSourcesAggregation(earliestDataTimeIDs)
		// 	t.outputDescription("AND latest data time PRESENT:")
		// 	t.outputMongoDataSourcesAggregation(earliestDataTimeIDs.Subtract(latestDataTimeIDs))
		// 	t.outputDescription("DATASET IDS:")
		// 	t.outputMongoOperationf("ids = db.data_sources.aggregate([{$match: {id: {$in: [%s]}}}, {$project: {dataSetIds: 1}}, {$unwind: '$dataSetIds'}, {$group: {_id: 'dataSetIds', dataSetIds: {$push: '$dataSetIds'}}}]).toArray()[0].dataSetIds", mongoIDs(earliestDataTimeIDs))
		// 	t.outputDescription("DATASETS:")
		// 	t.outputMongoOperation("use data")
		// 	t.outputMongoOperation("db.deviceDataSets.find({uploadId: {$in: ids}})")
		// 	t.outputDescription("DATA TYPE COUNT PER DATASET:")
		// 	t.outputMongoOperation("use data")
		// 	t.outputMongoOperation("db.deviceData.aggregate([{$match: {uploadId: {$in: ids}}}, {$group: {_id: {uploadId: '$uploadId', type: '$type'}, count: {$sum: 1}}}, {$group: {_id: '$_id.uploadId', types: {$addToSet: {type: '$_id.type', count: {$sum: '$count'}}}}}])")
		// 	t.outputDescription("DATA TYPE COUNT PER DATASET WITHOUT CGMSETTINGS:")
		// 	t.outputMongoOperation("use data")
		// 	t.outputMongoOperation("db.deviceData.aggregate([{$match: {uploadId: {$in: ids}, type: {$ne: 'cgmSettings'}}}, {$group: {_id: {uploadId: '$uploadId', type: '$type'}, count: {$sum: 1}}}, {$group: {_id: '$_id.uploadId', types: {$addToSet: {type: '$_id.type', count: {$sum: '$count'}}}}}])")
		// case Issue_DataSource_With_DataSetIDs_LastImportTime_Missing:
		// 	t.outputDescription("ALL:")
		// 	t.outputMongoDataSourcesAggregation(lastImportTimeIDs)
		// 	t.outputDescription("AND earliest data time OR latest data time PRESENT")
		// 	t.outputMongoDataSourcesAggregation(lastImportTimeIDs.Subtract(earliestDataTimeIDs, latestDataTimeIDs))
		// 	t.outputDescription("DATASET IDS:")
		// 	t.outputMongoOperationf("ids = db.data_sources.aggregate([{$match: {id: {$in: [%s]}}}, {$project: {dataSetIds: 1}}, {$unwind: '$dataSetIds'}, {$group: {_id: 'dataSetIds', dataSetIds: {$push: '$dataSetIds'}}}]).toArray()[0].dataSetIds", mongoIDs(lastImportTimeIDs))
		// 	t.outputDescription("DATASETS:")
		// 	t.outputMongoOperation("use data")
		// 	t.outputMongoOperation("db.deviceDataSets.find({uploadId: {$in: ids}})")
		// 	t.outputDescription("DATA TYPE COUNT PER DATASET:")
		// 	t.outputMongoOperation("use data")
		// 	t.outputMongoOperation("db.deviceData.aggregate([{$match: {uploadId: {$in: ids}}}, {$group: {_id: {uploadId: '$uploadId', type: '$type'}, count: {$sum: 1}}}, {$group: {_id: '$_id.uploadId', types: {$addToSet: {type: '$_id.type', count: {$sum: '$count'}}}}}])")
		// 	t.outputDescription("DATA TYPE COUNT PER DATASET WITHOUT CGMSETTINGS:")
		// 	t.outputMongoOperation("use data")
		// 	t.outputMongoOperation("db.deviceData.aggregate([{$match: {uploadId: {$in: ids}, type: {$ne: 'cgmSettings'}}}, {$group: {_id: {uploadId: '$uploadId', type: '$type'}, count: {$sum: 1}}}, {$group: {_id: '$_id.uploadId', types: {$addToSet: {type: '$_id.type', count: {$sum: '$count'}}}}}])")
		// case Issue_DataSource_With_DataSetIDs_LatestDataTime_Missing:
		// 	t.outputDescription("ALL:")
		// 	t.outputMongoDataSourcesAggregation(latestDataTimeIDs)
		// 	t.outputDescription("AND last import time PRESENT:")
		// 	t.outputMongoDataSourcesAggregation(latestDataTimeIDs.Subtract(lastImportTimeIDs))
		// 	t.outputDescription("DATASET IDS:")
		// 	t.outputMongoOperationf("ids = db.data_sources.aggregate([{$match: {id: {$in: [%s]}}}, {$project: {dataSetIds: 1}}, {$unwind: '$dataSetIds'}, {$group: {_id: 'dataSetIds', dataSetIds: {$push: '$dataSetIds'}}}]).toArray()[0].dataSetIds", mongoIDs(latestDataTimeIDs))
		// 	t.outputDescription("DATASETS:")
		// 	t.outputMongoOperation("use data")
		// 	t.outputMongoOperation("db.deviceDataSets.find({uploadId: {$in: ids}})")
		// 	t.outputDescription("DATA TYPE COUNT PER DATASET:")
		// 	t.outputMongoOperation("use data")
		// 	t.outputMongoOperation("db.deviceData.aggregate([{$match: {uploadId: {$in: ids}}}, {$group: {_id: {uploadId: '$uploadId', type: '$type'}, count: {$sum: 1}}}, {$group: {_id: '$_id.uploadId', types: {$addToSet: {type: '$_id.type', count: {$sum: '$count'}}}}}])")
		// 	t.outputDescription("DATA TYPE COUNT PER DATASET WITHOUT CGMSETTINGS:")
		// 	t.outputMongoOperation("use data")
		// 	t.outputMongoOperation("db.deviceData.aggregate([{$match: {uploadId: {$in: ids}, type: {$ne: 'cgmSettings'}}}, {$group: {_id: {uploadId: '$uploadId', type: '$type'}, count: {$sum: 1}}}, {$group: {_id: '$_id.uploadId', types: {$addToSet: {type: '$_id.type', count: {$sum: '$count'}}}}}])")
		// }
		return
	case Issue_DataSource_With_State_Connected_Error_Present:
	case Issue_DataSource_With_State_Connected_ProviderSession_Missing:
	case Issue_DataSource_With_State_Connected_ProviderSessionID_Missing:
		t.outputResolutionHeader("FIXED with BACK-3118. Will keep occurring until deployed. Will need to manually remove providerSessionId field from all data sources.")
		// t.outputMongoReadOperationsHeader()
		// t.outputMongoDataSourcesAggregation(marshalables.DataSources().IDs())
		return
	case Issue_DataSource_With_State_Connected_Task_Missing:
	case Issue_DataSource_With_State_Disconnected_Error_Present:
	case Issue_DataSource_With_State_Disconnected_ProviderSession_Present:
	case Issue_DataSource_With_State_Disconnected_ProviderSessionID_Present:
	case Issue_DataSource_With_State_Disconnected_Task_Present:
	case Issue_DataSource_With_State_Error_Error_Missing:
	case Issue_DataSource_With_State_Error_ProviderSession_Missing:
	case Issue_DataSource_With_State_Error_ProviderSessionID_Missing:
		t.outputResolutionHeader("FIXED with BACK-3118. Will keep occurring until deployed. Will need to manually remove providerSessionId field from all data sources.")
		// t.outputMongoReadOperationsHeader()
		// t.outputMongoDataSourcesAggregation(marshalables.DataSources().IDs())
		return
	case Issue_DataSource_With_State_Error_Task_Missing:
	case Issue_DataSource_Without_DataSetIDs_EarliestDataTime_Present:
	case Issue_DataSource_Without_DataSetIDs_LatestDataTime_Present:
	case Issue_ProviderSession_Task_Missing:
	case Issue_ProviderSession_UserID_Empty:
	case Issue_Task_Data_Missing:
	case Issue_Task_DataSourceID_DoesNotExist:
	case Issue_Task_DataSourceID_Empty:
	case Issue_Task_DataSourceID_Missing:
	case Issue_Task_DeviceHash_Invalid:
	case Issue_Task_DeviceHashes_Invalid:
	case Issue_Task_Name_Invalid:
	case Issue_Task_Name_Missing:
	case Issue_Task_ProviderSession_Missing:
		t.outputResolutionHeader("Examine each to determine why it is missing. Consider deleting if orphaned from data source and provider session.")
		t.outputMongoReadOperationsHeader()
		t.outputMongoTasksAggregation(marshalables.Tasks().IDs())
		t.outputMongoWriteOperationsHeader()
		t.outputMongoOperationf("db.tasks.deleteMany({id: {$in: [%s]}})", mongoIDs(marshalables.Tasks().IDs()))
		return
	case Issue_Task_ProviderSessionID_DoesNotExist:
		t.outputResolutionHeader("Examine each to determine why it does not exist. Consider deleting if orphaned from data source and provider session.")
		t.outputMongoReadOperationsHeader()
		t.outputMongoTasksAggregation(marshalables.Tasks().IDs())
		t.outputMongoWriteOperationsHeader()
		t.outputMongoOperationf("db.tasks.deleteMany({id: {$in: [%s]}})", mongoIDs(marshalables.Tasks().IDs()))
		return
	case Issue_Task_ProviderSessionID_Empty:
	case Issue_Task_ProviderSessionID_Missing:
	case Issue_Task_With_DeviceHashes_And_DataSource_DataSetIDs_Missing:
		t.outputResolutionHeader("Examine each to determine why it is missing")
		t.outputMongoReadOperationsHeader()
		t.outputMongoTasksAggregation(marshalables.Tasks().IDs())
		return
	case Issue_Task_With_DeviceHashes_And_DataSource_EarliestDataTime_Missing:
		t.outputResolutionHeader("FIXED with BACK-3113. Will keep occurring until deployed. May need to manually update failed tasks post-deploy.")
		t.outputMongoReadOperationsHeader()
		t.outputMongoTasksAggregation(marshalables.Tasks().IDs())
		t.outputMongoOperationf("db.tasks.aggregate([{$match: {id: {$in: [%s]}}}, {$lookup: {from: 'data_sources', localField: 'data.dataSourceId', foreignField: 'id', as: 'dataSourcesFromTasks'}}, {$addFields: {dataSetIds: '$dataSourcesFromTasks.dataSetIds'}}, {$project: {_id: 0, dataSetIds: 1}}, {$unwind: '$dataSetIds'}, {$unwind: '$dataSetIds'}, {$group: {_id: 'dataSetIds', dataSetIds: {$push: '$dataSetIds'}}}])", mongoIDs(marshalables.Tasks().IDs()))
		return
	case Issue_Task_With_DeviceHashes_And_DataSource_LastImportTime_Missing:
		t.outputResolutionHeader("FIXED with BACK-3113. Will keep occurring until deployed. May need to manually update failed tasks post-deploy.")
		// t.outputMongoReadOperationsHeader()
		// t.outputMongoTasksAggregation(marshalables.Tasks().IDs())
		// t.outputMongoOperationf("db.tasks.aggregate([{$match: {id: {$in: [%s]}}}, {$lookup: {from: 'data_sources', localField: 'data.dataSourceId', foreignField: 'id', as: 'dataSourcesFromTasks'}}, {$addFields: {dataSetIds: '$dataSourcesFromTasks.dataSetIds'}}, {$project: {_id: 0, dataSetIds: 1}}, {$unwind: '$dataSetIds'}, {$unwind: '$dataSetIds'}, {$group: {_id: 'dataSetIds', dataSetIds: {$push: '$dataSetIds'}}}])", mongoIDs(marshalables.Tasks().IDs()))
		return
	case Issue_Task_With_DeviceHashes_And_DataSource_LatestDataTime_Missing:
		t.outputResolutionHeader("FIXED with BACK-3113. Will keep occurring until deployed. May need to manually update failed tasks post-deploy.")
		// t.outputMongoReadOperationsHeader()
		// t.outputMongoTasksAggregation(marshalables.Tasks().IDs())
		// t.outputMongoOperationf("db.tasks.aggregate([{$match: {id: {$in: [%s]}}}, {$lookup: {from: 'data_sources', localField: 'data.dataSourceId', foreignField: 'id', as: 'dataSourcesFromTasks'}}, {$addFields: {dataSetIds: '$dataSourcesFromTasks.dataSetIds'}}, {$project: {_id: 0, dataSetIds: 1}}, {$unwind: '$dataSetIds'}, {$unwind: '$dataSetIds'}, {$group: {_id: 'dataSetIds', dataSetIds: {$push: '$dataSetIds'}}}])", mongoIDs(marshalables.Tasks().IDs()))
		return
	case Issue_Task_With_State_Failed_AvailableTime_Present:
		t.outputResolutionHeader("FIXED with BACK-3116. Will keep occurring until deployed. Will need to manually update failed tasks post-deploy.")
		t.outputMongoReadOperationsHeader()
		t.outputMongoOperationf("db.tasks.countDocuments({state: 'failed', availableTime: {$exists: true}})")
		t.outputMongoWriteOperationsHeader()
		t.outputMongoOperationf("db.tasks.updateMany({state: 'failed', availableTime: {$exists: true}}, {$unset: {availableTime: true}})")
		return
	case Issue_Task_With_State_Failed_DeadlineTime_Present:
		t.outputResolutionHeader("FIXED with BACK-3116. Will keep occurring until deployed. Will need to manually update failed tasks post-deploy.")
		t.outputMongoReadOperationsHeader()
		t.outputMongoOperationf("db.tasks.countDocuments({state: 'failed', deadlineTime: {$exists: true}})")
		t.outputMongoWriteOperationsHeader()
		t.outputMongoOperationf("db.tasks.updateMany({state: 'failed', deadlineTime: {$exists: true}}, {$unset: {deadlineTime: true}})")
		return
	case Issue_Task_With_State_Failed_Error_Missing:
	case Issue_Task_With_State_Failed_ExpirationTime_Present:
	case Issue_Task_With_State_Pending_DeadlineTime_Present:
	case Issue_Task_With_State_Pending_ExpirationTime_Present:
	case Issue_Task_With_State_Running_AvailableTime_Present:
		t.outputResolutionHeader("FIXED with BACK-3116. Will keep occurring until deployed. Will need to manually update failed tasks post-deploy.")
		// t.outputMongoReadOperationsHeader()
		// t.outputMongoTasksAggregation(marshalables.Tasks().IDs())
		return
	case Issue_Task_With_State_Running_DeadlineTime_Missing:
		t.outputResolutionHeader("Examine each to determine why it is missing.")
		t.outputMongoReadOperationsHeader()
		t.outputMongoTasksAggregation(marshalables.Tasks().IDs())
		t.outputMongoOperationf("db.tasks.find({id: {$in: [%s]}})", mongoIDs(marshalables.Tasks().IDs()))
		t.outputMongoWriteOperationsHeader()
		t.outputMongoOperationf("db.tasks.updateMany({id: {$in: [%s]}}, {$set: {deadlineTime: ISODate('%s')}})", mongoIDs(marshalables.Tasks().IDs()), time.Now().Format(time.RFC3339))
		return
	case Issue_Task_With_State_Running_Error_Present:
		t.outputResolutionHeader("Examine each to determine why it is present.")
		t.outputMongoReadOperationsHeader()
		t.outputMongoTasksAggregation(marshalables.Tasks().IDs())
		return
	case Issue_Task_With_State_Running_ExpirationTime_Present:
	default:
		t.analyzeIssueFormat(issue, marshalables)
		return
	}
	t.outputResolutionHeader("TBD")
}

func (t *Tool) analyzeIssueFormat(issueFormat string, marshalables Marshalables) {
	for _, issueFormatRE := range t.issueFormatREs {
		if matches := issueFormatRE.RE.FindStringSubmatch(issueFormat); matches != nil {
			switch issueFormatRE.IssueFormat {
			case IssueFormat_DataSource_Invalid:
				if matches[1] == "value does not exist" && matches[2] == "/revision" {
					t.outputResolutionHeader("Update data sources to contain revision.")
					t.outputMongoReadOperationsHeader()
					t.outputMongoOperationf("db.data_sources.find({id: {$in: [%s]}}).sort({modifiedTime: 1})", mongoIDs(marshalables.DataSources().IDs()))
					t.outputMongoWriteOperationsHeader()
					t.outputMongoOperationf("db.data_sources.updateMany({id: {$in: [%s]}}, {$set: {revision: 0}})", mongoIDs(marshalables.DataSources().IDs()))
					return
				}
			case IssueFormat_ProviderSession_Invalid:
			case IssueFormat_Task_Invalid:
			case IssueFormat_DataSource_ProviderSession_Mismatch:
				t.outputResolutionHeader("Examine each mismatch to determine which is valid and consider deleting invalid.")
				t.outputMongoReadOperationsHeader()
				t.outputMongoDataSourcesAggregation(marshalables.DataSources().IDs())
				t.outputMongoProviderSessionsAggregation(IDs{matches[1], matches[2]})
				return
			case IssueFormat_DataSource_Task_Mismatch:
				t.outputResolutionHeader("Examine each mismatch to determine which is valid and consider deleting invalid.")
				t.outputMongoReadOperationsHeader()
				t.outputMongoDataSourcesAggregation(marshalables.DataSources().IDs())
				t.outputMongoTasksAggregation(IDs{matches[1], matches[2]})
				return
			case IssueFormat_DataSource_User_Mismatch:
				t.outputResolutionHeader("Examine each mismatch to determine which is valid and consider deleting invalid.")
				t.outputMongoReadOperationsHeader()
				t.outputMongoDataSourcesAggregation(marshalables.DataSources().IDs())
				return
			case IssueFormat_ProviderSession_DataSource_Mismatch:
				t.outputResolutionHeader("Examine each mismatch to determine which is valid and consider deleting invalid.")
				t.outputMongoReadOperationsHeader()
				t.outputMongoProviderSessionsAggregation(marshalables.ProviderSessions().IDs())
				t.outputMongoDataSourcesAggregation(IDs{matches[1], matches[2]})
				return
			case IssueFormat_ProviderSession_Task_Mismatch:
				t.outputResolutionHeader("Examine each mismatch to determine which is valid and consider deleting invalid.")
				t.outputMongoReadOperationsHeader()
				t.outputMongoProviderSessionsAggregation(marshalables.ProviderSessions().IDs())
				t.outputMongoTasksAggregation(IDs{matches[1], matches[2]})
				return
			case IssueFormat_ProviderSession_User_Mismatch:
				t.outputResolutionHeader("Examine each mismatch to determine which is valid and consider deleting invalid.")
				t.outputMongoReadOperationsHeader()
				t.outputMongoProviderSessionsAggregation(marshalables.ProviderSessions().IDs())
				return
			case IssueFormat_Task_DataSource_Mismatch:
				t.outputResolutionHeader("Examine each mismatch to determine which is valid and consider deleting invalid.")
				t.outputMongoReadOperationsHeader()
				t.outputMongoTasksAggregation(marshalables.Tasks().IDs())
				t.outputMongoDataSourcesAggregation(IDs{matches[1], matches[2]})
				return
			case IssueFormat_Task_ProviderSession_Mismatch:
				t.outputResolutionHeader("Examine each mismatch to determine which is valid and consider deleting invalid.")
				t.outputMongoReadOperationsHeader()
				t.outputMongoTasksAggregation(marshalables.Tasks().IDs())
				t.outputMongoProviderSessionsAggregation(IDs{matches[1], matches[2]})
				return
			case IssueFormat_Task_User_Mismatch:
				t.outputResolutionHeader("Examine each mismatch to determine which is valid and consider deleting invalid.")
				t.outputMongoReadOperationsHeader()
				t.outputMongoTasksAggregation(marshalables.Tasks().IDs())
				return
			case IssueFormat_User_DataSource_Mismatch:
				t.outputResolutionHeader("Examine each mismatch to determine which is valid and consider deleting invalid.")
				t.outputMongoReadOperationsHeader()
				t.outputMongoDataSourcesAggregation(IDs{matches[1], matches[2]})
				return
			case IssueFormat_User_ProviderSession_Mismatch:
				t.outputResolutionHeader("Examine each mismatch to determine which is valid and consider deleting invalid.")
				t.outputMongoReadOperationsHeader()
				t.outputMongoProviderSessionsAggregation(IDs{matches[1], matches[2]})
				return
			case IssueFormat_User_Task_Mismatch:
				t.outputResolutionHeader("Examine each mismatch to determine which is valid and consider deleting invalid.")
				t.outputMongoReadOperationsHeader()
				t.outputMongoTasksAggregation(IDs{matches[1], matches[2]})
				return
			case IssueFormat_Task_ProviderSession_And_DataSource_Invalid:
			case IssueFormat_Task_State_Invalid:
			}
			t.outputResolutionHeader("TBD")
			return
		}
	}
	t.outputResolutionHeader("UNKNOWN ISSUE")
}

func (t *Tool) outputIssuesHeader(issueMarshalableMap IssueMarshalableMap) {
	fmt.Fprintf(t.output, "\nTotal issues: %d\n", len(issueMarshalableMap))
}

func (t *Tool) outputIssuesFooter() {
	fmt.Fprintln(t.output)
}

func (t *Tool) outputIssueHeader(issue string, marshalables Marshalables) {
	fmt.Fprintf(t.output, "\n%sISSUE: '%s' (%d)%s\n", magenta, issue, len(marshalables), clear)
}

func (t *Tool) outputResolutionHeader(resolution string) {
	fmt.Fprintf(t.output, "  %sRESOLUTION: %s%s\n", yellow, resolution, clear)
}

func (t *Tool) outputMongoReadOperationsHeader() {
	fmt.Fprintf(t.output, "    %sMongo read operations:%s\n", green, clear)
}

func (t *Tool) outputMongoWriteOperationsHeader() {
	fmt.Fprintf(t.output, "    %sMongo WRITE operations (HERE BE DRAGONS!!!):%s\n", red, clear)
}

func (t *Tool) outputDescription(description string) {
	fmt.Fprintf(t.output, "\n    %s%s%s\n", white, description, clear)
}

func (t *Tool) outputMongoOperation(mongo string) {
	fmt.Fprintf(t.output, "      %s\n", mongo)
}

func (t *Tool) outputMongoOperationf(format string, a ...any) {
	t.outputMongoOperation(fmt.Sprintf(format, a...))
}

func (t *Tool) outputMongoDataSourcesAggregation(ids IDs) {
	t.outputMongoOperationf("db.data_sources.aggregate([{$match: {id: {$in: [%s]}}}, {$lookup: {from: 'tasks', localField: 'id', foreignField: 'data.dataSourceId', as: 'tasks'}}, {$lookup: {from: 'provider_sessions', localField: 'tasks.data.providerSessionId', foreignField: 'id', as: 'providerSessionsFromTasks'}}, {$sort: {modifiedTime: 1}}])", mongoIDs(ids))
}

func (t *Tool) outputMongoProviderSessionsAggregation(ids IDs) {
	t.outputMongoOperationf("db.provider_sessions.aggregate([{$match: {id: {$in: [%s]}}}, {$lookup: {from: 'tasks', localField: 'id', foreignField: 'data.providerSessionId', as: 'tasks'}}, {$lookup: {from: 'data_sources', localField: 'tasks.data.dataSourceId', foreignField: 'id', as: 'dataSourcesFromTasks'}}, {$sort: {modifiedTime: 1}}])", mongoIDs(ids))
}

func (t *Tool) outputMongoTasksAggregation(ids IDs) {
	t.outputMongoOperationf("db.tasks.aggregate([{$match: {id: {$in: [%s]}}}, {$lookup: {from: 'data_sources', localField: 'data.dataSourceId', foreignField: 'id', as: 'dataSourcesFromTasks'}}, {$lookup: {from: 'provider_sessions', localField: 'data.providerSessionId', foreignField: 'id', as: 'providerSessionsFromTasks'}}, {$sort: {modifiedTime: 1}}])", mongoIDs(ids))
}

func mongoIDs(ids IDs) string {
	return strings.Join(arrayMap(ids, strconv.Quote), ", ")
}

func loadFile[T any](filename string) ([]*T, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, errors.New("unable to open file")
	}
	defer file.Close()

	data := []*T{}
	if err = json.NewDecoder(file).Decode(&data); err != nil {
		return nil, errors.New("unable to decode file")
	}
	return data, nil
}

func arrayMap[T any](ts []T, f func(T) T) []T {
	var fs []T
	for _, t := range ts {
		fs = append(fs, f(t))
	}
	return fs
}

const (
	clear   = "\033[m"
	black   = "\033[30m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	white   = "\033[37m"
)
