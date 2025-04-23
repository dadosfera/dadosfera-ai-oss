package orchestcomponent

import (
	orchestv1alpha1 "github.com/orchest/orchest/services/orchest-controller/pkg/apis/orchest/v1alpha1"
	"github.com/orchest/orchest/services/orchest-controller/pkg/controller"
	"github.com/orchest/orchest/services/orchest-controller/pkg/utils"
	"golang.org/x/net/context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CeleryWorkerReconciler struct {
	*OrchestComponentController
}

func NewCeleryWorkerReconciler(ctrl *OrchestComponentController) OrchestComponentReconciler {
	return &CeleryWorkerReconciler{
		ctrl,
	}
}

func (reconciler *CeleryWorkerReconciler) Reconcile(ctx context.Context, component *orchestv1alpha1.OrchestComponent) error {

	hash := utils.ComputeHash(component)
	matchLabels := controller.GetResourceMatchLables(controller.CeleryWorker, component)
	metadata := controller.GetMetadata(controller.CeleryWorker, hash, component, OrchestComponentKind)
	newDep := getCeleryWorkerDeployment(metadata, matchLabels, component)

	oldDep, err := reconciler.depLister.Deployments(component.Namespace).Get(component.Name)
	if err != nil {
		if !kerrors.IsAlreadyExists(err) {
			_, err = reconciler.Client().AppsV1().Deployments(component.Namespace).Create(ctx, newDep, metav1.CreateOptions{})
			reconciler.EnqueueAfter(component)
			return err
		}
		return err
	}

	if !isDeploymentUpdated(newDep, oldDep) {
		_, err := reconciler.Client().AppsV1().Deployments(component.Namespace).Update(ctx, newDep, metav1.UpdateOptions{})
		reconciler.EnqueueAfter(component)
		return err
	}

	if isDeploymentReady(oldDep) {
		err = reconciler.updatePhase(ctx, component, orchestv1alpha1.Running)
	} else {
		err = reconciler.updatePhase(ctx, component, orchestv1alpha1.OrchestPhase(orchestv1alpha1.DeployingCeleryWorker))
	}

	return err
}

func (reconciler *CeleryWorkerReconciler) Uninstall(ctx context.Context, component *orchestv1alpha1.OrchestComponent) (bool, error) {

	dep, err := reconciler.Client().AppsV1().Deployments(component.Namespace).Get(ctx, component.Name, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return true, nil
		}

	} else if dep.GetDeletionTimestamp().IsZero() {
		DeletePropagationForeground := metav1.DeletionPropagation("Foreground")
		err = reconciler.Client().AppsV1().Deployments(component.Namespace).Delete(ctx, component.Name, metav1.DeleteOptions{
			PropagationPolicy: &DeletePropagationForeground,
		})
	}

	return false, err
}

func getCeleryWorkerDeployment(metadata metav1.ObjectMeta,
	matchLabels map[string]string, component *orchestv1alpha1.OrchestComponent) *appsv1.Deployment {

	image := component.Spec.Template.Image

	dnsResolverTimeout := "10"
	dnsResolverAttempts := "5"

	livenessProbe := &corev1.Probe{
		ProbeHandler	: corev1.ProbeHandler{
			Exec: &corev1.ExecAction{
				Command: []string{"python3", "/celery_worker_healthcheck.py"},
			},
		},
		InitialDelaySeconds: 30,
		PeriodSeconds:       60,
		TimeoutSeconds:      10,
		FailureThreshold:    5,
	}
	template := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: matchLabels,
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: controller.CeleryWorker,
			NodeSelector:       component.Spec.Template.NodeSelector,
			Volumes: []corev1.Volume{
				{
					Name: controller.UserDirName,
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: controller.UserDirName,
							ReadOnly:  false,
						},
					},
				},
				{
					Name: "tls-secret",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "registry-tls-secret",
							Items: []corev1.KeyToPath{
								{
									Key:  "ca.crt",
									Path: "additional-ca-cert-bundle.crt",
								},
							},
						},
					},
				},
			},
			DNSConfig: &corev1.PodDNSConfig{
				Options: []corev1.PodDNSConfigOption{
					{Name: "timeout", Value: &dnsResolverTimeout},
					{Name: "attempts", Value: &dnsResolverAttempts},
				},
			},
			Containers: []corev1.Container{
				{
					Name:  controller.CeleryWorker,
					Image: image,
					Env:   component.Spec.Template.Env,
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("250m")},
					},
<<<<<<< HEAD
=======
					LivenessProbe: livenessProbe,
>>>>>>> parent/beta
					ImagePullPolicy: corev1.PullIfNotPresent,
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      controller.UserDirName,
							MountPath: controller.UserdirMountPath,
						},
						{
							Name:      "tls-secret",
							MountPath: "/usr/lib/ssl/certs/additional-ca-cert-bundle.crt",
							SubPath:   "additional-ca-cert-bundle.crt",
							ReadOnly:  true,
						},
					},
				},
			},
		},
	}

	if component.Spec.Template.CorePriorityClassName != "" {
		template.Spec.PriorityClassName = component.Spec.Template.CorePriorityClassName
		template.Spec.Containers[0].Env = append(template.Spec.Containers[0].Env,
			corev1.EnvVar{
				Name:  "PRIORITY_CLASS_NAME",
				Value: component.Spec.Template.CorePriorityClassName,
			})
	}

	if component.Spec.Template.StateVolumeName == controller.OrchestStateVolumeName {
		var celery_state_dir = "/celery-state"
		for i := 0; i < len(template.Spec.Containers[0].Env); i++ {
			if template.Spec.Containers[0].Env[i].Name == "CELERY_STATE_DIR" {
				template.Spec.Containers[0].Env[i].Value = celery_state_dir
			}
		}
		template.Spec.Volumes = append(template.Spec.Volumes,
			corev1.Volume{
				Name: component.Spec.Template.StateVolumeName,
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: component.Spec.Template.StateVolumeName,
						ReadOnly:  false,
					},
				},
			},
		)

		template.Spec.Containers[0].VolumeMounts = append(
			template.Spec.Containers[0].VolumeMounts,
			corev1.VolumeMount{
				Name:      component.Spec.Template.StateVolumeName,
				MountPath: celery_state_dir,
			},
		)
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metadata,
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: matchLabels,
			},
			Template: template,
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
		},
	}

	deployment.Labels = utils.CloneAndAddLabel(metadata.Labels, map[string]string{
		controller.DeploymentHashLabelKey: utils.ComputeHash(&deployment.Spec),
	})
	return deployment
}
