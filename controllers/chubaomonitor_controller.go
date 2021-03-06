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

	"github.com/go-logr/logr"
	"github.com/prometheus/common/log"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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

	//create desiredConfigmap
	desiredConfigmap := ConfigmapForChubaomonitor(chubaomonitor)
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
		if err != nil && !errors.IsAlreadyExists(err) {
			log.Error(err, "Failed to create new configmap", "Configmap.Namespace", desiredConfigmap.Namespace, "Configmap.Name", desiredConfigmap.Name)
			chubaomonitor.Status.Configmapstatus = false
			return ctrl.Result{}, err
		}
		//create the configmap successfully.
		chubaomonitor.Status.Configmapstatus = true
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get chubaomonitor configmap")
		chubaomonitor.Status.Configmapstatus = false
	}
	//fetch chubaomonitor configmap successfully

	//check if chubaomonitor configmap data is right
	if !reflect.DeepEqual(desiredConfigmap.Data, configmapchubaomonitor.Data) {
		configmapchubaomonitor.Data = desiredConfigmap.Data
		if err := controllerutil.SetControllerReference(chubaomonitor, configmapchubaomonitor, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		log.Info("Updating congfigmap")

		if err = r.Update(ctx, configmapchubaomonitor); err != nil && !errors.IsConflict(err) {
			chubaomonitor.Status.Configmapstatus = false
			return ctrl.Result{}, err
		}
		chubaomonitor.Status.Configmapstatus = true
	}

	//create desiredDeploymentPrometheus

	desiredDeploymentPrometheus := r.Deploymentforprometheus(chubaomonitor)
	if err := controllerutil.SetControllerReference(chubaomonitor, desiredDeploymentPrometheus, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	//check if the prometheus deployment exit. If not, create one
	deploymentPrometheus := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: "prometheus", Namespace: chubaomonitor.Namespace}, deploymentPrometheus)
	if err != nil && errors.IsNotFound(err) {
		//create prometheus deployment
		log.Info("Creating a new prometheus Deployment", "Deployment.Namespace", desiredDeploymentPrometheus.Namespace, "Deployment.Name", desiredDeploymentPrometheus.Name)
		err = r.Create(ctx, desiredDeploymentPrometheus)
		if err != nil && !errors.IsAlreadyExists(err) {
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
	if check := CompareDeployment(desiredDeploymentPrometheus, deploymentPrometheus); check {
		deploymentPrometheus.Spec = desiredDeploymentPrometheus.Spec
		if err := controllerutil.SetControllerReference(chubaomonitor, deploymentPrometheus, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		log.Info("Updating deploymentprometheus")

		if err = r.Update(ctx, deploymentPrometheus); err != nil && !errors.IsConflict(err) {
			return ctrl.Result{}, err
		}
	}

	//Update chubaomonitor.Status.PrometheusReplicas, if needed.
	if chubaomonitor.Status.PrometheusReplicas != *deploymentPrometheus.Spec.Replicas {
		chubaomonitor.Status.PrometheusReplicas = *deploymentPrometheus.Spec.Replicas
		err := r.Status().Update(ctx, chubaomonitor)
		if err != nil && !errors.IsConflict(err) {
			log.Error(err, "Failed to update ChubaoMonitor status")
			return ctrl.Result{}, err
		}
	}

	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(deploymentPrometheus.Namespace),
		client.MatchingLabels(labelsForChubaoMonitor(deploymentPrometheus.Name)),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods", "deploymentPrometheus.Namespace", deploymentPrometheus.Namespace, "deploymentPrometheus.Name", deploymentPrometheus.Name)
		return ctrl.Result{}, err
	}
	podNames := getPodNames(podList.Items)

	// Update chubaomonitor.Status.PrometheusPods, if needed.
	if !reflect.DeepEqual(podNames, chubaomonitor.Status.PrometheusPods) {
		chubaomonitor.Status.PrometheusPods = podNames
		err := r.Status().Update(ctx, chubaomonitor)
		if err != nil && !errors.IsConflict(err) {
			log.Error(err, "Failed to update ChubaoMonitor status")
			return ctrl.Result{}, err
		}
	}

	//create desiredServicePrometheus
	desiredServicePrometheus := Serviceforprometheus(chubaomonitor)
	if err := controllerutil.SetControllerReference(chubaomonitor, desiredServicePrometheus, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	//check if the prometheus service exit. If not, create one
	servicePrometheus := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: "prometheus-service", Namespace: chubaomonitor.Namespace}, servicePrometheus)

	if err != nil && errors.IsNotFound(err) {
		//create the prometheus service
		log.Info("Creating a new promethues Service", "Service.Namespace", desiredServicePrometheus.Namespace, "Service.Name", "prometheus-service")
		err = r.Create(ctx, desiredServicePrometheus)
		if err != nil && !errors.IsAlreadyExists(err) {
			log.Error(err, "Failed to create new promethues Service", "Service.Namespace", desiredServicePrometheus.Namespace, "Service.Name", "prometheus-service")
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get prometheus Service")
	}
	//fetch the serviceprometheus successfully

	//check if the service Prometheus is right
	if check := CompareService(servicePrometheus, desiredServicePrometheus); check {
		servicePrometheus.Spec.Ports = desiredServicePrometheus.Spec.Ports
		servicePrometheus.Spec.Type = corev1.ServiceTypeClusterIP
		servicePrometheus.Spec.Selector = desiredServicePrometheus.Spec.Selector
		if err := controllerutil.SetControllerReference(chubaomonitor, servicePrometheus, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		log.Info("Updating serviceprometheus")

		if err := r.Update(ctx, servicePrometheus); err != nil && !errors.IsConflict(err) {
			return ctrl.Result{}, err
		}
	}

	if chubaomonitor.Status.PrometheusclusterIP != servicePrometheus.Spec.ClusterIP {
		chubaomonitor.Status.PrometheusclusterIP = servicePrometheus.Spec.ClusterIP
		err := r.Status().Update(ctx, chubaomonitor)
		if err != nil && !errors.IsConflict(err) {
			log.Error(err, "Failed to update ChubaoMonitor status")
			return ctrl.Result{}, err
		}
	}

	//create desiredDeploymentGrafana

	desiredDeploymentGrafana := r.Deploymentforgrafana(chubaomonitor)
	if err := controllerutil.SetControllerReference(chubaomonitor, desiredDeploymentGrafana, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	//check whether the grafana deployment exit. If not, create one
	deploymentGrafana := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: "grafana", Namespace: chubaomonitor.Namespace}, deploymentGrafana)
	if err != nil && errors.IsNotFound(err) {
		//create the grafana deployment
		log.Info("Creating a new grafana Deployment", "Deployment.Namespace", desiredDeploymentGrafana.Namespace, "Deployment.Name", desiredDeploymentGrafana.Name)
		err = r.Create(ctx, desiredDeploymentGrafana)
		if err != nil && !errors.IsAlreadyExists(err) {
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
	if check := CompareDeployment(desiredDeploymentGrafana, deploymentGrafana); check {
		deploymentGrafana.Spec.Replicas = desiredDeploymentGrafana.Spec.Replicas
		deploymentGrafana.Spec = desiredDeploymentGrafana.Spec
		//		deploymentGrafana.Spec.Selector = desiredDeploymentGrafana.Spec.Selector
		//		deploymentGrafana.Spec.Template = desiredDeploymentGrafana.Spec.Template
		log.Info("Updating deploymentgrafana")

		if err := controllerutil.SetControllerReference(chubaomonitor, deploymentGrafana, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		if err = r.Update(ctx, deploymentGrafana); err != nil && !errors.IsConflict(err) {
			return ctrl.Result{}, err
		}
	}

	//Update chubaomonitor.Status.GrafanaReplicas, if needed.
	if chubaomonitor.Status.GrafanaReplicas != *deploymentGrafana.Spec.Replicas {
		chubaomonitor.Status.GrafanaReplicas = *deploymentGrafana.Spec.Replicas
		err := r.Status().Update(ctx, chubaomonitor)
		if err != nil && !errors.IsConflict(err) {
			log.Error(err, "Failed to update ChubaoMonitor status")
			return ctrl.Result{}, err
		}
	}

	podList = &corev1.PodList{}
	listOpts = []client.ListOption{
		client.InNamespace(deploymentGrafana.Namespace),
		client.MatchingLabels(labelsForChubaoMonitor(deploymentGrafana.Name)),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods", "deploymentGrafana.Namespace", deploymentGrafana.Namespace, "deploymentGrafana.Name", deploymentGrafana.Name)
		return ctrl.Result{}, err
	}
	podNames = getPodNames(podList.Items)

	// Update chubaomonitor.Status.GrafanaPods, if needed.
	if !reflect.DeepEqual(podNames, chubaomonitor.Status.GrafanaPods) {
		chubaomonitor.Status.GrafanaPods = podNames
		err := r.Status().Update(ctx, chubaomonitor)
		if err != nil && !errors.IsConflict(err) {
			log.Error(err, "Failed to update ChubaoMonitor status")
			return ctrl.Result{}, err
		}
	}

	//create desiredServiceGrafana

	desiredServiceGrafana := Serviceforgrafana(chubaomonitor)
	if err := controllerutil.SetControllerReference(chubaomonitor, desiredServiceGrafana, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	//check if the grafana service exit. If not, create one
	serviceGrafana := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: "grafana-service", Namespace: chubaomonitor.Namespace}, serviceGrafana)

	if err != nil && errors.IsNotFound(err) {
		//create the grafana service
		log.Info("Creating a new grafana Service", "Service.Namespace", desiredServiceGrafana.Namespace, "Service.Name", "grafana-service")
		err = r.Create(ctx, desiredServiceGrafana)
		if err != nil && !errors.IsAlreadyExists(err) {
			log.Error(err, "Failed to create new grafana Service", "Service.Namespace", desiredServiceGrafana.Namespace, "Service.Name", "grafana-service")
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get grafana Service")
	}
	//fetch the servicegrafana successful

	//check if the grafana service is right
	if check := CompareService(serviceGrafana, desiredServiceGrafana); check {
		serviceGrafana.Spec.Ports = desiredServiceGrafana.Spec.Ports
		serviceGrafana.Spec.Type = corev1.ServiceTypeClusterIP
		serviceGrafana.Spec.Selector = desiredServiceGrafana.Spec.Selector
		if err := controllerutil.SetControllerReference(chubaomonitor, serviceGrafana, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		log.Info("Updating servicegrafana")

		if err := r.Update(ctx, serviceGrafana); err != nil && !errors.IsConflict(err) {
			return ctrl.Result{}, err
		}
	}

	if chubaomonitor.Status.GrafanaclusterIP != serviceGrafana.Spec.ClusterIP {
		chubaomonitor.Status.GrafanaclusterIP = serviceGrafana.Spec.ClusterIP
		err := r.Status().Update(ctx, chubaomonitor)
		if err != nil && !errors.IsConflict(err) {
			log.Error(err, "Failed to update ChubaoMonitor status")
			return ctrl.Result{}, err
		}

	}

	//my logic finished
	return ctrl.Result{}, nil
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
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
