package imagepolicy

import (
	"errors"
	"strings"
	"testing"
)

const (
	ecrPattern   = "*.dkr.ecr.*.amazonaws.com/team/*"
	ghcrPattern  = "ghcr.io/example/*"
	matchAll     = "**"
	unqualified  = "nginx:latest"
	digestLength = 64

	// A reference that ecrPattern is expected to allow.
	allowedECRImage = "1234567890.dkr.ecr.us-west-2.amazonaws.com/team/web:abc123"
)

func TestNewIgnoresEmptyPatterns(t *testing.T) {
	p, err := New([]string{"", "   ", ghcrPattern})
	if err != nil {
		t.Fatalf("New() returned an unexpected error: %s", err)
	}
	if got := p.Patterns(); len(got) != 1 || got[0] != ghcrPattern {
		t.Errorf("Patterns() = %v, want [%s]", got, ghcrPattern)
	}
}

func TestNewTreatsRegexMetacharactersAsLiterals(t *testing.T) {
	// A stray "(" would be a regex syntax error if it were not escaped.
	if _, err := New([]string{"ghcr.io/example(1)/*"}); err != nil {
		t.Fatalf("New() should treat regex metacharacters as literals, got: %s", err)
	}
}

func TestDisabledPolicyRejectsEverything(t *testing.T) {
	for _, p := range []*Policy{nil, {}, mustNew(t, nil), mustNew(t, []string{""})} {
		if p.Enabled() {
			t.Errorf("Enabled() = true for a policy with no patterns")
		}
		if err := p.Validate(""); err != nil {
			t.Errorf("Validate(\"\") = %s, want nil - an empty image is not an override", err)
		}
		if err := p.Validate(unqualified); !errors.Is(err, ErrDisabled) {
			t.Errorf("Validate() = %v, want ErrDisabled", err)
		}
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		image    string
		wantErr  bool
	}{
		{
			name:     "empty image is always allowed",
			patterns: []string{ecrPattern},
			image:    "",
		},
		{
			name:     "matching registry and tag",
			patterns: []string{ecrPattern},
			image:    allowedECRImage,
		},
		{
			name:     "matching registry and digest",
			patterns: []string{ecrPattern},
			image: "1234567890.dkr.ecr.us-west-2.amazonaws.com/team/web@sha256:" +
				strings.Repeat("a", digestLength),
		},
		{
			name:     "matching registry with no tag",
			patterns: []string{ecrPattern},
			image:    "1234567890.dkr.ecr.us-west-2.amazonaws.com/team/web",
		},
		{
			name:     "any of several patterns may match",
			patterns: []string{ghcrPattern, ecrPattern},
			image:    "ghcr.io/example/debug:v1",
		},
		{
			name:     "wrong repository is rejected",
			patterns: []string{ecrPattern},
			image:    "1234567890.dkr.ecr.us-west-2.amazonaws.com/someone-else/web:abc123",
			wantErr:  true,
		},
		{
			name:     "wrong registry is rejected",
			patterns: []string{ecrPattern},
			image:    "docker.io/team/web:abc123",
			wantErr:  true,
		},
		{
			name:     "unqualified image is rejected",
			patterns: []string{ecrPattern},
			image:    unqualified,
			wantErr:  true,
		},
		{
			// The important one. A leading "*" must not swallow a path
			// separator, or an attacker-controlled registry host could be
			// prefixed onto an otherwise-trusted-looking reference.
			name:     "registry prefix smuggling is rejected",
			patterns: []string{ecrPattern},
			image:    "evil.example.com/" + allowedECRImage,
			wantErr:  true,
		},
		{
			// The same attack, but hiding the extra segment in the middle.
			name:     "extra path segment is rejected by a single-star pattern",
			patterns: []string{ecrPattern},
			image:    "1234567890.dkr.ecr.us-west-2.amazonaws.com/team/../evil/web:abc123",
			wantErr:  true,
		},
		{
			name:     "double star crosses path separators",
			patterns: []string{"*.dkr.ecr.*.amazonaws.com/team/**"},
			image:    "1234567890.dkr.ecr.us-west-2.amazonaws.com/team/sub/web:abc123",
		},
		{
			name:     "double star still anchors the registry host",
			patterns: []string{"*.dkr.ecr.*.amazonaws.com/team/**"},
			image:    "evil.example.com/x.dkr.ecr.us-west-2.amazonaws.com/team/web:abc123",
			wantErr:  true,
		},
		{
			name:     "single star does not cross path separators",
			patterns: []string{ecrPattern},
			image:    "1234567890.dkr.ecr.us-west-2.amazonaws.com/team/sub/web:abc123",
			wantErr:  true,
		},
		{
			name:     "dots in the pattern are literal",
			patterns: []string{ghcrPattern},
			image:    "ghcrxio/example/debug:v1",
			wantErr:  true,
		},
		{
			name:     "newline in the image reference is rejected",
			patterns: []string{matchAll},
			image:    "ghcr.io/example/debug:v1\nmalicious: true",
			wantErr:  true,
		},
		{
			name:     "leading dash in the image reference is rejected",
			patterns: []string{matchAll},
			image:    "-ghcr.io/example/debug:v1",
			wantErr:  true,
		},
		{
			name:     "over-long image reference is rejected",
			patterns: []string{matchAll},
			image:    "ghcr.io/example/" + strings.Repeat("a", maxImageLength),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mustNew(t, tt.patterns).Validate(tt.image)
			if tt.wantErr && err == nil {
				t.Errorf("Validate(%q) = nil, want an error", tt.image)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Validate(%q) = %s, want nil", tt.image, err)
			}
		})
	}
}

func TestActiveDefaultsToDenyAll(t *testing.T) {
	// Deliberately not calling SetActive() first - this is the state of the
	// process before main() configures anything.
	if err := Active().Validate(unqualified); !errors.Is(err, ErrDisabled) {
		t.Errorf("Active().Validate() = %v, want ErrDisabled", err)
	}
}

func TestSetActive(t *testing.T) {
	t.Cleanup(func() { SetActive(nil) })

	SetActive(mustNew(t, []string{ecrPattern}))
	if err := Active().Validate(allowedECRImage); err != nil {
		t.Errorf("Active().Validate() = %s, want nil", err)
	}

	// A nil Policy must fall back to deny-all rather than panic.
	SetActive(nil)
	if err := Active().Validate(unqualified); !errors.Is(err, ErrDisabled) {
		t.Errorf("Active().Validate() after SetActive(nil) = %v, want ErrDisabled", err)
	}
}

func mustNew(t *testing.T, patterns []string) *Policy {
	t.Helper()
	p, err := New(patterns)
	if err != nil {
		t.Fatalf("New(%v) returned an unexpected error: %s", patterns, err)
	}
	return p
}
