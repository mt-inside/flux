package resource

type DaemonSet struct {
	baseObject
	Spec DaemonSetSpec
}

type DaemonSetSpec struct {
	Template PodTemplate
}

func (ds DaemonSet) ContainerNames() []string {
	return ds.Spec.Template.ContainerNames()
}
