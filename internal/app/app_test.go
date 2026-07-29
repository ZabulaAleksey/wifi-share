package app

import (
	"path/filepath"
	"testing"
)

func TestIDsRoundTrip(t *testing.T) {
	input := filepath.Join("Music", "Artist", "track.mp3")
	id := encodeID(input)
	output, err := decodeID(id)
	if err != nil {
		t.Fatal(err)
	}
	if output != input {
		t.Fatalf("expected %q, got %q", input, output)
	}
}

func TestDecodeIDRejectsTraversal(t *testing.T) {
	id := encodeID(filepath.Join("..", "secret.txt"))
	if _, err := decodeID(id); err == nil {
		t.Fatal("expected traversal to be rejected")
	}
}

func TestSafeName(t *testing.T) {
	for _, name := range []string{"../secret", "folder/file", `folder\file`, "", ".."} {
		if _, err := safeName(name); err == nil {
			t.Errorf("expected %q to be rejected", name)
		}
	}
	if _, err := safeName("holiday photo.jpg"); err != nil {
		t.Fatal(err)
	}
}
