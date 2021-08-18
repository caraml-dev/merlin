package imagebuilder

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gojek/merlin/cluster"
	"github.com/gojek/merlin/log"
)

var (
	deleteJobsErrors = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "merlin_api",
		Name:      "image_builder_delete_jobs_error_count",
		Help:      "The total number of jobs deletion failed",
	})

	deleteJobsLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "merlin_api",
		Name:      "image_builder_delete_jobs_duration_ms",
		Help:      "Image builder delete jobs histogram",
		Buckets:   prometheus.ExponentialBuckets(64, 2, 5), // 64, 128, 256, 512, 1024, +Inf
	}, []string{"status"})
)

// Janitor cleans the finished image building jobs.
type Janitor struct {
	cc  cluster.Controller
	cfg JanitorConfig
}

// JanitorConfig stores the configuration for the Janitor.
type JanitorConfig struct {
	DryRun         bool
	BuildNamespace string
	Retention      time.Duration
}

// NewJanitor returns an initialized Janitor.
func NewJanitor(clusterController cluster.Controller, cfg JanitorConfig) *Janitor {
	return &Janitor{
		cc:  clusterController,
		cfg: cfg,
	}
}

// CleanJobs deletes the finished (succeeded or failed) image building jobs.
func (j *Janitor) CleanJobs() {
	log.Infof("Image Builder Janitor: Start cleaning jobs...")

	expiredJobs, err := j.getExpiredJobs()
	if err != nil {
		log.Errorf("failed to get expired jobs: %s", err)
		return
	}

	if err := j.deleteJobs(expiredJobs); err != nil {
		log.Errorf("failed to delete jobs: %s", err)
		return
	}

	log.Infof("Image Builder Janitor: Cleaning jobs finish...")
	return
}

func (j *Janitor) getExpiredJobs() ([]batchv1.Job, error) {
	jobs, err := j.cc.ListJobs(j.cfg.BuildNamespace, labelOrchestratorName+"=merlin")
	if err != nil {
		return nil, err
	}

	expiredJobs := []batchv1.Job{}

	now := time.Now()
	for _, job := range jobs.Items {
		if now.Sub(job.Status.CompletionTime.Time) > j.cfg.Retention {
			expiredJobs = append(expiredJobs, job)
		}
	}

	return expiredJobs, nil
}

func (j *Janitor) deleteJobs(expiredJobs []batchv1.Job) error {
	for _, job := range expiredJobs {
		startTime := time.Now()

		logMsg := fmt.Sprintf("Image Builder Janitor: Deleting an image builder job (%s)", job.Name)

		deleteOptions := &metav1.DeleteOptions{}
		if j.cfg.DryRun {
			deleteOptions.DryRun = []string{"All"}
			logMsg = "Dry run: All. " + logMsg
		}

		log.Debugf(logMsg)

		err := j.cc.DeleteJob(j.cfg.BuildNamespace, job.Name, deleteOptions)
		durationMs := time.Now().Sub(startTime).Milliseconds()

		if err != nil {
			// Failed deletion would be picked up by the next clean up job.
			log.Errorf("failed to delete an image builder job (%s): %s", job.Name, err)

			deleteJobsErrors.Inc()
			deleteJobsLatency.WithLabelValues("error").Observe(float64(durationMs))

			continue
		}

		deleteJobsLatency.WithLabelValues("success").Observe(float64(durationMs))
	}

	return nil
}
