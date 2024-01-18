package k8snodedecorator_test

import (
	"reflect"
	"testing"

	decorator "github.com/linode/k8s-node-decorator/k8snodedecorator"
)

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

	parsedTags := decorator.ParseTags(invalidTags)
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
		parsedTag := decorator.ParseTag(tag)
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
	parsedTags := decorator.ParseTags(tags)
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
		decorator.TagLabelPrefix + "foo",
		decorator.TagLabelPrefix + "bar",
		decorator.TagLabelPrefix + "a",
	}
	expectedValues := []string{"", "", ""}
	testParseTag(t, expectedKeys, expectedValues, keyOnlyTags)

	expectedResults := map[string]string{
		decorator.TagLabelPrefix + "foo": "",
		decorator.TagLabelPrefix + "bar": "",
		decorator.TagLabelPrefix + "a":   "",
	}
	testParseTags(t, expectedResults, keyOnlyTags)
}

func TestParseKeyValueTags(t *testing.T) {
	keyValueTags := []string{"foo=bar", "foo:bar", "a=b", "a:b"}
	expectedKeys := []string{
		decorator.TagLabelPrefix + "foo",
		decorator.TagLabelPrefix + "foo",
		decorator.TagLabelPrefix + "a",
		decorator.TagLabelPrefix + "a",
	}
	expectedValues := []string{"bar", "bar", "b", "b"}
	testParseTag(t, expectedKeys, expectedValues, keyValueTags)

	expectedResults := map[string]string{
		decorator.TagLabelPrefix + "foo": "bar",
		decorator.TagLabelPrefix + "a":   "b",
	}
	testParseTags(t, expectedResults, keyValueTags)
}

func TestOutOfOrderTags(t *testing.T) {
	tags1 := []string{"foo", "bar", "baz=qux", "bar:quux", "foo=bar"}
	tags2 := []string{"foo=bar", "baz=qux", "bar:quux", "foo", "bar"}
	tags3 := []string{"foo", "foo=bar", "bar:quux", "bar", "baz=qux"}

	parsedTags1 := decorator.ParseTags(tags1)
	parsedTags2 := decorator.ParseTags(tags2)
	parsedTags3 := decorator.ParseTags(tags3)

	if !(reflect.DeepEqual(parsedTags1, parsedTags2) && reflect.DeepEqual(parsedTags1, parsedTags3)) {
		t.Error("Tags parser should return consistent result regardless of order of raw tags")
	}

	expectedResults := map[string]string{
		decorator.TagLabelPrefix + "bar": "quux",
		decorator.TagLabelPrefix + "foo": "bar",
		decorator.TagLabelPrefix + "baz": "qux",
	}
	testParseTags(t, expectedResults, tags1)
	testParseTags(t, expectedResults, tags2)
	testParseTags(t, expectedResults, tags3)
}
