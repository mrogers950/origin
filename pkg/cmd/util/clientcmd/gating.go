package clientcmd

import (
	"fmt"

	"github.com/openshift/origin/pkg/authorization/apis/authorization"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
)

// LegacyPolicyResourceGate returns err if the server does not support any of the legacy policy objects (< 3.7)
func LegacyPolicyResourceGate(client discovery.DiscoveryInterface) error {
	res, err := DiscoverGroupVersionResources(client,
		schema.GroupVersionResource{
			Group:    authorization.LegacySchemeGroupVersion.Group,
			Version:  authorization.LegacySchemeGroupVersion.Version,
			Resource: "clusterpolicies",
		},
		schema.GroupVersionResource{
			Group:    authorization.LegacySchemeGroupVersion.Group,
			Version:  authorization.LegacySchemeGroupVersion.Version,
			Resource: "clusterpolicybindings",
		},
		schema.GroupVersionResource{
			Group:    authorization.LegacySchemeGroupVersion.Group,
			Version:  authorization.LegacySchemeGroupVersion.Version,
			Resource: "policies",
		},
		schema.GroupVersionResource{
			Group:    authorization.LegacySchemeGroupVersion.Group,
			Version:  authorization.LegacySchemeGroupVersion.Version,
			Resource: "policybindings",
		},
		schema.GroupVersionResource{
			Group:    authorization.SchemeGroupVersion.Group,
			Version:  authorization.SchemeGroupVersion.Version,
			Resource: "clusterpolicies",
		},
		schema.GroupVersionResource{
			Group:    authorization.SchemeGroupVersion.Group,
			Version:  authorization.SchemeGroupVersion.Version,
			Resource: "clusterpolicybindings",
		},
		schema.GroupVersionResource{
			Group:    authorization.SchemeGroupVersion.Group,
			Version:  authorization.SchemeGroupVersion.Version,
			Resource: "policies",
		},
		schema.GroupVersionResource{
			Group:    authorization.SchemeGroupVersion.Group,
			Version:  authorization.SchemeGroupVersion.Version,
			Resource: "policybindings",
		})

	if err != nil {
		return err
	}

	if len(res) != 4 {
		return fmt.Errorf("the server does not support legacy policy resources")
	}

	return nil
}

// DiscoverGroupVersionResources performs a server resource discovery for each GroupVersionResource in
// filterGVR, returning a slice of GroupVersionResources containing the discovered resources.
func DiscoverGroupVersionResources(client discovery.ServerResourcesInterface, filterGVR ...schema.GroupVersionResource) ([]schema.GroupVersionResource, error) {
	if len(filterGVR) == 0 {
		return nil, fmt.Errorf("at least one GroupVersionResource must be provided")
	}

	ret := []schema.GroupVersionResource{}
	for i := range filterGVR {
		// Discover the list of resources for this GVR
		gv := filterGVR[i].GroupVersion()
		serverResources, err := client.ServerResourcesForGroupVersion(gv.String())
		if err != nil {
			return nil, err
		}

		for _, resource := range serverResources.APIResources {
			// If a resource name was given, return the matching GVR
			if filterGVR[i].Resource == resource.Name {
				ret = append(ret, filterGVR[i])
				break
			}
		}
	}

	return ret, nil
}
