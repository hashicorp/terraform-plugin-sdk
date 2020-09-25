package acctest

import (
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// SkipOnErrorContains helps skip tests based on a portion of an error message
// which is contained in a complete error message.
func SkipOnErrorContains(s string) resource.SkipOnErrorFunc {
	return func(err error) bool {
		if err == nil {
			return false
		}

		return strings.Contains(err.Error(), s)
	}
}

// SkipOnErrorRegexp helps skip tests based on a regexp that matches an error
// message.
func SkipOnErrorRegexp(re *regexp.Regexp) resource.SkipOnErrorFunc {
	return func(err error) bool {
		if err == nil {
			return false
		}

		return re.MatchString(err.Error())
	}
}
