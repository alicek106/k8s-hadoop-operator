package hadoopservice

import (
	"context"
	alicek106v1alpha1 "hadoop-operator/pkg/apis/alicek106/v1alpha1"
	"math/rand"
	"strconv"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_hadoopservice")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new HadoopService Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileHadoopService{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("hadoopservice-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource HadoopService
	err = c.Watch(&source.Kind{Type: &alicek106v1alpha1.HadoopService{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner HadoopService
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &alicek106v1alpha1.HadoopService{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileHadoopService implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileHadoopService{}

// ReconcileHadoopService reconciles a HadoopService object
type ReconcileHadoopService struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a HadoopService object and makes changes based on the state read
// and what is in the HadoopService.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileHadoopService) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling HadoopService")

	// Fetch the HadoopService instance
	instance := &alicek106v1alpha1.HadoopService{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("HadoopService resource not found. Ignoring since object must be deleted.")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get HadoopService.")
		return reconcile.Result{}, err
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////////
	/// Create Slave StatefulSet and Service, Secret
	slaveFound := &appsv1.StatefulSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-hadoop-slave", Namespace: instance.Namespace}, slaveFound)

	if err != nil && errors.IsNotFound(err) {
		// 1. Define a new StatefulSet for slaves
		slaveStatefulset := r.statefulSetForSlave(instance)
		reqLogger.Info("Creating a new statefulset and services.", "StatefulSet.Namespace", slaveStatefulset.Namespace, "StatefulSet.Name", slaveStatefulset.Name)
		err = r.client.Create(context.TODO(), slaveStatefulset)
		if err != nil {
			reqLogger.Error(err, "Failed to create new StatefulSet for slaves.", "StatefulSet.Namespace", slaveStatefulset.Namespace, "StatefulSet.Name", slaveStatefulset.Name)
			return reconcile.Result{}, err
		}

		// 2. Define a new Headless Service for slaves
		slaveService := r.serviceForSlave(instance)
		err = r.client.Create(context.TODO(), slaveService)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Service for slaves.", "Service.Namespace", slaveService.Namespace, "Service.Name", slaveService.Name)
			return reconcile.Result{}, err
		}

		// 3. Define a new Secret for master (SSH access password)
		passwordGen := randomString(4)
		log.Info("Generated password is " + passwordGen)
		masterSecret := r.secretForSlave(instance, passwordGen)
		err = r.client.Create(context.TODO(), masterSecret)
		if err != nil {
			reqLogger.Error(err, "Failed to create new secret for master.", "Secret.Namespace", masterSecret.Namespace, "Secret.Name", masterSecret.Name)
			return reconcile.Result{}, err
		}

		// 4. Define a new StatefulSet for master
		masterStatefulset := r.statefulSetForMaster(instance)
		err = r.client.Create(context.TODO(), masterStatefulset)
		if err != nil {
			reqLogger.Error(err, "Failed to create new StatefulSet for master.", "StatefulSet.Namespace", masterStatefulset.Namespace, "StatefulSet.Name", masterStatefulset.Name)
			return reconcile.Result{}, err
		}

		// 5. Define a new Headless Service for master
		masterService := r.serviceForMaster(instance)
		err = r.client.Create(context.TODO(), masterService)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Service for master.", "Service.Namespace", masterService.Namespace, "Service.Name", slaveService.Name)
			return reconcile.Result{}, err
		}

		// 6. Define a new external Service for master
		masterExternalService := r.externalServiceForMaster(instance)
		err = r.client.Create(context.TODO(), masterExternalService)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Service for master.", "Service.Namespace", masterExternalService.Namespace, "Service.Name", masterExternalService.Name)
			return reconcile.Result{}, err
		}

		// All resources are created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get StatefulSet for slaves.")
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// Reference : https://github.com/kubernetes/kubernetes/blob/master/test/e2e/framework/service_util.go
func (r *ReconcileHadoopService) serviceForSlave(h *alicek106v1alpha1.HadoopService) *corev1.Service {
	slaveServiceName := h.Name + "-hadoop-slave-svc"
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      slaveServiceName,
			Namespace: h.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name: "ssh",
				Port: 22,
			}},
			ClusterIP: "None",
			Selector:  labelsForHadoopSlave(h.Name),
		},
	}

	// Register HadoopService instance as the owner and controller of slaves service
	controllerutil.SetControllerReference(h, service, r.scheme)
	return service
}

func (r *ReconcileHadoopService) statefulSetForSlave(h *alicek106v1alpha1.HadoopService) *appsv1.StatefulSet {
	ls := labelsForHadoopSlave(h.Name)
	replicas := h.Spec.ClusterSize

	slaveServiceName := h.Name + "-hadoop-slave-svc"
	slaveStatefulsetName := h.Name + "-hadoop-slave"

	masterServiceName := h.Name + "-hadoop-master-svc"
	masterEndpoint := h.Name + "-hadoop-master-0." + masterServiceName + "." + h.Namespace + ".svc.cluster.local"
	// masterEndpoint example : hadoop-master-0.hadoop-master-svc.default.svc.cluster.local

	statefulset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      slaveStatefulsetName,
			Namespace: h.Namespace,
			Labels:    ls,
		},
		Spec: appsv1.StatefulSetSpec{
			PodManagementPolicy: appsv1.ParallelPodManagement,
			Replicas:            &replicas,
			ServiceName:         slaveServiceName,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						ImagePullPolicy: corev1.PullAlways,
						Image:           "alicek106/hadoop:2.6.0-k8s-slave",
						Name:            "hadoop-slave",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 22,
							Name:          "ssh",
						}},
						Env: []corev1.EnvVar{{
							Name:  "MASTER_ENDPOINT",
							Value: masterEndpoint,
						}},
					}},
				},
			},
		},
	}

	// Register HadoopService instance as the owner and controller of slaves StatefulSet
	controllerutil.SetControllerReference(h, statefulset, r.scheme)
	return statefulset
}

func (r *ReconcileHadoopService) secretForSlave(h *alicek106v1alpha1.HadoopService, password string) *corev1.Secret {
	masterSecretName := h.Name + "-hadoop-master-secret"
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      masterSecretName,
			Namespace: h.Namespace,
		},
		Data: map[string][]byte{
			"password": []byte(password),
		},
		Type: "Opaque",
	}

	// Register HadoopService instance as the owner and controller of slaves service
	controllerutil.SetControllerReference(h, secret, r.scheme)
	return secret
}

func (r *ReconcileHadoopService) serviceForMaster(h *alicek106v1alpha1.HadoopService) *corev1.Service {
	masterServiceName := h.Name + "-hadoop-master-svc"
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      masterServiceName,
			Namespace: h.Namespace,
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "None",
			Selector:  labelsForHadoopMaster(h.Name),
		},
	}

	// Register HadoopService instance as the owner and controller of slaves service
	controllerutil.SetControllerReference(h, service, r.scheme)
	return service
}

func (r *ReconcileHadoopService) statefulSetForMaster(h *alicek106v1alpha1.HadoopService) *appsv1.StatefulSet {
	ls := labelsForHadoopMaster(h.Name)
	var replicas int32 = 1
	masterSecretName := h.Name + "-hadoop-master-secret"

	slaveServiceName := h.Name + "-hadoop-slave-svc"
	slaveStatefulsetName := h.Name + "-hadoop-slave"

	masterName := h.Name + "-hadoop-master"
	masterServiceName := h.Name + "-hadoop-master-svc"
	masterEndpoint := h.Name + "-hadoop-master-0." + masterServiceName + "." + h.Namespace + ".svc.cluster.local"

	slaveCount := int(h.Spec.ClusterSize - 1)

	statefulset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      masterName,
			Namespace: h.Namespace,
			Labels:    ls,
		},
		Spec: appsv1.StatefulSetSpec{
			PodManagementPolicy: appsv1.ParallelPodManagement,
			Replicas:            &replicas,
			ServiceName:         masterServiceName,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{{
						Name: masterSecretName,
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{SecretName: masterSecretName},
						},
					}},
					Containers: []corev1.Container{{
						ImagePullPolicy: corev1.PullAlways,
						Image:           "alicek106/hadoop:2.6.0-k8s-master",
						Name:            "hadoop-master",
						ReadinessProbe: &corev1.Probe{
							Handler: corev1.Handler{
								TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt(50070)}, // hadoop dashboard port
							},
							InitialDelaySeconds: 5,
							PeriodSeconds:       2,
						},
						Ports: []corev1.ContainerPort{
							{ContainerPort: 22, Name: "ssh"},
							{ContainerPort: 50070, Name: "dashboard"},
							{ContainerPort: 8088, Name: "yarn"},
						},
						Env: []corev1.EnvVar{
							{Name: "MASTER_ENDPOINT", Value: masterEndpoint},
							{Name: "SLAVES_SVC_NAME", Value: slaveServiceName},
							{Name: "SLAVES_SS_NAME", Value: slaveStatefulsetName},
							{Name: "SLAVES_COUNT", Value: strconv.Itoa(slaveCount)},
							{Name: "NAMESPACE", Value: h.Namespace},
						},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      masterSecretName,
							MountPath: "/etc/rootpwd",
							ReadOnly:  true,
						}},
					}},
				},
			},
		},
	}

	// Register HadoopService instance as the owner and controller of slaves StatefulSet
	controllerutil.SetControllerReference(h, statefulset, r.scheme)
	return statefulset
}

func (r *ReconcileHadoopService) externalServiceForMaster(h *alicek106v1alpha1.HadoopService) *corev1.Service {
	masterServiceName := h.Name + "-hadoop-master-svc-external"
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      masterServiceName,
			Namespace: h.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{
				{Name: "ssh", Port: 22},
				{Name: "yarn", Port: 8088},
				{Name: "dashboard", Port: 50070},
			},
			Selector: labelsForHadoopMaster(h.Name),
		},
	}

	// Register HadoopService instance as the owner and controller of slaves service
	controllerutil.SetControllerReference(h, service, r.scheme)
	return service
}

// labelsForHadoopService returns the labels for selecting the resources
// belonging to the given HadoopService custom resource.
func labelsForHadoopSlave(name string) map[string]string {
	return map[string]string{"app": "hadoopservice", "hadoop_cr": name, "role": "slave"}
}

func labelsForHadoopMaster(name string) map[string]string {
	return map[string]string{"app": "hadoopservice", "hadoop_cr": name, "role": "master"}
}

func randomString(len int) string {
	bytes := make([]byte, len)
	rand.Seed(time.Now().UTC().UnixNano())
	for i := 0; i < len; i++ {
		bytes[i] = byte(65 + rand.Intn(25)) //A=65 and Z = 65+25
	}
	return string(bytes)
}
