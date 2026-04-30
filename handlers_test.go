package main

import "testing"

func TestValidateEnqueue(t *testing.T) {
	cases := []struct {
		name string
		in   enqueueRequest
		want string
	}{
		{"valid full", enqueueRequest{"hi", 1, "send_email", "high"}, ""},
		{"valid no priority", enqueueRequest{"hi", 1, "send_email", ""}, ""},
		{"valid generate_invoice", enqueueRequest{"hi", 0, "generate_invoice", "low"}, ""},
		{"empty payload", enqueueRequest{"", 1, "send_email", "high"}, "payload is required"},
		{"unknown type", enqueueRequest{"hi", 1, "bogus", "high"}, "type must be send_email or generate_invoice"},
		{"empty type", enqueueRequest{"hi", 1, "", "high"}, "type must be send_email or generate_invoice"},
		{"negative duration", enqueueRequest{"hi", -1, "send_email", "high"}, "duration must be between 0 and 600 seconds"},
		{"duration over max", enqueueRequest{"hi", 601, "send_email", "high"}, "duration must be between 0 and 600 seconds"},
		{"unknown priority", enqueueRequest{"hi", 1, "send_email", "urgent"}, "priority must be high, medium, or low"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := validateEnqueue(&tc.in)
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
