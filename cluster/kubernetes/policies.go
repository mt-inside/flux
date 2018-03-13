package kubernetes

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/weaveworks/flux"
	"github.com/weaveworks/flux/cluster/kubernetes/resource"
	"github.com/weaveworks/flux/policy"
)

func (m *Manifests) UpdatePolicies(in []byte, resourceID flux.ResourceID, update policy.Update) ([]byte, error) {
	ns, kind, name := resourceID.Components()

	add := update.Add
	del := update.Remove

	// We may be sent the pseudo-policy `policy.TagAll`, which means
	// apply this filter to all containers. To do so, we need to know
	// what all the containers are.
	if tagAll, ok := add.Get(policy.TagAll); ok {
		add = add.Without(policy.TagAll)
		containers, err := parseForContainers(resourceID, in)
		if err != nil {
			return nil, err
		}
		for _, c := range containers {
			// Special case: glob:* is the same as "allow anything",
			// i.e., don't have a filter.
			if tagAll == "glob:*" {
				del = del.Add(policy.TagPrefix(c))
			} else {
				add = add.Set(policy.TagPrefix(c), tagAll)
			}
		}
	}

	args := []string{}
	for pol, val := range add {
		args = append(args, fmt.Sprintf("%s%s=%s", resource.PolicyPrefix, pol, val))
	}
	for pol, _ := range del {
		args = append(args, fmt.Sprintf("%s%s=", resource.PolicyPrefix, pol))
	}

	return (KubeYAML{}).Annotate(in, ns, kind, name, args...)
}

func parseForContainers(id flux.ResourceID, def []byte) ([]string, error) {
	all, err := resource.ParseMultidoc(def, "stdin")
	if err != nil {
		return nil, err
	}
	res, ok := all[id.String()]
	if !ok {
		return nil, errors.New("resource " + id.String() + " not found")
	}
	workload, ok := res.(resource.Workload)
	if !ok {
		return nil, errors.New("resource " + id.String() + " does not have containers")
	}

	return workload.ContainerNames(), nil
}

func (m *Manifests) ServicesWithPolicies(root string) (policy.ResourceMap, error) {
	resources, err := m.LoadManifests(root)
	if err != nil {
		return nil, err
	}

	polMap := policy.ResourceMap{}
	for _, res := range resources {
		polMap[res.ResourceID()] = res.Policy()
	}
	return polMap, nil
}
