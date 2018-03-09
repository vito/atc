package metric

import (
	"errors"
	"strconv"
	"time"

	"code.cloudfoundry.org/lager"
	"github.com/concourse/atc/db"
)

var Databases []db.Conn
var DatabaseQueries = Meter(0)

var ContainersCreated = Meter(0)
var VolumesCreated = Meter(0)

var FailedContainers = Meter(0)
var FailedVolumes = Meter(0)

var ContainersDeleted = Meter(0)
var VolumesDeleted = Meter(0)

type SchedulingFullDuration struct {
	PipelineName string
	Duration     time.Duration
}

func (event SchedulingFullDuration) Emit(logger lager.Logger) {
	state := EventStateOK

	if event.Duration > time.Second {
		state = EventStateWarning
	}

	if event.Duration > 5*time.Second {
		state = EventStateCritical
	}

	emit(
		logger.Session("full-scheduling-duration"),
		Event{
			Name:  "scheduling: full duration (ms)",
			Value: ms(event.Duration),
			State: state,
			Attributes: map[string]string{
				"pipeline": event.PipelineName,
			},
		},
	)
}

type SchedulingLoadVersionsDuration struct {
	PipelineName string
	Duration     time.Duration
}

func (event SchedulingLoadVersionsDuration) Emit(logger lager.Logger) {
	state := EventStateOK

	if event.Duration > time.Second {
		state = EventStateWarning
	}

	if event.Duration > 5*time.Second {
		state = EventStateCritical
	}

	emit(
		logger.Session("loading-versions-duration"),
		Event{
			Name:  "scheduling: loading versions duration (ms)",
			Value: ms(event.Duration),
			State: state,
			Attributes: map[string]string{
				"pipeline": event.PipelineName,
			},
		},
	)
}

type SchedulingJobDuration struct {
	PipelineName string
	JobName      string
	Duration     time.Duration
}

func (event SchedulingJobDuration) Emit(logger lager.Logger) {
	state := EventStateOK

	if event.Duration > time.Second {
		state = EventStateWarning
	}

	if event.Duration > 5*time.Second {
		state = EventStateCritical
	}

	emit(
		logger.Session("job-scheduling-duration"),
		Event{
			Name:  "scheduling: job duration (ms)",
			Value: ms(event.Duration),
			State: state,
			Attributes: map[string]string{
				"pipeline": event.PipelineName,
				"job":      event.JobName,
			},
		},
	)
}

type WorkerContainers struct {
	WorkerName string
	Containers int
}

func (event WorkerContainers) Emit(logger lager.Logger) {
	emit(
		logger.Session("worker-containers"),
		Event{
			Name:  "worker containers",
			Value: event.Containers,
			State: EventStateOK,
			Attributes: map[string]string{
				"worker": event.WorkerName,
			},
		},
	)
}

type WorkerVolumes struct {
	WorkerName string
	Volumes    int
}

func (event WorkerVolumes) Emit(logger lager.Logger) {
	emit(
		logger.Session("worker-volumes"),
		Event{
			Name:  "worker volumes",
			Value: event.Volumes,
			State: EventStateOK,
			Attributes: map[string]string{
				"worker": event.WorkerName,
			},
		},
	)
}

type CreatingContainersToBeGarbageCollected struct {
	Containers int
}

func (event CreatingContainersToBeGarbageCollected) Emit(logger lager.Logger) {
	emit(
		logger.Session("gc-found-creating-containers-for-deletion"),
		Event{
			Name:       "creating containers to be garbage collected",
			Value:      event.Containers,
			State:      EventStateOK,
			Attributes: map[string]string{},
		},
	)
}

type CreatedContainersToBeGarbageCollected struct {
	Containers int
}

func (event CreatedContainersToBeGarbageCollected) Emit(logger lager.Logger) {
	emit(
		logger.Session("gc-found-created-ccontainers-for-deletion"),
		Event{
			Name:       "created containers to be garbage collected",
			Value:      event.Containers,
			State:      EventStateOK,
			Attributes: map[string]string{},
		},
	)
}

type DestroyingContainersToBeGarbageCollected struct {
	Containers int
}

func (event DestroyingContainersToBeGarbageCollected) Emit(logger lager.Logger) {
	emit(
		logger.Session("gc-found-destroying-containers-for-deletion"),
		Event{
			Name:       "destroying containers to be garbage collected",
			Value:      event.Containers,
			State:      EventStateOK,
			Attributes: map[string]string{},
		},
	)
}

type FailedContainersToBeGarbageCollected struct {
	Containers int
}

func (event FailedContainersToBeGarbageCollected) Emit(logger lager.Logger) {
	emit(
		logger.Session("gc-found-failed-containers-for-deletion"),
		Event{
			Name:       "failed containers to be garbage collected",
			Value:      event.Containers,
			State:      EventStateOK,
			Attributes: map[string]string{},
		},
	)
}

type CreatedVolumesToBeGarbageCollected struct {
	Volumes int
}

func (event CreatedVolumesToBeGarbageCollected) Emit(logger lager.Logger) {
	emit(
		logger.Session("gc-found-created-volumes-for-deletion"),
		Event{
			Name:       "created volumes to be garbage collected",
			Value:      event.Volumes,
			State:      EventStateOK,
			Attributes: map[string]string{},
		},
	)
}

type DestroyingVolumesToBeGarbageCollected struct {
	Volumes int
}

func (event DestroyingVolumesToBeGarbageCollected) Emit(logger lager.Logger) {
	emit(
		logger.Session("gc-found-destroying-volumes-for-deletion"),
		Event{
			Name:       "destroying volumes to be garbage collected",
			Value:      event.Volumes,
			State:      EventStateOK,
			Attributes: map[string]string{},
		},
	)
}

type FailedVolumesToBeGarbageCollected struct {
	Volumes int
}

func (event FailedVolumesToBeGarbageCollected) Emit(logger lager.Logger) {
	emit(
		logger.Session("gc-found-failed-volumes-for-deletion"),
		Event{
			Name:       "failed volumes to be garbage collected",
			Value:      event.Volumes,
			State:      EventStateOK,
			Attributes: map[string]string{},
		},
	)
}

type GarbageCollectionContainerCollectorJobDropped struct {
	WorkerName string
}

func (event GarbageCollectionContainerCollectorJobDropped) Emit(logger lager.Logger) {
	emit(
		logger.Session("gc-container-collector-dropped"),
		Event{
			Name:  "GC container collector job dropped",
			Value: 1,
			State: EventStateOK,
			Attributes: map[string]string{
				"worker": event.WorkerName,
			},
		},
	)
}

type GarbageCollectionVolumeCollectorJobDropped struct {
	WorkerName string
}

func (event GarbageCollectionVolumeCollectorJobDropped) Emit(logger lager.Logger) {
	emit(
		logger.Session("gc-volume-collector-dropped"),
		Event{
			Name:  "GC volume collector job dropped",
			Value: 1,
			State: EventStateOK,
			Attributes: map[string]string{
				"worker": event.WorkerName,
			},
		},
	)
}

type BuildStarted struct {
	PipelineName string
	JobName      string
	BuildName    string
	BuildID      int
	TeamName     string
}

func (event BuildStarted) Emit(logger lager.Logger) {
	emit(
		logger.Session("build-started"),
		Event{
			Name:  "build started",
			Value: event.BuildID,
			State: EventStateOK,
			Attributes: map[string]string{
				"pipeline":   event.PipelineName,
				"job":        event.JobName,
				"build_name": event.BuildName,
				"build_id":   strconv.Itoa(event.BuildID),
				"team_name":  event.TeamName,
			},
		},
	)
}

type BuildFinished struct {
	PipelineName  string
	JobName       string
	BuildName     string
	BuildID       int
	BuildStatus   db.BuildStatus
	BuildDuration time.Duration
	TeamName      string
}

func (event BuildFinished) Emit(logger lager.Logger) {
	emit(
		logger.Session("build-finished"),
		Event{
			Name:  "build finished",
			Value: ms(event.BuildDuration),
			State: EventStateOK,
			Attributes: map[string]string{
				"pipeline":     event.PipelineName,
				"job":          event.JobName,
				"build_name":   event.BuildName,
				"build_id":     strconv.Itoa(event.BuildID),
				"build_status": string(event.BuildStatus),
				"team_name":    event.TeamName,
			},
		},
	)
}

type SlowQuery struct {
<<<<<<< HEAD
	AvgTime      float64
	Calls        int
	TotalTime    float64
=======
	AvgTime      time.Duration
	Calls        int
	TotalTime    time.Duration
>>>>>>> 7a2382a8888ea3de791f6711e98b68f99c9d150f
	Rows         int
	HitPercent   float64
	SqlStatement string
}

func (event SlowQuery) Emit(logger lager.Logger) {
	emit(
		logger.Session("slow-queries"),
		Event{
			Name:  "slow queries",
<<<<<<< HEAD
			Value: event.AvgTime,
			State: EventStateOK,
			Attributes: map[string]string{
				"calls":          strconv.Itoa(event.Calls),
				"total_time_sec": strconv.FormatFloat(event.TotalTime, 'f', -1, 64),
=======
			Value: ms(event.AvgTime),
			State: EventStateOK,
			Attributes: map[string]string{
				"calls":          strconv.Itoa(event.Calls),
				"total_time_sec": strconv.FormatFloat(sec(event.TotalTime), 'f', -1, 64),
>>>>>>> 7a2382a8888ea3de791f6711e98b68f99c9d150f
				"rows":           strconv.Itoa(event.Rows),
				"hit_percent":    strconv.FormatFloat(event.HitPercent, 'f', -1, 64),
				"sql_statement":  event.SqlStatement,
			},
		},
	)
}

func ms(duration time.Duration) float64 {
	return float64(duration) / 1000000
}

func sec(duration time.Duration) float64 {
	return float64(duration) / 1000000000
}

type HTTPResponseTime struct {
	Route    string
	Path     string
	Method   string
	Duration time.Duration
}

func (event HTTPResponseTime) Emit(logger lager.Logger) {
	state := EventStateOK

	if event.Duration > 100*time.Millisecond {
		state = EventStateWarning
	}

	if event.Duration > 1*time.Second {
		state = EventStateCritical
	}

	emit(
		logger.Session("http-response-time"),
		Event{
			Name:  "http response time",
			Value: ms(event.Duration),
			State: state,
			Attributes: map[string]string{
				"route":  event.Route,
				"path":   event.Path,
				"method": event.Method,
			},
		},
	)
}

func collectSlowQueries(logger lager.Logger) ([]SlowQuery, error) {
	if len(Databases) < 1 {
		err := errors.New("No database available")
		logger.Error("slow-query-no-database-found", err)
		return nil, err
	}
	dbConn := Databases[0]

	slowQueries := []SlowQuery{}

	results, err := dbConn.Query(
		`SELECT
		total_time / calls as avg_time,
		calls,
		total_time,
		rows,
		100.0 * shared_blks_hit / nullif(shared_blks_hit + shared_blks_read, 0) AS hit_percent,
	  regexp_replace(query, '[\s\t\n]+', ' ', 'g') as sanitized_sql
		FROM pg_stat_statements
		WHERE query NOT LIKE '%EXPLAIN%'
		AND query NOT LIKE '%INDEX%'
		AND query NOT LIKE '%pg_stat_statements%'
		AND calls > 1
		ORDER BY avg_time DESC LIMIT 5
	`)
	if err != nil || results == nil {
		logger.Error("slow-query-collection-failed", err)
		return nil, err
	}

	defer db.Close(results)

	for results.Next() {
		var slowQuery SlowQuery
		err = results.Scan(
			&slowQuery.AvgTime,
			&slowQuery.Calls,
			&slowQuery.TotalTime,
			&slowQuery.Rows,
			&slowQuery.HitPercent,
			&slowQuery.SqlStatement,
		)
		if err != nil {
			logger.Error("slow-query-scanrow-failed", err)
		} else {
			slowQueries = append(slowQueries, slowQuery)
		}
	}
	return slowQueries, nil
}
