package hadoopservice

import (
	"context"

	alicek106v1alpha1 "hadoop-operator/pkg/apis/alicek106/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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
	/// 1. Create Slave StatefulSet and Service
	slaveFound := &appsv1.StatefulSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-hadoop-slave", Namespace: instance.Namespace}, slaveFound)

	if err != nil && errors.IsNotFound(err) {
		// Define a new StatefulSet for slaves
		slaveStatefulset := r.statefulSetForSlave(instance)
		reqLogger.Info("Creating a new statefulset for slave.", "StatefulSet.Namespace", slaveStatefulset.Namespace,
			"StatefulSet.Name", slaveStatefulset.Name)
		err = r.client.Create(context.TODO(), slaveStatefulset)
		if err != nil {
			reqLogger.Error(err, "Failed to create new StatefulSet for slaves.", "StatefulSet.Namespace",
				slaveStatefulset.Namespace, "StatefulSet.Name", slaveStatefulset.Name)
			return reconcile.Result{}, err
		}

		slaveService := r.serviceForSlave(instance)
		err = r.client.Create(context.TODO(), slaveService)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Service for slaves.", "Service.Namespace",
				slaveService.Namespace, "Service.Name", slaveService.Name)
			return reconcile.Result{}, err
		}

		// StatefulSet and Service created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get StatefulSet for slaves.")
		return reconcile.Result{}, err
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////////
	/// 2. Create Master StatefulSet and Service

	return reconcile.Result{}, nil
}

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
			Selector:  labelsForHadoopService(h.Name),
		},
	}

	// Register HadoopService instance as the owner and controller of slaves service
	controllerutil.SetControllerReference(h, service, r.scheme)
	return service
}

func (r *ReconcileHadoopService) statefulSetForSlave(h *alicek106v1alpha1.HadoopService) *appsv1.StatefulSet {
	ls := labelsForHadoopService(h.Name)
	replicas := h.Spec.ClusterSize

	slaveServiceName := h.Name + "-hadoop-slave-svc"
	slaveStatefulsetName := h.Name + "-hadoop-slave"

	masterServiceName := h.Name + "-hadoop-master-svc"
	masterEndpoint := h.Name + "hadoop-master-0." + masterServiceName + "." + h.Namespace + ".svc.cluster.local"
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

// labelsForHadoopService returns the labels for selecting the resources
// belonging to the given HadoopService custom resource.
func labelsForHadoopService(name string) map[string]string {
	return map[string]string{"app": "hadoopservice", "hadoop_cr": name}
}
