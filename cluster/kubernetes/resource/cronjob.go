package resource

type CronJob struct {
	baseObject
	Spec CronJobSpec
}

type CronJobSpec struct {
	JobTemplate struct {
		Spec struct {
			Template PodTemplate
		}
	}
}

func (c CronJob) ContainerNames() []string {
	return c.Spec.JobTemplate.Spec.Template.ContainerNames()
}
