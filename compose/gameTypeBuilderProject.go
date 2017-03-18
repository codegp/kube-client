package compose

import (
	"fmt"

	"github.com/codegp/cloud-persister/models"
	"github.com/codegp/env"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/docker/ctx"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/yaml"
)

const (
	dockerSockVolName  = "dockersock"
	dockerSockHostPath = "/var/run/docker.sock"

	gameTypeBuilderName = "game-type-builder"
	cgpNetworkName      = "appclientts_cgp-network"
)

func createGameTypeBuilderProject(gameType *models.GameType) (project.APIProject, error) {
	proj, err := docker.NewProject(&ctx.Context{
		Context: project.Context{
			ProjectName: fmt.Sprint("game-type-builder-", gameType.ID),
		},
	}, nil)

	if err != nil {
		return nil, err
	}

	// projPtr := proj.(*project.Project)
	// projPtr.AddNetworkConfig(fmt.Sprint("appclientts_", cgpNetworkName), cgpNetworkConfig())
	proj.AddConfig(gameTypeBuilderName, gameTypeBuilderConfig(gameType))
	proj.Parse()
	return proj, nil
}

func gameTypeBuilderConfig(gameType *models.GameType) *config.ServiceConfig {
	return &config.ServiceConfig{
		Image:       registry(gameTypeBuilderName),
		Environment: gameTypeBuilderEnv(gameType),
		Volumes:     gameTypeBuilderVolumes(),
		Networks:    cgpNetwork(),
	}
}

func gameTypeBuilderVolumes() *yaml.Volumes {
	return &yaml.Volumes{
		Volumes: []*yaml.Volume{
			localStoreVolume(),
			dockerHostVolume(),
		},
	}
}

func dockerHostVolume() *yaml.Volume {
	return &yaml.Volume{
		Source:      dockerSockHostPath,
		Destination: dockerSockHostPath,
	}
}

func gameTypeBuilderEnv(gameType *models.GameType) yaml.MaporEqualSlice {
	return append(cgpEnv(), fmt.Sprint("GAME_TYPE_ID=", gameType.ID))
}

func localStoreVolumeConfig() *config.VolumeConfig {
	return &config.VolumeConfig{
		External: yaml.External{
			External: true,
		},
	}
}

func cgpEnv() []string {
	return []string{
		fmt.Sprint("GCLOUD_PROJECT_ID=", env.GCloudProjectID()),
		fmt.Sprint("IS_LOCAL=", env.IsLocal()),
		fmt.Sprint("DATASTORE_EMULATOR_HOST=", env.DataStoreHost()),
	}
}

func cgpNetwork() *yaml.Networks {
	return &yaml.Networks{
		Networks: []*yaml.Network{
			&yaml.Network{
				RealName: cgpNetworkName,
			},
		},
	}
}

func cgpNetworkConfig() *config.NetworkConfig {
	return &config.NetworkConfig{
		External: yaml.External{
			External: true,
		},
	}
}
