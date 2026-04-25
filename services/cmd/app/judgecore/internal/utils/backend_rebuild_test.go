package utils

import "testing"

func TestNormalizeWritebackResult(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "Accepted", want: "accepted"},
		{input: "Wrong Answer", want: "wrong_answer"},
		{input: "Compile Error", want: "compile_error"},
		{input: "Time Limit Exceeded", want: "time_limit_exceeded"},
		{input: "Memory Limit Exceeded", want: "memory_limit_exceeded"},
		{input: "runtime_error", want: "runtime_error"},
		{input: "system_error", want: "system_error"},
	}

	for _, tc := range tests {
		if got := normalizeWritebackResult(tc.input); got != tc.want {
			t.Fatalf("normalizeWritebackResult(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
