package controllers

import (
	"context"
	toxiv1alpha1 "github.com/claudioed/toxi-operator/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	eventReason = "Killed by Pod Killer"
)

// delete the desired pod list
func (r *KillerReconciler) EnsurePodsKilled(instance *toxiv1alpha1.Killer) error {
	r.Log.Info("Finding pods to delete ", "killer_name", instance.Name, "killer_namespace", instance.Namespace)
	sel := labels.NewSelector()
	for key, value := range instance.Spec.Selector.MatchLabels {
		r, _ := labels.NewRequirement(key, selection.Equals, []string{value})
		sel.Add(*r)
	}
	pods := &v1.PodList{}
	r.Client.List(context.TODO(), pods, &client.ListOptions{
		LabelSelector: client.MatchingLabelsSelector{Selector: sel},
		Namespace:     instance.Namespace,
	})

	if len(pods.Items) > 0 {
		r.Log.Info("There are pods to delete. ", "Number:", len(pods.Items))
	}
	for _, pod := range pods.Items {
		r.Log.Info("Deleting POD ", "POD name", pod.Name, "pod_namespace", pod.Namespace)
		if v1.PodRunning == pod.Status.Phase {
			err := r.Client.Delete(context.Background(), &pod)
			if err != nil {
				r.Log.Error(err, "Error on pod deletion", "pod_name", pod.Name, "pod_namespace", pod.Namespace)
				return err
			}
			r.Er.Eventf(instance, v1.EventTypeNormal, eventReason, "Pod name %s was killed by Killer", pod.Name)
		} else {
			r.Log.Info("POD is not Running Phase ", "POD name", pod.Name, "pod_namespace", pod.Namespace)
		}
		r.Log.Info("POD name deleted", "POD name", pod.Name, "pod_namespace", pod.Namespace)
	}
	return nil
}
