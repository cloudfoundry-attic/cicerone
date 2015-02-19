package functions

import (
	"fmt"

	. "github.com/onsi/sommelier/dsl"
)

func FezzikTasks(e Entries) error {
	byTaskGuid := e.GroupBy(DataGetter("task-guid", "container-guid", "guid"))

	timelineDescription := TimelineDescription{
		{"Receptor-Creating", MatchMessage(`create\.creating-task`)},
		{"Receptor-Persisting-Done", MatchMessage(`create\.requesting-task-auction`)},
		{"Receptor-Auction-Submitted", MatchMessage(`create\.created`)},
		{"Rep-Starting-Task", MatchMessage(`task-processor\.starting-task`)},
		{"Rep-Succeeded-Starting-Task-In-BBS", MatchMessage(`task-processor\.succeeded-starting-task`)},
		{"Executor-Created-Container", MatchMessage(`run-container\.run\.started`)},
		{"Executor-Spawned-Process", MatchMessage(`run-step-process\.succeeded-transitioning-to-running`)},
		{"Rep-Start-Completing-Task", MatchMessage(`task-processor\.completing-task`)},
		{"Rep-Succeeded-Completing-Task-In-BBS", MatchMessage(`task-processor\.succeeded-completing-task`)},
		{"Receptor-Resolved-Task", MatchMessage(`resolved-task`)},
	}

	byTaskGuid = byTaskGuid.Filter(Or(
		MatchMessage(`task-handler\.create`),
		MatchMessage(`task-processor\.starting-task`),
		MatchMessage(`task-processor\.succeeded-starting-task`),
		MatchMessage(`run-container\.run\.started`),
		MatchMessage(`run-step-process\.succeeded-transitioning-to-running`),
		MatchMessage(`task-processor\.completing-task`),
		MatchMessage(`task-processor\.succeeded-completing-task`),
		MatchMessage(`resolving-task`),
		MatchMessage(`resolved-task`),
	))

	timelines := byTaskGuid.ConstructTimelines(timelineDescription, e[0])
	fmt.Println(timelines)

	fmt.Println(timelines.DTStatsSlice())

	//[X] timeline description => timeline (text)
	//timeline description => pair stats (text)
	//timeline description => timeline (visualization)
	//timeline description => pair stats (histograms)

	return nil
}
