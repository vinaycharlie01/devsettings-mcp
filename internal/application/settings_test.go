package application

import (
	"testing"
)

func TestDeepMerge(t *testing.T) {
	tests := []struct {
		name    string
		base    map[string]any
		overlay map[string]any
		want    map[string]any
	}{
		{
			name:    "empty base",
			base:    map[string]any{},
			overlay: map[string]any{"a": 1},
			want:    map[string]any{"a": 1},
		},
		{
			name:    "empty overlay",
			base:    map[string]any{"a": 1},
			overlay: map[string]any{},
			want:    map[string]any{"a": 1},
		},
		{
			name:    "simple override",
			base:    map[string]any{"a": 1},
			overlay: map[string]any{"a": 2},
			want:    map[string]any{"a": 2},
		},
		{
			name:    "add new key",
			base:    map[string]any{"a": 1},
			overlay: map[string]any{"b": 2},
			want:    map[string]any{"a": 1, "b": 2},
		},
		{
			name: "nested merge",
			base: map[string]any{
				"editor": map[string]any{
					"tabSize":      4,
					"formatOnSave": false,
				},
			},
			overlay: map[string]any{
				"editor": map[string]any{
					"formatOnSave": true,
				},
			},
			want: map[string]any{
				"editor": map[string]any{
					"tabSize":      4,
					"formatOnSave": true,
				},
			},
		},
		{
			name: "overlay replaces non-map with map",
			base: map[string]any{
				"editor": "simple",
			},
			overlay: map[string]any{
				"editor": map[string]any{"tabSize": 4},
			},
			want: map[string]any{
				"editor": map[string]any{"tabSize": 4},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeepMerge(tt.base, tt.overlay)
			assertMapsEqual(t, tt.want, got)
		})
	}
}

func TestComputeDiff(t *testing.T) {
	tests := []struct {
		name        string
		before      map[string]any
		after       map[string]any
		wantAdded   int
		wantChanged int
		wantRemoved int
	}{
		{
			name:        "no changes",
			before:      map[string]any{"a": 1},
			after:       map[string]any{"a": 1},
			wantAdded:   0,
			wantChanged: 0,
			wantRemoved: 0,
		},
		{
			name:        "added key",
			before:      map[string]any{},
			after:       map[string]any{"a": 1},
			wantAdded:   1,
			wantChanged: 0,
			wantRemoved: 0,
		},
		{
			name:        "changed value",
			before:      map[string]any{"a": 1},
			after:       map[string]any{"a": 2},
			wantAdded:   0,
			wantChanged: 1,
			wantRemoved: 0,
		},
		{
			name:        "removed key",
			before:      map[string]any{"a": 1},
			after:       map[string]any{},
			wantAdded:   0,
			wantChanged: 0,
			wantRemoved: 1,
		},
		{
			name:        "mixed changes",
			before:      map[string]any{"a": 1, "b": 2},
			after:       map[string]any{"a": 3, "c": 4},
			wantAdded:   1,
			wantChanged: 1,
			wantRemoved: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := ComputeDiff(tt.before, tt.after)
			if len(diff.Added) != tt.wantAdded {
				t.Errorf("added: got %d, want %d", len(diff.Added), tt.wantAdded)
			}
			if len(diff.Changed) != tt.wantChanged {
				t.Errorf("changed: got %d, want %d", len(diff.Changed), tt.wantChanged)
			}
			if len(diff.Removed) != tt.wantRemoved {
				t.Errorf("removed: got %d, want %d", len(diff.Removed), tt.wantRemoved)
			}
		})
	}
}

func assertMapsEqual(t *testing.T, want, got map[string]any) {
	t.Helper()
	for k, wv := range want {
		gv, ok := got[k]
		if !ok {
			t.Errorf("missing key %q", k)
			continue
		}
		wm, wIsMap := wv.(map[string]any)
		gm, gIsMap := gv.(map[string]any)
		if wIsMap && gIsMap {
			assertMapsEqual(t, wm, gm)
		} else if !jsonEqual(wv, gv) {
			t.Errorf("key %q: got %v, want %v", k, gv, wv)
		}
	}
	for k := range got {
		if _, ok := want[k]; !ok {
			t.Errorf("unexpected key %q", k)
		}
	}
}
