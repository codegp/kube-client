package jobclient

import (
	"github.com/codegp/cloud-persister/models"
)

// Job is a running job. You can monitor a pod as it runs
type Job interface {
	WatchToCompletion() error
	WatchToStartup() error
}

// JobClient starts games and builds
type JobClient interface {

	// StartGame starts a game, requires a game
	// @returns error if client fails to start the game runner job
	StartGame(game *models.Game, projects []*models.Project) (Job, error)

	// BuildGameType kicks off a gametype build
	// @returns error if client fails to start the gametype builder job
	BuildGameType(gameType *models.GameType) (Job, error)
}
