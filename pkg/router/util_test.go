/*
Copyright 2020 The Flux authors

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

package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIncludeLabelsByPrefix(t *testing.T) {
	labels := map[string]string{
		"foo":   "foo-value",
		"bar":   "bar-value",
		"lorem": "ipsum",
	}
	includeLabelPrefix := []string{"foo", "lor"}

	filteredLabels := includeLabelsByPrefix(labels, includeLabelPrefix)

	assert.Equal(t, filteredLabels, map[string]string{
		"foo":   "foo-value",
		"lorem": "ipsum",
		// bar excluded
	})
}

func TestIncludeLabelsByPrefixWithWildcard(t *testing.T) {
	labels := map[string]string{
		"foo":                                  "foo-value",
		"bar":                                  "bar-value",
		"lorem":                                "ipsum",
		"kustomize.toolkit.fluxcd.io/checksum": "some",
	}
	includeLabelPrefix := []string{"*"}

	filteredLabels := includeLabelsByPrefix(labels, includeLabelPrefix)

	assert.Equal(t, filteredLabels, map[string]string{
		"foo":   "foo-value",
		"bar":   "bar-value",
		"lorem": "ipsum",
	})
}

func TestIncludeLabelsNoIncludes(t *testing.T) {
	labels := map[string]string{
		"foo":   "foo-value",
		"bar":   "bar-value",
		"lorem": "ipsum",
	}
	includeLabelPrefix := []string{""}

	filteredLabels := includeLabelsByPrefix(labels, includeLabelPrefix)

	assert.Equal(t, map[string]string{}, filteredLabels)
}

func TestFilterMetadata(t *testing.T) {
	// Test with nil map
	result := filterMetadata(nil)
	assert.NotNil(t, result)
	assert.Equal(t, "disabled", result["kustomize.toolkit.fluxcd.io/reconcile"])
	assert.Equal(t, "disabled", result["helm.toolkit.fluxcd.io/driftDetection"])

	// Test with existing map
	meta := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	result = filterMetadata(meta)
	assert.Equal(t, "value1", result["key1"])
	assert.Equal(t, "value2", result["key2"])
	assert.Equal(t, "disabled", result["kustomize.toolkit.fluxcd.io/reconcile"])
	assert.Equal(t, "disabled", result["helm.toolkit.fluxcd.io/driftDetection"])

	// Test that it overrides existing toolkit keys
	meta = map[string]string{
		"kustomize.toolkit.fluxcd.io/reconcile": "enabled",
		"key1":                                  "value1",
	}

	result = filterMetadata(meta)
	assert.Equal(t, "value1", result["key1"])
	assert.Equal(t, "disabled", result["kustomize.toolkit.fluxcd.io/reconcile"])
}

func TestIncludeLabelsByPrefixEdgeCases(t *testing.T) {
	// Test with empty labels
	labels := map[string]string{}
	includeLabelPrefix := []string{"foo"}

	filteredLabels := includeLabelsByPrefix(labels, includeLabelPrefix)
	assert.Equal(t, map[string]string{}, filteredLabels)

	// Test with empty prefixes
	labels = map[string]string{
		"foo": "foo-value",
		"bar": "bar-value",
	}
	includeLabelPrefix = []string{}

	filteredLabels = includeLabelsByPrefix(labels, includeLabelPrefix)
	assert.Equal(t, map[string]string{}, filteredLabels)

	// Test with nil prefixes
	filteredLabels = includeLabelsByPrefix(labels, nil)
	assert.Equal(t, map[string]string{}, filteredLabels)

	// Test with multiple matching prefixes
	labels = map[string]string{
		"foo-key": "foo-value",
		"bar-key": "bar-value",
		"baz-key": "baz-value",
	}
	includeLabelPrefix = []string{"foo", "bar"}

	filteredLabels = includeLabelsByPrefix(labels, includeLabelPrefix)
	assert.Equal(t, map[string]string{
		"foo-key": "foo-value",
		"bar-key": "bar-value",
	}, filteredLabels)
}
