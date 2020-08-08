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

package controllers

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	"github.com/prometheus/common/log"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilintstr "k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	cachev1alpha1 "ChubaoMonitorOperator/api/v1alpha1"
)

// ChubaoMonitorReconciler reconciles a ChubaoMonitor object
type ChubaoMonitorReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=cache.example.com,resources=chubaomonitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cache.example.com,resources=chubaomonitors/status,verbs=get;update;patch

func (r *ChubaoMonitorReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)

	// my logic starts here

	//fetch ChubaoMonitor instance.
	chubaomonitor := &cachev1alpha1.ChubaoMonitor{}
	err := r.Get(ctx, req.NamespacedName, chubaomonitor)
	if err != nil {
		if errors.IsNotFound(err) {
			//can't find ChubaoMonitor instance
			fmt.Println("ChubaoMonitor resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to fetch ChubaoMonitor")
		return ctrl.Result{}, err
	}
	//fetch the ChubaoMonitor instance successfully

	desiredDeploymentPrometheus := r.deploymentforprometheus(chubaomonitor)
	desiredServicePrometheus := serviceforprometheus(chubaomonitor)
	desiredDeploymentGrafana := r.deploymentforgrafana(chubaomonitor)
	desiredServiceGrafana := serviceforgrafana(chubaomonitor)
	desiredConfigmap := ConfigmapForChubaomonitor(chubaomonitor)

	if err := controllerutil.SetControllerReference(chubaomonitor, desiredDeploymentPrometheus, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	if err := controllerutil.SetControllerReference(chubaomonitor, desiredServicePrometheus, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	if err := controllerutil.SetControllerReference(chubaomonitor, desiredDeploymentGrafana, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	if err := controllerutil.SetControllerReference(chubaomonitor, desiredServiceGrafana, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	if err := controllerutil.SetControllerReference(chubaomonitor, desiredConfigmap, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	//check if the configmap exit. If not,create one.
	configmapchubaomonitor := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: "monitor-config", Namespace: chubaomonitor.Namespace}, configmapchubaomonitor)
	if err != nil && errors.IsNotFound(err) {
		//create configmap
		log.Info("Creating a new chubaomonitor configmap", "Configmap.Namespace", desiredConfigmap.Namespace, "Configmap.Name", desiredConfigmap.Name)
		err = r.Create(ctx, desiredConfigmap)
		if err != nil {
			log.Error(err, "Failed to create new configmap", "Configmap.Namespace", desiredConfigmap.Namespace, "Configmap.Name", desiredConfigmap.Name)
			return ctrl.Result{}, err
		}
		//create the configmap successfully.
		return ctrl.Result{RequeueAfter: time.Second * 1}, nil
	} else if err != nil {
		log.Error(err, "Failed to get chubaomonitor configmap")
	}
	//fetch chubaomonitor configmap successfully

	//check if chubaomonitor configmap data is right
	if !reflect.DeepEqual(desiredConfigmap, configmapchubaomonitor) {
		configmapchubaomonitor = desiredConfigmap
		if err = r.Update(ctx, configmapchubaomonitor); err != nil {
			return ctrl.Result{}, err
		}
	}

	//check if the prometheus deployment exit. If not, create one
	deploymentPrometheus := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: "prometheus", Namespace: chubaomonitor.Namespace}, deploymentPrometheus)
	if err != nil && errors.IsNotFound(err) {
		//create prometheus deployment
		log.Info("Creating a new prometheus Deployment", "Deployment.Namespace", desiredDeploymentPrometheus.Namespace, "Deployment.Name", desiredDeploymentPrometheus.Name)
		err = r.Create(ctx, desiredDeploymentPrometheus)
		if err != nil {
			log.Error(err, "Failed to create new prometheus Deployment", "Deployment.Namespace", desiredDeploymentPrometheus.Namespace, "Deployment.Name", desiredDeploymentPrometheus.Name)
			return ctrl.Result{}, err
		}
		//create the deployment successfully.
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get promethues Deployment")
	}
	//fetch the deploymentprometheus successfully

	//check if the deploymentprometheus is right
	if !reflect.DeepEqual(desiredDeploymentPrometheus.Spec, deploymentPrometheus.Spec) {
		deploymentPrometheus.Spec = desiredDeploymentPrometheus.Spec
		if err = r.Update(ctx, deploymentPrometheus); err != nil {
			return ctrl.Result{}, err
		}
	}

	//check if the prometheus service exit. If not, create one
	servicePrometheus := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: "prometheus-service", Namespace: chubaomonitor.Namespace}, servicePrometheus)

	if err != nil && errors.IsNotFound(err) {
		//create the prometheus service
		log.Info("Creating a new promethues Service", "Service.Namespace", desiredServicePrometheus.Namespace, "Service.Name", "prometheus-service")
		err = r.Create(ctx, desiredServicePrometheus)
		if err != nil {
			log.Error(err, "Failed to create new promethues Service", "Service.Namespace", desiredServicePrometheus.Namespace, "Service.Name", "prometheus-service")
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get promethues Service")
	}
	//fetch the serviceprometheus successfully

	//check if the service Prometheus is right
	if !reflect.DeepEqual(desiredServicePrometheus.Spec.Ports, servicePrometheus.Spec.Ports) {
		servicePrometheus.Spec.Ports = desiredServicePrometheus.Spec.Ports
		if err := r.Update(ctx, servicePrometheus); err != nil {
			return ctrl.Result{}, err
		}
	}

	//check whether the grafana deployment exit. If not, create one
	deploymentGrafana := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: "grafana", Namespace: chubaomonitor.Namespace}, deploymentGrafana)
	if err != nil && errors.IsNotFound(err) {
		//create the grafana deployment
		log.Info("Creating a new grafana Deployment", "Deployment.Namespace", desiredDeploymentGrafana.Namespace, "Deployment.Name", desiredDeploymentGrafana.Name)
		err = r.Create(ctx, desiredDeploymentGrafana)
		if err != nil {
			log.Error(err, "Failed to create new grafana Deployment", "Deployment.Namespace", desiredDeploymentGrafana.Namespace, "Deployment.Name", desiredDeploymentGrafana.Name)
			return ctrl.Result{}, err
		}
		//create the deployment successfully.

		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get grafana Deployment")
		return ctrl.Result{}, err
	}
	//fetch the deploymentgrafana successfully

	//check if the deploymentgrafana is right
	if !reflect.DeepEqual(desiredDeploymentGrafana.Spec, deploymentGrafana.Spec) {
		deploymentGrafana.Spec = desiredDeploymentGrafana.Spec
		if err = r.Update(ctx, deploymentGrafana); err != nil {
			return ctrl.Result{}, err
		}
	}

	//check if the grafana service exit. If not, create one
	serviceGrafana := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: "grafana-service", Namespace: chubaomonitor.Namespace}, serviceGrafana)

	if err != nil && errors.IsNotFound(err) {
		//create the grafana service
		log.Info("Creating a new grafana Service", "Service.Namespace", desiredServiceGrafana.Namespace, "Service.Name", "grafana-service")
		err = r.Create(ctx, desiredServiceGrafana)
		if err != nil {
			log.Error(err, "Failed to create new grafana Service", "Service.Namespace", desiredServiceGrafana.Namespace, "Service.Name", "grafana-service")
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get grafana Service")
	}
	//fetch the servicegrafana successful

	//check if the grafana service is right
	if !reflect.DeepEqual(desiredServiceGrafana.Spec.Ports, serviceGrafana.Spec.Ports) {
		serviceGrafana.Spec.Ports = desiredServiceGrafana.Spec.Ports
		if err := r.Update(ctx, serviceGrafana); err != nil {
			return ctrl.Result{}, err
		}
	}

	//my logic finished
	return ctrl.Result{}, nil
}

func (r *ChubaoMonitorReconciler) deploymentforprometheus(m *cachev1alpha1.ChubaoMonitor) *appsv1.Deployment {
	name := "prometheus"
	labels := labelsForChubaoMonitor(name)
	selector := &metav1.LabelSelector{MatchLabels: labels}
	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "prometheus",
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &m.Spec.Sizep,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: containerforprometheus(m),
					Volumes:    volumeforprometheus(m),
				},
			},
			Selector: selector,
		},
	}
	return dep
}

func containerforprometheus(m *cachev1alpha1.ChubaoMonitor) []corev1.Container {
	containerPorts := []corev1.ContainerPort{}
	for _, svcPort := range m.Spec.Portsp {
		cport := corev1.ContainerPort{}
		cport.ContainerPort = svcPort.TargetPort.IntVal
		containerPorts = append(containerPorts, cport)
	}

	return []corev1.Container{
		{
			Name:            "prometheus-pod",
			Image:           m.Spec.Imagep,
			Ports:           containerPorts,
			ImagePullPolicy: m.Spec.ImagePullPolicyp,
			Env: []corev1.EnvVar{
				{Name: "TZ", Value: "Asia/Shanghai"},
			},
			VolumeMounts: volumemountsforprometheus(),
		},
	}
}

func serviceforprometheus(m *cachev1alpha1.ChubaoMonitor) *corev1.Service {
	name := "prometheus"
	labels := labelsForChubaoMonitor(name)

	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "prometheus-service",
			Namespace: m.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:       m.Spec.Portsp[0].Port,
					TargetPort: m.Spec.Portsp[0].TargetPort,
					Protocol:   "TCP",
				},
			},
			Selector: labels,
		},
	}

}

func volumeforprometheus(m *cachev1alpha1.ChubaoMonitor) []corev1.Volume {
	var defaultmode int32 = 0555
	return []corev1.Volume{
		{
			Name: "monitor-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "monitor-config",
					},
					DefaultMode: &defaultmode,
				},
			},
		},
		{
			Name: "prometheus-data",
			VolumeSource: corev1.VolumeSource{
				HostPath: m.Spec.HostPath,
			},
		},
	}
}

func volumemountsforprometheus() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      "monitor-config",
			MountPath: "/etc/prometheus/prometheus.yml",
			SubPath:   "prometheus.yml",
		},
		{
			Name:      "prometheus-data",
			MountPath: "/prometheus-data",
		},
	}
}

func (r *ChubaoMonitorReconciler) deploymentforgrafana(m *cachev1alpha1.ChubaoMonitor) *appsv1.Deployment {
	name := "grafana"
	labels := labelsForChubaoMonitor(name)
	selector := &metav1.LabelSelector{MatchLabels: labels}
	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "grafana",
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &m.Spec.Sizeg,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: containerforgrafana(m),
					Volumes:    volumesforgrafana(m),
				},
			},
			Selector: selector,
		},
	}
	return dep
}

func containerforgrafana(m *cachev1alpha1.ChubaoMonitor) []corev1.Container {
	containerPorts := []corev1.ContainerPort{}
	for _, svcPort := range m.Spec.Portsg {
		cport := corev1.ContainerPort{}
		cport.ContainerPort = svcPort.TargetPort.IntVal
		containerPorts = append(containerPorts, cport)
	}
	var privileged bool = true
	return []corev1.Container{
		{
			Name:            "grafana-pod",
			Image:           m.Spec.Imageg,
			Ports:           containerPorts,
			ImagePullPolicy: m.Spec.ImagePullPolicyg,
			SecurityContext: &corev1.SecurityContext{
				Privileged: &privileged,
			},
			Env: envforgrafana(),
			// If grafana pod show the err "back-off restarting failed container", run this command to keep the container running ang then run ./run.sh in the container to check the really error.
			//			Command:        []string{"/bin/bash", "-ce", "tail -f /dev/null"},
			ReadinessProbe: readinessforgrafana(),
			VolumeMounts:   volumemountsforgrafana(),
		},
	}
}

func serviceforgrafana(m *cachev1alpha1.ChubaoMonitor) *corev1.Service {
	name := "grafana"
	labels := labelsForChubaoMonitor(name)
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "grafana-service",
			Namespace: m.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:       m.Spec.Portsg[0].Port,
					TargetPort: m.Spec.Portsg[0].TargetPort,
					Protocol:   "TCP",
				},
			},
			Selector: labels,
		},
	}
}

func volumesforgrafana(m *cachev1alpha1.ChubaoMonitor) []corev1.Volume {
	var defaultmode int32 = 0555

	return []corev1.Volume{
		{
			Name: "monitor-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "monitor-config",
					},
					DefaultMode: &defaultmode,
				},
			},
		},
		{Name: "grafana-persistent-storage",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}
}

func volumemountsforgrafana() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      "grafana-persistent-storage",
			MountPath: "/var/lib/grafana",
		},
		{
			Name:      "monitor-config",
			MountPath: "/grafana/init.sh",
			SubPath:   "init.sh",
		},
		{
			Name:      "monitor-config",
			MountPath: "/etc/grafana/grafana.ini",
			SubPath:   "grafana.ini",
		},
		{
			Name:      "monitor-config",
			MountPath: "/etc/grafana/provisioning/dashboards/chubaofs.json",
			SubPath:   "chubaofs.json",
		},
		{
			Name:      "monitor-config",
			MountPath: "/etc/grafana/provisioning/dashboards/dashboard.yml",
			SubPath:   "dashboard.yml",
		},
		{
			Name:      "monitor-config",
			MountPath: "/etc/grafana/provisioning/datasources/datasource.yml",
			SubPath:   "datasource.yml",
		},
	}
}

func envforgrafana() []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "GF_AUTH_BASIC_ENABLED",
			Value: "true",
		},
		{
			Name:  "GF_AUTH_ANONYMOUS_ENABLED",
			Value: "false",
		},
		{
			Name:  "GF_SECURITY_ADMIN_PASSWORD",
			Value: "123456",
		},
	}
}

func readinessforgrafana() *corev1.Probe {
	return &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/login",
				Port: utilintstr.IntOrString{
					IntVal: 3000,
				},
			},
		},
	}
}

// labelsForChubaoMonitorOperator returns the labels for selecting the resources
// belonging to the given chubaomonitor CR name.
func labelsForChubaoMonitor(name string) map[string]string {
	return map[string]string{"app": name}
}

func (r *ChubaoMonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {

	c, err := controller.New("chubaomonitor-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		log.Error(err, "unable to setup chubaomonitor-controller")
		os.Exit(1)
	}

	err = c.Watch(&source.Kind{Type: &cachev1alpha1.ChubaoMonitor{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	log.Info("ChubaoMonitor being watched")

	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &cachev1alpha1.ChubaoMonitor{},
	})
	if err != nil {
		return err
	}
	log.Info("Deployment being watched")

	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &cachev1alpha1.ChubaoMonitor{},
	})
	if err != nil {
		return err
	}
	log.Info("Service being watched")
	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &cachev1alpha1.ChubaoMonitor{},
	})
	if err != nil {
		return err
	}
	log.Info("ConfigMap being watched")

	return nil
}
