package sensu

import (
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestValidEvent_EventKey(t *testing.T) {
	event := &types.Event{
		Entity: &types.Entity{
			ObjectMeta: types.ObjectMeta{
				Name: "EntityName",
			},
		},
		Check: &types.Check{
			ObjectMeta: types.ObjectMeta{
				Name: "CheckName",
			},
		},
	}

	eventKey := EventKey(event)
	assert.Equal(t, "EntityName/CheckName", eventKey, "Invalid key")
}

func TestEmptyEntityName_EventKey(t *testing.T) {
	event := &types.Event{
		Entity: &types.Entity{},
		Check: &types.Check{
			ObjectMeta: types.ObjectMeta{
				Name: "CheckName",
			},
		},
	}

	eventKey := EventKey(event)
	assert.Equal(t, "nil/CheckName", eventKey, "Invalid key")
}

func TestEmptyCheckName_EventKey(t *testing.T) {
	event := &types.Event{
		Entity: &types.Entity{
			ObjectMeta: types.ObjectMeta{
				Name: "EntityName",
			},
		},
		Check: &types.Check{},
	}

	eventKey := EventKey(event)
	assert.Equal(t, "EntityName/nil", eventKey, "Invalid key")
}

func TestNilEntity_EventKey(t *testing.T) {
	event := &types.Event{
		Check: &types.Check{
			ObjectMeta: types.ObjectMeta{
				Name: "CheckName",
			},
		},
	}

	eventKey := EventKey(event)
	assert.Equal(t, "nil/CheckName", eventKey, "Invalid key")
}

func TestNilCheck_EventKey(t *testing.T) {
	event := &types.Event{
		Entity: &types.Entity{
			ObjectMeta: types.ObjectMeta{
				Name: "EntityName",
			},
		},
	}

	eventKey := EventKey(event)
	assert.Equal(t, "EntityName/nil", eventKey, "Invalid key")
}

func TestNullEvent_EventKey(t *testing.T) {
	eventKey := EventKey(nil)
	assert.Equal(t, "nil/nil", eventKey, "Invalid key")
}

func TestValidEventZeroTrim_EventSummaryWithTrim(t *testing.T) {
	event := &types.Event{
		Entity: &types.Entity{
			ObjectMeta: types.ObjectMeta{
				Name: "EntityName",
			},
		},
		Check: &types.Check{
			ObjectMeta: types.ObjectMeta{
				Name: "CheckName",
			},
			Output: "CheckOutput",
		},
	}

	eventSummary := EventSummaryWithTrim(event, 0)
	assert.Equal(t, "EntityName/CheckName : CheckOutput", eventSummary)
}

func TestValidEventNoTrim_EventSummaryWithTrim(t *testing.T) {
	event := &types.Event{
		Entity: &types.Entity{
			ObjectMeta: types.ObjectMeta{
				Name: "EntityName",
			},
		},
		Check: &types.Check{
			ObjectMeta: types.ObjectMeta{
				Name: "CheckName",
			},
			Output: "CheckOutput",
		},
	}

	eventSummary := EventSummaryWithTrim(event, 1000)
	assert.Equal(t, "EntityName/CheckName : CheckOutput", eventSummary)
}

func TestValidEventWithTrim_EventSummaryWithTrim(t *testing.T) {
	event := &types.Event{
		Entity: &types.Entity{
			ObjectMeta: types.ObjectMeta{
				Name: "EntityName",
			},
		},
		Check: &types.Check{
			ObjectMeta: types.ObjectMeta{
				Name: "CheckName",
			},
			Output: "CheckOutput withaverylongstringthatwillbetruncatedtoonehundresscharactertomakesurethedestinationsystendoesntoverflow",
		},
	}

	eventKey := EventKey(event)

	eventSummary := EventSummaryWithTrim(event, 100)
	expectedLen := len(eventKey) + len(" : ") + 100
	assert.Len(t, eventSummary, expectedLen, "Event summary not trimmed at %d characters", expectedLen)
	assert.Equal(t, "EntityName/CheckName : CheckOutput withaverylongstringthatwillbetruncatedtoonehundresscharactertomakesurethedestinationsyst",
		eventSummary)
}

func TestEmptyOutputNoTrim_EventSummaryWithTrim(t *testing.T) {
	event := &types.Event{
		Entity: &types.Entity{
			ObjectMeta: types.ObjectMeta{
				Name: "EntityName",
			},
		},
		Check: &types.Check{
			ObjectMeta: types.ObjectMeta{
				Name: "CheckName",
			},
		},
	}

	eventSummary := EventSummaryWithTrim(event, 100)
	assert.Equal(t, "EntityName/CheckName : nil", eventSummary)
}

func TestNilCheck_EventSummaryWithTrim(t *testing.T) {
	event := &types.Event{
		Entity: &types.Entity{
			ObjectMeta: types.ObjectMeta{
				Name: "EntityName",
			},
		},
	}

	eventSummary := EventSummaryWithTrim(event, 100)
	assert.Equal(t, "EntityName/nil : nil",
		eventSummary)
}

func TestNilEvent_EventSummaryWithTrim(t *testing.T) {
	eventSummary := EventSummaryWithTrim(nil, 100)
	assert.Equal(t, "nil/nil : nil",
		eventSummary)
}

func TestEventSummary(t *testing.T) {
	event := &types.Event{
		Entity: &types.Entity{
			ObjectMeta: types.ObjectMeta{
				Name: "EntityName",
			},
		},
		Check: &types.Check{
			ObjectMeta: types.ObjectMeta{
				Name: "CheckName",
			},
			Output: "CheckOutput",
		},
	}

	eventSummary := EventSummary(event)
	assert.Equal(t, "EntityName/CheckName : CheckOutput", eventSummary)
}

func Test_FormattedMessage(t *testing.T) {
	event := &types.Event{
		Entity: &types.Entity{
			ObjectMeta: types.ObjectMeta{
				Name: "EntityName",
			},
		},
		Check: &types.Check{
			ObjectMeta: types.ObjectMeta{
				Name: "CheckName",
			},
			Output: "CheckOutput",
			Status: 0,
		},
	}

	expectedAlert := "ALERT - EntityName/CheckName : CheckOutput"
	expectedResolve := "RESOLVE - EntityName/CheckName : CheckOutput"

	for i := uint32(0); i < 10; i++ {
		event.Check.Status = i
		formattedMessage := FormattedMessage(event)
		if i == 0 {
			assert.Equal(t, expectedResolve, formattedMessage)
		} else {
			assert.Equal(t, expectedAlert, formattedMessage)
		}
	}
}

func TestNilCheck_FormattedMessage(t *testing.T) {
	event := &types.Event{
		Entity: &types.Entity{
			ObjectMeta: types.ObjectMeta{
				Name: "EntityName",
			},
		},
	}

	formattedMessage := FormattedMessage(event)
	assert.Equal(t, "ALERT - EntityName/nil : nil", formattedMessage)
}

func TestNilEntity_FormattedMessage(t *testing.T) {
	event := &types.Event{
		Check: &types.Check{
			ObjectMeta: types.ObjectMeta{
				Name: "CheckName",
			},
			Output: "CheckOutput",
			Status: 0,
		},
	}

	expectedAlert := "ALERT - nil/CheckName : CheckOutput"
	expectedResolve := "RESOLVE - nil/CheckName : CheckOutput"

	for i := uint32(0); i < 10; i++ {
		event.Check.Status = i
		formattedMessage := FormattedMessage(event)
		if i == 0 {
			assert.Equal(t, expectedResolve, formattedMessage)
		} else {
			assert.Equal(t, expectedAlert, formattedMessage)
		}
	}
}

func TestNilEvent_FormattedMessage(t *testing.T) {
	formattedMessage := FormattedMessage(nil)
	assert.Equal(t, "ALERT - nil/nil : nil", formattedMessage)
}
