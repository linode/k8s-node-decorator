// Copyright 2024 Akamai Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package decorator_test

import (
	"reflect"
	"testing"

	"github.com/linode/k8s-node-decorator/pkg/decorator"
)

const defaultTagLabelPrefix = "tags.decorator.linode.com/"

func TestParseInvalidTag(t *testing.T) {
	invalidTags := []string{
		":foo",
		"=bar",
		":f",
		"=b",
		":",
		"=",
		":::",
		"===",
		"=foo=bar",
		":foo:bar",
		"=foo:bar",
		":foo=bar",
	}

	parsedTags := decorator.ParseTags(invalidTags, defaultTagLabelPrefix)
	if len(parsedTags) > 0 {
		t.Errorf(
			"None of the invalid tags (%v) should be parsed but got parsed as %v.",
			invalidTags, parsedTags,
		)
	}
}

func testParseTag(
	t *testing.T, expectedKeys, expectedValues, tags []string,
) {
	t.Helper()

	for i, tag := range tags {
		parsedTag := decorator.ParseTag(tag, defaultTagLabelPrefix)
		if parsedTag.Key != expectedKeys[i] {
			t.Errorf("Expected key '%s' but got '%s'", expectedKeys[i], parsedTag.Key)
		}
		if parsedTag.Value != expectedValues[i] {
			t.Errorf("Expected value '%s' but got '%s'", expectedValues[i], parsedTag.Value)
		}
	}
}

func testParseTags(
	t *testing.T, expectedResults map[string]string, tags []string,
) {
	t.Helper()

	parsedTags := decorator.ParseTags(tags, defaultTagLabelPrefix)
	if len(parsedTags) != len(expectedResults) {
		t.Errorf(
			"Length of parsed tags (%d) doesn't equal to length of expected results (%d)",
			len(parsedTags), len(expectedResults),
		)
	}

	for key, value := range parsedTags {
		if value != expectedResults[key] {
			t.Errorf(
				"Expected value '%s' of key '%s' but got '%s'",
				expectedResults[key], key, value,
			)
		}
	}
}

func TestParseKeyOnlyTags(t *testing.T) {
	keyOnlyTags := []string{"foo=", "bar", "a"}
	expectedKeys := []string{
		defaultTagLabelPrefix + "foo",
		defaultTagLabelPrefix + "bar",
		defaultTagLabelPrefix + "a",
	}
	expectedValues := []string{"", "", ""}
	testParseTag(t, expectedKeys, expectedValues, keyOnlyTags)

	expectedResults := map[string]string{
		defaultTagLabelPrefix + "foo": "",
		defaultTagLabelPrefix + "bar": "",
		defaultTagLabelPrefix + "a":   "",
	}
	testParseTags(t, expectedResults, keyOnlyTags)
}

func TestParseKeyValueTags(t *testing.T) {
	keyValueTags := []string{"foo=bar", "foo:bar", "a=b", "a:b"}
	expectedKeys := []string{
		defaultTagLabelPrefix + "foo",
		defaultTagLabelPrefix + "foo",
		defaultTagLabelPrefix + "a",
		defaultTagLabelPrefix + "a",
	}
	expectedValues := []string{"bar", "bar", "b", "b"}
	testParseTag(t, expectedKeys, expectedValues, keyValueTags)

	expectedResults := map[string]string{
		defaultTagLabelPrefix + "foo": "bar",
		defaultTagLabelPrefix + "a":   "b",
	}
	testParseTags(t, expectedResults, keyValueTags)
}

func TestOutOfOrderTags(t *testing.T) {
	tags1 := []string{"foo", "bar", "baz=qux", "bar:quux", "foo=bar"}
	tags2 := []string{"foo=bar", "baz=qux", "bar:quux", "foo", "bar"}
	tags3 := []string{"foo", "foo=bar", "bar:quux", "bar", "baz=qux"}

	parsedTags1 := decorator.ParseTags(tags1, defaultTagLabelPrefix)
	parsedTags2 := decorator.ParseTags(tags2, defaultTagLabelPrefix)
	parsedTags3 := decorator.ParseTags(tags3, defaultTagLabelPrefix)

	if !(reflect.DeepEqual(parsedTags1, parsedTags2) && reflect.DeepEqual(parsedTags1, parsedTags3)) {
		t.Error("Tags parser should return consistent result regardless of order of raw tags")
	}

	expectedResults := map[string]string{
		defaultTagLabelPrefix + "bar": "quux",
		defaultTagLabelPrefix + "foo": "bar",
		defaultTagLabelPrefix + "baz": "qux",
	}
	testParseTags(t, expectedResults, tags1)
	testParseTags(t, expectedResults, tags2)
	testParseTags(t, expectedResults, tags3)
}

func TestIsValidName(t *testing.T) {
	testCases := []struct {
		name    string
		isValid bool
	}{
		{"valid-name-123", true},
		{"InvalidName-1", false},
		{"invalid_name-2", false},
		{"inv@lid-name-3", false},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			isValid := decorator.IsValidObjectName(tc.name)
			if tc.isValid != isValid {
				t.Errorf("%s validity should be: %v, got: %v", tc.name, tc.isValid, isValid)
			}
		})
	}
}
