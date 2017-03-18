package compose

import (
	"log"

	"golang.org/x/net/context"

	"k8s.io/kubernetes/pkg/client/unversioned"

	"github.com/codegp/cloud-persister/models"
	jobs "github.com/codegp/job-client/jobclient"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/project/options"
)

var _ jobs.JobClient = (*ComposeClient)(nil)

// ComposeClient creats jobs that are kube projects
type ComposeClient struct{}

var _ jobs.Job = (*ComposeJob)(nil)

// ComposeJob is a watchable job
type ComposeJob struct {
	project project.APIProject
	client  *unversioned.Client
}

func NewClient() *ComposeClient {
	return &ComposeClient{}
}

// StartGame implements JobClient.StartGame
func (cc *ComposeClient) StartGame(game *models.Game, projects []*models.Project) (jobs.Job, error) {
	project, err := createGameProject(game, projects)
	if err != nil {
		return nil, err
	}
	log.Printf("Starting Game! Spec:\n%v", project)
	err = project.Up(context.Background(), options.Up{})
	return cc.projectToComposeJob(project), err
}

// BuildGameType implements JobClient.BuildGameType
func (cc *ComposeClient) BuildGameType(gameType *models.GameType) (jobs.Job, error) {
	project, err := createGameTypeBuilderProject(gameType)
	if err != nil {
		return nil, err
	}
	log.Printf("Building Gametype Builder! Spec:\n%v", project)
	err = project.Up(context.Background(), options.Up{})
	return cc.projectToComposeJob(project), err
}

func (cc *ComposeClient) projectToComposeJob(project project.APIProject) *ComposeJob {
	return &ComposeJob{
		project: project,
	}
}

type DesiredProjectState func(project.APIProject) bool

func (cj *ComposeJob) watchProject(isDesired DesiredProjectState) error {
	events, err := cj.project.Events(context.Background())
	if err != nil {
		return err

	}

	for {
		e := <-events
		log.Printf("event %v dasfd %v", e, e.Event)
		if e.Event == "die" {
			return nil
		}
		// project := e.Object.(project.APIProject)
		// log.Printf("Update received for project %s, EventType %v\n", project.Name, e.Type)
		// if isDesired(project) {
		// 	log.Printf("Reached desired state for project %s", project.Name)
		// 	return nil
		// }
		// if project.Status.Phase == project.ProjectFailed {
		// 	return fmt.Errorf("Project failed.\nProject:\n%v", project)
		// }
		// if e.Type == watch.Deleted {
		// 	return fmt.Errorf("Project was deleted before reaching desired state.\nProject:\n%v", project)
		// }
		// if e.Type == watch.Error {
		// 	return fmt.Errorf("Project errored before reaching desired state.\nProject:\n%v", project)
		// }
	}
}

func (cj *ComposeJob) WatchToCompletion() error {
	// log.Printf("Watching %s to completion\n", cj.project.Name)
	return cj.watchProject(func(project project.APIProject) bool {
		return false
	})
}

func (cj *ComposeJob) WatchToStartup() error {
	// log.Printf("Watching %s to startup\n", cj.project.Name)
	return cj.watchProject(func(project project.APIProject) bool {
		return false
	})
}
