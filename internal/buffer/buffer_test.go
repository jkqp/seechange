package buffer

import (
	"testing"
)

func TestNew(t *testing.T) {
	got := New(3)

	if got.ch == nil {
		t.Fatal("New(3): ch is nil, want an initalised channel")
	}

	if c := cap(got.ch); c != 3 {
		t.Errorf("New(3): cap(ch) = %d, want 3", c)
	}

	if l := len(got.ch); l != 0 {
		t.Errorf("New(3): len(ch) = %d, want 0", l)
	}
}

func TestPush(t *testing.T) {
	tests := []struct {
		name         string
		prefill      []string
		input        string
		want         bool
		wantContents []string
		capacity     int
	}{
		{name: "push to empty", prefill: nil, input: "A", want: true, wantContents: []string{"A"}, capacity: 1},
		{name: "push to full", prefill: []string{"A"}, input: "B", want: false, wantContents: []string{"A"}, capacity: 1},
		{name: "push empty string", prefill: nil, input: "", want: true, wantContents: []string{""}, capacity: 1},
		{name: "push to partly full buffer", prefill: []string{"A"}, input: "B", want: true, wantContents: []string{"A", "B"}, capacity: 2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b := New(tc.capacity)
			for _, s := range tc.prefill {
				b.ch <- s
			}
			got := b.Push(tc.input)
			if got != tc.want {
				t.Errorf("Push(%q) = %v, want %v", tc.input, got, tc.want)
			}

			if l := len(b.ch); l != len(tc.wantContents) {
				t.Fatalf("after Push(%q): len(ch) = %d, want %d", tc.input, l, len(tc.wantContents))
			}
			for i, want := range tc.wantContents {
				if s := <-b.ch; s != want {
					t.Errorf("after Push(%q): item %d = %q, want %q", tc.input, i, s, want)
				}
			}
		})
	}
}

func TestPull(t *testing.T) {
	tests := []struct {
		name         string
		prefill      []string
		wantOK       bool
		wantString   string
		wantContents []string
		capacity     int
	}{
		{name: "pull from empty", prefill: []string{}, wantOK: false, wantString: "", wantContents: []string{}, capacity: 1},
		{name: "pull from full with one string", prefill: []string{"A"}, wantOK: true, wantString: "A", wantContents: []string{}, capacity: 1},
		{name: "pull from full with multiple string", prefill: []string{"A", "B"}, wantOK: true, wantString: "A", wantContents: []string{"B"}, capacity: 2},
		{name: "pull from empty string", prefill: []string{""}, wantOK: true, wantString: "", wantContents: []string{}, capacity: 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b := New(tc.capacity)
			for _, s := range tc.prefill {
				b.ch <- s
			}
			s, ok := b.Pull()
			if ok != tc.wantOK {
				t.Errorf("Pull() ok = %v, want %v", ok, tc.wantOK)
			}
			if s != tc.wantString {
				t.Errorf("Pull() = %q, want %q", s, tc.wantString)
			}

			if l := len(b.ch); l != len(tc.wantContents) {
				t.Fatalf("after Pull(): len(ch) = %d, want %d", l, len(tc.wantContents))
			}
			for i, want := range tc.wantContents {
				if s := <-b.ch; s != want {
					t.Errorf("after Pull(): item %d = %q, want %q", i, s, want)
				}
			}

		})
	}
}

func TestSize(t *testing.T) {
	tests := []struct {
		name     string
		prefill  []string
		capacity int
		want     int
	}{
		{name: "empty", prefill: []string{}, capacity: 2, want: 0},
		{name: "one item", prefill: []string{"A"}, capacity: 2, want: 1},
		{name: "full", prefill: []string{"A", "B"}, capacity: 2, want: 2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b := New(tc.capacity)
			for _, s := range tc.prefill {
				b.ch <- s
			}

			if got := b.Size(); got != tc.want {
				t.Errorf("Size() = %d, want %d", got, tc.want)
			}
		})
	}
}
