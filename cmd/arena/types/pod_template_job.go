package types

import (
	appsv1 "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodTemplateJob struct {
	metav1.TypeMeta
	metav1.ObjectMeta
	Selector *metav1.LabelSelector
	Template v1.PodTemplateSpec
}

func PodTemplateJobFromJob(job batch.Job) *PodTemplateJob {
	return &PodTemplateJob{
		ObjectMeta: job.ObjectMeta,
		TypeMeta:   job.TypeMeta,
		Template:   job.Spec.Template,
		Selector:   job.Spec.Selector,
	}
}

func PodTemplateJobFromStatefulSet(statefulSet appsv1.StatefulSet) *PodTemplateJob {
	return &PodTemplateJob{
		ObjectMeta: statefulSet.ObjectMeta,
		TypeMeta:   statefulSet.TypeMeta,
		Template:   statefulSet.Spec.Template,
		Selector:   statefulSet.Spec.Selector,
	}
}

func PodTemplateJobFromReplicaSet(replicaSet appsv1.ReplicaSet) *PodTemplateJob {
	return &PodTemplateJob{
		ObjectMeta: replicaSet.ObjectMeta,
		TypeMeta:   replicaSet.TypeMeta,
		Template:   replicaSet.Spec.Template,
		Selector:   replicaSet.Spec.Selector,
	}
}
