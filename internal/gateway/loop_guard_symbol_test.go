package gateway

import (
	"strings"
	"testing"
)

// A reply can loop on symbols or whitespace alone. unitKey ignores a unit
// with no word, so the line window never fills, and runOnWords ignores it
// too. The real tail below is that shape: a short sentence, then the same
// asterisk line thousands of times.
func TestLoopGuardFindsASymbolLoop(t *testing.T) {
	prefix := "The lock budget is twenty seconds. I am checking the wait.\n"
	for name, text := range map[string]string{
		"real asterisk tail": prefix + symbolLoopTail,
		"asterisk lines":     prefix + strings.Repeat("* \n", 80),
		"asterisk run":       prefix + strings.Repeat("*", 800),
		"whitespace lines":   prefix + strings.Repeat(" \n", 600),
		"blank lines":        prefix + strings.Repeat("\n", 600),
	} {
		g := &loopGuard{}
		trip, ok := feedAll(t, g, "content", text)
		if !ok || trip.reasoning || !trip.runOn {
			t.Errorf("%s: trip %+v %v", name, trip, ok)
			continue
		}
		msg := trip.message()
		if !strings.Contains(msg, "without a break") || !strings.Contains(msg, "Stop looping replies") {
			t.Errorf("%s: message %q", name, msg)
		}
	}
	// the same channel as reasoning
	g := &loopGuard{}
	trip, ok := feedAll(t, g, "reasoning_content", strings.Repeat("*\n", 80))
	if !ok || !trip.reasoning {
		t.Fatalf("reasoning: %+v %v", trip, ok)
	}
}

// Ordinary markdown is not a symbol loop: a rule, a fence, a table, a
// short asterisk tail, and the same marks broken up by a sentence.
func TestLoopGuardLeavesMarkdownAlone(t *testing.T) {
	texts := []string{
		"intro\n\n---\n\n## Next\n\nA sentence about the change.\n",
		"```\nfmt.Println(1)\n```\n",
		strings.Repeat("```go\nfmt.Println(\"x\")\n```\n\n", 40),
		"| name | value |\n| --- | --- |\n" + strings.Repeat("| alpha | beta |\n", 80),
		strings.Repeat("A section of the notes.\n\n---\n\n", 40),
		strings.Repeat("> \n", 30) + "A paragraph follows.\n",
		"See *this* and **that** in a sentence.\n",
		strings.Repeat("│ │ │\n", 40),
		strings.Repeat("*** ---\n", 80),
		"done\n" + strings.Repeat("*\n", 20),
		"done\n" + strings.Repeat("\n", 20),
		"done\n" + strings.Repeat("*", 80),
		strings.Repeat("* \n", 40) + "The sentence resumes here.\n" + strings.Repeat("* \n", 40),
	}
	for i, text := range texts {
		g := &loopGuard{}
		if trip, ok := feedAll(t, g, "content", text); ok {
			t.Errorf("case %d: taken for a loop: %s", i, trip.message())
		}
	}
}
