package functions

import (
	"fmt"
	"strings"

	. "github.com/onsi/sommelier/dsl"
)

func Playground(e Entries) error {
	return threads(e)
}

func threads(e Entries) error {
	groups := e.GroupBy(DataGetter("task-guid", "container-guid", "guid"))
	startingMatcher := Or(MatchMessage(`task-processor\.succeeded-starting-task`), MatchMessage(`task-processor\.starting-task`))
	groups = groups.Filter(Or(
		MatchMessage(`create\.created`),
		MatchMessage(`task-processor\.succeeded-starting-task`),
		MatchMessage(`task-processor\.starting-task`),
		MatchMessage(`transitioning-to-complete`),
		MatchMessage(`succeeded-transitioning-to-complete`),
		MatchMessage(`resolving-task`),
		MatchMessage(`resolved-task`),
	)).Dedupe(startingMatcher)

	firstTime := e[0].Timestamp

	groups.EachGroup(func(key interface{}, entries Entries) error {
		s := []string{fmt.Sprintf("%s", key)}
		for _, entry := range entries {
			components := strings.Split(entry.Message, ".")
			s = append(s, fmt.Sprintf("%s %s", components[len(components)-1], entry.Timestamp.Sub(firstTime)))
		}
		fmt.Println(strings.Join(s, " "))
		return nil
	})

	startStats := groups.FindPairs(MatchMessage(`create\.created`), startingMatcher).DTStats()
	completeStats := groups.FindPairs(startingMatcher, MatchMessage(`transitioning-to-complete`)).DTStats()
	transitioningToCompleteStats := groups.FindPairs(MatchMessage(`transitioning-to-complete`), MatchMessage(`succeeded-transitioning-to-complete`)).DTStats()
	resolvingStats := groups.FindPairs(MatchMessage(`succeeded-transitioning-to-complete`), MatchMessage(`resolving-task`)).DTStats()
	resolvedStats := groups.FindPairs(MatchMessage(`resolving-task`), MatchMessage(`resolved-task`)).DTStats()
	roundTripStats := groups.FindPairs(MatchMessage(`create\.created`), MatchMessage(`resolved-task`)).DTStats()

	createStats := e.Filter(MatchMessage(`create\.created`)).PairAllWith(e[0], DataGetter("task-guid")).DTStats()

	fmt.Println("0->Create             - ", createStats)
	fmt.Println("Create->Starting      - ", startStats)
	fmt.Println("Starting->Complete    - ", completeStats)
	fmt.Println("Complete->Completed   - ", transitioningToCompleteStats)
	fmt.Println("Completed->Resolving  - ", resolvingStats)
	fmt.Println("Resolving->Resolved   - ", resolvedStats)
	fmt.Println("Create->Resolved      - ", roundTripStats)

	vmGroups := e.Filter(Or(
		MatchMessage(`task-processor\.succeeded-starting-task`),
		MatchMessage(`task-processor\.starting-task`),
		MatchMessage(`transitioning-to-complete`),
		MatchMessage(`succeeded-transitioning-to-complete`),
	)).GroupBy(GetVM)

	vmGroups.EachGroup(func(key interface{}, entries Entries) error {
		groups := entries.GroupBy(DataGetter("task-guid", "container-guid", "guid"))
		groups.Dedupe(startingMatcher)
		completeStats := groups.FindPairs(startingMatcher, MatchMessage(`transitioning-to-complete`)).DTStats()
		transitioningToCompleteStats := groups.FindPairs(MatchMessage(`transitioning-to-complete`), MatchMessage(`succeeded-transitioning-to-complete`)).DTStats()
		fmt.Println(key)
		fmt.Println("  Starting->Complete    - ", completeStats)
		fmt.Println("  Complete->Completed   - ", transitioningToCompleteStats)
		return nil
	})

	return nil
}
