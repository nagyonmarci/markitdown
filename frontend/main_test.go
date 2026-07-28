package main

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
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
			LLMAPIKey:    "sk-test",
		})
		if len(args) == 0 || args[len(args)-1] != value {
			t.Fatalf("commandArgs did not preserve the input location as the final argument: %#v", args)
		}
		if !slices.Contains(args, "--llm-model") {
			t.Fatalf("commandArgs did not add --llm-model for a non-empty LLMAPIKey: %#v", args)
		}
	})
}

func TestLLMEnv(t *testing.T) {
	if env := llmEnv(conversionOptions{}); env != nil {
		t.Fatalf("llmEnv should return nil with no LLM options set: %#v", env)
	}

	env := llmEnv(conversionOptions{LLMBaseURL: "http://localhost:11434/v1"})
	if !slices.Contains(env, "OPENAI_API_KEY=ollama") {
		t.Fatalf("llmEnv did not default an API key for Ollama: %#v", env)
	}
	if !slices.Contains(env, "OPENAI_BASE_URL=http://localhost:11434/v1") {
		t.Fatalf("llmEnv did not set OPENAI_BASE_URL: %#v", env)
	}

	env = llmEnv(conversionOptions{LLMAPIKey: "sk-test"})
	if !slices.Contains(env, "OPENAI_API_KEY=sk-test") {
		t.Fatalf("llmEnv did not use the provided API key: %#v", env)
	}
}

// newFileHeader builds a *multipart.FileHeader the way an incoming upload would produce one.
func newFileHeader(t *testing.T, filename, content string) *multipart.FileHeader {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("files", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", w.FormDataContentType())
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		t.Fatal(err)
	}
	return r.MultipartForm.File["files"][0]
}

func TestLogStoreCapacityAndOrder(t *testing.T) {
	store := &logStore{}
	for i := 0; i < maxLogEntries+10; i++ {
		store.Add(logEntry{Name: strconv.Itoa(i)})
	}

	entries := store.Snapshot()
	if len(entries) != maxLogEntries {
		t.Fatalf("len = %d, want %d", len(entries), maxLogEntries)
	}
	if entries[0].Name != "10" {
		t.Errorf("oldest entry Name = %q, want %q (first 10 should have been evicted)", entries[0].Name, "10")
	}
	if want := strconv.Itoa(maxLogEntries + 9); entries[len(entries)-1].Name != want {
		t.Errorf("newest entry Name = %q, want %q", entries[len(entries)-1].Name, want)
	}
}

func TestLogConversion(t *testing.T) {
	t.Setenv("PATH", "") // guarantees convertFile/convertLocation fail fast, with no network or external binary needed
	conversionLog = &logStore{}

	header := newFileHeader(t, "report.pdf", "dummy")
	fileResult := convertFile(context.Background(), header, 0, conversionOptions{})
	if fileResult.Error == "" {
		t.Fatal("expected convertFile to fail with an empty PATH")
	}

	locResult := convertLocation(context.Background(), "https://example.com/doc.html", 0, conversionOptions{})
	if locResult.Error == "" {
		t.Fatal("expected convertLocation to fail with an empty PATH")
	}

	entries := conversionLog.Snapshot()
	if len(entries) != 2 {
		t.Fatalf("expected 2 log entries, got %d", len(entries))
	}
	if entries[0].Source != "file" || entries[0].Name != "report.pdf" || entries[0].Error != fileResult.Error {
		t.Errorf("file entry = %+v", entries[0])
	}
	if entries[1].Source != "url" || entries[1].Name != "https://example.com/doc.html" || entries[1].Error != locResult.Error {
		t.Errorf("url entry = %+v", entries[1])
	}
}

func TestRequireAdminAuth(t *testing.T) {
	next := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}

	tests := []struct {
		name        string
		password    string
		sendAuth    bool
		authUser    string
		authPass    string
		wantStatus  int
		wantWWWAuth bool
	}{
		{name: "no password set", password: "", wantStatus: http.StatusServiceUnavailable},
		{name: "missing credentials", password: "secret", wantStatus: http.StatusUnauthorized, wantWWWAuth: true},
		{name: "wrong password", password: "secret", sendAuth: true, authUser: "admin", authPass: "wrong", wantStatus: http.StatusUnauthorized, wantWWWAuth: true},
		{name: "wrong username", password: "secret", sendAuth: true, authUser: "root", authPass: "secret", wantStatus: http.StatusUnauthorized, wantWWWAuth: true},
		{name: "correct credentials", password: "secret", sendAuth: true, authUser: "admin", authPass: "secret", wantStatus: http.StatusTeapot},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("ADMIN_PASSWORD", tt.password)
			r := httptest.NewRequest(http.MethodGet, "/admin", nil)
			if tt.sendAuth {
				r.SetBasicAuth(tt.authUser, tt.authPass)
			}
			w := httptest.NewRecorder()
			requireAdminAuth(next)(w, r)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if got := w.Header().Get("WWW-Authenticate"); tt.wantWWWAuth && got == "" {
				t.Error("expected a WWW-Authenticate header, got none")
			}
		})
	}
}

func TestAdminHandlerRendersEntries(t *testing.T) {
	t.Setenv("PATH", "")
	conversionLog = &logStore{}

	convertLocation(context.Background(), "https://example.com/doc.html", 0, conversionOptions{LLMAPIKey: "sk-should-never-leak"})
	conversionLog.Add(logEntry{Time: time.Now(), Duration: time.Second, Source: "file", Name: "report.pdf"})

	r := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()
	adminHandler(w, r)

	body := w.Body.String()
	if !strings.Contains(body, "report.pdf") {
		t.Error("response missing the successful entry")
	}
	if !strings.Contains(body, "https://example.com/doc.html") {
		t.Error("response missing the failed entry")
	}
	if !strings.Contains(body, "Conversion failed.") {
		t.Error("response missing the error status")
	}
	if strings.Contains(body, "sk-should-never-leak") {
		t.Error("response leaked the LLM API key")
	}
}
