package controller

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"os"
	"fmt"
	orchestv1alpha1 "github.com/orchest/orchest/services/orchest-controller/pkg/apis/orchest/v1alpha1"
	"github.com/orchest/orchest/services/orchest-controller/pkg/utils"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	netsv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	//Components
	OrchestDatabase   = "orchest-database"
	OrchestApi        = "orchest-api"
	OrchestApiCleanup = "orchest-api-cleanup"
	Rabbitmq          = "rabbitmq-server"
	CeleryWorker      = "celery-worker"
	AuthServer        = "auth-server"
	OrchestWebserver  = "orchest-webserver"
	NodeAgent         = "node-agent"
	BuildKitDaemon    = "buildkit-daemon"

	// PVC names
	UserDirName            = "userdir-pvc"
	OrchestStateVolumeName = "orchest-state"

	// Deprecated, the PVC is removed.
	OldBuilderDirName = "image-builder-cache-pvc"

	// Mount paths
	UserdirMountPath      = "/userdir"
	OrchestStateMountPath = "/orchest/state"

	// database paths
	DBMountPath = "/userdir/.orchest/database/data"
	DBSubPath   = ".orchest/database/data"

	// rabbitmq paths
	RabbitmountPath = "/var/lib/rabbitmq/mnesia"
	RabbitSubPath   = ".orchest/rabbitmq-mnesia"

	// Ingress constants
	PrefixPathType            = netsv1.PathType("Prefix")
	k8sIngressClassController = "k8s.io/ingress-nginx"
	ngxIngressClassController = "nginx.org/ingress-controller"

	// Labels and Annotations
	OrchestHashLabelKey    = "orchest.io/orchest-hash"
	DeploymentHashLabelKey = "orchest.io/deployment-hash"
	GenerationKey          = "controller.orchest.io/generation"
	OwnerLabelKey          = "controller.orchest.io/owner"
	ControllerPartOfLabel  = "controller.orchest.io/part-of"
	ComponentLabelKey      = "controller.orchest.io/component"
	K8sDistroAnnotationKey = "controller.orchest.io/k8s"
	IngressAnnotationKey   = "controller.orchest.io/deploy-ingress"
	RestartAnnotationKey   = "orchest.io/restart"

	// Runtime annotations
	KubeAdmCRISocketAnnotationKey           = "kubeadm.alpha.kubernetes.io/cri-socket"
	ContainerRuntimeSocketPathAnnotationKey = "orchest.io/container-runtime-socket"
)

// AddFinalizer adds specified finalizer string to object
func AddFinalizerIfNotPresent(ctx context.Context,
	generalClient client.Client,
	object client.Object,
	finalizer string) (bool, error) {

	if utils.Contains(object.GetFinalizers(), finalizer) {
		return false, nil
	}

	klog.V(2).Infof("Object does not have finalizer, Object: %s", object.GetName())
	copy := object.DeepCopyObject().(client.Object)

	copy.SetFinalizers(append(copy.GetFinalizers(), finalizer))

	err := generalClient.Update(ctx, copy, &client.UpdateOptions{})
	if err != nil {
		return false, errors.Wrapf(err, "failed to add finalizer %q to %q", finalizer, copy.GetName())
	}

	return true, nil
}

// RemoveFinalizers removes finalizersfrom object
func RemoveFinalizerIfPresent(ctx context.Context,
	generalClient client.Client,
	object client.Object,
	finalizer string) (bool, error) {

	finalizers := utils.Remove(object.GetFinalizers(), finalizer)

	copy := object.DeepCopyObject().(client.Object)
	copy.SetFinalizers(finalizers)

	err := generalClient.Update(ctx, copy, &client.UpdateOptions{})
	if err != nil {
		return false, errors.Wrapf(err, "failed to remove finalizer %q from %q", finalizer, object.GetName())
	}

	return true, nil
}

func AnnotateIfNotPresent(ctx context.Context, generalClient client.Client, object client.Object, key, value string) (bool, error) {

	accessor, err := meta.Accessor(object)
	if err != nil {
		return false, err
	}

	if accessor.GetAnnotations() == nil {
		_, err = AnnotateObject(ctx, generalClient, object, key, value)
		return true, err
	}

	if _, ok := accessor.GetAnnotations()[key]; !ok {
		_, err = AnnotateObject(ctx, generalClient, object, key, value)
		return true, err
	}

	return false, nil
}

func AnnotateObject(ctx context.Context, generalClient client.Client, object client.Object, key, value string) (bool, error) {

	patchData := map[string]interface{}{"metadata": map[string]map[string]string{"annotations": {
		key: value,
	}}}

	patchBytes, err := json.Marshal(patchData)
	if err != nil {
		return false, err
	}

	patch := client.RawPatch(types.MergePatchType, patchBytes)

	err = generalClient.Patch(ctx, object, patch)
	if err != nil {
		return false, err
	}

	return true, err
}

func RemoveAnnotation(ctx context.Context, generalClient client.Client, object client.Object, key string) (bool, error) {

	if _, ok := object.GetAnnotations()[key]; !ok {
		return false, nil
	}

	copy := object.DeepCopyObject().(client.Object)

	delete(copy.GetAnnotations(), key)

	err := generalClient.Update(ctx, copy, &client.UpdateOptions{})

	if err != nil {
		klog.Errorf("Failed to remove annotation: %s from OObject %s, error %v", key, object.GetName(), err)
		return false, err
	}

	return true, nil
}

func UpsertObject(ctx context.Context, generalClient client.Client, object client.Object) error {

	err := generalClient.Create(ctx, object, &client.CreateOptions{})
	if err != nil && !kerrors.IsAlreadyExists(err) {
		return errors.Wrapf(err, "failed to create object %v", object)
	} else if err == nil {
		return nil
	}

	// If we are here, it means the object already exist.
	oldObj := utils.GetInstanceOfObj(object)
	err = generalClient.Get(ctx, client.ObjectKeyFromObject(object), oldObj)
	if err != nil {
		return errors.Wrapf(err, "failed to get the object %v", object)
	}

	if reflect.DeepEqual(oldObj, object) {
		return nil
	}

	patchData, err := utils.GetPatchData(oldObj, object)
	if err != nil {
		return err
	}

	patch := client.RawPatch(types.StrategicMergePatchType, patchData)

	err = generalClient.Patch(ctx, oldObj, patch)
	if err != nil {
		return errors.Wrapf(err, "failed to patch the object %v", object)
	}

	return nil
}

func GetMetadata(resourceName, hash string,
	object client.Object, kind schema.GroupVersionKind) metav1.ObjectMeta {

	labels := GetResourceLables(resourceName, hash, object)

	metadata := metav1.ObjectMeta{
		Name:            resourceName,
		Namespace:       object.GetNamespace(),
		Labels:          labels,
		Annotations:     object.GetAnnotations(),
		OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(object, kind)},
	}

	return metadata
}

func GetResourceLables(resourceName, hash string,
	object client.Object) map[string]string {
	labels := GetResourceMatchLables(resourceName, object)

	utils.AddLabel(labels, map[string]string{
		OrchestHashLabelKey: hash,
	})

	return labels
}

func GetResourceMatchLables(resourceName string, object client.Object) map[string]string {
	labels := GetOrchestMatchLabels(object)

	utils.AddLabel(labels, map[string]string{
		ComponentLabelKey: resourceName,
	})

	if _, ok := labels[OrchestHashLabelKey]; ok {
		delete(labels, OrchestHashLabelKey)
	}

	return labels
}

func GetOrchestMatchLabels(object client.Object) map[string]string {
	labels := utils.CloneAndAddLabel(object.GetLabels(), map[string]string{
		OwnerLabelKey:         object.GetName(),
		ControllerPartOfLabel: "orchest",
	})
	return labels
}

func GetOrchestLabelSelector(object client.Object) (labels.Selector, error) {

	labels := GetOrchestMatchLabels(object)
	labelSelector := metav1.SetAsLabelSelector(labels)

	selector, err := metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get Selector from LabelSelector")
	}

	return selector, nil
}

func GetNamespacedRbacManifest(metadata metav1.ObjectMeta) []client.Object {
    sa := &corev1.ServiceAccount{ObjectMeta: metadata}

    role := &rbacv1.ClusterRole{
        ObjectMeta: metadata,
        Rules: []rbacv1.PolicyRule{{
            APIGroups: []string{"*"},
            Resources: []string{"*"},
            Verbs:     []string{"*"},
        }},
    }

	rbMeta := *metadata.DeepCopy()
	rbMeta.Name = fmt.Sprintf("%s-%s", metadata.Namespace, metadata.Name)
	
    roleBinding := &rbacv1.ClusterRoleBinding{
        ObjectMeta: rbMeta,
        Subjects: []rbacv1.Subject{{
            Kind:      "ServiceAccount",
            Name:      metadata.Name,
            Namespace: metadata.Namespace,
        }},
        RoleRef: rbacv1.RoleRef{
            APIGroup: "rbac.authorization.k8s.io",
            Kind:     "ClusterRole",
            Name:     metadata.Name,
        },
    }

    return []client.Object{role, roleBinding, sa}
}


func GetRbacManifest(metadata metav1.ObjectMeta) []client.Object {

	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metadata,
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"*"},
				Resources: []string{"*"},
				Verbs:     []string{"*"},
			},
		},
	}

	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metadata,
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      metadata.Name,
				Namespace: metadata.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Name:     metadata.Name,
			Kind:     "ClusterRole",
		},
	}

	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metadata,
	}

	return []client.Object{
		clusterRole,
		clusterRoleBinding,
		serviceAccount,
	}
}

// IsComponentReady checks if the OrchestComponent is ready or not
func IsComponentReady(component orchestv1alpha1.OrchestComponent) bool {
	return component.Status != nil && component.Status.Phase == orchestv1alpha1.Running
}

// IsComponentReady checks if the OrchestComponent is ready or not
func IsNginxIngressClass(name string) bool {
	defaultControllers := []string{
		k8sIngressClassController,
		ngxIngressClassController,
	}

	// Check if a custom controller is defined in an environment variable
	customControllers := os.Getenv("CUSTOM_INGRESS_CONTROLLERS")
	if customControllers != "" {
		// Add custom ingress controllers to the default list
		defaultControllers = append(defaultControllers, strings.Split(customControllers, ",")...)
	}
	for _, controller := range defaultControllers {
		if name == controller {
			return true
		}
	}

	return false
}
