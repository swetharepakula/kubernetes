/*
Copyright 2020 The Kubernetes Authors.

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

package endpointslice

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	featuregatetesting "k8s.io/component-base/featuregate/testing"
	"k8s.io/kubernetes/pkg/apis/discovery"
	"k8s.io/kubernetes/pkg/features"
	utilpointer "k8s.io/utils/pointer"
)

func Test_dropTopologyOnV1(t *testing.T) {
	testcases := []struct {
		name        string
		v1Request   bool
		eps         *discovery.EndpointSlice
		expectedEPS *discovery.EndpointSlice
	}{
		{
			name:      "v1 request, without deprecated topology",
			v1Request: true,
			eps: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hostname: utilpointer.StringPtr("hostname-1"),
					},
					{
						Hostname: utilpointer.StringPtr("hostname-1"),
					},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hostname: utilpointer.StringPtr("hostname-1"),
					},
					{
						Hostname: utilpointer.StringPtr("hostname-1"),
					},
				},
			},
		},
		{
			name: "v1beta1 request, without deprecated topology",
			eps: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hostname: utilpointer.StringPtr("hostname-1"),
					},
					{
						Hostname: utilpointer.StringPtr("hostname-1"),
					},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hostname: utilpointer.StringPtr("hostname-1"),
					},
					{
						Hostname: utilpointer.StringPtr("hostname-1"),
					},
				},
			},
		},
		{
			name:      "v1 request, with deprecated topology",
			v1Request: true,
			eps: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						DeprecatedTopology: map[string]string{
							"key": "value",
						},
					},
					{
						DeprecatedTopology: map[string]string{
							"key": "value",
						},
					},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{{}, {}},
			},
		},
		{
			name: "v1beta1 request, with deprecated topology",
			eps: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						DeprecatedTopology: map[string]string{
							"key": "value",
						},
					},
					{
						DeprecatedTopology: map[string]string{
							"key": "value",
						},
					},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						DeprecatedTopology: map[string]string{
							"key": "value",
						},
					},
					{
						DeprecatedTopology: map[string]string{
							"key": "value",
						},
					},
				},
			},
		},
		{
			name:      "v1 request, with nodeName",
			v1Request: true,
			eps: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						NodeName: utilpointer.StringPtr("node-1"),
						DeprecatedTopology: map[string]string{
							corev1.LabelHostname: "node-2",
						},
					},
					{
						NodeName: utilpointer.StringPtr("node-1"),
						DeprecatedTopology: map[string]string{
							corev1.LabelHostname: "node-2",
						},
					},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						NodeName: utilpointer.StringPtr("node-1"),
						DeprecatedTopology: map[string]string{
							corev1.LabelHostname: "node-1",
						},
					},
					{
						NodeName: utilpointer.StringPtr("node-1"),
						DeprecatedTopology: map[string]string{
							corev1.LabelHostname: "node-1",
						},
					},
				},
			},
		},
		{
			name: "v1beta1 request, with nodeName",
			eps: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						NodeName: utilpointer.StringPtr("node-1"),
						DeprecatedTopology: map[string]string{
							corev1.LabelHostname: "node-2",
						},
					},
					{
						NodeName: utilpointer.StringPtr("node-1"),
						DeprecatedTopology: map[string]string{
							corev1.LabelHostname: "node-2",
						},
					},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						NodeName: utilpointer.StringPtr("node-1"),
						DeprecatedTopology: map[string]string{
							corev1.LabelHostname: "node-2",
						},
					},
					{
						NodeName: utilpointer.StringPtr("node-1"),
						DeprecatedTopology: map[string]string{
							corev1.LabelHostname: "node-2",
						},
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := genericapirequest.WithRequestInfo(genericapirequest.NewContext(), &genericapirequest.RequestInfo{APIGroup: "discovery.k8s.io", APIVersion: "v1beta1", Resource: "endpointslices"})
			if tc.v1Request {
				ctx = genericapirequest.WithRequestInfo(genericapirequest.NewContext(), &genericapirequest.RequestInfo{APIGroup: "discovery.k8s.io", APIVersion: "v1", Resource: "endpointslices"})
			}

			dropTopologyOnV1(ctx, tc.eps)
			if !apiequality.Semantic.DeepEqual(tc.eps, tc.expectedEPS) {
				t.Logf("actual endpointslice: %v", tc.eps)
				t.Logf("expected endpointslice: %v", tc.expectedEPS)
				t.Errorf("unexpected EndpointSlice on create API strategy")
			}
		})
	}
}

func Test_dropDisabledFieldsOnCreate(t *testing.T) {
	testcases := []struct {
		name                   string
		terminatingGateEnabled bool
		eps                    *discovery.EndpointSlice
		expectedEPS            *discovery.EndpointSlice
	}{
		{
			name:                   "terminating gate enabled, field should be allowed",
			terminatingGateEnabled: true,
			eps: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Conditions: discovery.EndpointConditions{
							Serving:     utilpointer.BoolPtr(true),
							Terminating: utilpointer.BoolPtr(false),
						},
					},
					{
						Conditions: discovery.EndpointConditions{
							Serving:     utilpointer.BoolPtr(true),
							Terminating: utilpointer.BoolPtr(true),
						},
					},
					{
						Conditions: discovery.EndpointConditions{
							Serving:     nil,
							Terminating: nil,
						},
					},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Conditions: discovery.EndpointConditions{
							Serving:     utilpointer.BoolPtr(true),
							Terminating: utilpointer.BoolPtr(false),
						},
					},
					{
						Conditions: discovery.EndpointConditions{
							Serving:     utilpointer.BoolPtr(true),
							Terminating: utilpointer.BoolPtr(true),
						},
					},
					{
						Conditions: discovery.EndpointConditions{
							Serving:     nil,
							Terminating: nil,
						},
					},
				},
			},
		},
		{
			name:                   "terminating gate disabled, field should be set to nil",
			terminatingGateEnabled: false,
			eps: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Conditions: discovery.EndpointConditions{
							Serving:     utilpointer.BoolPtr(true),
							Terminating: utilpointer.BoolPtr(false),
						},
					},
					{
						Conditions: discovery.EndpointConditions{
							Serving:     utilpointer.BoolPtr(true),
							Terminating: utilpointer.BoolPtr(true),
						},
					},
					{
						Conditions: discovery.EndpointConditions{
							Serving:     nil,
							Terminating: nil,
						},
					},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Conditions: discovery.EndpointConditions{
							Serving:     nil,
							Terminating: nil,
						},
					},
					{
						Conditions: discovery.EndpointConditions{
							Serving:     nil,
							Terminating: nil,
						},
					},
					{
						Conditions: discovery.EndpointConditions{
							Serving:     nil,
							Terminating: nil,
						},
					},
				},
			},
		},
		{
			name: "node name gate enabled, field should be allowed",
			eps: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						NodeName: utilpointer.StringPtr("node-1"),
					},
					{
						NodeName: utilpointer.StringPtr("node-2"),
					},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						NodeName: utilpointer.StringPtr("node-1"),
					},
					{
						NodeName: utilpointer.StringPtr("node-2"),
					},
				},
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			defer featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.EndpointSliceTerminatingCondition, testcase.terminatingGateEnabled)()

			dropDisabledFieldsOnCreate(testcase.eps)
			if !apiequality.Semantic.DeepEqual(testcase.eps, testcase.expectedEPS) {
				t.Logf("actual endpointslice: %v", testcase.eps)
				t.Logf("expected endpointslice: %v", testcase.expectedEPS)
				t.Errorf("unexpected EndpointSlice on create API strategy")
			}
		})
	}
}

func Test_dropDisabledFieldsOnUpdate(t *testing.T) {
	testcases := []struct {
		name                   string
		terminatingGateEnabled bool
		oldEPS                 *discovery.EndpointSlice
		newEPS                 *discovery.EndpointSlice
		expectedEPS            *discovery.EndpointSlice
	}{
		{
			name:                   "terminating gate enabled, field should be allowed",
			terminatingGateEnabled: true,
			oldEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Conditions: discovery.EndpointConditions{
							Serving:     utilpointer.BoolPtr(true),
							Terminating: utilpointer.BoolPtr(false),
						},
					},
					{
						Conditions: discovery.EndpointConditions{
							Serving:     utilpointer.BoolPtr(true),
							Terminating: utilpointer.BoolPtr(true),
						},
					},
					{
						Conditions: discovery.EndpointConditions{
							Serving:     nil,
							Terminating: nil,
						},
					},
				},
			},
			newEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Conditions: discovery.EndpointConditions{
							Serving:     utilpointer.BoolPtr(true),
							Terminating: utilpointer.BoolPtr(false),
						},
					},
					{
						Conditions: discovery.EndpointConditions{
							Serving:     utilpointer.BoolPtr(true),
							Terminating: utilpointer.BoolPtr(true),
						},
					},
					{
						Conditions: discovery.EndpointConditions{
							Serving:     nil,
							Terminating: nil,
						},
					},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Conditions: discovery.EndpointConditions{
							Serving:     utilpointer.BoolPtr(true),
							Terminating: utilpointer.BoolPtr(false),
						},
					},
					{
						Conditions: discovery.EndpointConditions{
							Serving:     utilpointer.BoolPtr(true),
							Terminating: utilpointer.BoolPtr(true),
						},
					},
					{
						Conditions: discovery.EndpointConditions{
							Serving:     nil,
							Terminating: nil,
						},
					},
				},
			},
		},
		{
			name:                   "terminating gate disabled, and not set on existing EPS",
			terminatingGateEnabled: false,
			oldEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Conditions: discovery.EndpointConditions{
							Serving:     nil,
							Terminating: nil,
						},
					},
					{
						Conditions: discovery.EndpointConditions{
							Serving:     nil,
							Terminating: nil,
						},
					},
					{
						Conditions: discovery.EndpointConditions{
							Serving:     nil,
							Terminating: nil,
						},
					},
				},
			},
			newEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Conditions: discovery.EndpointConditions{
							Serving:     utilpointer.BoolPtr(true),
							Terminating: utilpointer.BoolPtr(false),
						},
					},
					{
						Conditions: discovery.EndpointConditions{
							Serving:     utilpointer.BoolPtr(true),
							Terminating: utilpointer.BoolPtr(true),
						},
					},
					{
						Conditions: discovery.EndpointConditions{
							Serving:     nil,
							Terminating: nil,
						},
					},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Conditions: discovery.EndpointConditions{
							Serving:     nil,
							Terminating: nil,
						},
					},
					{
						Conditions: discovery.EndpointConditions{
							Serving:     nil,
							Terminating: nil,
						},
					},
					{
						Conditions: discovery.EndpointConditions{
							Serving:     nil,
							Terminating: nil,
						},
					},
				},
			},
		},
		{
			name:                   "terminating gate disabled, and set on existing EPS",
			terminatingGateEnabled: false,
			oldEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Conditions: discovery.EndpointConditions{
							Serving:     utilpointer.BoolPtr(true),
							Terminating: utilpointer.BoolPtr(false),
						},
					},
					{
						Conditions: discovery.EndpointConditions{
							Serving:     utilpointer.BoolPtr(true),
							Terminating: utilpointer.BoolPtr(true),
						},
					},
					{
						Conditions: discovery.EndpointConditions{
							Serving:     nil,
							Terminating: nil,
						},
					},
				},
			},
			newEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Conditions: discovery.EndpointConditions{
							Serving:     utilpointer.BoolPtr(true),
							Terminating: utilpointer.BoolPtr(false),
						},
					},
					{
						Conditions: discovery.EndpointConditions{
							Serving:     utilpointer.BoolPtr(true),
							Terminating: utilpointer.BoolPtr(true),
						},
					},
					{
						Conditions: discovery.EndpointConditions{
							Serving:     nil,
							Terminating: nil,
						},
					},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Conditions: discovery.EndpointConditions{
							Serving:     utilpointer.BoolPtr(true),
							Terminating: utilpointer.BoolPtr(false),
						},
					},
					{
						Conditions: discovery.EndpointConditions{
							Serving:     utilpointer.BoolPtr(true),
							Terminating: utilpointer.BoolPtr(true),
						},
					},
					{
						Conditions: discovery.EndpointConditions{
							Serving:     nil,
							Terminating: nil,
						},
					},
				},
			},
		},
		{
			name:                   "terminating gate disabled, and set on existing EPS with new values",
			terminatingGateEnabled: false,
			oldEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Conditions: discovery.EndpointConditions{
							Serving:     utilpointer.BoolPtr(false),
							Terminating: utilpointer.BoolPtr(false),
						},
					},
					{
						Conditions: discovery.EndpointConditions{
							Serving:     utilpointer.BoolPtr(true),
							Terminating: utilpointer.BoolPtr(true),
						},
					},
					{
						Conditions: discovery.EndpointConditions{
							Terminating: nil,
						},
					},
				},
			},
			newEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Conditions: discovery.EndpointConditions{
							Serving:     utilpointer.BoolPtr(true),
							Terminating: utilpointer.BoolPtr(true),
						},
					},
					{
						Conditions: discovery.EndpointConditions{
							Serving:     utilpointer.BoolPtr(false),
							Terminating: utilpointer.BoolPtr(false),
						},
					},
					{
						Conditions: discovery.EndpointConditions{
							Terminating: utilpointer.BoolPtr(false),
						},
					},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Conditions: discovery.EndpointConditions{
							Serving:     utilpointer.BoolPtr(true),
							Terminating: utilpointer.BoolPtr(true),
						},
					},
					{
						Conditions: discovery.EndpointConditions{
							Serving:     utilpointer.BoolPtr(false),
							Terminating: utilpointer.BoolPtr(false),
						},
					},
					{
						Conditions: discovery.EndpointConditions{
							Terminating: utilpointer.BoolPtr(false),
						},
					},
				},
			},
		},
		{
			name: "node name gate enabled, set on new EPS",
			oldEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						NodeName: nil,
					},
					{
						NodeName: nil,
					},
				},
			},
			newEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						NodeName: utilpointer.StringPtr("node-1"),
					},
					{
						NodeName: utilpointer.StringPtr("node-2"),
					},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						NodeName: utilpointer.StringPtr("node-1"),
					},
					{
						NodeName: utilpointer.StringPtr("node-2"),
					},
				},
			},
		},
		{
			name: "node name gate disabled, set on old and updated EPS",
			oldEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						NodeName: utilpointer.StringPtr("node-1-old"),
					},
					{
						NodeName: utilpointer.StringPtr("node-2-old"),
					},
				},
			},
			newEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						NodeName: utilpointer.StringPtr("node-1"),
					},
					{
						NodeName: utilpointer.StringPtr("node-2"),
					},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						NodeName: utilpointer.StringPtr("node-1"),
					},
					{
						NodeName: utilpointer.StringPtr("node-2"),
					},
				},
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			defer featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.EndpointSliceTerminatingCondition, testcase.terminatingGateEnabled)()

			dropDisabledFieldsOnUpdate(testcase.oldEPS, testcase.newEPS)
			if !apiequality.Semantic.DeepEqual(testcase.newEPS, testcase.expectedEPS) {
				t.Logf("actual endpointslice: %v", testcase.newEPS)
				t.Logf("expected endpointslice: %v", testcase.expectedEPS)
				t.Errorf("unexpected EndpointSlice from update API strategy")
			}
		})
	}
}
