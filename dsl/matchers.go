package dsl

import (
	"regexp"

	"github.com/pivotal-golang/lager"
)

type Matcher interface {
	Match(entry Entry) bool
}

type MatcherFunc func(Entry) bool

func (m MatcherFunc) Match(entry Entry) bool {
	return m(entry)
}

func True() Matcher {
	return MatcherFunc(func(entry Entry) bool {
		return true
	})
}

func False() Matcher {
	return MatcherFunc(func(entry Entry) bool {
		return false
	})
}

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

func Not(matcher Matcher) Matcher {
	return MatcherFunc(func(entry Entry) bool {
		return !matcher.Match(entry)
	})
}

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

func MatchVM(vm string) Matcher {
	return RegExpMatcher(GetVM, vm)
}

func MatchJob(job string) Matcher {
	return RegExpMatcher(GetJob, job)
}

func MatchSource(source string) Matcher {
	return RegExpMatcher(GetSource, source)
}

func MatchMessage(message string) Matcher {
	return RegExpMatcher(GetMessage, message)
}

func MatchIndex(index int) Matcher {
	return MatcherFunc(func(entry Entry) bool {
		return entry.Index == index
	})
}

func MatchLogLevel(logLevel lager.LogLevel) Matcher {
	return MatcherFunc(func(entry Entry) bool {
		return entry.LogLevel == logLevel
	})
}
