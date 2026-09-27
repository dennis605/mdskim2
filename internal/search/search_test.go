package search

import "testing"

func TestFindEinfach(t *testing.T) {
	lines := []string{"hello world", "world hello", "nichts"}
	ms := Find(lines, "world", Options{})
	if len(ms) != 2 {
		t.Fatalf("erwartet 2 matches, got %d", len(ms))
	}
	if ms[0].Line != 0 || ms[0].Col != 6 {
		t.Errorf("Match 0 falsch: %+v", ms[0])
	}
	if ms[1].Line != 1 || ms[1].Col != 0 {
		t.Errorf("Match 1 falsch: %+v", ms[1])
	}
}

func TestFindCaseInsensitive(t *testing.T) {
	lines := []string{"Hello", "HELLO"}
	ms := Find(lines, "hello", Options{CaseSensitive: false})
	if len(ms) != 2 {
		t.Errorf("case-insensitive sollte 2 finden, got %d", len(ms))
	}
}

func TestFindCaseSensitive(t *testing.T) {
	lines := []string{"Hello", "HELLO"}
	ms := Find(lines, "hello", Options{CaseSensitive: true})
	if len(ms) != 0 {
		t.Errorf("case-sensitive sollte 0 finden, got %d", len(ms))
	}
}

func TestFindWholeWord(t *testing.T) {
	lines := []string{"helloworld", "hello world"}
	ms := Find(lines, "hello", Options{WholeWord: true})
	if len(ms) != 1 {
		t.Fatalf("erwartet 1 whole-word match, got %d", len(ms))
	}
	if ms[0].Line != 1 {
		t.Errorf("Match sollte in Zeile 1 sein, got Line=%d", ms[0].Line)
	}
}

func TestFindLeererPattern(t *testing.T) {
	lines := []string{"foo", "bar"}
	ms := Find(lines, "", Options{})
	if ms != nil {
		t.Errorf("leerer pattern sollte nil sein, got %v", ms)
	}
}

func TestReplaceAll(t *testing.T) {
	lines := []string{"foo bar foo", "foo baz"}
	out, count := ReplaceAll(lines, "foo", "FOO", Options{})
	if count != 3 {
		t.Errorf("erwartet 3 replaces, got %d", count)
	}
	if out[0] != "FOO bar FOO" {
		t.Errorf("Zeile 0 falsch: %q", out[0])
	}
}

func TestReplaceAllNichtsGefunden(t *testing.T) {
	lines := []string{"foo", "bar"}
	out, count := ReplaceAll(lines, "baz", "QUX", Options{})
	if count != 0 {
		t.Errorf("erwartet 0 replaces, got %d", count)
	}
	if out[0] != "foo" || out[1] != "bar" {
		t.Errorf("Inhalt sollte unverändert sein")
	}
}
