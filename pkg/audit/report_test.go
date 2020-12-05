package audit

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReport(t *testing.T) {
	expectedReport := `{"ready":{"pre_ready":{"result":"Success","message":"Successfully completed pre ready hook"},"post_ready":{"result":"","message":""},"result":"Success","message":"Successfully completed ready stage"},"load":{"pre_load":{"result":"","message":""},"post_load":{"result":"Success","message":"Successfully completed post load hook"},"result":"Success","message":"Successfully completed load stage"},"scenarios":{"pre_chaos_tests":{"result":"Success","message":"Successfully completed pre chaos test hook"},"post_chaos_tests":{"result":"Success","message":"Successfully completed post chaos hook"},"scenarios":[{"identifiers":"id1","result":"Success","message":"Successfully completed scenario 1"},{"identifiers":"id2","result":"Failure","message":"Failed to complete scenario 2"}]},"miscellaneous":[{"result":"Success","message":"All tests passes successfully"}]}`
	reporter := &Reporter{
		Ready: Ready{
			PreReady: Message{
				Result:  SuccessResult,
				Message: "Successfully completed pre ready hook",
			},
			Message: Message{
				Result:  SuccessResult,
				Message: "Successfully completed ready stage",
			},
		},
		Load: Load{
			PostLoad: Message{
				Result:  SuccessResult,
				Message: "Successfully completed post load hook",
			},
			Message: Message{
				Result:  SuccessResult,
				Message: "Successfully completed load stage",
			},
		},
		Scenarios: Scenarios{
			PreChaosTests: Message{
				Result:  SuccessResult,
				Message: "Successfully completed pre chaos test hook",
			},
			PostChaosTests: Message{
				Result:  SuccessResult,
				Message: "Successfully completed post chaos hook",
			},
			Scenarios: []Scenario{
				{
					Identifiers: "id1",
					Message: Message{
						Result:  SuccessResult,
						Message: "Successfully completed scenario 1",
					},
				},
				{
					Identifiers: "id2",
					Message: Message{
						Result:  FailureResult,
						Message: "Failed to complete scenario 2",
					},
				},
			},
		},
		Miscellaneous: []Message{
			{
				Result:  SuccessResult,
				Message: "All tests passes successfully",
			},
		},
	}
	writer := &bytes.Buffer{}

	err := reporter.Report(writer)
	require.NoError(t, err)
	require.Equal(t, expectedReport, writer.String())
}
