package ports

import "testing"

func TestParseAndContains(t *testing.T) {
	s, err := Parse("80,445,5000-5005")
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []int{80, 445, 5000, 5003, 5005} {
		if !s.Contains(p) {
			t.Fatalf("expected %d in set", p)
		}
	}
	if s.Contains(5006) {
		t.Fatal("did not expect 5006")
	}
}

func TestParseInvalid(t *testing.T) {
	if _, err := Parse("70000"); err == nil {
		t.Fatal("expected error")
	}
}
