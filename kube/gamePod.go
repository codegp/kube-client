package kube

import (
	"fmt"

	"github.com/codegp/cloud-persister/models"
	"github.com/codegp/env"
	"k8s.io/kubernetes/pkg/api"
)

const sourceVolumeName string = "source-vol"

func createGamePod(game *models.Game) *api.Pod {
	return &api.Pod{
		TypeMeta:   podTypeMeta(),
		ObjectMeta: gamePodMeta(game),
		Spec:       gamePodSpec(game),
	}
}

func gamePodMeta(game *models.Game) api.ObjectMeta {
	return api.ObjectMeta{
		Name: fmt.Sprintf("game-%d", game.ID),
	}
}

func gamePodSpec(game *models.Game) api.PodSpec {
	return api.PodSpec{
		Volumes:       gamePodSourceVolumes(),
		Containers:    gamePodContainers(game),
		RestartPolicy: "Never",
	}
}

func gamePodSourceVolumes() []api.Volume {
	volumes := []api.Volume{
		api.Volume{
			Name: sourceVolumeName,
			VolumeSource: api.VolumeSource{
				HostPath: &api.HostPathVolumeSource{
					Path: env.SourcePath,
				},
			},
		},
	}

	if env.IsLocal() {
		volumes = append(volumes, localStoreVolume())
	}
	return volumes
}

func gamePodContainers(game *models.Game) []api.Container {
	// TODO: establish sane resource requirements
	return []api.Container{
		api.Container{
			Name:            fmt.Sprintf("game-%d", game.ID),
			Image:           registry(fmt.Sprintf("game-runner-%d", game.GameTypeID)),
			Ports:           gamePodContainerPorts(),
			Env:             gameRunnerEnv(game),
			VolumeMounts:    gamePodVolumeMounts(),
			ImagePullPolicy: "IfNotPresent",
		},
	}
}

func gamePodVolumeMounts() []api.VolumeMount {
	mounts := []api.VolumeMount{
		api.VolumeMount{
			Name:      sourceVolumeName,
			ReadOnly:  true,
			MountPath: env.SourcePath,
		},
	}
	if env.IsLocal() {
		mounts = append(mounts, localStoreVolumeMount())
	}
	return mounts
}

func gamePodContainerPorts() []api.ContainerPort {
	return []api.ContainerPort{
		api.ContainerPort{
			ContainerPort: 9000,
		},
	}
}

func gameRunnerEnv(game *models.Game) []api.EnvVar {
	gameIDEnv := envVar("GAME_ID", fmt.Sprintf("%d", game.ID))
	return append(configEnv(), gameIDEnv, ipEnvVar())
}
