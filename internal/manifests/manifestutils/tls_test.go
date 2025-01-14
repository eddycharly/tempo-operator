package manifestutils

import (
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"

	"github.com/grafana/tempo-operator/apis/tempo/v1alpha1"
)

func TestConfigureTLSVolumes(t *testing.T) {
	tlsSpec := v1alpha1.TLSSpec{
		Enabled:    true,
		CA:         "custom-ca",
		Cert:       "custom-cert",
		MinVersion: "123",
	}

	pod := corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name: "tempo",
			},
		},
	}

	ConfigureTLSVolumes(&pod, 0, tlsSpec, "/var/ca", "/var/cert", "prefix")

	require.Equal(t, []corev1.VolumeMount{
		{
			Name:      "prefix-ca",
			MountPath: "/var/ca",
			ReadOnly:  true,
		},
		{
			Name:      "prefix-cert",
			MountPath: "/var/cert",
			ReadOnly:  true,
		},
	}, pod.Containers[0].VolumeMounts)

	require.Equal(t, []corev1.Volume{
		{
			Name: "prefix-ca",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "custom-ca",
					},
				},
			},
		},
		{
			Name: "prefix-cert",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "custom-cert",
				},
			},
		},
	}, pod.Volumes)
}
