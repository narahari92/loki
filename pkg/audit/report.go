package audit

import (
	"encoding/json"
	"io"

	"github.com/pkg/errors"
)

// Reporter represents the complete report of chaos test execution.
type Reporter struct {
	// Ready represents ready phase.
	Ready `json:"ready"`
	// Load represents system load phase.
	Load `json:"load"`
	// Scenarios represents actual chaos testing phase.
	Scenarios `json:"scenarios"`
	// Miscellaneous contains any other information.
	Miscellaneous []Message `json:"miscellaneous"`
}

// Message is the basic message structure in chaos test report.
type Message struct {
	// Result indicates success or failure.
	Result string `json:"result"`
	// Message gives more context to the user about what happened.
	Message string `json:"message"`
}

// Ready contains the report information of ready phase.
type Ready struct {
	// PreReady contains report information of pre ready hook.
	PreReady Message `json:"pre_ready"`
	// PostReady contains report information of post ready hook.
	PostReady Message `json:"post_ready"`
	// Message contains report information of actual ready phase.
	Message
}

// Load contains the report information of load phase.
type Load struct {
	// PreLoad contains report information of pre system load hook.
	PreLoad Message `json:"pre_load"`
	// PostLoad contains report information of post system load hook.
	PostLoad Message `json:"post_load"`
	// Message contains report information of actual system load phase.
	Message
}

// Scenario represents a single chaos test scenario.
type Scenario struct {
	// Identifiers indicate the identifiers involved in the chaos test scenario.
	Identifiers string `json:"identifiers"`
	// Message contains report information of chaos test scenario.
	Message
}

// Scenarios contains the report information of actual chaos testing phase.
type Scenarios struct {
	// PreChaosTests contains report information of pre chaos test hook.
	PreChaosTests Message `json:"pre_chaos_tests"`
	// PostChaosTests contains report information of post chaos test hook.
	PostChaosTests Message `json:"post_chaos_tests"`
	// Scenarios contain report information of all chaos test scenarios.
	Scenarios []Scenario `json:"scenarios"`
}

// Report writes the json representation of Reporter into the writer passed.
func (r *Reporter) Report(writer io.Writer) error {
	report, err := json.Marshal(r)
	if err != nil {
		return errors.Wrap(err, "failed to marshall report")
	}

	if _, err := writer.Write(report); err != nil {
		return errors.Wrap(err, "failed to write report")
	}

	return nil
}
