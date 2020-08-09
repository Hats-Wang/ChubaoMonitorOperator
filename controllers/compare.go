package controllers

import (
	appsv1 "k8s.io/api/apps/v1"
	"reflect"
)

func CompareDeployment(a *appsv1.Deployment, b *appsv1.Deployment) bool {

	if a.Spec.Replicas != b.Spec.Replicas || a.Spec.Selector != b.Spec.Selector || !reflect.DeepEqual(a.Spec.Template, b.Spec.Template) {
		return false
	}
	return true
}
