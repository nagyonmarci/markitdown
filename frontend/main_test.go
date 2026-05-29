package main

import (
	"strings"
	"testing"
	"unicode"
)

func FuzzFrontendHelpers(f *testing.F) {
	seeds := []string{
		"https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		"https://youtu.be/dQw4w9WgXcQ",
		"https://example.com/docs/report.pdf",
		"Quarterly Report 2026.pdf",
		"../../etc/passwd",
		"emoji 🚀 file name",
		"",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, value string) {
		safe := safeFileName(value)
		if safe == "" {
			t.Fatal("safeFileName returned an empty name")
		}
		if strings.ContainsAny(safe, `/\:`) {
			t.Fatalf("safeFileName returned path separator: %q", safe)
		}
		for _, char := range safe {
			if !(unicode.IsLetter(char) || unicode.IsDigit(char) || char == '-') {
				t.Fatalf("safeFileName returned unsafe character %q in %q", char, safe)
			}
		}

		markdownName := markdownFileName(value)
		if !strings.HasSuffix(markdownName, ".md") {
			t.Fatalf("markdownFileName did not produce markdown file: %q", markdownName)
		}

		urlName := markdownURLFileName(value)
		if !strings.HasSuffix(urlName, ".md") {
			t.Fatalf("markdownURLFileName did not produce markdown file: %q", urlName)
		}

		args := commandArgs(value, conversionOptions{
			UsePlugins:   true,
			Extension:    "pdf",
			Charset:      "utf-8",
			KeepDataURIs: true,
		})
		if len(args) == 0 || args[len(args)-1] != value {
			t.Fatalf("commandArgs did not preserve the input location as the final argument: %#v", args)
		}
	})
}
