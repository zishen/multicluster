package object

import (
	"fmt"
	mcs "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"

	"github.com/coredns/coredns/plugin/pkg/log"
	discovery "k8s.io/api/discovery/v1"
	discoveryV1beta1 "k8s.io/api/discovery/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	LabelClusterId = "multicluster.kubernetes.io/source-cluster"
)

// Endpoints is a stripped down api.Endpoints with only the items we need for CoreDNS.
type Endpoints struct {
	// Don't add new fields to this struct without talking to the CoreDNS maintainers.
	Version   string
	ClusterId string
	Name      string
	Namespace string
	Index     string
	IndexIP   []string
	Subsets   []EndpointSubset

	*Empty
}

// EndpointSubset is a group of addresses with a common set of ports. The
// expanded set of endpoints is the Cartesian product of Addresses x Ports.
type EndpointSubset struct {
	Addresses []EndpointAddress
	Ports     []EndpointPort
}

// EndpointAddress is a tuple that describes single IP address.
type EndpointAddress struct {
	IP            string
	Hostname      string
	NodeName      string
	TargetRefName string
}

// EndpointPort is a tuple that describes a single port.
type EndpointPort struct {
	Port     int32
	Name     string
	Protocol string
}

// EndpointsKey returns a string using for the index.
func EndpointsKey(name, namespace string) string { return name + "." + namespace }

// EndpointSliceToEndpoints converts a *discovery.EndpointSlice to a *Endpoints.
func EndpointSliceToEndpoints(obj meta.Object) (meta.Object, error) {
	ends, ok := obj.(*discovery.EndpointSlice)
	log.Warningf("hehe===1==MultiCluster EndpointSliceToEndpoints obj:%+v", obj)
	if !ok {
		return nil, fmt.Errorf("unexpected object %v", obj)
	}
	e := &Endpoints{
		Version:   ends.GetResourceVersion(),
		ClusterId: ends.Labels[LabelClusterId],
		Name:      ends.GetName(),
		Namespace: ends.GetNamespace(),
		Index:     EndpointsKey(ends.Labels[mcs.LabelServiceName], ends.GetNamespace()),
		Subsets:   make([]EndpointSubset, 1),
	}
	log.Warningf("hehe===1==MultiCluster EndpointSliceToEndpoints Endpoints:%+v", e)
	if len(ends.Ports) == 0 {
		// Add sentinel if there are no ports.
		e.Subsets[0].Ports = []EndpointPort{{Port: -1}}
	} else {
		e.Subsets[0].Ports = make([]EndpointPort, len(ends.Ports))
		for k, p := range ends.Ports {
			ep := EndpointPort{Port: *p.Port, Name: *p.Name, Protocol: string(*p.Protocol)}
			e.Subsets[0].Ports[k] = ep
		}
	}

	for _, end := range ends.Endpoints {
		if !endpointsliceReady(end.Conditions.Ready) {
			continue
		}
		for _, a := range end.Addresses {
			ea := EndpointAddress{IP: a}
			if end.Hostname != nil {
				ea.Hostname = *end.Hostname
			}
			if end.TargetRef != nil {
				ea.TargetRefName = end.TargetRef.Name
			}
			if end.NodeName != nil {
				ea.NodeName = *end.NodeName
			}
			e.Subsets[0].Addresses = append(e.Subsets[0].Addresses, ea)
			e.IndexIP = append(e.IndexIP, a)
		}
	}

	*ends = discovery.EndpointSlice{}
	log.Warningf("hehe===1==MultiCluster EndpointSliceV1beta1ToEndpoints Endpoints:%+v", e)
	return e, nil
}

// EndpointSliceV1beta1ToEndpoints converts a v1beta1 *discovery.EndpointSlice to a *Endpoints.
func EndpointSliceV1beta1ToEndpoints(obj meta.Object) (meta.Object, error) {
	ends, ok := obj.(*discoveryV1beta1.EndpointSlice)
	log.Warningf("hehe===1==MultiCluster EndpointSliceV1beta1ToEndpoints obj:%+v", obj)
	if !ok {
		return nil, fmt.Errorf("unexpected object %v", obj)
	}
	e := &Endpoints{
		Version:   ends.GetResourceVersion(),
		Name:      ends.GetName(),
		Namespace: ends.GetNamespace(),
		Index:     EndpointsKey(ends.Labels[mcs.LabelServiceName], ends.GetNamespace()),
		Subsets:   make([]EndpointSubset, 1),
	}
	log.Warningf("hehe===1==MultiCluster EndpointSliceV1beta1ToEndpoints Endpoints:%+v", e)
	if len(ends.Ports) == 0 {
		// Add sentinel if there are no ports.
		e.Subsets[0].Ports = []EndpointPort{{Port: -1}}
	} else {
		e.Subsets[0].Ports = make([]EndpointPort, len(ends.Ports))
		for k, p := range ends.Ports {
			ep := EndpointPort{Port: *p.Port, Name: *p.Name, Protocol: string(*p.Protocol)}
			e.Subsets[0].Ports[k] = ep
		}
	}

	for _, end := range ends.Endpoints {
		if !endpointsliceReady(end.Conditions.Ready) {
			continue
		}
		for _, a := range end.Addresses {
			ea := EndpointAddress{IP: a}
			if end.Hostname != nil {
				ea.Hostname = *end.Hostname
			}
			if end.TargetRef != nil {
				ea.TargetRefName = end.TargetRef.Name
			}
			// EndpointSlice does not contain NodeName, leave blank
			e.Subsets[0].Addresses = append(e.Subsets[0].Addresses, ea)
			e.IndexIP = append(e.IndexIP, a)
		}
	}

	*ends = discoveryV1beta1.EndpointSlice{}
	log.Warningf("hehe===1==MultiCluster EndpointSliceV1beta1ToEndpoints Endpoints:%+v", e)
	return e, nil
}

func endpointsliceReady(ready *bool) bool {
	// Per API docs: a nil value indicates an unknown state. In most cases consumers
	// should interpret this unknown state as ready.
	if ready == nil {
		return true
	}
	return *ready
}

// CopyWithoutSubsets copies e, without the subsets.
func (e *Endpoints) CopyWithoutSubsets() *Endpoints {
	e1 := &Endpoints{
		Version:   e.Version,
		Name:      e.Name,
		Namespace: e.Namespace,
		Index:     e.Index,
		IndexIP:   make([]string, len(e.IndexIP)),
	}
	copy(e1.IndexIP, e.IndexIP)
	return e1
}

var _ runtime.Object = &Endpoints{}

// DeepCopyObject implements the ObjectKind interface.
func (e *Endpoints) DeepCopyObject() runtime.Object {
	e1 := &Endpoints{
		Version:   e.Version,
		Name:      e.Name,
		Namespace: e.Namespace,
		Index:     e.Index,
		IndexIP:   make([]string, len(e.IndexIP)),
		Subsets:   make([]EndpointSubset, len(e.Subsets)),
	}
	copy(e1.IndexIP, e.IndexIP)

	for i, eps := range e.Subsets {
		sub := EndpointSubset{
			Addresses: make([]EndpointAddress, len(eps.Addresses)),
			Ports:     make([]EndpointPort, len(eps.Ports)),
		}
		for j, a := range eps.Addresses {
			ea := EndpointAddress{IP: a.IP, Hostname: a.Hostname, NodeName: a.NodeName, TargetRefName: a.TargetRefName}
			sub.Addresses[j] = ea
		}
		for k, p := range eps.Ports {
			ep := EndpointPort{Port: p.Port, Name: p.Name, Protocol: p.Protocol}
			sub.Ports[k] = ep
		}

		e1.Subsets[i] = sub
	}
	return e1
}

// GetNamespace implements the metav1.Object interface.
func (e *Endpoints) GetNamespace() string { return e.Namespace }

// SetNamespace implements the metav1.Object interface.
func (e *Endpoints) SetNamespace(namespace string) {}

// GetName implements the metav1.Object interface.
func (e *Endpoints) GetName() string { return e.Name }

// SetName implements the metav1.Object interface.
func (e *Endpoints) SetName(name string) {}

// GetResourceVersion implements the metav1.Object interface.
func (e *Endpoints) GetResourceVersion() string { return e.Version }

// SetResourceVersion implements the metav1.Object interface.
func (e *Endpoints) SetResourceVersion(version string) {}
