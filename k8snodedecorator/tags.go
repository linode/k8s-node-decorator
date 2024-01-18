package k8snodedecorator

import (
	"fmt"
	"sort"
	"strings"
)

var TagSeparators = []rune{':', '='}

const TagLabelPrefix = "tags.decorator.linode.com/"

func isSeparator(r rune) bool {
	for _, s := range TagSeparators {
		if r == s {
			return true
		}
	}
	return false
}

type KeyValueTag struct {
	Key   string
	Value string
}

func ParseTag(tag string) (result *KeyValueTag) {
	separatorIndex := strings.IndexFunc(tag, isSeparator)

	if separatorIndex == 0 {
		// tag with separator at index 0 is considered as invalid
		// e.g. ":foo", "=bar", and "===:::==="
		result = nil
	} else if separatorIndex == -1 {
		result = &KeyValueTag{
			Key: fmt.Sprintf(TagLabelPrefix + tag),
		}
	} else {
		result = &KeyValueTag{
			Key:   fmt.Sprintf(TagLabelPrefix + tag[:separatorIndex]),
			Value: tag[separatorIndex+1:],
		}
	}

	return result
}

func ParseTags(tags []string) (result []KeyValueTag) {
	sort.Strings(tags)
	for _, tag := range tags {
		parsedTag := ParseTag(tag)
		if parsedTag != nil {
			result = append(result, *parsedTag)
		}
	}

	return result
}
