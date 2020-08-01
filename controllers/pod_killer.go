package controllers

import (
	"context"
	v1 "k8s.io/api/core/v1"
)

// delete the desired pod list
func (r *KillerReconciler) KillPods(podList *v1.PodList) error {
	for _, pod := range podList.Items {
		r.Log.Info("deleting pod_name", pod.Name, "pod_namespace", pod.Namespace)
		err := r.Client.Delete(context.Background(), &pod)
		if err != nil {
			r.Log.Error(err, "Error on pod deletion", "pod_name", pod.Name, "pod_namespace", pod.Namespace)
			return err
		}
		r.Log.Info("deleted pod_name", pod.Name, "pod_namespace", pod.Namespace)
	}
	return nil
}
