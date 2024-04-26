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

package decorator

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var TagSeparators = []rune{':', '='}

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

func ParseTag(tag string, tagLabelPrefix string) (result *KeyValueTag) {
	separatorIndex := strings.IndexFunc(tag, isSeparator)

	if separatorIndex == 0 {
		// tag with separator at index 0 is considered as invalid
		// e.g. ":foo", "=bar", and "===:::==="
		result = nil
	} else if separatorIndex == -1 {
		result = &KeyValueTag{
			Key: fmt.Sprintf(tagLabelPrefix + tag),
		}
	} else {
		result = &KeyValueTag{
			Key:   fmt.Sprintf(tagLabelPrefix + tag[:separatorIndex]),
			Value: tag[separatorIndex+1:],
		}
	}

	return result
}

func ParseTags(tags []string, tagLabelPrefix string) map[string]string {
	sort.Strings(tags)

	result := make(map[string]string)
	for _, tag := range tags {
		parsedTag := ParseTag(tag, tagLabelPrefix)
		if parsedTag != nil {
			result[parsedTag.Key] = parsedTag.Value
		}
	}

	return result
}

func IsValidObjectName(name string) bool {
	pattern := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)
	return pattern.MatchString(name)
}
