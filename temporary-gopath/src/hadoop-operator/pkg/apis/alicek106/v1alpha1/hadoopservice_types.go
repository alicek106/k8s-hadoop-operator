package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HadoopServiceSpec defines the desired state of HadoopService
// +k8s:openapi-gen=true
type HadoopServiceSpec struct {
	// Size is the size of the memcached deployment
	ClusterSize int32 `json:"clusterSize"`
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// HadoopServiceStatus defines the observed state of HadoopService
// +k8s:openapi-gen=true
type HadoopServiceStatus struct {
	// Nodes are the names of the memcached pods
	Nodes []string `json:"nodes"`
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HadoopService is the Schema for the hadoopservices API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type HadoopService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HadoopServiceSpec   `json:"spec,omitempty"`
	Status HadoopServiceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HadoopServiceList contains a list of HadoopService
type HadoopServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HadoopService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HadoopService{}, &HadoopServiceList{})
}
