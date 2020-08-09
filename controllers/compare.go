package controllers

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"reflect"
)

func CompareDeployment(a *appsv1.Deployment, b *appsv1.Deployment) bool {

	if *a.Spec.Replicas != *b.Spec.Replicas {
		return true
	}
	return false
}

func CompareService(a *corev1.Service, b *corev1.Service) bool {
	if !reflect.DeepEqual(a.Spec.Ports, b.Spec.Ports) || a.Spec.Type != corev1.ServiceTypeClusterIP || !reflect.DeepEqual(a.Spec.Selector, b.Spec.Selector) {
		return true
	}
	return false
}
