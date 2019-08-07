package sensu

import (
	"fmt"

	"github.com/sensu/sensu-go/types"
)

const nilStr = "nil"

// EventKey returns the event key using the event's entity name and check name
func EventKey(event *types.Event) string {
	entityName := nilStr
	if event != nil && event.Entity != nil && len(event.Entity.Name) > 0 {
		entityName = event.Entity.Name
	}

	checkName := nilStr
	if event != nil && event.Check != nil && len(event.Check.Name) > 0 {
		checkName = event.Check.Name
	}
	return entityName + "/" + checkName
}

// EventSummaryWithTrim generates the event summary, trimming the output at trimAt if necessary
func EventSummaryWithTrim(event *types.Event, trimAt int) string {
	// TODO: Ruby code uses check[:notification] and check[:description] which are not present in the GO code.
	// Skipping them from now.
	output := nilStr
	if event != nil && event.Check != nil && len(event.Check.Output) > 0 {
		output = event.Check.Output
	}
	if trimAt > 0 && len(output) > trimAt {
		output = string(([]rune(output))[0:trimAt])
	}
	return EventKey(event) + " : " + output
}

// EventSummary generates the event summary, without trimming any of the output
func EventSummary(event *types.Event) string {
	return EventSummaryWithTrim(event, 0)
}

// FormattedMessage creates a formatted message, intended for chat rooms etc.
func FormattedMessage(event *types.Event) string {
	action := "ALERT"
	if event != nil && event.Check != nil && event.Check.Status == 0 {
		action = "RESOLVE"
	}
	return fmt.Sprintf("%s - %s", action, EventSummary(event))
}
