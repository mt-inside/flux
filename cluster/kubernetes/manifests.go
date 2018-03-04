package kubernetes

import (
	kresource "github.com/weaveworks/flux/cluster/kubernetes/resource"
	"github.com/weaveworks/flux/image"
	"github.com/weaveworks/flux"
	"github.com/weaveworks/flux/resource"
)

type Manifests struct {
}

// FindDefinedServices implementation in files.go

func (c *Manifests) LoadManifests(base, first string, rest ...string) (map[string]resource.Resource, error) {
	return kresource.Load(base, first, rest...)
}

func (c *Manifests) ParseManifests(allDefs []byte) (map[string]resource.Resource, error) {
	return kresource.ParseMultidoc(allDefs, "exported")
}

func (c *Manifests) UpdateDefinition(def []byte, id flux.ResourceID, container string, image image.Ref) ([]byte, error) {
	return updatePodController(def, id, container, image)
}

// UpdatePolicies and ServicesWithPolicies in policies.go
