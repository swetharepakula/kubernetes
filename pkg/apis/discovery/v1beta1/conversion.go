/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	"k8s.io/api/discovery/v1beta1"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/kubernetes/pkg/apis/discovery"
)

type EndpointSiceTopology struct {
	Strings  []string
	Topology []map[int]int
}

func Convert_v1beta1_EndpointSlice_To_discovery_EndpointSlice(in *v1beta1.EndpointSlice, out *discovery.EndpointSlice, s conversion.Scope) error {
	// check with autoconversion first and then do conversion
	return nil
}

func Convert_discovery_EndpointSlice_To_v1beta1_EndpointSlice(in *discovery.EndpointSlice, out *v1beta1.EndpointSlice, s conversion.Scope) error {
	return nil
}

func Convert_v1beta1_Endpoint_To_discovery_Endpoint(in *v1beta1.Endpoint, out *discovery.Endpoint, s conversion.Scope) error {

	return nil
}

func Convert_discovery_Endpoint_To_v1beta1_Endpoint(in *discovery.Endpoint, out *v1beta1.Endpoint, s conversion.Scope) error {

	return nil
}

func add_endpoint_topology_map(endpointTop map[string]string, epsTop EndpointSliceTopology) EndpointSliceTopology {
	return ""
}
