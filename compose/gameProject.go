package compose

import (
	"fmt"
	"log"

	"github.com/codegp/cloud-persister/models"
	"github.com/codegp/env"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/docker/ctx"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/yaml"
)

const (
	gameRunnerImagePrefix = "game-runner-"
	gameRunnerName        = "game-runner"
	teamRunnerPrefix      = "team-runner-"

	localVolumeName  = "localstore"
	sourceVolumeName = "source-vol"
)

func createGameProject(game *models.Game, projects []*models.Project) (project.APIProject, error) {
	project, err := docker.NewProject(&ctx.Context{
		Context: project.Context{
			ProjectName: fmt.Sprint("game-", game.ID),
		},
	}, nil)

	if err != nil {
		return nil, err
	}

	// project.AddVolumeConfig(localVolumeName, localVolumeConfig())
	for _, proj := range projects {
		log.Printf("Proj config for %v %d", proj, proj.ID)
		project.AddConfig(teamRunnerName(proj.ID), teamRunnerConfig(game, proj))
	}
	project.AddConfig(gameRunnerName, gameRunnerConfig(game))
	return project, nil
}

func gameRunnerConfig(game *models.Game) *config.ServiceConfig {
	return &config.ServiceConfig{
		Image:       registry(fmt.Sprint(gameRunnerImagePrefix, game.GameTypeID)),
		Ports:       []string{"9000"},
		Environment: gameRunnerEnv(game),
		Networks:    defaultNetwork(),
		Volumes:     gameRunnerProjectSourceVolumes(),
	}
}

func teamRunnerConfig(game *models.Game, project *models.Project) *config.ServiceConfig {
	return &config.ServiceConfig{
		Image:       registry(fmt.Sprint(teamRunnerPrefix, game.GameTypeID, "-", project.Language)),
		Ports:       []string{"9000"},
		Environment: teamRunnerEnv(game, project),
		Networks:    defaultNetwork(),
		Volumes:     teamRunnerProjectSourceVolumes(project),
	}
}

func gameRunnerProjectSourceVolumes() *yaml.Volumes {
	return &yaml.Volumes{
		Volumes: []*yaml.Volume{
			localStoreVolume(),
		},
	}
}

func teamRunnerProjectSourceVolumes(project *models.Project) *yaml.Volumes {
	vols := &yaml.Volumes{
		Volumes: []*yaml.Volume{
			localStoreVolume(),
		},
	}
	// if it is a local project mount the source directory
	if project.Directory != "" {
		vols.Volumes = append(vols.Volumes, sourceVolume(project))
	}
	return vols
}

func localStoreVolume() *yaml.Volume {
	return &yaml.Volume{
		Source:      "appclientts_local-store",
		Destination: env.LocalStorePath,
	}
}

func sourceVolume(project *models.Project) *yaml.Volume {
	return &yaml.Volume{
		Source:      project.Directory,
		Destination: env.SourcePath,
	}
}

func gameRunnerEnv(game *models.Game) yaml.MaporEqualSlice {
	return append(cgpEnv(),
		fmt.Sprint("GAME_ID=", game.ID),
		fmt.Sprint("POD_IP=", gameRunnerName))
}

func teamRunnerEnv(game *models.Game, project *models.Project) yaml.MaporEqualSlice {
	return append(cgpEnv(),
		fmt.Sprint("GAME_ID=", game.ID),
		fmt.Sprint("PROJECT_ID=", project.ID),
		fmt.Sprint("POD_IP=", teamRunnerName(project.ID)),
	)
}

func teamRunnerLinks(game *models.Game) []string {
	links := make([]string, len(game.ProjectIDs))
	for i, projID := range game.ProjectIDs {
		links[i] = fmt.Sprint("default_", teamRunnerName(projID), ":", teamRunnerName(projID))
	}
	return links
}

func teamRunnerName(projID int64) string {
	return fmt.Sprint(teamRunnerPrefix, projID)
}

func registry(image string) string {
	return fmt.Sprintf("gcr.io/%s/%s:latest", env.GCloudProjectID(), image)
}

func defaultNetwork() *yaml.Networks {
	return &yaml.Networks{
		Networks: []*yaml.Network{
			&yaml.Network{
				Name: "default",
			},
			&yaml.Network{
				RealName: cgpNetworkName,
				Aliases:  []string{"dsemulator"},
			},
		},
	}
}
