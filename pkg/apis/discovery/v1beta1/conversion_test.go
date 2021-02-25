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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/api/discovery/v1beta1"
	"k8s.io/kubernetes/pkg/apis/discovery"
)

func TestEndpointTopologyConverstion(t *testing.T) {

	testcases := []struct {
		desc     string
		external v1beta1.EndpointSlice
		internal discovery.EndpointSlice
	}{}

	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			convertedInternal := discovery.EndpointSlice{}
			require.NoError(t, Convert_v1beta1_EndpointSlice_To_discovery_EndpointSlice(&tc.external, &convertedInternal, nil))
			assert.Equal(t, tc.internal, convertedInternal, "v1beta1.EndpointSlice -> discovery.EndpointSlice")

			convertedV1beta1 := v1beta1.EndpointSlice{}
			require.NoError(t, Convert_discovery_EndpointSlice_To_v1beta1_EndpointSlice(&tc.internal, &convertedV1beta1, nil))
			assert.Equal(t, tc.external, convertedV1beta1, "discovery.EndpointSlice -> v1beta1.EndpointSlice")
		})

	}

}

func TestEndpointZoneConverstion(t *testing.T) {
	testcases := []struct {
		desc     string
		external v1beta1.Endpoint
		internal discovery.Endpoint
	}{
		{
			desc:     "no topology field",
			external: v1beta1.Endpoint{},
			internal: discovery.Endpoint{},
		},
		{
			desc: "non empty topology map, but no zone",
			external: v1beta1.Endpoint{
				Topology: map[string]string{
					"key1": "val1",
				},
			},
			internal: discovery.Endpoint{},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			convertedInternal := discovery.Endpoint{}
			require.NoError(t, Convert_v1beta1_Endpoint_To_discovery_Endpoint(&tc.external, &convertedInternal, nil))
			assert.Equal(t, tc.internal, convertedInternal, "v1beta1.Endpoint -> discovery.Endpoint")

			convertedV1beta1 := v1beta1.Endpoint{}
			require.NoError(t, Convert_discovery_Endpoint_To_v1beta1_Endpoint(&tc.internal, &convertedV1beta1, nil))
			assert.Equal(t, tc.external, convertedV1beta1, "discovery.Endpoint -> v1beta1.Endpoint")
		})

	}

}

func TestTopologyMapConversion(t *testing.T) {
	testcases := []struct {
		desc      string
		topMap    map[string]string
		inEPSTop  EndpointSliceTopology
		outEPSTop EndpointSliceTopology
	}{
		{
			desc:     "empty epsTopology input, topology map has zone",
			inEPSTop: EndpointSliceTopology{},
			outEPSTop: EndpointSliceTopology{
				Strings: []string{"key1", "val1", "key2", "val2"},
				Topology: []map[int]int{
					{0: 1, 2: 3},
				},
			},
			topMap: map[string]string{
				"topology.kubernetes.io/zone": "zoneA",
				"key1":                        " val1",
				"key2":                        "val2",
			},
		},
		{
			desc:     "empty epsTopology input, topology map has zone",
			inEPSTop: EndpointSliceTopology{},
			outEPSTop: EndpointSliceTopology{
				Strings: []string{"key1", "val1", "key2", "val2"},
				Topology: []map[int]int{
					{0: 1, 2: 3},
				},
			},
			topMap: map[string]string{
				"key1": " val1",
				"key2": "val2",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			add_endpoint_topology_map(tc.topMap, epsTop)
		})

	}

}
