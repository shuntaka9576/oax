package oax

import (
	"os/user"
	"strings"
	"testing"
)

func TestReplaceTildeWithHomedir(t *testing.T) {
	usr, err := user.Current()
	if err != nil {
		t.Fatalf("Error: Cannot get the current user: %v", err)
	}
	homedir := usr.HomeDir

	testCases := []struct {
		input    string
		expected string
	}{
		{"~/example/path", strings.Join([]string{homedir, "/example/path"}, "")},
		{"/example/path", "/example/path"},
		{"~", homedir},
		{"/", "/"},
		{"", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result, err := replaceTildeWithHomedir(tc.input)
			if err != nil {
				t.Fatalf("Error: Return err func: %v", err)
			}

			if result != tc.expected {
				t.Errorf("Expected %q but got %q", tc.expected, result)
			}
		})
	}
}

func TestReplaceHomedirWithTilde(t *testing.T) {
	usr, err := user.Current()
	if err != nil {
		t.Fatalf("Error: Cannot get the current user: %v", err)
	}
	homedir := usr.HomeDir

	testCases := []struct {
		input    string
		expected string
	}{
		{strings.Join([]string{homedir, "/example/path"}, ""), "~/example/path"},
		{"/example/path", "/example/path"},
		{homedir, "~"},
		{"/", "/"},
		{"", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result, err := replaceHomedirWithTilde(tc.input)
			if err != nil {
				t.Fatalf("Error: Return err func: %v", err)
			}

			if result != tc.expected {
				t.Errorf("Expected %q but got %q", tc.expected, result)
			}
		})
	}
}
