package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFormatSlack(t *testing.T) {
	body, err := FormatPayload("slack", "lobby", Message{Seq: 1, From: "nick", Text: "hi"})
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]string
	json.Unmarshal(body, &m)
	if !strings.Contains(m["text"], "hi") || !strings.Contains(m["text"], "#lobby") {
		t.Fatalf("unexpected slack body: %s", body)
	}
}

func TestFormatDiscord(t *testing.T) {
	body, _ := FormatPayload("discord", "lobby", Message{Seq: 1, From: "nick", Text: "yo"})
	var m map[string]string
	json.Unmarshal(body, &m)
	if !strings.Contains(m["content"], "yo") {
		t.Fatalf("unexpected discord body: %s", body)
	}
}

func TestFormatGenericHasStructuredMessage(t *testing.T) {
	body, _ := FormatPayload("generic", "lobby", Message{Seq: 7, From: "did:key:z", Text: "hey"})
	if !strings.Contains(string(body), "\"seq\":7") {
		t.Fatalf("generic payload should include structured message: %s", body)
	}
}

func TestShortTruncatesDid(t *testing.T) {
	got := short("did:key:z6MkExampleExampleExampleExample")
	if len(got) >= 40 || !strings.Contains(got, "…") {
		t.Fatalf("expected truncated did, got %q", got)
	}
}
