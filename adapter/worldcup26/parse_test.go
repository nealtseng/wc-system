package worldcup26

import (
	"testing"
	"time"
)

func TestResolveMatchStatus_InferredFinished(t *testing.T) {
	kickoff := time.Date(2026, 6, 11, 18, 0, 0, 0, time.UTC)
	now := time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)

	status, finished := ResolveMatchStatus(false, "notstarted", kickoff, now)
	if status != "finished" || !finished {
		t.Fatalf("expected inferred finished, got status=%q finished=%v", status, finished)
	}
}

func TestResolveMatchStatus_InferredLive(t *testing.T) {
	kickoff := time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)
	now := time.Date(2026, 6, 19, 13, 0, 0, 0, time.UTC)

	status, finished := ResolveMatchStatus(false, "notstarted", kickoff, now)
	if status != "scheduled" || finished {
		t.Fatalf("notstarted should stay scheduled, got status=%q finished=%v", status, finished)
	}
}

func TestResolveMatchStatus_CDNLiveElapsed(t *testing.T) {
	kickoff := time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)
	now := time.Date(2026, 6, 19, 13, 0, 0, 0, time.UTC)

	status, finished := ResolveMatchStatus(false, "67'", kickoff, now)
	if status != "live" || finished {
		t.Fatalf("expected CDN live, got status=%q finished=%v", status, finished)
	}
}

func TestResolveMatchStatus_Upcoming(t *testing.T) {
	kickoff := time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC)
	now := time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)

	status, finished := ResolveMatchStatus(false, "notstarted", kickoff, now)
	if status != "scheduled" || finished {
		t.Fatalf("expected scheduled, got status=%q finished=%v", status, finished)
	}
}

func TestResolveMatchStatus_CDNLive(t *testing.T) {
	kickoff := time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)
	now := time.Date(2026, 6, 19, 12, 30, 0, 0, time.UTC)

	status, finished := ResolveMatchStatus(false, "67'", kickoff, now)
	if status != "live" || finished {
		t.Fatalf("expected CDN live, got status=%q finished=%v", status, finished)
	}
}
