package markdown

import (
	"reflect"
	"testing"
)

func TestTags_ExtractsSimpleTags(t *testing.T) {
	text := "Hello #world and #foo\nNew line with #bar\n"
	tags := Tags(text)
	got := []string{}
	for _, t := range tags {
		got = append(got, t.Name)
	}
	want := []string{"#world", "#foo", "#bar"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func TestTags_IgnoresHeadings(t *testing.T) {
	text := "# Heading 1\nSome text with #tag\n## Heading 2\nAnother #another\n"
	tags := Tags(text)
	got := []string{}
	for _, t := range tags {
		got = append(got, t.Name)
	}
	want := []string{"#tag", "#another"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func TestTags_IgnoresCodeFences(t *testing.T) {
	text := "Some #real-tag\n```\n#not-a-tag-inside-code\n```\n#also-real\n"
	tags := Tags(text)
	got := []string{}
	for _, t := range tags {
		got = append(got, t.Name)
	}
	want := []string{"#real-tag", "#also-real"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func TestHighlightLine_HighlightsTags(t *testing.T) {
	out := HighlightLine("Some text with #tag here", HighlightParams{})
	if !contains(out, "tag") {
		t.Errorf("expected tag in highlighted output, got %q", out)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
