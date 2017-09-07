package clientcmd

import (
	"fmt"

	"github.com/openshift/origin/pkg/authorization/apis/authorization"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kclientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
)

// LegacyPolicyResourceGate returns err if the server does not support any of the legacy policy objects (< 3.7)
func LegacyPolicyResourceGate(client kclientset.Interface) error {
	legacyResources := []string{"clusterpolicies", "clusterpolicybindings", "policies", "policybindings"}
	schemeResources := []schema.GroupVersionResource{}
	for _, resource := range legacyResources {
		// We want to check legacy and current group versions for each resource
		schemeResources = append(schemeResources, schema.GroupVersionResource{
			Group:    authorization.LegacySchemeGroupVersion.Group,
			Version:  authorization.LegacySchemeGroupVersion.Version,
			Resource: resource,
		})
		schemeResources = append(schemeResources, schema.GroupVersionResource{
			Group:    authorization.SchemeGroupVersion.Group,
			Version:  authorization.SchemeGroupVersion.Version,
			Resource: resource,
		})
	}

	res, err := DiscoverGroupVersionResources(client, schemeResources)
	if err != nil {
		return err
	}

	if len(res) != len(legacyResources) {
		return fmt.Errorf("the server does not support legacy policy resources")
	}

	return nil
}

// DiscoverGroupVersionResources performs a server resource discovery for each GroupVersionResource in
// filterGVR, returning a slice of GroupVersionResources containing the discovered resources. If a filterGVR
// element has a non-empty Resource, then only the matching resource is added to the returning slice. A
// filterGVR element with an empty Resource returns all resources discovered for the group and version.
func DiscoverGroupVersionResources(client kclientset.Interface, filterGVR []schema.GroupVersionResource) ([]schema.GroupVersionResource, error) {
	if len(filterGVR) == 0 {
		return nil, fmt.Errorf("at least one GroupVersionResource must be provided")
	}

	ret := []schema.GroupVersionResource{}
	for _, filter := range filterGVR {
		// Discover the list of resources for this GVR
		serverResources, err := client.Discovery().ServerResourcesForGroupVersion(filter.String())
		if err != nil {
			continue
		}

		for _, resource := range serverResources.APIResources {
			// If a resource name was given, return the matching GVR
			if len(filter.Resource) > 0 {
				if filter.Resource == resource.Name {
					ret = append(ret, filter)
					break
				}
			} else { // No resource, return a GVR for each resource
				new := schema.GroupVersionResource{
					Group:    filter.Group,
					Version:  filter.Version,
					Resource: resource.Name,
				}
				ret = append(ret, new)
			}
		}
	}

	return ret, nil
}
