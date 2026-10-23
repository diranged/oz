/*
Copyright 2022 Matt Wise.

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

// Package imagepolicy implements the operator-wide allow-list that governs
// which container images a PodAccessRequest may run via its `spec.image`
// override.
//
// The allow-list is deliberately a deployment-time setting on the controller
// itself (see the `--allowed-image-patterns` flag) rather than a field on a
// PodAccessTemplate. The set of registries an organization is willing to run
// code from is a cluster-wide security boundary, and letting template authors
// widen it would defeat the purpose of having one.
//
// The zero value of a Policy allows nothing, so an operator that is deployed
// without any configuration rejects every image override.
package imagepolicy

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
)

const (
	// maxImageLength is an upper bound on the length of an image reference we
	// are willing to evaluate. The OCI distribution spec caps a repository
	// name at 255 characters; 512 leaves generous room for a registry host,
	// a port and a digest without accepting unbounded input.
	maxImageLength = 512
)

// validImageChars matches the characters that may legally appear in an image
// reference. This is checked before any pattern matching so that neither a
// hostile nor a fat-fingered value can smuggle newlines or shell
// metacharacters into a PodSpec.
var validImageChars = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._:/@+-]*$`)

// ErrDisabled is returned when a request asks for an image override but the
// controller has no allow-list patterns configured at all.
var ErrDisabled = errors.New(
	"image overrides are disabled on this cluster: the Oz controller was deployed without any --allowed-image-patterns",
)

// Policy is an immutable, compiled set of allow-list patterns.
type Policy struct {
	patterns []string
	matchers []*regexp.Regexp
}

// New compiles a Policy from a list of glob patterns.
//
// Patterns are matched against the image reference exactly as the user wrote
// it - Oz does not normalize a bare `nginx` into `docker.io/library/nginx`.
// Operators should therefore write patterns for the fully-qualified form their
// developers actually use.
//
// Two wildcards are supported:
//
//	"*"  matches any run of characters except the path separator "/"
//	"**" matches any run of characters, including "/"
//
// The distinction matters for security. Under `*`-matches-everything
// semantics, a pattern like `*.dkr.ecr.*.amazonaws.com/team/*` would also
// match `evil.example.com/x.dkr.ecr.us-west-2.amazonaws.com/team/backdoor`,
// because the leading wildcard would happily swallow a registry host that the
// operator never intended to trust. Confining `*` to a single path segment
// closes that hole; `**` remains available when an operator genuinely wants to
// match across segments.
//
// Empty and whitespace-only patterns are ignored, which keeps a Helm value of
// `[]` (rendered as an empty flag) from being mistaken for a real pattern.
func New(patterns []string) (*Policy, error) {
	p := &Policy{}
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}

		matcher, err := compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid image pattern %q: %w", pattern, err)
		}

		p.patterns = append(p.patterns, pattern)
		p.matchers = append(p.matchers, matcher)
	}
	return p, nil
}

// compile turns a glob pattern into an anchored regular expression.
func compile(pattern string) (*regexp.Regexp, error) {
	var b strings.Builder
	b.WriteString("^")
	for i := 0; i < len(pattern); i++ {
		switch {
		case pattern[i] == '*' && i+1 < len(pattern) && pattern[i+1] == '*':
			// "**" - cross path separators.
			b.WriteString(".*")
			i++
		case pattern[i] == '*':
			// "*" - stay within a single path segment.
			b.WriteString("[^/]*")
		default:
			b.WriteString(regexp.QuoteMeta(string(pattern[i])))
		}
	}
	b.WriteString("$")
	return regexp.Compile(b.String())
}

// Enabled reports whether any patterns are configured. When false, every
// non-empty image override is rejected.
func (p *Policy) Enabled() bool {
	return p != nil && len(p.matchers) > 0
}

// Patterns returns the configured patterns, for logging and error messages.
func (p *Policy) Patterns() []string {
	if p == nil {
		return nil
	}
	return append([]string(nil), p.patterns...)
}

// Validate checks an image reference against the policy.
//
// An empty image is always allowed: it means the caller did not ask for an
// override, and the image inherited from the target controller's PodSpec will
// be used instead.
func (p *Policy) Validate(image string) error {
	if image == "" {
		return nil
	}

	if !p.Enabled() {
		return ErrDisabled
	}

	if len(image) > maxImageLength {
		return fmt.Errorf(
			"invalid image reference: must be %d characters or fewer, got %d",
			maxImageLength,
			len(image),
		)
	}

	if !validImageChars.MatchString(image) {
		return fmt.Errorf(
			"invalid image reference %q: must begin with an alphanumeric character and contain only alphanumerics and the characters . _ : / @ + -",
			image,
		)
	}

	for _, matcher := range p.matchers {
		if matcher.MatchString(image) {
			return nil
		}
	}

	return fmt.Errorf(
		"image %q is not permitted on this cluster: it must match one of the allowed patterns %v",
		image,
		p.patterns,
	)
}

// The active Policy is process-global state, set once during controller
// startup. A global is used because the admission webhook handlers are
// dispatched on the API type itself (see internal/webhook) and have no
// constructor to inject configuration through.
var (
	activeMu sync.RWMutex
	active   = &Policy{}
)

// SetActive installs the Policy used by the admission webhooks and the access
// builders. It is intended to be called once, from the controller's main()
// before the manager starts.
func SetActive(p *Policy) {
	activeMu.Lock()
	defer activeMu.Unlock()
	if p == nil {
		p = &Policy{}
	}
	active = p
}

// Active returns the Policy installed by SetActive. It never returns nil; the
// default denies every image override.
func Active() *Policy {
	activeMu.RLock()
	defer activeMu.RUnlock()
	return active
}
