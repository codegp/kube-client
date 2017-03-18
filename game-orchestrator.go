package jobclient

import (
	"github.com/codegp/env"
	"github.com/codegp/job-client/compose"
	jobs "github.com/codegp/job-client/jobclient"
	"github.com/codegp/job-client/kube"
)

// GetJobClient -
func GetJobClient() (jobs.JobClient, error) {
	if env.IsLocal() {
		return compose.NewClient(), nil
	}
	return kube.NewClient()
}
