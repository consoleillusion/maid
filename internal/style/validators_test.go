package style

import "testing"

func rules(fs []Finding) map[string]int {
	// Purpose: count findings by rule for compact assertions.
	m := map[string]int{}
	for _, f := range fs {
		m[f.Rule]++
	}
	return m
}

// REQ-STY-003
func TestSTESentenceLength(t *testing.T) {
	// Purpose: flag a sentence over 25 words and ignore tables and code.
	long := "This sentence has far too many words in it because it keeps going on and on without stopping for any good reason at all on this long day here."
	body := "Short sentence. " + long + "\n\n| a very long table cell that has many many many many many many many many many many many many words |\n\n```\n" + long + "\n```\n"
	got := rules(ValidateSTE(body))
	if got["STE-2"] != 1 {
		t.Fatalf("STE-2 = %d, want 1: %v", got["STE-2"], ValidateSTE(body))
	}
}

func TestSTEListItemsAreSentences(t *testing.T) {
	// Purpose: treat each list item as its own sentence, even without a period.
	body := "- one two three\n- four five six\n"
	if fs := ValidateSTE(body); len(fs) != 0 {
		t.Fatalf("unexpected findings: %v", fs)
	}
}

// REQ-STY-003
func TestEARS(t *testing.T) {
	// Purpose: accept valid EARS patterns and reject the common mistakes.
	body := "### REQ-AB-001 Good\nWhen an agent claims a work item, maid shall return the claim.\n\n" +
		"### REQ-AB-002 Should\nThe merge queue should reject it.\n\n" +
		"### REQ-AB-003 Many\nThe merge queue shall a, shall b and shall c.\n\n" +
		"### REQ-AB-004 Empty\n\n### REQ-AB-005 Pattern\nRejected submissions shall be logged.\n"
	got := rules(ValidateEARS(body))
	want := map[string]int{"EARS-1": 2, "EARS-3": 1, "EARS-0": 1, "EARS-P": 1}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %d, want %d (all: %v)", k, got[k], v, got)
		}
	}
}

func TestSBVR(t *testing.T) {
	// Purpose: require rule keywords in rules and bold terms in definitions.
	body := "## Terms\n\n**claim**: a right.\nlease: a time.\n\n## Business rules\n\n- It is obligatory that each **claim** expires.\n- Claims expire.\n"
	got := rules(ValidateSBVR(body))
	if got["SBVR-1"] != 1 || got["SBVR-2"] != 1 {
		t.Fatalf("got %v", got)
	}
}

func TestGherkin(t *testing.T) {
	// Purpose: require exactly one When and at least one Then per scenario.
	ok := "Feature: x\n  Scenario: a\n    Given s\n    When w\n    Then t\n"
	if fs := ValidateGherkin(ok); len(fs) != 0 {
		t.Fatalf("valid block flagged: %v", fs)
	}
	bad := "Feature: x\n  Scenario: a\n    Given s\n    When w\n    When v\n"
	got := rules(ValidateGherkin(bad))
	if got["GH-2"] != 1 || got["GH-3"] != 1 {
		t.Fatalf("got %v", got)
	}
}
