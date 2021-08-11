package imagebuilder

import (
	"time"

	"github.com/gojek/merlin/cluster"
	"github.com/gojek/merlin/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	if err := j.deleteJobs(); err != nil {
		log.Errorf("failed to delete jobs: %s", err)
		return
	}

	log.Infof("Image Builder Janitor: Cleaning jobs finish...")
	return
}

func (j *Janitor) deleteJobs() error {
	deleteOptions := &metav1.DeleteOptions{}
	if j.cfg.DryRun {
		deleteOptions.DryRun = []string{"All"}
	}

	listOptions := metav1.ListOptions{LabelSelector: labelOrchestratorName + "=merlin"}

	return j.cc.DeleteJobs(j.cfg.BuildNamespace, deleteOptions, listOptions)
}
