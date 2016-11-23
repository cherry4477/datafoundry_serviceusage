package openshift

import (
	"fmt"
	"regexp"
	"strings"
)

const DNS1123LabelFmt string = "[a-z0-9]([-a-z0-9]*[a-z0-9])?"
//const DNS1123LabelMaxLength int = 63
const DNS1123LabelMaxLength int = 30
const EmailAddressFmt string = `[-._\w]*\w@[-._\w]*\w\.\w{2,3}` //`(\w[-._\w]*\w@\w[-._\w]*\w\.\w{2,3})`

var NameMayNotBe = []string{".", ".."}
var NameMayNotContain = []string{"/", "%"}
var DNS1123LabelErrorMsg string = fmt.Sprintf(`must be a DNS label (at most %d characters, matching regex %s): e.g. "my-name"`, DNS1123LabelMaxLength, DNS1123LabelFmt)
var dns1123LabelRegexp = regexp.MustCompile("^" + DNS1123LabelFmt + "$")

func ValidateEmail(email string) (bool, string) {
	emailRegexp := regexp.MustCompile(EmailAddressFmt)

	if !emailRegexp.MatchString(email) {
		return false, "email not valid."
	}
	return true, ""

}

func ValidateUserName(name string) (bool, string) {
	if ok, reason := MinimalNameRequirements(name); !ok {
		return ok, reason
	}

	if len(name) < 2 {
		return false, "name must be at least 2 characters long"
	}

	if ok, msg := ValidateNamespaceName(name, false); !ok {
		return ok, msg
	}

	return true, ""
}

func MinimalNameRequirements(name string) (bool, string) {
	for _, illegalName := range NameMayNotBe {
		if name == illegalName {
			return false, fmt.Sprintf(`name may not be %q`, illegalName)
		}
	}

	for _, illegalContent := range NameMayNotContain {
		if strings.Contains(name, illegalContent) {
			return false, fmt.Sprintf(`name may not contain %q`, illegalContent)
		}
	}

	return true, ""
}

func ValidateNamespaceName(name string, prefix bool) (bool, string) {
	return NameIsDNSLabel(name, prefix)
}

// NameIsDNSLabel is a ValidateNameFunc for names that must be a DNS 1123 label.
func NameIsDNSLabel(name string, prefix bool) (bool, string) {
	if prefix {
		name = maskTrailingDash(name)
	}
	if IsDNS1123Label(name) {
		return true, ""
	}
	return false, DNS1123LabelErrorMsg
}

// IsDNS1123Label tests for a string that conforms to the definition of a label in
// DNS (RFC 1123).
func IsDNS1123Label(value string) bool {
	return len(value) <= DNS1123LabelMaxLength && dns1123LabelRegexp.MatchString(value)
}

// maskTrailingDash replaces the final character of a string with a subdomain safe
// value if is a dash.
func maskTrailingDash(name string) string {
	if strings.HasSuffix(name, "-") {
		return name[:len(name)-2] + "a"
	}
	return name
}
