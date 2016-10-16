package kubeclient

import (
	"fmt"

	"github.com/codegp/cloud-persister/models"
	"github.com/codegp/env"

	"k8s.io/kubernetes/pkg/api"
)

const (
	dockerSockVolName  = "dockersock"
	dockerSockHostPath = "/var/run/docker.sock"
)

func createGameTypeBuilderPod(gameType *models.GameType) *api.Pod {
	return &api.Pod{
		TypeMeta:   podTypeMeta(),
		ObjectMeta: gameTypePodMeta(gameType),
		Spec:       gameTypePodSpec(gameType),
	}
}

func gameTypePodMeta(gameType *models.GameType) api.ObjectMeta {
	return api.ObjectMeta{
		Name: fmt.Sprintf("builder-%d", gameType.ID),
	}
}

func gameTypePodSpec(gameType *models.GameType) api.PodSpec {
	return api.PodSpec{
		Volumes:       gameTypeBuilderSourceVolumes(),
		Containers:    gameTypeBuilderPodContainers(gameType),
		RestartPolicy: "Never",
	}
}

func gameTypeBuilderSourceVolumes() []api.Volume {
	volumes := []api.Volume{api.Volume{
		Name: dockerSockVolName,
		VolumeSource: api.VolumeSource{
			HostPath: &api.HostPathVolumeSource{
				Path: dockerSockHostPath,
			},
		},
	}}
	if env.IsLocal() {
		volumes = append(volumes, localStoreVolume())
	}
	return volumes
}

func gameTypeBuilderPodContainers(gameType *models.GameType) []api.Container {
	return []api.Container{gameTypeBuilderContainer(gameType)}
}

func gameTypeBuilderContainer(gameType *models.GameType) api.Container {
	return api.Container{
		Name:            "builder",
		Image:           registry("builder"),
		SecurityContext: privilegedSecurityContext(),
		Env:             gameTypeBuilderEnv(gameType),
		VolumeMounts:    gameTypeBuilderPodVolumeMounts(),
		ImagePullPolicy: "IfNotPresent",
	}
}

func gameTypeBuilderEnv(gameType *models.GameType) []api.EnvVar {
	gameTypeIDEnv := envVar("GAME_TYPE_ID", fmt.Sprintf("%d", gameType.ID))
	return append(configEnv(), gameTypeIDEnv)
}

func privilegedSecurityContext() *api.SecurityContext {
	// SecurityContext wants pointers to bools, i'm not crazy
	privileged := true
	return &api.SecurityContext{
		Privileged: &privileged,
	}
}

func gameTypeBuilderPodVolumeMounts() []api.VolumeMount {
	mounts := []api.VolumeMount{
		api.VolumeMount{
			Name:      dockerSockVolName,
			MountPath: dockerSockHostPath,
		},
	}
	if env.IsLocal() {
		mounts = append(mounts, localStoreVolumeMount())
	}
	return mounts
}
