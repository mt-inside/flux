package resource

type StatefulSet struct {
	baseObject
	Spec StatefulSetSpec
}

type StatefulSetSpec struct {
	Replicas int
	Template PodTemplate
}

func (ss StatefulSet) ContainerNames() []string {
	return ss.Spec.Template.ContainerNames()
}
