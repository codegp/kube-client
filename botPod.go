package kubeclient

import (
	"fmt"
	"log"

	"github.com/codegp/cloud-persister/models"
	"github.com/codegp/env"
	"k8s.io/kubernetes/pkg/api"
)

const botSourceVolumeName string = "bot-source-vol"

func createBotPod(ip string, botID int32, proj *models.Project, game *models.Game) *api.Pod {
	return &api.Pod{
		TypeMeta:   podTypeMeta(),
		ObjectMeta: botPodMeta(game, botID),
		Spec:       botPodSpec(ip, botID, proj, game),
	}
}

func botPodMeta(game *models.Game, botID int32) api.ObjectMeta {
	return api.ObjectMeta{
		Name: fmt.Sprintf("bot-%d-%d", game.ID, botID),
	}
}

func botPodSpec(ip string, botID int32, proj *models.Project, game *models.Game) api.PodSpec {
	return api.PodSpec{
		Volumes:       botPodSourceVolumes(proj.ID),
		Containers:    botPodContainers(ip, botID, proj, game),
		RestartPolicy: "Never",
	}
}

func botPodSourceVolumes(projID int64) []api.Volume {
	volumes := []api.Volume{
		api.Volume{
			Name: botSourceVolumeName,
			VolumeSource: api.VolumeSource{
				HostPath: &api.HostPathVolumeSource{
					Path: fmt.Sprintf("/%s/%d", env.SourcePath(), projID),
				},
			},
		},
	}

	if env.IsLocal() {
		volumes = append(volumes, localStoreVolume())
	}

	return volumes
}

func botPodContainers(ip string, botID int32, proj *models.Project, game *models.Game) []api.Container {
	// TODO: establish sane resource requirements
	log.Println("asdf")
	return []api.Container{
		api.Container{
			Name:         fmt.Sprintf("bot-%d", botID),
			Image:        registry(fmt.Sprintf("botrunner-%d-%s", game.GameTypeID, proj.Language)),
			Ports:        botPodContainerPorts(),
			Env:          botRunnerEnv(ip, botID, proj),
			VolumeMounts: botPodVolumeMounts(),
			// SecurityContext: botContainerSecurityContext(),
			ImagePullPolicy: "IfNotPresent",
		},
	}
}

func botPodVolumeMounts() []api.VolumeMount {
	mounts := []api.VolumeMount{
		api.VolumeMount{
			Name:      botSourceVolumeName,
			ReadOnly:  false,
			MountPath: env.SourcePath(),
		},
	}

	if env.IsLocal() {
		mounts = append(mounts, localStoreVolumeMount())
	}

	return mounts
}

func botPodContainerPorts() []api.ContainerPort {
	return []api.ContainerPort{
		api.ContainerPort{
			ContainerPort: 9000,
		},
	}
}

func botRunnerEnv(ip string, botID int32, proj *models.Project) []api.EnvVar {
	envVars := []api.EnvVar{
		envVar("GAME_RUNNER_IP", ip),
		envVar("BOT_ID", fmt.Sprintf("%d", botID)),
		envVar("PROJECT_ID", fmt.Sprintf("%d", proj.ID)),
		ipEnvVar(),
	}

	return append(envVars, configEnv()...)
}

func botContainerSecurityContext() *api.SecurityContext {
	// SecurityContext wants pointers to bools, i'm not crazy
	readOnly := false
	log.Println("NOT READ ONLY")
	// runAsNonRoot := true
	return &api.SecurityContext{
		ReadOnlyRootFilesystem: &readOnly,
	}
}
