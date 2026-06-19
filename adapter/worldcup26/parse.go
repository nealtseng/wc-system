package worldcup26

import (
	"strconv"
	"strings"
	"time"
)

// Typical group-stage match window including stoppage and half-time buffer.
const matchDuration = 2*time.Hour + 20*time.Minute

// ParseScore converts a CDN score string to an int pointer; nil when not a valid score.
func ParseScore(s string) *int {
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "null") {
		return nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &n
}

// ParseFinished interprets worldcup26 "TRUE"/"FALSE" finished flags.
func ParseFinished(s string) bool {
	return strings.EqualFold(strings.TrimSpace(s), "TRUE")
}

// MatchStatus derives a UI-friendly status from finished and time_elapsed fields.
func MatchStatus(finished bool, timeElapsed string) string {
	status, _ := ResolveMatchStatus(finished, timeElapsed, time.Time{}, time.Time{})
	return status
}

// ResolveMatchStatus combines CDN flags with kickoff-time inference.
// The worldcup2026 CDN often keeps finished=FALSE before/during early tournament;
// once kickoff + matchDuration has passed we treat the fixture as finished.
func ResolveMatchStatus(finished bool, timeElapsed string, kickoff, now time.Time) (status string, inferredFinished bool) {
	if finished {
		return "finished", true
	}

	te := strings.ToLower(strings.TrimSpace(timeElapsed))
	switch {
	case te == "" || te == "notstarted":
		// fall through to kickoff inference
	case strings.Contains(te, "half") || strings.Contains(te, "'") || te == "live":
		return "live", false
	case strings.Contains(te, "ft") || strings.Contains(te, "full"):
		return "finished", true
	default:
		return te, false
	}

	if kickoff.IsZero() || now.IsZero() {
		return "scheduled", false
	}

	endEstimate := kickoff.Add(matchDuration)
	if now.After(endEstimate) {
		return "finished", true
	}
	// CDN still says notstarted — do not infer live from kickoff alone (avoids false
	// live when kickoff timestamps are wrong or the feed has not caught up yet).
	if te == "" || te == "notstarted" {
		if now.Before(kickoff) {
			return "scheduled", false
		}
		return "scheduled", false
	}
	if !now.Before(kickoff) {
		return "live", false
	}
	return "scheduled", false
}
