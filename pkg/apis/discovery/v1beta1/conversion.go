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
	"encoding/json"
	"fmt"
	"strconv"

	"k8s.io/api/discovery/v1beta1"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/kubernetes/pkg/apis/discovery"
)

const (
	v1beta1Topology     = "endpointslice.kubernetes.io/v1beta1-topology"
	v1beta1ZoneTopology = "topology.kubernetes.io/zone"
)

type EndpointSliceTopology struct {
	Strings  []string      `json:"s"`
	Topology []map[int]int `json:"t"`
}

func Convert_v1beta1_EndpointSlice_To_discovery_EndpointSlice(in *v1beta1.EndpointSlice, out *discovery.EndpointSlice, s conversion.Scope) error {
	// zone conversion is taken care by endpoint conversion
	if err := autoConvert_v1beta1_EndpointSlice_To_discovery_EndpointSlice(in, out, s); err != nil {
		return err
	}

	// Use maps initially to keep track of unique keys and counts of the values
	// keyMap is keep track of unique topology keys. The map values (struct {}) in this map will not be used
	// valueCounts maps a topology value to the number of instances that value is used in all of the Endpoint topologies
	keyMap := make(map[string]struct{})
	valueCounts := make(map[string]int)

	// keysSize and valuesSize keep track of the number of bytes that will be required just for the unique
	// keys and values. These values will be used in approximating the size of the annotation.
	keysSize := 0
	valuesSize := 0

	// gather unique keys, values and value counts
	for _, endpoint := range in.Endpoints {
		for k, v := range endpoint.Topology {
			// Don't include zone, as it will be covered by the zone field on the Endpoint
			if k == v1beta1ZoneTopology {
				continue
			}

			if _, ok := keyMap[k]; !ok {
				keyMap[k] = struct{}{}
				keysSize += len(k)
			}

			if _, ok := valueCounts[v]; !ok {
				valueCounts[v] = 0
				valuesSize += len(v)
			}
			valueCounts[v]++
		}
	}

	// If keyMap is empty, then topology fields is empty and no conversion is needed
	if len(keyMap) == 0 {
		return nil
	}

	approximateSize := approximateAnnotationSize(keyMap, valueCounts, len(in.Endpoints), keysSize, valuesSize)
	maxSize := 256*(1<<10) - sizeOfAnnotations(out.Annotations)

	// Until the approximation is smaller than the maxSize allowed for the annotation, remove a key at a time
	// from the keyMap. keyMap and valueCounts becomes the source of truth of the keys and values that will
	// be preserved and converted into an annotation. The list of strings in the final annotation will be the
	// generated from the keys and values that exist in keyMap and valueCounts.
	for approximateSize > maxSize {
		keyToRemove := ""
		for k := range keyMap {
			// pick a key from the keyMap that is not one of the standard keys
			if k != "kubernetes.io/hostname" && k != "topology.kubernetes.io/region" {
				keyToRemove = k
				break
			}
		}
		// Either no more keys exist or remaining keys are standard keys
		// No other way to preserve the data, so throw an error
		if keyToRemove == "" {
			return fmt.Errorf("unable to convert topology fields on v1beta1.EndpointSlice")
		}

		delete(keyMap, keyToRemove)
		keysSize -= len(keyToRemove)

		// Go through each Endpoint's topology field and remove the key by finding the corresponding value.
		// Reduce the valueCount, and if the count becomes zero, the value is not being used in any of the
		// topology maps and can be removed from the valueCounts map.
		for _, endpoint := range in.Endpoints {
			val, ok := endpoint.Topology[keyToRemove]
			if !ok {
				continue
			}
			valueCounts[val]--
			if valueCounts[val] == 0 {
				delete(valueCounts, val)
				valuesSize -= len(val)
			}
		}

		approximateSize = approximateAnnotationSize(keyMap, valueCounts, len(in.Endpoints), keysSize, valuesSize)
	}

	// strings contains the unique list of topology keys and topology values that will be converted into an
	// annotation.
	var strings []string
	// The `stringIndex` maps the string to its index in the `strings` slice
	stringIndex := make(map[string]int)
	// topologies are the condensend Endpoint topologies that map a the index of the topology key to the
	// index of the topology value. The index of the topology in `topologies` is the same as the the index
	// of the topology's corresponding endpoint resource
	var topologies []map[int]int
	for _, endpoint := range in.Endpoints {
		topology := make(map[int]int)
		for k, v := range endpoint.Topology {
			var kIndex, vIndex int
			var ok bool
			if kIndex, ok = stringIndex[k]; !ok {
				if _, ok := keyMap[k]; !ok {
					continue
				}
				kIndex = len(strings)
				strings = append(strings, k)
				stringIndex[k] = kIndex
			}

			if vIndex, ok = stringIndex[v]; !ok {
				vIndex = len(strings)
				strings = append(strings, v)
			}
			topology[kIndex] = vIndex
		}
		topologies = append(topologies, topology)
	}

	epsTopology := EndpointSliceTopology{Strings: strings, Topology: topologies}
	topologyBytes, err := json.Marshal(epsTopology)
	if err != nil {
		return fmt.Errorf("errored marshaling endpoint slice topology fields: %q", err)
	}

	if out.Annotations == nil {
		out.Annotations = make(map[string]string)
	}
	out.Annotations[v1beta1Topology] = string(topologyBytes)

	return nil
}

func Convert_discovery_EndpointSlice_To_v1beta1_EndpointSlice(in *discovery.EndpointSlice, out *v1beta1.EndpointSlice, s conversion.Scope) error {
	// zone field conversion is taken care by endpoint conversion
	if err := autoConvert_discovery_EndpointSlice_To_v1beta1_EndpointSlice(in, out, s); err != nil {
		return err
	}

	annotation, ok := in.Annotations[v1beta1Topology]
	if !ok {
		return nil
	}

	var epsTop EndpointSliceTopology
	if err := json.Unmarshal([]byte(annotation), &epsTop); err != nil {
		return fmt.Errorf("errored unmarshaling annotation %s : %q", annotation, err)
	}

	for i, endpoint := range out.Endpoints {
		topology := epsTop.Topology[i]
		if endpoint.Topology == nil {
			endpoint.Topology = make(map[string]string)
		}
		for k, v := range topology {
			endpoint.Topology[epsTop.Strings[k]] = epsTop.Strings[v]
		}
		out.Endpoints[i] = endpoint
	}
	if len(out.Annotations) > 1 {
		delete(out.Annotations, v1beta1Topology)
	} else {
		out.Annotations = nil
	}

	return nil
}

func Convert_v1beta1_Endpoint_To_discovery_Endpoint(in *v1beta1.Endpoint, out *discovery.Endpoint, s conversion.Scope) error {
	if err := autoConvert_v1beta1_Endpoint_To_discovery_Endpoint(in, out, s); err != nil {
		return err
	}

	if zone, ok := in.Topology[v1beta1ZoneTopology]; ok {
		out.Zone = &zone
	}

	return nil
}

func Convert_discovery_Endpoint_To_v1beta1_Endpoint(in *discovery.Endpoint, out *v1beta1.Endpoint, s conversion.Scope) error {
	if err := autoConvert_discovery_Endpoint_To_v1beta1_Endpoint(in, out, s); err != nil {
		return err
	}

	if in.Zone != nil {
		if out.Topology == nil {
			out.Topology = make(map[string]string)
		}
		out.Topology[v1beta1ZoneTopology] = *in.Zone
	}

	return nil
}

// siseOfAnnotations returns the size of all the annotations in bytes. The size is calculated
// in the same way as the annotations are validated
func sizeOfAnnotations(annotations map[string]string) int {
	totalSize := 0
	for k, v := range annotations {
		totalSize = len(k) + len(v)
	}
	return totalSize
}

// approximateAnnotationSize approximates the number of bytes the topology annotation would be
func approximateAnnotationSize(keys map[string]struct{}, values map[string]int, numEndpoints, numKeyBytes, numValueBytes int) int {

	numKeys := len(keys)
	numValues := len(values)

	digitsPerKey := len(strconv.Itoa(numKeys))
	digitsPerValue := len(strconv.Itoa(numValues))

	// "endpointslice.kubernetes.io/v1beta1-topology" 44 bytes
	// {"s":[],"t":[]} -> 15 bytes
	// each topology needs {} -> 3 bytes
	// each topology "":"", + keyDigits + valueDigits per key -> 6 + keyDigits + valueDigits

	return 44 + 16 + 3*(numKeys+numValues) + numKeyBytes + numValueBytes + numEndpoints*(3+numKeys*(6+digitsPerKey+digitsPerValue))
}
