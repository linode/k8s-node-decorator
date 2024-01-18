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

func checkParsedTags(
	t *testing.T, expectedKeys, expectedValues []string, parsedTags []decorator.KeyValueTag,
) {
	t.Helper()

	if len(expectedKeys) != len(parsedTags) || len(expectedValues) != len(parsedTags) {
		t.Fatalf(
			"length of expected keys (%d) or expected keys (%d) mismatches with parsed tags (%d)",
			len(expectedKeys), len(expectedValues), len(parsedTags),
		)
	}

	for i, parsedTag := range parsedTags {
		if parsedTag.Key != expectedKeys[i] {
			t.Errorf(
				"Expected key '%s' but got '%s'",
				expectedKeys[i], parsedTag.Key,
			)
		}

		if parsedTag.Value != expectedValues[i] {
			t.Errorf(
				"Expected value '%s' but got '%s'",
				expectedValues[i], parsedTag.Value,
			)
		}
	}
}

func TestParseKeyOnlyTag(t *testing.T) {
	keyOnlyTags := []string{"foo=", "bar", "a"}
	expectedKeys := []string{
		decorator.TagLabelPrefix + "foo",
		decorator.TagLabelPrefix + "bar",
		decorator.TagLabelPrefix + "a",
	}
	expectedValues := []string{"", "", ""}

	parsedTags := decorator.ParseTags(keyOnlyTags)
	if len(parsedTags) < len(keyOnlyTags) {
		t.Errorf(
			"All valid key only tags (%v) should be parsed but some were missing after parsing (%v)",
			keyOnlyTags, parsedTags,
		)
	}

	checkParsedTags(t, expectedKeys, expectedValues, parsedTags)
}

func TestParseKeyValueTag(t *testing.T) {
	keyValueTags := []string{"foo=bar", "foo:bar", "a=b", "a:b"}
	expectedKeys := []string{
		decorator.TagLabelPrefix + "foo",
		decorator.TagLabelPrefix + "foo",
		decorator.TagLabelPrefix + "a",
		decorator.TagLabelPrefix + "a",
	}
	expectedValues := []string{"bar", "bar", "b", "b"}

	parsedTags := decorator.ParseTags(keyValueTags)
	if len(parsedTags) < len(keyValueTags) {
		t.Errorf(
			"All valid key only tags (%v) should be parsed but some were missing after parsing (%v)",
			keyValueTags, parsedTags,
		)
	}

	checkParsedTags(t, expectedKeys, expectedValues, parsedTags)
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
}
