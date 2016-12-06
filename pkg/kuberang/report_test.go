package kuberang

import "testing"

func TestSimpleReportIsSuccess(t *testing.T) {
	tests := []struct {
		name          string
		results       []CheckResult
		expectSuccess bool
	}{
		{
			name: "one successful",
			results: []CheckResult{
				{
					Success: true,
					Ignored: false,
				},
			},
			expectSuccess: true,
		},
		{
			name: "one failure",
			results: []CheckResult{
				{
					Success: false,
					Ignored: false,
				},
			},
			expectSuccess: false,
		},
		{
			name: "one ignored failure",
			results: []CheckResult{
				{
					Success: false,
					Ignored: true,
				},
			},
			expectSuccess: true,
		},
		{
			name: "one success, one failure",
			results: []CheckResult{
				{
					Success: true,
				},
				{
					Success: false,
				},
			},
			expectSuccess: false,
		},
		{
			name: "one failure, one ignored failure",
			results: []CheckResult{
				{
					Success: false,
				},
				{
					Success: false,
					Ignored: true,
				},
			},
			expectSuccess: false,
		},
		{
			name: "one failure, one ignored failure, one success",
			results: []CheckResult{
				{
					Success: false,
				},
				{
					Success: false,
					Ignored: true,
				},
				{
					Success: true,
				},
			},
			expectSuccess: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			report := simpleReport{CheckResults: test.results}
			success := report.isSuccess()
			if success != test.expectSuccess {
				t.Errorf("Expected %v, but got %v", test.expectSuccess, success)
			}
		})
	}
}
