/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ChubaoMonitorSpec defines the desired state of ChubaoMonitor
type ChubaoMonitorSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of ChubaoMonitor. Edit ChubaoMonitor_types.go to remove/update
	Sizeprom               int32                        `json:"sizeprom"`
	Sizegrafana            int32                        `json:"sizegrafana"`
	Imageprom              string                       `json:"imageprom"`
	Imagegrafana           string                       `json:"imagegrafana"`
	Portprom               int32                        `json:"portprom,omitempty"`
	Portgrafana            int32                        `json:"portgrafana,omitempty"`
	ImagePullPolicyprom    corev1.PullPolicy            `json:"imagePullPolicyprom,omitempty"`
	ImagePullPolicygrafana corev1.PullPolicy            `json:"imagePullPolicygrafana,omitempty"`
	HostPath               *corev1.HostPathVolumeSource `json:"hostPath,omitempty"`
	Resourcesprom          corev1.ResourceRequirements  `json:"resourcesprom,omitempty"`
	Resourcesgrafana       corev1.ResourceRequirements  `json:"resourcesgrafana,omitempty"`
}

// ChubaoMonitorStatus defines the observed state of ChubaoMonitor
type ChubaoMonitorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	//	Deploymentforprometheus bool `json:"deploymentforprometheus"`
	//	Depploymentforgrafana   bool `json:"deploymentforgrafana"`
	//	Serviceforprometheus    bool `json:"serviceforprometheus"`
	//	Serviceforgrafana       bool `json:"serviceforgrafana"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ChubaoMonitor is the Schema for the chubaomonitors API
type ChubaoMonitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ChubaoMonitorSpec   `json:"spec,omitempty"`
	Status ChubaoMonitorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ChubaoMonitorList contains a list of ChubaoMonitor
type ChubaoMonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ChubaoMonitor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ChubaoMonitor{}, &ChubaoMonitorList{})
}
