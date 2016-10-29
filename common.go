package kubeclient

import (
	"fmt"

	"github.com/codegp/env"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

const localStoreVolumeName string = "local-store-vol"

func registry(image string) string {
	return fmt.Sprintf("gcr.io/%s/%s:latest", env.GCloudProjectID(), image)
}

func configEnv() []api.EnvVar {
	return []api.EnvVar{
		configEnvVar("GCLOUD_PROJECT_ID", "gcloud-project-id"),
		configEnvVar("IS_LOCAL", "is-local"),
		configEnvVar("DATASTORE_EMULATOR_HOST", "datastore-emulator-host"),
	}
}

func configEnvVar(envVarName, configName string) api.EnvVar {
	return api.EnvVar{
		Name: envVarName,
		ValueFrom: &api.EnvVarSource{
			ConfigMapKeyRef: &api.ConfigMapKeySelector{
				LocalObjectReference: api.LocalObjectReference{
					Name: "codegp-config",
				},
				Key: configName,
			},
		},
	}
}

func ipEnvVar() api.EnvVar {
	return api.EnvVar{
		Name: "POD_IP",
		ValueFrom: &api.EnvVarSource{
			FieldRef: &api.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "status.podIP",
			},
		},
	}
}

func envVar(key, value string) api.EnvVar {
	return api.EnvVar{
		Name:  key,
		Value: value,
	}
}

func podTypeMeta() unversioned.TypeMeta {
	return unversioned.TypeMeta{
		Kind:       "Pod",
		APIVersion: "v1",
	}
}

func localStoreVolume() api.Volume {
	return api.Volume{
		Name: localStoreVolumeName,
		VolumeSource: api.VolumeSource{
			HostPath: &api.HostPathVolumeSource{
				Path: env.LocalStorePath,
			},
		},
	}
}

func localStoreVolumeMount() api.VolumeMount {
	return api.VolumeMount{
		Name:      localStoreVolumeName,
		ReadOnly:  false,
		MountPath: env.LocalStorePath,
	}
}
