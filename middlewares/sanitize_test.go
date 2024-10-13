package middlewares

import (
	"testing"
)

func TestSanitizeInput(t *testing.T) {
	type args struct {
		input        string
		maxWordCount int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Empty input",
			args: args{
				input:        "",
				maxWordCount: 10,
			},
			want: "",
		},
		{
			name: "Input within word limit",
			args: args{
				input:        "This is a valid input.",
				maxWordCount: 5,
			},
			want: "This is a valid input.",
		},
		{
			name: "Input exceeding word limit",
			args: args{
				input:        "This input has more words than allowed by the limit.",
				maxWordCount: 5,
			},
			want: "This input has more words",
		},
		{
			name: "Input with extra spaces",
			args: args{
				input:        "   Too   many   spaces   ",
				maxWordCount: 4,
			},
			want: "Too many spaces",
		},
		{
			name: "Exact word limit",
			args: args{
				input:        "This is five words exactly",
				maxWordCount: 5,
			},
			want: "This is five words exactly",
		},
		{
			name: "Zero word limit",
			args: args{
				input:        "Some input text",
				maxWordCount: 0,
			},
			want: "",
		},
		{
			name: "Negative word limit (invalid case)",
			args: args{
				input:        "Some input text",
				maxWordCount: -1,
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeInput(tt.args.input, tt.args.maxWordCount); got != tt.want {
				t.Errorf("SanitizeInput() = %v, want %v", got, tt.want)
			}
		})
	}
}
