package orchestcluster

import (
	"fmt"
	"reflect"
	"strings"
	"time"
	"github.com/orchest/orchest/services/orchest-controller/pkg/addons"
	orchestv1alpha1 "github.com/orchest/orchest/services/orchest-controller/pkg/apis/orchest/v1alpha1"
	"github.com/orchest/orchest/services/orchest-controller/pkg/client/clientset/versioned"
	orchestinformers "github.com/orchest/orchest/services/orchest-controller/pkg/client/informers/externalversions/orchest/v1alpha1"
	orchestlisters "github.com/orchest/orchest/services/orchest-controller/pkg/client/listers/orchest/v1alpha1"
	"github.com/orchest/orchest/services/orchest-controller/pkg/controller"
	"github.com/orchest/orchest/services/orchest-controller/pkg/utils"
	"github.com/orchest/orchest/services/orchest-controller/pkg/version"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	OrchestClusterKind = orchestv1alpha1.SchemeGroupVersion.WithKind("OrchestCluster")
)

type ControllerConfig struct {
	// Postgres default image to use if not provided
	PostgresDefaultImage string
	// Rabbitmq default image to use if not provided
	RabbitmqDefaultImage string
	// Orchest Default version to use if not provided
	OrchestDefaultVersion string
	// celery-worker default image to use if not provided
	CeleryWorkerImageName string
	// orchest-api default image to use if not provided
	OrchestApiImageName string
	// orchest-webserver default image to use if not provided
	OrchestWebserverImageName string
	// auth-server default image to use if not provided
	AuthServerImageName string
	// Default storage class when not specified for a volume.
	DefaultStorageClass string
	// Orchest state default volume size if not specified.
	OrchestStateDefaultVolumeSize string
	// User-dir default volume size if not specified.
	UserdirDefaultVolumeSize string
	// build-dir volume size if not provided
	BuilddirDefaultVolumeSize string
	// default env vars for orchest services
	OrchestDefaultEnvVars map[string]string
	// default env vars for orchest-api
	OrchestApiDefaultEnvVars map[string]string
	// default env vars for orchest-webserver
	OrchestWebserverDefaultEnvVars map[string]string
	// default env vars for auth-server
	AuthServerDefaultEnvVars map[string]string
	// default env vars for celery-worker
	CeleryWorkerDefaultEnvVars map[string]string
	// default env vars for orchest-database
	OrchestDatabaseDefaultEnvVars map[string]string
	// default env vars for rabbitmq
	RabbitmqDefaultEnvVars map[string]string
	// default third-party applications
	DefaultApplications []orchestv1alpha1.ApplicationSpec
	Threadiness         int
	InCluster           bool
	DefaultPause        bool
}

func NewDefaultControllerConfig() ControllerConfig {
	return ControllerConfig{
		PostgresDefaultImage:          "postgres:13.1",
		RabbitmqDefaultImage:          "rabbitmq:3",
		OrchestDefaultVersion:         version.Version,
		CeleryWorkerImageName:         "orchest/celery-worker",
		OrchestApiImageName:           "orchest/orchest-api",
		OrchestWebserverImageName:     "orchest/orchest-webserver",
		AuthServerImageName:           "orchest/auth-server",
		UserdirDefaultVolumeSize:      "50Gi",
		OrchestStateDefaultVolumeSize: "5Gi",
		BuilddirDefaultVolumeSize:     "25Gi",
		OrchestDefaultEnvVars: map[string]string{
			"PYTHONUNBUFFERED":  "TRUE",
			"ORCHEST_LOG_LEVEL": "INFO",
			"ORCHEST_HOST_GID":  "1",
		},
		OrchestApiDefaultEnvVars: map[string]string{
			"ORCHEST_GPU_ENABLED_INSTANCE": "FALSE",
			"FLASK_ENV":                    "production",
			"JUPYTER_SERVER_IMAGE":         "docker.io/orchest/jupyter-server",
		},
		OrchestWebserverDefaultEnvVars: map[string]string{
			"ORCHEST_GPU_ENABLED_INSTANCE": "FALSE",
			"FLASK_ENV":                    "production",
			"ORCHEST_PORT":                 "8000",
			"USERDIR_PVC":                  "userdir-pvc",
			"HOST_CONFIG_DIR":              "/var/lib/orchest/config",
			"HOST_REPO_DIR":                "/var/lib/orchest/repo",
			"HOST_OS":                      "linux",
		},
		AuthServerDefaultEnvVars: map[string]string{
			"FLASK_ENV": "production",
		},
		CeleryWorkerDefaultEnvVars: map[string]string{
			"ORCHEST_GPU_ENABLED_INSTANCE": "FALSE",
			"FLASK_ENV":                    "production",
			"CELERY_STATE_DIR":             "/userdir/.orchest",
		},
		OrchestDatabaseDefaultEnvVars: map[string]string{
			"PGDATA":                    controller.DBMountPath,
			"POSTGRES_HOST_AUTH_METHOD": "trust",
		},
		DefaultApplications: []orchestv1alpha1.ApplicationSpec{
			{
				Name: addons.ArgoWorkflow,
				Config: orchestv1alpha1.ApplicationConfig{
					Helm: &orchestv1alpha1.ApplicationConfigHelm{
						Parameters: []orchestv1alpha1.HelmParameter{
							{
								Name:  "singleNamespace",
								Value: "true",
							},
						},
					},
				},
			},
			{
				Name: addons.DockerRegistry,
			},
		},
		RabbitmqDefaultEnvVars: make(map[string]string, 0),
		Threadiness:            1,
		InCluster:              true,
		DefaultPause:           false,
	}
}

// OrchestClusterController reconciles OrchestCluster CRD.
type OrchestClusterController struct {
	*controller.Controller[*orchestv1alpha1.OrchestCluster]

	oClient versioned.Interface

	gClient client.Client

	scheme *runtime.Scheme

	config ControllerConfig

	k8sDistro utils.KubernetesDistros

	oClusterLister orchestlisters.OrchestClusterLister

	oComponentLister orchestlisters.OrchestComponentLister

	addonManager *addons.AddonManager
}

// NewOrchestClusterController returns a new *OrchestClusterController.
func NewOrchestClusterController(kClient kubernetes.Interface,
	oClient versioned.Interface,
	gClient client.Client,
	scheme *runtime.Scheme,
	config ControllerConfig,
	k8sDistro utils.KubernetesDistros,
	oClusterInformer orchestinformers.OrchestClusterInformer,
	oComponentInformer orchestinformers.OrchestComponentInformer,
	addonManager *addons.AddonManager,
) *OrchestClusterController {

	informerSyncedList := make([]cache.InformerSynced, 0)

	ctrl := controller.NewController[*orchestv1alpha1.OrchestCluster](
		"orchest-cluster",
		1,
		kClient,
		&OrchestClusterKind,
	)

	occ := OrchestClusterController{
		oClient:      oClient,
		gClient:      gClient,
		scheme:       scheme,
		config:       config,
		k8sDistro:    k8sDistro,
		addonManager: addonManager,
	}

	// OrchestCluster event handlers
	oClusterInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    occ.addOrchestCluster,
		UpdateFunc: occ.updateOrchestCluster,
		DeleteFunc: occ.deleteOrchestCluster,
	})
	informerSyncedList = append(informerSyncedList, oClusterInformer.Informer().HasSynced)
	occ.oClusterLister = oClusterInformer.Lister()

	// OrchestComponent event handlers
	oComponentWatcher := controller.NewControlleeWatcher[*orchestv1alpha1.OrchestComponent](ctrl)
	oComponentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    oComponentWatcher.AddObject,
		UpdateFunc: oComponentWatcher.UpdateObject,
		DeleteFunc: oComponentWatcher.DeleteObject,
	})
	informerSyncedList = append(informerSyncedList, oComponentInformer.Informer().HasSynced)
	occ.oComponentLister = oComponentInformer.Lister()

	ctrl.InformerSyncedList = informerSyncedList
	ctrl.SyncHandler = occ.syncOrchestCluster
	ctrl.ControleeGetter = occ.getOrchestCluster

	occ.Controller = ctrl

	if occ.config.DefaultStorageClass == "" {
		sc, err := detectDefaultStorageClass(context.Background(), kClient)
		if err == nil {
			occ.config.DefaultStorageClass = sc
		}
	}

	return &occ
}

func (occ *OrchestClusterController) addOrchestCluster(obj interface{}) {
	oc := obj.(*orchestv1alpha1.OrchestCluster)
	klog.V(4).Infof("Adding OrchestCluster %s", oc.Name)
	occ.Enqueue(oc)
}

func (occ *OrchestClusterController) updateOrchestCluster(cur, old interface{}) {
	oldOc := old.(*orchestv1alpha1.OrchestCluster)
	curOc := cur.(*orchestv1alpha1.OrchestCluster)

	if curOc.UID != oldOc.UID {
		key, err := controller.KeyFunc(oldOc)
		if err != nil {
			utilruntime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", oldOc, err))
			return
		}
		occ.deleteOrchestCluster(cache.DeletedFinalStateUnknown{
			Key: key,
			Obj: oldOc,
		})
	}

	klog.V(4).Infof("Updating OrchestCluster %s", oldOc.Name)
	occ.Enqueue(curOc)
}

func (occ *OrchestClusterController) deleteOrchestCluster(obj interface{}) {
	oc, ok := obj.(*orchestv1alpha1.OrchestCluster)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		oc, ok = tombstone.Obj.(*orchestv1alpha1.OrchestCluster)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("tombstone contained object that is not a OrchestCluster %#v", obj))
			return
		}
	}
	klog.V(4).Infof("Deleting OrchestCluster %s", oc.Name)

	occ.Enqueue(oc)
}

func (occ *OrchestClusterController) getOrchestCluster(namespace, name string) (
	interface{}, error) {
	return occ.oClusterLister.OrchestClusters(namespace).Get(name)
}

func (occ *OrchestClusterController) syncOrchestCluster(ctx context.Context, key string) error {

	startTime := time.Now()
	klog.V(3).Infof("Started syncing OrchestCluster: %s.", key)
	defer func() {
		klog.V(3).Infof("Finished syncing OrchestCluster: %s. duration: (%v)", key, time.Since(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	orchest, err := occ.oClient.OrchestV1alpha1().OrchestClusters(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			klog.V(2).Info("OrchestCluster %s resource not found.", key)
			return nil
		}
		// Error reading OrchestCluster - The request will be requeued.
		return errors.Wrapf(err, "failed to get to OrchestCluster %s", key)
	}

	if !orchest.GetDeletionTimestamp().IsZero() {
		// The cluster is deleted, we need to stop it first, then remove the finalizer
		stopped, err := occ.stopOrchest(ctx, orchest)
		if err != nil {
			return nil
		}

		if stopped {
			_, err = controller.RemoveFinalizerIfPresent(ctx, occ.gClient, orchest, orchestv1alpha1.Finalizer)
			return err
		}
		return nil
	}

	// Set a finalizer so we can do cleanup before the object goes away
	changed, err := controller.AddFinalizerIfNotPresent(ctx, occ.gClient, orchest, orchestv1alpha1.Finalizer)
	if changed || err != nil {
		return err
	}

	// Add kubernetes distro annotation
	changed, err = controller.AnnotateIfNotPresent(ctx, occ.gClient, orchest, controller.K8sDistroAnnotationKey, string(occ.k8sDistro))
	if changed || err != nil {
		return err
	}

	// If Status struct is not initialized yet, the cluster is new, create it
	if orchest.Status == nil {
		// Set the default values in CR if not specified
		err = occ.updatePhase(ctx, namespace, name, orchestv1alpha1.Initializing, "Initializing Orchest Cluster")
		if err != nil {
			klog.Error(err)
			return err
		}

	}

	ok, err := occ.validateOrchestCluster(ctx, orchest)
	if err != nil {
		klog.Error(err)
		return err
	}

	if !ok {
		err = occ.updatePhase(ctx, namespace, name, orchestv1alpha1.Error, fmt.Sprintf("OrchestCluster object is not valid %s", key))
		if err != nil {
			klog.Error(err)
			return err
		}
	}

	changed, err = occ.setDefaultIfNotSpecified(ctx, orchest)
	if err != nil {
		return err
	}

	// If the object is changed, we return and the object will be requeued.
	if changed {
		return nil
	}

	return occ.manageOrchestCluster(ctx, orchest)
}

func (occ *OrchestClusterController) validateOrchestCluster(ctx context.Context, orchest *orchestv1alpha1.OrchestCluster) (bool, error) {

	var err error
	if occ.config.DefaultStorageClass != "" {
		_, err := occ.Client().StorageV1().StorageClasses().Get(ctx, occ.config.DefaultStorageClass, metav1.GetOptions{})
		if err != nil && kerrors.IsNotFound(err) {
			return false, nil
		}
	}

	// EKS_TODO: temporarily disabled because it will lead to an "Error"
	// state during installation if the storage class is being installed
	// during the Orchest installation by an addon.
	// if orchest.Spec.Orchest.Resources.UserDirVolume != nil {
	// 	_, err := occ.Client().StorageV1().StorageClasses().Get(ctx, orchest.Spec.Orchest.Resources.UserDirVolume.StorageClass, metav1.GetOptions{})
	// 	if err != nil && kerrors.IsNotFound(err) {
	// 		return false, nil
	// 	}
	// }

	// if orchest.Spec.Orchest.Resources.OrchestStateVolume != nil {
	// 	_, err := occ.Client().StorageV1().StorageClasses().Get(ctx, orchest.Spec.Orchest.Resources.OrchestStateVolume.StorageClass, metav1.GetOptions{})
	// 	if err != nil && kerrors.IsNotFound(err) {
	// 		return false, nil
	// 	}
	// }

	// Detect runtime environment
	runtime, socketPath, err := detectContainerRuntime(ctx, occ.Client(), orchest)
	if err != nil {
		return false, err
	}
	occ.config.OrchestDefaultEnvVars["CONTAINER_RUNTIME"] = runtime
	// Some k8s flavours place "unix://" in front of the path.
	occ.config.OrchestDefaultEnvVars["CONTAINER_RUNTIME_SOCKET"] = strings.Replace(socketPath, "unix://", "", -1)
	occ.config.OrchestDefaultEnvVars["CONTAINER_RUNTIME_IMAGE"] = utils.GetFullImageName("docker.io",
		"image-puller", occ.config.OrchestDefaultVersion)
	occ.config.OrchestDefaultEnvVars["K8S_DISTRO"] = string(occ.k8sDistro)

	return true, err
}

func (occ *OrchestClusterController) setDefaultIfNotSpecified(ctx context.Context,
	orchest *orchestv1alpha1.OrchestCluster) (bool, error) {

	orchest, err := occ.oClient.OrchestV1alpha1().OrchestClusters(orchest.Namespace).Get(ctx, orchest.Name, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			klog.V(2).Info("OrchestCluster %s resource not found.", orchest.Name)
			return false, nil
		}
		// Error reading OrchestCluster - The request will be requeued.
		return false, errors.Wrap(err, "failed to get OrchestCluster")
	}

	if orchest.Status.ObservedGeneration == orchest.Generation {
		return false, nil
	}

	copy := orchest.DeepCopy()

	changed := false

	var controlNodeSelector = copy.Spec.ControlNodeSelector
	if controlNodeSelector == nil {
		controlNodeSelector = map[string]string{}
	}

	// Orchest configs
	if copy.Spec.Orchest.Version == "" {
		changed = true
		copy.Spec.Orchest.Version = occ.config.OrchestDefaultVersion
	}

	if copy.Spec.SingleNode == nil {
		changed = true
		True := true
		copy.Spec.SingleNode = &True
	}

	if copy.Spec.Orchest.Pause == nil {
		changed = true
		copy.Spec.Orchest.Pause = &occ.config.DefaultPause
	}

	if copy.Spec.Orchest.Env == nil {
		changed = true
		copy.Spec.Orchest.Env = make([]corev1.EnvVar, 0, len(occ.config.OrchestDefaultEnvVars))
	}

	envChanged := utils.UpsertEnvVariable(&copy.Spec.Orchest.Env,
		occ.config.OrchestDefaultEnvVars, false)
	changed = changed || envChanged

	envChanged = utils.UpsertEnvVariable(&copy.Spec.Orchest.Env,
		map[string]string{
			"ORCHEST_CLUSTER":   orchest.Name,
			"ORCHEST_NAMESPACE": orchest.Namespace,
		}, false)
	changed = changed || envChanged

	if copy.Spec.Orchest.OrchestHost != nil {
		envChanged := utils.UpsertEnvVariable(&copy.Spec.Orchest.Env,
			map[string]string{"ORCHEST_FQDN": *copy.Spec.Orchest.OrchestHost}, true)
		changed = changed || envChanged
	}

	// Orchest-API configs
	newImage, update := isUpdateRequired(copy, controller.OrchestApi, copy.Spec.Orchest.OrchestApi.Image)
	if update {
		changed = true
		copy.Spec.Orchest.OrchestApi.Image = newImage
	}

	if copy.Spec.Orchest.OrchestApi.Env == nil {
		changed = true
		copy.Spec.Orchest.OrchestApi.Env = make([]corev1.EnvVar, 0, len(occ.config.OrchestApiDefaultEnvVars))
	}

	envChanged = utils.UpsertEnvVariable(&copy.Spec.Orchest.OrchestApi.Env,
		occ.config.OrchestApiDefaultEnvVars, false)
	envChanged = utils.UpsertEnvVariable(
		&copy.Spec.Orchest.OrchestApi.Env,
		map[string]string{"SINGLE_NODE": strings.ToUpper(fmt.Sprintf("%t", *copy.Spec.SingleNode))},
		true,
	) || envChanged
	changed = changed || envChanged

	// Orchest-Webserver configs
	newImage, update = isUpdateRequired(copy, controller.OrchestWebserver, copy.Spec.Orchest.OrchestWebServer.Image)
	if update {
		changed = true
		copy.Spec.Orchest.OrchestWebServer.Image = newImage
	}

	if copy.Spec.Orchest.OrchestWebServer.Env == nil {
		changed = true
		copy.Spec.Orchest.OrchestWebServer.Env = make([]corev1.EnvVar, 0, len(occ.config.OrchestWebserverDefaultEnvVars))
	}

	envChanged = utils.UpsertEnvVariable(&copy.Spec.Orchest.OrchestWebServer.Env,
		occ.config.OrchestWebserverDefaultEnvVars, false)
	changed = changed || envChanged

	// Celery-Worker configs
	newImage, update = isUpdateRequired(copy, controller.CeleryWorker, copy.Spec.Orchest.CeleryWorker.Image)
	if update {
		changed = true
		copy.Spec.Orchest.CeleryWorker.Image = newImage
	}

	if copy.Spec.Orchest.CeleryWorker.Env == nil {
		changed = true
		copy.Spec.Orchest.CeleryWorker.Env = make([]corev1.EnvVar, 0, len(occ.config.CeleryWorkerDefaultEnvVars))
	}

	envChanged = utils.UpsertEnvVariable(&copy.Spec.Orchest.CeleryWorker.Env,
		occ.config.CeleryWorkerDefaultEnvVars, false)
	envChanged = utils.UpsertEnvVariable(
		&copy.Spec.Orchest.CeleryWorker.Env,
		map[string]string{"SINGLE_NODE": strings.ToUpper(fmt.Sprintf("%t", *copy.Spec.SingleNode))},
		true,
	) || envChanged
	changed = changed || envChanged

	// Auth-Server configs
	newImage, update = isUpdateRequired(copy, controller.AuthServer, copy.Spec.Orchest.AuthServer.Image)
	if update {
		changed = true
		copy.Spec.Orchest.AuthServer.Image = newImage
	}

	if copy.Spec.Orchest.AuthServer.Env == nil {
		changed = true
		copy.Spec.Orchest.AuthServer.Env = make([]corev1.EnvVar, 0, len(occ.config.AuthServerDefaultEnvVars))
	}

	envChanged = utils.UpsertEnvVariable(&copy.Spec.Orchest.AuthServer.Env,
		occ.config.AuthServerDefaultEnvVars, false)
	if envChanged {
		changed = true
	}

	newImage, update = isUpdateRequired(copy, controller.NodeAgent, copy.Spec.Orchest.NodeAgent.Image)
	if update {
		changed = true
		copy.Spec.Orchest.NodeAgent.Image = newImage
	}

	// Postgres configs
	if copy.Spec.Postgres.Image == "" {
		changed = true
		copy.Spec.Postgres.Image = occ.config.PostgresDefaultImage
	}

	if copy.Spec.Postgres.Env == nil && len(occ.config.OrchestDatabaseDefaultEnvVars) != 0 {
		changed = true
		copy.Spec.Postgres.Env = utils.GetEnvVarFromMap(occ.config.OrchestDatabaseDefaultEnvVars)
	}

	// RabbitMq configs
	if copy.Spec.RabbitMq.Image == "" {
		changed = true
		copy.Spec.RabbitMq.Image = occ.config.RabbitmqDefaultImage
	}

	if copy.Spec.RabbitMq.Env == nil && len(occ.config.RabbitmqDefaultEnvVars) != 0 {
		changed = true
		copy.Spec.RabbitMq.Env = utils.GetEnvVarFromMap(occ.config.RabbitmqDefaultEnvVars)
	}

	if copy.Spec.Orchest.Resources.UserDirVolumeSize == "" {
		changed = true
		copy.Spec.Orchest.Resources.UserDirVolumeSize = occ.config.UserdirDefaultVolumeSize
	}

	if copy.Spec.Orchest.Resources.OrchestStateVolumeSize == "" {
		changed = true
		copy.Spec.Orchest.Resources.OrchestStateVolumeSize = occ.config.OrchestStateDefaultVolumeSize
	}

	// If not specified assume that we are in the mode which splits
	// userdir and orchest state.
	if copy.Spec.Orchest.Resources.SeparateOrchestStateFromUserDir {

		if copy.Spec.Orchest.Resources.UserDirVolume == nil {
			changed = true
			copy.Spec.Orchest.Resources.UserDirVolume = &orchestv1alpha1.Volume{
				StorageClass: occ.config.DefaultStorageClass,
				VolumeSize:   copy.Spec.Orchest.Resources.UserDirVolumeSize,
				MountPath:    controller.UserdirMountPath,
			}
		} else {
			if copy.Spec.Orchest.Resources.UserDirVolume.VolumeSize == "" {
				changed = true
				copy.Spec.Orchest.Resources.UserDirVolume.VolumeSize = copy.Spec.Orchest.Resources.UserDirVolumeSize
			}
			if copy.Spec.Orchest.Resources.UserDirVolume.StorageClass == "" {
				changed = true
				copy.Spec.Orchest.Resources.UserDirVolume.StorageClass = occ.config.DefaultStorageClass
			}
		}

		if copy.Spec.Orchest.Resources.OrchestStateVolume == nil {
			changed = true
			copy.Spec.Orchest.Resources.OrchestStateVolume = &orchestv1alpha1.Volume{
				StorageClass: occ.config.DefaultStorageClass,
				VolumeSize:   copy.Spec.Orchest.Resources.OrchestStateVolumeSize,
				MountPath:    controller.OrchestStateMountPath,
			}
		} else {
			if copy.Spec.Orchest.Resources.OrchestStateVolume.VolumeSize == "" {
				changed = true
				copy.Spec.Orchest.Resources.OrchestStateVolume.VolumeSize = copy.Spec.Orchest.Resources.OrchestStateVolumeSize
			}
			if copy.Spec.Orchest.Resources.OrchestStateVolume.StorageClass == "" {
				changed = true
				copy.Spec.Orchest.Resources.OrchestStateVolume.StorageClass = occ.config.DefaultStorageClass
			}
		}
	}

	if copy.Spec.Applications == nil {
		changed = true
		copy.Spec.Applications = occ.config.DefaultApplications
	}

	// BuildKitDaemon configs
	buildKitDaemonImage := utils.GetFullImageName(copy.Spec.Orchest.Registry, controller.BuildKitDaemon, copy.Spec.Orchest.Version)
	if copy.Spec.Orchest.BuildKitDaemon.Image != buildKitDaemonImage {
		changed = true
		copy.Spec.Orchest.BuildKitDaemon.Image = buildKitDaemonImage
	}

	if isIngressDisabled(copy) {
		ingressEnabled := false
		applications := make([]orchestv1alpha1.ApplicationSpec, 0, 0)
		for _, app := range copy.Spec.Applications {
			if app.Name == addons.IngressNginx {
				ingressEnabled = true
				continue
			}
			applications = append(applications, app)
		}

		// If ingress is enabled, it has to be removed from the .Spec.Applications,
		// k8s ensures we are using the latest object version.
		if ingressEnabled {
			copy.Spec.Applications = applications
			changed = true
		}
	} else if isIngressAddonRequired(ctx, occ.k8sDistro, occ.Client()) {
		ingressEnabled := false
		for _, app := range copy.Spec.Applications {
			if app.Name == addons.IngressNginx {
				ingressEnabled = true
			}
		}

		if !ingressEnabled {
			copy.Spec.Applications = append(copy.Spec.Applications, orchestv1alpha1.ApplicationSpec{
				Name: addons.IngressNginx,
			})
			changed = true
		}
	}

	for i := 0; i < len(copy.Spec.Applications); i++ {
		app := &copy.Spec.Applications[i]
		// Docker-registry config.
		if app.Name == addons.DockerRegistry {

			registryIpChanged, err := setRegistryServiceIP(ctx, occ.Client(), copy.Namespace, app)
			if err != nil {
				klog.Error(err)
				return changed, err
			}

			registryNodeSelectorChanged, err := setHelmParamNodeSelector(
				ctx, occ.Client(), copy.Namespace,
				app, controlNodeSelector, "")

			if err != nil {
				klog.Error(err)
				return changed, err
			}

			changed = changed || registryIpChanged || registryNodeSelectorChanged
			// Argo config.
		} else if app.Name == addons.ArgoWorkflow {
			argoNodeSelectorChanged, err := setHelmParamNodeSelector(
				ctx, occ.Client(), copy.Namespace,
				app, controlNodeSelector, "controller.")

			if err != nil {
				klog.Error(err)
				return changed, err
			}

			changed = changed || argoNodeSelectorChanged
		} else if app.Name == addons.IngressNginx {
			nginxNodeSelectorChanged, err := setHelmParamNodeSelector(
				ctx, occ.Client(), copy.Namespace,
				app, controlNodeSelector, "controller.")

			if err != nil {
				klog.Error(err)
				return changed, err
			}

			changed = changed || nginxNodeSelectorChanged
		} else if app.Name == addons.EfsCsiDriver {
			efsCsiControllerNodeSelectorChanged, err := setHelmParamNodeSelector(
				ctx, occ.Client(), copy.Namespace,
				app, controlNodeSelector, "controller.")

			if err != nil {
				klog.Error(err)
				return changed, err
			}

			changed = changed || efsCsiControllerNodeSelectorChanged
		}
	}

	if changed || !reflect.DeepEqual(copy.Spec, orchest.Spec) {
		_, err := occ.oClient.OrchestV1alpha1().OrchestClusters(orchest.Namespace).Update(ctx, copy, metav1.UpdateOptions{})

		if err != nil {
			return false, errors.Wrapf(err, "failed to update orchest with default values  %q", orchest.Name)
		}

		klog.Infof("OrchestCluster is updated with default values %s", orchest.Name)
		return true, nil

	}
	return false, nil
}

func (occ *OrchestClusterController) ensureThirdPartyDependencies(ctx context.Context,
	orchest *orchestv1alpha1.OrchestCluster) (err error) {

	updateConditionPreInstall := func(app *orchestv1alpha1.ApplicationSpec) error {
		err = occ.updateCondition(ctx, orchest.Namespace, orchest.Name,
			orchestv1alpha1.OrchestClusterEvent(fmt.Sprintf("Deploying %s", app.Name)))
		if err != nil {
			klog.Error(err)
			return err
		}
		return nil
	}

	registryPreInstall := func(app *orchestv1alpha1.ApplicationSpec) error {
		err = occ.updateCondition(ctx, orchest.Namespace, orchest.Name, orchestv1alpha1.CreatingCertificates)
		if err != nil {
			klog.Error(err)
			return err
		}

		if err != nil {
			return errors.Wrapf(err, "failed to update orchest with registry service ip  %q", orchest.Name)
		}

		serviceIP, err := getRegistryServiceIP(&app.Config)
		if err != nil {
			return err
		}

		err = registryCertgen(ctx, occ.Client(), serviceIP, orchest)
		if err != nil {
			klog.Error(err)
			return err
		}

		return nil
	}

	for _, application := range orchest.Spec.Applications {
		preInstallHooks := []addons.PreInstallHookFn{
			updateConditionPreInstall,
		}

		namespace := orchest.Namespace
		if application.Name == addons.DockerRegistry {
			preInstallHooks = append(preInstallHooks, registryPreInstall)
		} else if application.Name == addons.EfsCsiDriver {
			namespace = "kube-system"
		}

		err = occ.addonManager.Get(application.Name).Enable(ctx, preInstallHooks, namespace, &application)
		if err != nil {
			klog.Error(err)
			return err
		}

	}

	return nil

}

// Installs deployer if the config is changed
func (occ *OrchestClusterController) manageOrchestCluster(ctx context.Context, orchest *orchestv1alpha1.OrchestCluster) (err error) {

	if orchest.Status == nil {
		return errors.Errorf("status object is not initialzed yet %s", orchest.Name)
	}
	stopped := false
	nextPhase, endPhase := determineNextPhase(orchest)

	if nextPhase == endPhase {
		return
	}

	err = occ.updatePhase(ctx, orchest.Namespace, orchest.Name, nextPhase, "")
	defer func() {
		if err == nil && (stopped || endPhase == orchestv1alpha1.DeployedThirdParties) {
			err = occ.updatePhase(ctx, orchest.Namespace, orchest.Name, endPhase, "")
		}
	}()
	if err != nil {
		return err
	}

	// we should check for third parties to see if they need to be updated
	err = occ.ensureThirdPartyDependencies(ctx, orchest)
	if err != nil {
		return err
	}

	// If endPhase is DeployedThirdParties return early to update phase
	if endPhase == orchestv1alpha1.DeployedThirdParties {
		return
	}

	// If endPhase is Paused or nextPhase is Pausing the cluster should be paused first
	if endPhase == orchestv1alpha1.Stopped || nextPhase == orchestv1alpha1.Stopping {
		stopped, err = occ.stopOrchest(ctx, orchest)
		return err
	}

	generation := fmt.Sprint(orchest.Generation)

	// This is to avoid breaking changes w.r.t. the previous setup where
	// the userdir and Orchest state were in the same pvc.
	if !orchest.Spec.Orchest.Resources.SeparateOrchestStateFromUserDir {
		var madeUpVolume = orchestv1alpha1.Volume{
			VolumeSize: orchest.Spec.Orchest.Resources.UserDirVolumeSize,
		}

		// The specification of access mode here is also to avoid
		// breaking changes.
		accessMode := corev1.ReadWriteMany
		if orchest.Spec.SingleNode != nil && *orchest.Spec.SingleNode {
			accessMode = corev1.ReadWriteOnce
		}

		err = occ.ensurePvc(ctx, generation, controller.UserDirName,
			madeUpVolume, accessMode, orchest)
	} else {
		// TODO: the choice of ReadWriteMany here pretty much forces the
		// resource backing the pvc to support that, perhaps it should
		// be specifiable or detected automatically based on the storage
		// class?
		if orchest.Spec.Orchest.Resources.UserDirVolume != nil {
			err = occ.ensurePvc(ctx, generation, controller.UserDirName,
				*orchest.Spec.Orchest.Resources.UserDirVolume,
				corev1.ReadWriteMany, orchest)

			if err != nil {
				return err
			}
		}

		if orchest.Spec.Orchest.Resources.OrchestStateVolume != nil {
			err = occ.ensurePvc(ctx, generation, controller.OrchestStateVolumeName,
				*orchest.Spec.Orchest.Resources.OrchestStateVolume,
				corev1.ReadWriteOnce,
				orchest)
			if err != nil {
				return err
			}
		}
	}

	err = occ.ensureRbacs(ctx, generation, orchest)
	if err != nil {
		return err
	}

	// Deploy and Update
	components, err := GetOrchestComponents(ctx, orchest, occ.oComponentLister)
	if err != nil {
		return err
	}

	for _, componentName := range orderOfDeployment {
		component, ok := components[componentName]

		// Do not create the buildkit-daemon when not needed.
		if componentName == controller.BuildKitDaemon {
			containerRuntime, _, _ := detectContainerRuntime(ctx, occ.Client(), orchest)
			if containerRuntime != "containerd" && containerRuntime != "cri-o" {
				continue
			}
		}

		if ok {
			// If component is not ready, the key will be requeued to be checked later
			if !controller.IsComponentReady(*component) {
				occ.EnqueueAfter(orchest)
				return
			}
		} else {

			// component does not exist, let't create it
			componentTemplate, err := GetComponentTemplate(componentName, orchest)
			if err != nil {
				return err
			}

			err = occ.updateCondition(ctx, orchest.Namespace, orchest.Name,
				orchestv1alpha1.OrchestClusterEvent(fmt.Sprintf("deploying %s", componentName)))
			if err != nil {
				return errors.Wrapf(err, "failed to update status while changing the state to DeployingOrchest")
			}

			newComponent := getOrchestComponent(componentName, generation, componentTemplate, orchest)

			klog.Infof("New component %s", newComponent)
			// Creating Orchest Component
			_, err = occ.oClient.OrchestV1alpha1().OrchestComponents(orchest.Namespace).
				Create(ctx, newComponent, metav1.CreateOptions{})
			return err
		}
	}

	klog.V(4).Infof("Deleting deprecated PVC %s", controller.OldBuilderDirName)
	occ.deletePvc(ctx, controller.OldBuilderDirName, orchest)

	stopped = true
	return err
}

func (occ *OrchestClusterController) stopOrchest(ctx context.Context, orchest *orchestv1alpha1.OrchestCluster) (bool, error) {

	stopped := false

	// Get the current Components
	components, err := GetOrchestComponents(ctx, orchest, occ.oComponentLister)
	if err != nil {
		return stopped, err
	}

	if len(components) > 0 {
		for i := len(orderOfDeployment) - 1; i >= 0; i-- {
			component, ok := components[orderOfDeployment[i]]

			// component exist
			if ok {
				// If component is already deleted we wait for OrchestComponentController to gracefully delete it
				// and we requeue the OrchestCluster for continution
				if !component.GetDeletionTimestamp().IsZero() {
					occ.EnqueueAfter(orchest)
					return stopped, nil
				} else {
					// The Component is not deleted, we will delete it
					return stopped, occ.oClient.OrchestV1alpha1().OrchestComponents(orchest.Namespace).
						Delete(ctx, component.Name, metav1.DeleteOptions{})
				}
			}
		}
	}
	// The cluster is paused, remove the restart annotation if present
	_, err = controller.RemoveAnnotation(ctx, occ.gClient, orchest, controller.RestartAnnotationKey)

	stopped = true
	return stopped, err
}

func (occ *OrchestClusterController) ensurePvc(ctx context.Context, curHash, name string,
	volume orchestv1alpha1.Volume, am corev1.PersistentVolumeAccessMode, orchest *orchestv1alpha1.OrchestCluster) error {

	// Retrive the created pvcs
	oldPvc, err := occ.Client().CoreV1().PersistentVolumeClaims(orchest.Namespace).Get(ctx, name, metav1.GetOptions{})
	newPvc := getPersistentVolumeClaim(name, curHash, volume, am, orchest)
	// userdir is not created or is removed, we have to recreate it
	if err != nil && kerrors.IsNotFound(err) {
		_, err := occ.Client().CoreV1().PersistentVolumeClaims(orchest.Namespace).Create(ctx, newPvc, metav1.CreateOptions{})
		if err != nil {
			return errors.Wrapf(err, "failed to create %s pvc", name)
		}
		return nil
	} else if err != nil {
		return err
	}

	return occ.adoptPVC(ctx, oldPvc, newPvc)

}

func (occ *OrchestClusterController) deletePvc(ctx context.Context, name string, orchest *orchestv1alpha1.OrchestCluster) error {

	err := occ.Client().CoreV1().PersistentVolumeClaims(orchest.Namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			klog.Infof("PVC %s not found, nothing to delete.", name)
			return nil
		}
		klog.Errorf("Failed to delete PVC %s.", name)
		return err
	}
	return nil
}

func (occ *OrchestClusterController) adoptPVC(ctx context.Context, oldPvc, newPvc *corev1.PersistentVolumeClaim) error {

	if oldPvc.OwnerReferences == nil || !reflect.DeepEqual(oldPvc.OwnerReferences[0], newPvc.OwnerReferences[0]) {
		oldPvc.OwnerReferences = newPvc.OwnerReferences
		_, err := occ.Client().CoreV1().PersistentVolumeClaims(oldPvc.Namespace).Update(ctx, oldPvc, metav1.UpdateOptions{})
		return err
	}
	return nil
}

func (occ *OrchestClusterController) ensureRbacs(ctx context.Context, hash string, orchest *orchestv1alpha1.OrchestCluster) error {

	objects := make([]client.Object, 0, 6)
	apiMetadata := controller.GetMetadata(controller.OrchestApi, hash, orchest, OrchestClusterKind)
	// Get the rbac manifest
	objects = append(objects, controller.GetNamespacedRbacManifest(apiMetadata)...)

	celeryMetadata := controller.GetMetadata(controller.CeleryWorker, hash, orchest, OrchestClusterKind)
	objects = append(objects, controller.GetNamespacedRbacManifest(celeryMetadata)...)

	for _, obj := range objects {
		err := controller.UpsertObject(ctx, occ.gClient, obj)
		if err != nil {
			return err
		}
	}

	return nil
}

func (occ *OrchestClusterController) updatePhase(ctx context.Context,
	namespace, name string,
	phase orchestv1alpha1.OrchestPhase, reason string) error {

	orchest, err := occ.oClient.OrchestV1alpha1().OrchestClusters(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			klog.V(2).Info("OrchestCluster %s resource not found.", name)
			return nil
		}
		// Error reading OrchestCluster - The request will be requeued.
		return errors.Wrap(err, "failed to get OrchestCluster")
	}

	if orchest.Status != nil && orchest.Status.Phase == phase {
		return nil
	} else if orchest.Status != nil {
		orchest.Status.Phase = phase
		orchest.Status.LastHeartbeatTime = metav1.NewTime(time.Now())
		orchest.Status.Conditions = nil
	} else {
		orchest.Status = &orchestv1alpha1.OrchestClusterStatus{
			Phase:             phase,
			LastHeartbeatTime: metav1.NewTime(time.Now()),
		}

	}

	if orchest.Status.Phase == orchestv1alpha1.Running ||
		orchest.Status.Phase == orchestv1alpha1.Starting ||
		orchest.Status.Phase == orchestv1alpha1.DeployingOrchest ||
		orchest.Status.Phase == orchestv1alpha1.Stopped {

		orchest.Status.ObservedHash = utils.ComputeHash(&orchest.Spec)

		if orchest.Status.Phase != orchestv1alpha1.Starting &&
			orchest.Status.Phase != orchestv1alpha1.DeployingOrchest {
			orchest.Status.ObservedGeneration = orchest.Generation
		}
	}

	_, err = occ.oClient.OrchestV1alpha1().OrchestClusters(orchest.Namespace).UpdateStatus(ctx, orchest, metav1.UpdateOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to update orchest with phase %q", orchest.Status.Phase)
	}

	return nil
}

func (occ *OrchestClusterController) updateCondition(ctx context.Context,
	namespace, name string,
	event orchestv1alpha1.OrchestClusterEvent) error {

	orchest, err := occ.oClient.OrchestV1alpha1().OrchestClusters(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			klog.V(2).Info("OrchestCluster %s resource not found.", name)
			return nil
		}
		// Error reading OrchestCluster - The request will be requeued.
		return errors.Wrap(err, "failed to get OrchestCluster")
	}

	return occ.updateClusterCondition(ctx, orchest, event)
}

// UpdateClusterCondition function will export each condition into the cluster custom resource
func (occ *OrchestClusterController) updateClusterCondition(ctx context.Context,
	orchest *orchestv1alpha1.OrchestCluster,
	event orchestv1alpha1.OrchestClusterEvent) error {

	if orchest.Status == nil {
		return occ.updatePhase(ctx, orchest.Namespace, orchest.Name, orchestv1alpha1.Initializing, "")
	}

	conditions := make([]orchestv1alpha1.Condition, 0, len(orchest.Status.Conditions)+1)
	var currentCondition *orchestv1alpha1.Condition
	for _, condition := range orchest.Status.Conditions {
		if event == condition.Event {
			condition.LastHeartbeatTime = metav1.NewTime(time.Now())
			currentCondition = &condition
		}
		conditions = append(conditions, condition)
	}

	if currentCondition == nil {
		currentCondition = &orchestv1alpha1.Condition{
			Event:              event,
			LastTransitionTime: metav1.NewTime(time.Now()),
			LastHeartbeatTime:  metav1.NewTime(time.Now()),
		}

		conditions = append(conditions, *currentCondition)
	}

	orchest.Status.Conditions = conditions

	_, err := occ.oClient.OrchestV1alpha1().OrchestClusters(orchest.Namespace).UpdateStatus(ctx, orchest, metav1.UpdateOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to update orchest with phase %q", orchest.Status.Phase)
	}

	return nil
}
