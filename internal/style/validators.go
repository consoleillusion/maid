package style

import (
	"fmt"
	"regexp"
	"strings"
)

// STE limits (ASD-STE100 rules STE-1, STE-2, STE-7).
const (
	steMaxSentenceWords  = 25
	steMaxParagraphSents = 6
)

// ValidateSTE checks sentence and paragraph length.
func ValidateSTE(body string) []Finding {
	// Purpose: apply the machine-checkable STE rules to a chapter body.
	prose, _ := Prose(body)
	var out []Finding
	for _, s := range sentences(prose) {
		if n := words(s.Text); n > steMaxSentenceWords {
			out = append(out, Finding{s.Line, "STE-2", Error,
				fmt.Sprintf("sentence has %d words (max %d): %s", n, steMaxSentenceWords, clip(s.Text))})
		}
	}
	for _, p := range paragraphs(prose) {
		if n := len(sentences(p.text)); n > steMaxParagraphSents {
			out = append(out, Finding{p.line, "STE-7", Error,
				fmt.Sprintf("paragraph has %d sentences (max %d)", n, steMaxParagraphSents)})
		}
	}
	return out
}

type para struct {
	line int
	text string
}

func paragraphs(prose string) []para {
	// Purpose: split prose into paragraphs that are not lists.
	var out []para
	var cur []string
	start := 0
	flush := func() {
		if len(cur) > 0 {
			out = append(out, para{start, strings.Join(cur, "\n")})
		}
		cur = nil
	}
	for i, l := range strings.Split(prose, "\n") {
		t := strings.TrimSpace(l)
		if t == "" || listItem(t) {
			flush()
			continue
		}
		if len(cur) == 0 {
			start = i + 1
		}
		cur = append(cur, t)
	}
	flush()
	return out
}

var reqHeading = regexp.MustCompile(`^#{2,4}\s+REQ-[A-Z]+-\d{3}\b`)
var earsStart = regexp.MustCompile(`^(The |When |While |If |Where |Each |maid |A |An )`)
var earsBanned = regexp.MustCompile(`\b(should|may|must)\b`)

// ValidateEARS checks the first sentence after each REQ heading.
func ValidateEARS(body string) []Finding {
	// Purpose: apply EARS-1 to EARS-3 to every requirement in a body.
	lines := strings.Split(body, "\n")
	var out []Finding
	inFence := false
	for i := 0; i < len(lines); i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "```") {
			inFence = !inFence
		}
		if inFence || !reqHeading.MatchString(lines[i]) {
			continue
		}
		j := i + 1
		for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
			j++
		}
		if j >= len(lines) || strings.HasPrefix(lines[j], "#") {
			out = append(out, Finding{i + 1, "EARS-0", Error, "requirement has no text"})
			continue
		}
		var req []string
		for k := j; k < len(lines) && strings.TrimSpace(lines[k]) != ""; k++ {
			req = append(req, strings.TrimSpace(lines[k]))
		}
		out = append(out, checkEARS(j+1, strings.Join(req, " "))...)
	}
	return out
}

func checkEARS(line int, s string) []Finding {
	// Purpose: check one requirement sentence against the EARS card.
	var out []Finding
	n := strings.Count(" "+s+" ", " shall ")
	if n == 0 {
		out = append(out, Finding{line, "EARS-1", Error, "requirement has no \"shall\": " + clip(s)})
	}
	if n > 2 {
		out = append(out, Finding{line, "EARS-3", Error, fmt.Sprintf("requirement has %d responses (max 2): %s", n, clip(s))})
	}
	if m := earsBanned.FindString(stripCode(s)); m != "" {
		out = append(out, Finding{line, "EARS-1", Error, fmt.Sprintf("use \"shall\", not %q", m)})
	}
	if !earsStart.MatchString(s) {
		out = append(out, Finding{line, "EARS-P", Error, "requirement does not start with an EARS pattern: " + clip(s)})
	}
	return out
}

var sbvrRule = regexp.MustCompile(`^It is (obligatory|prohibited|permitted|necessary|possible) that `)

// ValidateSBVR checks the rule keywords in the "Business rules" section and
// the definition form in the "Terms" section.
func ValidateSBVR(body string) []Finding {
	// Purpose: apply SBVR-2 and the definition form to a glossary body.
	var out []Finding
	section := ""
	for i, l := range strings.Split(body, "\n") {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "## ") {
			section = strings.ToLower(strings.TrimPrefix(t, "## "))
			continue
		}
		switch {
		case strings.Contains(section, "rules") && strings.HasPrefix(t, "- "):
			if !sbvrRule.MatchString(strings.TrimPrefix(t, "- ")) {
				out = append(out, Finding{i + 1, "SBVR-2", Error, "rule does not start with a rule keyword: " + clip(t)})
			}
		case section == "terms" && t != "" && !strings.HasPrefix(t, "**"):
			out = append(out, Finding{i + 1, "SBVR-1", Error, "term definition must start with **term**: " + clip(t)})
		}
	}
	return out
}

// ValidateGherkin checks the structure of a Gherkin block.
func ValidateGherkin(text string) []Finding {
	// Purpose: require scenarios, steps in order, and exactly one When per scenario.
	var out []Finding
	scenarioLine, whens, seenThen := 0, 0, false
	end := func() {
		if scenarioLine > 0 {
			if whens != 1 {
				out = append(out, Finding{scenarioLine, "GH-2", Error, fmt.Sprintf("scenario has %d When steps (want 1)", whens)})
			}
			if !seenThen {
				out = append(out, Finding{scenarioLine, "GH-3", Error, "scenario has no Then step"})
			}
		}
	}
	for i, l := range strings.Split(text, "\n") {
		t := strings.TrimSpace(l)
		switch {
		case strings.HasPrefix(t, "Scenario:"), strings.HasPrefix(t, "Scenario Outline:"):
			end()
			scenarioLine, whens, seenThen = i+1, 0, false
		case strings.HasPrefix(t, "When "):
			whens++
		case strings.HasPrefix(t, "Then "):
			seenThen = true
		}
	}
	end()
	if !strings.Contains(text, "Scenario") {
		out = append(out, Finding{1, "GH-1", Error, "block has no Scenario"})
	}
	return out
}

func stripCode(s string) string {
	// Purpose: remove inline code spans before word checks.
	return regexp.MustCompile("`[^`]*`").ReplaceAllString(s, "")
}

func clip(s string) string {
	// Purpose: shorten text for a finding message.
	if len(s) > 90 {
		return s[:87] + "..."
	}
	return s
}
