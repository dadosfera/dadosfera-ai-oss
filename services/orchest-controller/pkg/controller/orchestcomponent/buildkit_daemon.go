package orchestcomponent

import (
	"strings"

	orchestv1alpha1 "github.com/orchest/orchest/services/orchest-controller/pkg/apis/orchest/v1alpha1"
	"github.com/orchest/orchest/services/orchest-controller/pkg/controller"
	"github.com/orchest/orchest/services/orchest-controller/pkg/utils"
	"golang.org/x/net/context"
	"k8s.io/utils/pointer"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type BuildKitDaemonReconciler struct {
	*OrchestComponentController
}

func NewBuildKitDaemonReconciler(ctrl *OrchestComponentController) OrchestComponentReconciler {
	return &BuildKitDaemonReconciler{
		ctrl,
	}
}

func (reconciler *BuildKitDaemonReconciler) Reconcile(ctx context.Context, component *orchestv1alpha1.OrchestComponent) error {
	hash := utils.ComputeHash(component)
	matchLabels := controller.GetResourceMatchLables(controller.BuildKitDaemon, component)
	metadata := controller.GetMetadata(controller.BuildKitDaemon, hash, component, OrchestComponentKind)
	newDs, err := getBuildKitDaemonDaemonset(metadata, matchLabels, component)
	if err != nil {
		return err
	}

	_, err = reconciler.dsLister.DaemonSets(component.Namespace).Get(component.Name)
	if err != nil {
		if !kerrors.IsAlreadyExists(err) {
			_, err = reconciler.Client().AppsV1().DaemonSets(component.Namespace).Create(ctx, newDs, metav1.CreateOptions{})
			reconciler.EnqueueAfter(component)
			return err
		}
		return err
	}

	return reconciler.updatePhase(ctx, component, orchestv1alpha1.Running)

}

func (reconciler *BuildKitDaemonReconciler) Uninstall(ctx context.Context, component *orchestv1alpha1.OrchestComponent) (bool, error) {

	err := reconciler.Client().AppsV1().DaemonSets(component.Namespace).Delete(ctx, component.Name, metav1.DeleteOptions{})
	if err != nil && !kerrors.IsNotFound(err) {
		return false, err
	}

	return true, nil
}

func getBuildKitDaemonCrioDaemonset(
    metadata metav1.ObjectMeta,
    matchLabels map[string]string,
    component *orchestv1alpha1.OrchestComponent,
) (*appsv1.DaemonSet, error) {
    env := component.Spec.Template.Env
    socket := utils.GetKeyFromEnvVar(env, "CONTAINER_RUNTIME_SOCKET")

    hostPathDirOrCreate := corev1.HostPathDirectoryOrCreate
    bidir := corev1.MountPropagationBidirectional
    privileged := true

    runCrio := "/var/run/crio"
    storageCrio := "/var/lib/containers/storage"

    vols := []corev1.Volume{
        {
            Name: "runtime-socket",
            VolumeSource: corev1.VolumeSource{
                HostPath: &corev1.HostPathVolumeSource{
                    Path: socket,
                    Type: func() *corev1.HostPathType {
                        t := corev1.HostPathSocket
                        return &t
                    }(),
                },
            },
        },
        {
            Name: "run-containerd",
            VolumeSource: corev1.VolumeSource{
                HostPath: &corev1.HostPathVolumeSource{
                    Path: runCrio,
                    Type: &hostPathDirOrCreate,
                },
            },
        },
        {
            Name: "var-lib-containerd",
            VolumeSource: corev1.VolumeSource{
                HostPath: &corev1.HostPathVolumeSource{
                    Path: storageCrio,
                    Type: &hostPathDirOrCreate,
                },
            },
        },
        {
            Name: "run-orchest-buildkit",
            VolumeSource: corev1.VolumeSource{
                HostPath: &corev1.HostPathVolumeSource{
                    Path: "/run/orchest_buildkit",
                    Type: &hostPathDirOrCreate,
                },
            },
        },
        {
            Name: "orchest-buildkit-storage",
            VolumeSource: corev1.VolumeSource{
                HostPath: &corev1.HostPathVolumeSource{
                    Path: "/var/lib/orchest_buildkit",
                    Type: &hostPathDirOrCreate,
                },
            },
        },
    }

    mounts := []corev1.VolumeMount{
        {Name: "runtime-socket", MountPath: socket},
        {Name: "run-containerd", MountPath: runCrio},
        {Name: "var-lib-containerd", MountPath: storageCrio},
        {Name: "run-orchest-buildkit", MountPath: "/run/orchest_buildkit"},
        {
            Name:             "orchest-buildkit-storage",
            MountPath:        "/var/lib/orchest_buildkit",
            MountPropagation: &bidir,
        },
    }

    container := corev1.Container{
        Name:            controller.BuildKitDaemon,
        Image:           component.Spec.Template.Image,
        ImagePullPolicy: corev1.PullIfNotPresent,
        Command:         []string{"/usr/bin/buildkitd"},
        Args: []string{
            "--debug",
            "--root=/var/lib/orchest_buildkit",
            "--addr=unix:///run/orchest_buildkit/buildkitd.sock",
            "--oci-worker=true",
            "--containerd-worker=false",
        },
        Env: append(env,
            corev1.EnvVar{Name: "CONTAINER_RUNTIME", Value: "cri-o"},
            corev1.EnvVar{Name: "CONTAINER_RUNTIME_SOCKET", Value: socket},
            corev1.EnvVar{Name: "CONTAINERD_NAMESPACE", Value: "k8s.io"},
        ),
        SecurityContext: &corev1.SecurityContext{Privileged: &privileged},
        VolumeMounts:    mounts,
        ReadinessProbe: &corev1.Probe{
            ProbeHandler: corev1.ProbeHandler{
                Exec: &corev1.ExecAction{
                    Command: []string{
                        "buildctl", "--addr",
                        "unix:///run/orchest_buildkit/buildkitd.sock",
                        "debug", "workers",
                    },
                },
            },
            InitialDelaySeconds: 5,
            PeriodSeconds:       30,
            TimeoutSeconds:      1,
        },
        LivenessProbe: &corev1.Probe{
            ProbeHandler: corev1.ProbeHandler{
                Exec: &corev1.ExecAction{
                    Command: []string{
                        "buildctl", "--addr",
                        "unix:///run/orchest_buildkit/buildkitd.sock",
                        "debug", "workers",
                    },
                },
            },
            InitialDelaySeconds: 5,
            PeriodSeconds:       30,
            TimeoutSeconds:      1,
        },
    }

    return &appsv1.DaemonSet{
        ObjectMeta: metadata,
        Spec: appsv1.DaemonSetSpec{
            Selector: &metav1.LabelSelector{MatchLabels: matchLabels},
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{Labels: matchLabels},
                Spec: corev1.PodSpec{
                    TerminationGracePeriodSeconds: pointer.Int64Ptr(1),
                    Volumes:                       vols,
                    NodeSelector:                  component.Spec.Template.NodeSelector,
                    Containers:                    []corev1.Container{container},
                },
            },
        },
    }, nil
}


func getBuildKitDaemonDaemonset(metadata metav1.ObjectMeta,
	matchLabels map[string]string, component *orchestv1alpha1.OrchestComponent) (
	*appsv1.DaemonSet, error) {

	socketPath := utils.GetKeyFromEnvVar(component.Spec.Template.Env, "CONTAINER_RUNTIME_SOCKET")
	hostPathSocket := corev1.HostPathSocket
	hostPathDirectoryOrCreate := corev1.HostPathDirectoryOrCreate
	bidirectional_propagation := corev1.MountPropagationBidirectional

	runContainerdPath := "/run/containerd"
	varLibContainerdPath := "/var/lib/containerd"

	// TODO: use distribution from https://github.com/orchest/orchest/pull/1267.
	if strings.Contains(socketPath, "microk8s") {
		runContainerdPath = "/var/snap/microk8s/common/run/containerd"
		varLibContainerdPath = "/var/snap/microk8s/common/var/lib/containerd"
	} else if strings.Contains(socketPath, "k3s") {
		runContainerdPath = "/var/run/k3s/containerd"
		varLibContainerdPath = "/var/lib/rancher/k3s/agent/containerd"
	} else if strings.Contains(socketPath, "crio") {
		return getBuildKitDaemonCrioDaemonset(metadata, matchLabels, component)
	}

	volumes := []corev1.Volume{
		{
			Name: "containerd-socket",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: socketPath,
					Type: &hostPathSocket,
				},
			},
		},
		{
			Name: "run-containerd",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: runContainerdPath,
					Type: &hostPathDirectoryOrCreate,
				},
			},
		},
		{
			Name: "var-lib-containerd",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: varLibContainerdPath,
					Type: &hostPathDirectoryOrCreate,
				},
			},
		},
		// Use a separate buildkit socket location for safety.
		{
			Name: "run-orchest-buildkit",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/run/orchest_buildkit",
					Type: &hostPathDirectoryOrCreate,
				},
			},
		},
		// Use a separate buildkit storage for safety
		// and in case we want to cleanup in the future.
		{
			Name: "orchest-buildkit-storage",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/var/lib/orchest_buildkit",
					Type: &hostPathDirectoryOrCreate,
				},
			},
		},
	}

	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "containerd-socket",
			MountPath: "/run/containerd/containerd.sock",
		},
		{
			Name:      "run-containerd",
			MountPath: "/run/containerd",
		},
		{
			Name:      "var-lib-containerd",
			MountPath: "/var/lib/containerd",
		},
		{
			Name:      "run-orchest-buildkit",
			MountPath: "/run/orchest_buildkit",
		},
		{
			Name:             "orchest-buildkit-storage",
			MountPath:        "/var/lib/orchest_buildkit",
			MountPropagation: &bidirectional_propagation,
		},
	}

	if strings.Contains(socketPath, "microk8s") || strings.Contains(socketPath, "k3s") {
		volumes = append(volumes, corev1.Volume{
			Name: "run-containerd-fifo",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/run/containerd/fifo",
					Type: &hostPathDirectoryOrCreate,
				},
			},
		})

		volumeMounts = append(volumeMounts,
			corev1.VolumeMount{
				Name:      "run-containerd-fifo",
				MountPath: "/run/containerd/fifo",
			},
			corev1.VolumeMount{
				Name:      "run-containerd",
				MountPath: runContainerdPath,
			},
			corev1.VolumeMount{
				Name:      "var-lib-containerd",
				MountPath: varLibContainerdPath,
			},
		)
	}

	template := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: matchLabels,
		},
		Spec: corev1.PodSpec{
			TerminationGracePeriodSeconds: pointer.Int64Ptr(1),
			Volumes:                       volumes,
			NodeSelector:                  component.Spec.Template.NodeSelector,
			Containers: []corev1.Container{
				{
					Name:            controller.BuildKitDaemon,
					Image:           component.Spec.Template.Image,
					Env:             component.Spec.Template.Env,
					ImagePullPolicy: corev1.PullIfNotPresent,
					SecurityContext: &corev1.SecurityContext{
						Privileged: pointer.BoolPtr(true),
					},
					// Probes from buildkitd k8s examples.
					ReadinessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							Exec: &corev1.ExecAction{
								Command: []string{
									"buildctl",
									"--addr",
									"unix:///run/orchest_buildkit/buildkitd.sock",
									"debug",
									"workers",
								},
							},
						},
						InitialDelaySeconds: 5,
						PeriodSeconds:       30,
					},
					LivenessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							Exec: &corev1.ExecAction{
								Command: []string{
									"buildctl",
									"--addr",
									"unix:///run/orchest_buildkit/buildkitd.sock",
									"debug",
									"workers",
								},
							},
						},
						InitialDelaySeconds: 5,
						PeriodSeconds:       30,
					},
					VolumeMounts: volumeMounts},
			},
		},
	}

	if component.Spec.Template.CorePriorityClassName != "" {
		template.Spec.PriorityClassName = component.Spec.Template.CorePriorityClassName
	}

	daemonSet := &appsv1.DaemonSet{
		ObjectMeta: metadata,
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: matchLabels,
			},
			Template: template,
		},
	}

	return daemonSet, nil

}
