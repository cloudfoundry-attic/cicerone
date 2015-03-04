package dsl

import (
	"regexp"
	"time"

	"github.com/pivotal-golang/lager"
)

//Matcher objects can test an Entry, returning true/false
type Matcher interface {
	Match(entry Entry) bool
}

//MatcherFunc makes it easy to create Matchers from bare functions
type MatcherFunc func(Entry) bool

//Match satisifes the Matcher interface
func (m MatcherFunc) Match(entry Entry) bool {
	return m(entry)
}

//True is a trivial Matcher that always returns True
func True() Matcher {
	return MatcherFunc(func(entry Entry) bool {
		return true
	})
}

//False is a trivial Matcher that always returns False
func False() Matcher {
	return MatcherFunc(func(entry Entry) bool {
		return false
	})
}

//And combines Matchers - the output is the logical && of the output of each component Matcher
func And(matchers ...Matcher) Matcher {
	return MatcherFunc(func(entry Entry) bool {
		for _, matcher := range matchers {
			if !matcher.Match(entry) {
				return false
			}
		}
		return true
	})
}

//Or combines Matchers - the output is the logical || of the output of each component Matcher
func Or(matchers ...Matcher) Matcher {
	return MatcherFunc(func(entry Entry) bool {
		for _, matcher := range matchers {
			if matcher.Match(entry) {
				return true
			}
		}
		return false
	})
}

//Not negates the passed-in Matcher
func Not(matcher Matcher) Matcher {
	return MatcherFunc(func(entry Entry) bool {
		return !matcher.Match(entry)
	})
}

//RegExpMatcher takes a Getter (presumed to return a string) and a regular expression (encoded as a string)
//RegExpMatcher returns true of the string returned by the Getter matches the passed-in regular expression.
func RegExpMatcher(getter Getter, regExp string) Matcher {
	re := regexp.MustCompile(regExp)
	return MatcherFunc(func(entry Entry) bool {
		value, ok := getter.Get(entry)
		if !ok {
			return false
		}
		stringValue, ok := value.(string)
		if !ok {
			return false
		}
		return re.MatchString(stringValue)
	})
}

//MatchVM matches true if the Entry's VM matches the passed-in string (interpreted as a regular expression)
func MatchVM(vm string) Matcher {
	return RegExpMatcher(GetVM, vm)
}

//MatchJob matches true if the Entry's Job matches the passed-in string (interpreted as a regular expression)
func MatchJob(job string) Matcher {
	return RegExpMatcher(GetJob, job)
}

//MatchSource matches true if the Entry's Source matches the passed-in string (interpreted as a regular expression)
func MatchSource(source string) Matcher {
	return RegExpMatcher(GetSource, source)
}

//MatchMessage matches true if the Entry's Message matches the passed-in string (interpreted as a regular expression)
func MatchMessage(message string) Matcher {
	return RegExpMatcher(GetMessage, message)
}

//MatchSession matches true if the Entry's Session matches the passed-in string (interpreted as a regular expression)
func MatchSession(session string) Matcher {
	return RegExpMatcher(GetSession, session)
}

//MatchIndex matches true if the Entry's Index matches the passed-in integer
func MatchIndex(index int) Matcher {
	return MatcherFunc(func(entry Entry) bool {
		return entry.Index == index
	})
}

//MatchLogLEvel matches true if the Entry's LogLevel matches the passed-in lager.LogLevel
func MatchLogLevel(logLevel lager.LogLevel) Matcher {
	return MatcherFunc(func(entry Entry) bool {
		return entry.LogLevel == logLevel
	})
}

//MatchAfter returns true if the Entry's timestamp is after the passed-in time
func MatchAfter(t time.Time) Matcher {
	return MatcherFunc(func(entry Entry) bool {
		return entry.Timestamp.After(t)
	})
}

//MatchBefore returns true if hte Entry's timestamp is before the passed-in time
func MatchBefore(t time.Time) Matcher {
	return MatcherFunc(func(entry Entry) bool {
		return entry.Timestamp.Before(t)
	})
}
