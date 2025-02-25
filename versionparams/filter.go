package versionparams

import (
	"regexp"
)

// filter holds sets of regular expressions (RE). If a given value matches
// any of the REs in the matches set and does not match any in the does not
// match set then it passes the filter.
type filter struct {
	matches      []*regexp.Regexp
	doesNotMatch []*regexp.Regexp
}

// makeFilterFromMap returns a filter and a list of any errors found in the
// filterMap. Entries in the filter map must compile to valid regular
// expressions and if the map entry is true, the RE will be added to the list
// of matches filters, otherwise to the doesNotMatch filters
func makeFilterFromMap(filterMap map[string]bool) (filter, []error) {
	f := filter{}
	errs := []error{}

	for filter, mustMatch := range filterMap {
		re, err := regexp.Compile(filter)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		if mustMatch {
			f.matches = append(f.matches, re)
		} else {
			f.doesNotMatch = append(f.doesNotMatch, re)
		}
	}

	return f, errs
}

// makeFilter returns a filter populated with the supplied filter lists. For
// a value to pass the filter it must match any of the matches REs and none
// of the does REs
/*
func makeFilter(matches, doesNotMatch []*regexp.Regexp) filter {
	return filter{
		matches:      matches,
		doesNotMatch: doesNotMatch,
	}
}
*/

// AddMatchFilter adds a filter to the 'matches' set.
func (f *filter) AddMatchFilter(re *regexp.Regexp) {
	f.matches = append(f.matches, re)
}

// AddNoMatchFilter adds a filter to the 'matches' set.
func (f *filter) AddNoMatchFilter(re *regexp.Regexp) {
	f.doesNotMatch = append(f.doesNotMatch, re)
}

// HasFilters returns true if any filters have been given, false otherwise.
func (f filter) HasFilters() bool {
	return len(f.matches) > 0 || len(f.doesNotMatch) > 0
}

// Passes checks that the value satisfies all the filters. A value will pass
// if it matches any of the matches filters and does not match any of the
// doesNotMatch filters. If there are no filters then any value will match.
func (f filter) Passes(val string) bool {
	match := false

	for _, re := range f.matches {
		if match = re.MatchString(val); match {
			break
		}
	}

	if len(f.matches) == 0 {
		match = true
	}

	for _, re := range f.doesNotMatch {
		if !match {
			break
		}

		match = !re.MatchString(val)
	}

	return match
}
