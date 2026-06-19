package filter

import (
	"fmt"
	"regexp"
	"strings"
)

// UnlessExclude drops a title when pattern matches but unless does not.
type UnlessExclude struct {
	Pattern string `yaml:"pattern"`
	Unless  string `yaml:"unless"`
}

// Rules decides which RSS items are kept based on their title.
//
// When include is non-empty, the title must match at least one include pattern.
// When exclude is non-empty, the title must not match any exclude pattern.
// exclude_unless drops a title when pattern matches but unless does not.
// Patterns are case-insensitive substrings by default; set regex: true to treat
// them as regular expressions.
type Rules struct {
	Include       []string        `yaml:"include"`
	Exclude       []string        `yaml:"exclude"`
	ExcludeUnless []UnlessExclude `yaml:"exclude_unless"`
	Regex         bool            `yaml:"regex"`
}

func (f Rules) compile() (include, exclude []*regexp.Regexp, err error) {
	if !f.Regex {
		return nil, nil, nil
	}
	for _, p := range f.Include {
		re, err := regexp.Compile(p)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid include regex %q: %w", p, err)
		}
		include = append(include, re)
	}
	for _, p := range f.Exclude {
		re, err := regexp.Compile(p)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid exclude regex %q: %w", p, err)
		}
		exclude = append(exclude, re)
	}
	return include, exclude, nil
}

func (f Rules) compileUnless() (patterns, unlesses []*regexp.Regexp, err error) {
	if !f.Regex {
		return nil, nil, nil
	}
	for _, rule := range f.ExcludeUnless {
		re, err := regexp.Compile(rule.Pattern)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid exclude_unless pattern regex %q: %w", rule.Pattern, err)
		}
		patterns = append(patterns, re)
		u, err := regexp.Compile(rule.Unless)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid exclude_unless unless regex %q: %w", rule.Unless, err)
		}
		unlesses = append(unlesses, u)
	}
	return patterns, unlesses, nil
}

// Validate checks that regex patterns compile.
func (f Rules) Validate() error {
	if _, _, err := f.compile(); err != nil {
		return err
	}
	_, _, err := f.compileUnless()
	return err
}

func (f Rules) matchesPattern(title, pattern string, compiled *regexp.Regexp) bool {
	if f.Regex {
		return compiled.MatchString(title)
	}
	return strings.Contains(strings.ToLower(title), strings.ToLower(pattern))
}

// Matches reports whether title passes the filter rules.
func (f Rules) Matches(title string) (bool, error) {
	includeRE, excludeRE, err := f.compile()
	if err != nil {
		return false, err
	}
	unlessPatternRE, unlessUnlessRE, err := f.compileUnless()
	if err != nil {
		return false, err
	}

	if len(f.Include) > 0 {
		matched := false
		for i, p := range f.Include {
			var re *regexp.Regexp
			if f.Regex {
				re = includeRE[i]
			}
			if f.matchesPattern(title, p, re) {
				matched = true
				break
			}
		}
		if !matched {
			return false, nil
		}
	}

	for i, p := range f.Exclude {
		var re *regexp.Regexp
		if f.Regex {
			re = excludeRE[i]
		}
		if f.matchesPattern(title, p, re) {
			return false, nil
		}
	}

	for i, rule := range f.ExcludeUnless {
		var patternRE, unlessRE *regexp.Regexp
		if f.Regex {
			patternRE = unlessPatternRE[i]
			unlessRE = unlessUnlessRE[i]
		}
		if f.matchesPattern(title, rule.Pattern, patternRE) &&
			!f.matchesPattern(title, rule.Unless, unlessRE) {
			return false, nil
		}
	}

	return true, nil
}

// Normalize tidies configured patterns.
func (f *Rules) Normalize() {
	f.Include = trimPatterns(f.Include)
	f.Exclude = trimPatterns(f.Exclude)

	rules := f.ExcludeUnless[:0]
	for _, rule := range f.ExcludeUnless {
		rule.Pattern = strings.TrimSpace(rule.Pattern)
		rule.Unless = strings.TrimSpace(rule.Unless)
		if rule.Pattern != "" && rule.Unless != "" {
			rules = append(rules, rule)
		}
	}
	f.ExcludeUnless = rules
}

func trimPatterns(patterns []string) []string {
	out := patterns[:0]
	for _, p := range patterns {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}