/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

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

package cmd

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"
)

const (
	errMalformedURL     = "malformed URL"
	errMissingURLScheme = "missing URL Scheme"
	errMissingURLHost   = "missing URL Host"
	errMissingURLPort   = "missing URL Port"
	errUnexpectedScheme = "unexpected URL Scheme"
	errFQDN             = "not a valid FQDN"
	errFQDNInvalid      = "only letters(a-z|A-Z), digits(0-9) and hyphens('-') for labels"
	errFQDNTrailing     = "trailing dot or hyphens are not allowed"
	errFQDNLen          = "max FQDN length is 253"
	errFQDNLabelLen     = "max FQDN length is 63"
	errFQDNEmptyDot     = "start or end with dots is not allowed"
	errFQDNHyp          = "labels cannot start or end with hyphens"
	errTLDLen           = "TLD length is not valid"
	errTLDLet           = "TLD allow only letters"
)

/*
Hostname FQDN validation
Hostnames are composed of a series of labels concatenated with dots.
Each label is 1 to 63 characters long, and may contain: the ASCII letters
a-z and A-Z, the digits 0-9, and the hyphen ('-'). Additionally: labels
cannot start or end with hyphens (RFC 952) labels can start with
numbers (RFC 1123) trailing dot is not allowed max length of ascii
hostname including dots is 253 characters TLD (last label) is at least 2
characters and only ASCII letters we want at least 1 level above TLD
Source: https://stackoverflow.com/questions/11809631/fully-qualified-domain-
name-validation, answer from JdeBP

<domain> ::= <subdomain> | " "
<subdomain> ::= <label> | <subdomain> "." <label>
<label> ::= <letter> [ [ <ldh-str> ] <let-dig> ]
<ldh-str> ::= <let-dig-hyp> | <let-dig-hyp> <ldh-str>
<let-dig-hyp> ::= <let-dig> | "-"
<let-dig> ::= <letter> | <digit>
<letter> ::= any one of the 52 alphabetic characters A through Z in
upper case and a through z in lower case
<digit> ::= any one of the ten digits 0 through 9
*/

func isHyphen(s rune) bool {
	return s == '-'
}

func isLet(s []rune) bool {
	for l := range s {
		return unicode.IsLetter(s[l])
	}
	return true
}

func isLetDig(s rune) bool {
	return unicode.IsLetter(s) || unicode.IsDigit(s)
}

func isLetDigHyp(s rune) bool {
	return isLetDig(s) || isHyphen(s)
}

func maxFQDNLen(s string) bool {
	return 63 > len(s) || 253 > len(s)
}

func trailingDotHyp(s string) bool {
	return strings.Contains(s, "..") || strings.Contains(s, "--")
}

func isLabel(label []rune) bool {

	for _, s := range label {
		if !isLetDigHyp(rune(s)) {
			return false
		}
	}

	if !isLetDigHyp(label[0]) {
		return false
	}

	return true
}

func isSubDomain(names []string) bool {
	for _, s := range names {
		if len(s) < 1 || !isLabel([]rune(s)) {
			return false
		}
	}
	return true
}

func isValidFQDN(fqdn string) error {

	components := strings.Split(fqdn, ".")
	label := []rune(components[len(components)-1])

	// Check TLD length.
	if 1 == len(label) {
		return fmt.Errorf("%s in %s", errTLDLen, fqdn)
	}

	// Check if TLD valid. Only letters ar allowed.
	if !isLet(label) {
		return fmt.Errorf("%s in %s", errTLDLet, fqdn)
	}

	// Check for invalid trailings.
	if trailingDotHyp(fqdn) {
		return fmt.Errorf("%s in %s", errFQDNTrailing, fqdn)
	}

	// Check if FQDN have valid length.
	if 253 < len(fqdn) {
		return fmt.Errorf("%s in %s", errFQDNLen, fqdn)
	}

	for c := range components {
		l := components[c]

		// Check if label have valid length.
		if 63 < len(l) {
			return fmt.Errorf("%s in %s", errFQDNLabelLen, fqdn)
		}

		// Check for empty values in components. Eg: acme.net.
		if l == "" {
			return fmt.Errorf("%s in %s", errFQDNEmptyDot, fqdn)
		}

		// Check for hyphens at begining|end.
		if isHyphen(rune(l[0])) || isHyphen(rune(l[(len(l)-1)])) {
			return fmt.Errorf("%s in %s", errFQDNHyp, fqdn)
		}
	}

	if !isSubDomain(components) {
		return fmt.Errorf("%s in %s", errFQDNInvalid, fqdn)
	}

	return nil
}

// urlParse function checks that given string parameters are valid URLs for
// REST requests: schema + hostname [ + port ]
func urlParse(endpoints ...string) error {
	for _, endpoint := range endpoints {
		url, err := url.Parse(endpoint)

		if err != nil {
			return fmt.Errorf("%s in %s", errMalformedURL, endpoint)
		}

		if url.Scheme == "" {
			return fmt.Errorf("%s in %s", errMissingURLScheme, endpoint)
		}

		if url.Hostname() == "" {
			return fmt.Errorf("%s in %s", errMissingURLHost, endpoint)
		}
	}
	return nil
}

// urlParse function checks that given string parameters are valid IPs for
// binding services: hostname + port. No schema is required.
func urlParseNoSchemaRequired(endpoints ...string) error {
	for _, endpoint := range endpoints {

		if strings.Contains(endpoint, "://") {
			return fmt.Errorf("%s in %s", errUnexpectedScheme, endpoint)
		}

		// Add fake scheme to get an expected result from url.Parse
		url, err := url.Parse("http://" + endpoint)

		if err != nil {
			return fmt.Errorf("%s in %s", errMalformedURL, endpoint)
		}

		if url.Hostname() == "" {
			return fmt.Errorf("%s in %s", errMissingURLHost, endpoint)
		}

		if url.Port() == "" {
			return fmt.Errorf("%s in %s", errMissingURLPort, endpoint)
		}
	}
	return nil
}
