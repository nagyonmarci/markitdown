package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const maxUploadSize = 50 << 20

type pageData struct {
	Results []conversionResult
	Error   string
}

type conversionOptions struct {
	UsePlugins   bool
	Extension    string
	Charset      string
	KeepDataURIs bool
}

type conversionResult struct {
	ID           string
	FileName     string
	DownloadName string
	Markdown     string
	Error        string
}

type youtubeMetadata struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Duration    int    `json:"duration"`
	Uploader    string `json:"uploader"`
	WebpageURL  string `json:"webpage_url"`
}

type youtubeSubtitle struct {
	Events []struct {
		Segments []struct {
			Text string `json:"utf8"`
		} `json:"segs"`
	} `json:"events"`
}

var page = template.Must(template.New("page").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>MarkItDown</title>
  <style>
    :root {
      color-scheme: light dark;
      font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      background: #f6f7f9;
      color: #20242a;
    }
    body {
      margin: 0;
      min-height: 100vh;
    }
    main {
      width: min(960px, calc(100% - 32px));
      margin: 0 auto;
      padding: 40px 0;
    }
    header {
      margin-bottom: 24px;
    }
    h1 {
      margin: 0 0 8px;
      font-size: 32px;
      line-height: 1.15;
      letter-spacing: 0;
    }
    p {
      margin: 0;
      color: #5d6673;
    }
    form {
      display: grid;
      gap: 16px;
      padding: 20px;
      border: 1px solid #d7dce2;
      border-radius: 8px;
      background: #ffffff;
    }
    input[type="file"] {
      padding: 12px;
      border: 1px dashed #aab3bf;
      border-radius: 6px;
      background: #fafbfc;
    }
    input[type="text"], input[type="url"] {
      box-sizing: border-box;
      min-height: 44px;
      padding: 0 12px;
      border: 1px solid #aab3bf;
      border-radius: 6px;
      background: #fafbfc;
      color: inherit;
      font: inherit;
    }
    label {
      display: grid;
      gap: 6px;
      color: #404852;
      font-size: 14px;
    }
    label.checkbox {
      display: flex;
      align-items: center;
      gap: 8px;
    }
    details {
      display: grid;
      gap: 12px;
    }
    summary {
      cursor: pointer;
      color: #404852;
      font-weight: 600;
    }
    .advanced {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
      gap: 12px;
      margin-top: 12px;
    }
    button {
      justify-self: start;
      min-height: 40px;
      padding: 0 16px;
      border: 0;
      border-radius: 6px;
      background: #106ebe;
      color: #fff;
      font: inherit;
      cursor: pointer;
    }
    button:hover {
      background: #0b5ea8;
    }
    .actions {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
    }
    .secondary {
      background: #eef2f7;
      color: #20242a;
    }
    .secondary:hover {
      background: #dfe6ef;
    }
    .message {
      margin-top: 20px;
      padding: 12px 14px;
      border-radius: 6px;
      background: #fff2f0;
      color: #a4262c;
      border: 1px solid #f3b8b5;
    }
    section {
      margin-top: 24px;
    }
    .result {
      margin-top: 18px;
    }
    .result-head {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 16px;
      margin-bottom: 10px;
    }
    h2 {
      margin: 0;
      font-size: 18px;
      letter-spacing: 0;
    }
    .file-error {
      margin-top: 10px;
      color: #a4262c;
      font-size: 14px;
    }
    textarea {
      box-sizing: border-box;
      width: 100%;
      min-height: 420px;
      padding: 16px;
      border: 1px solid #d7dce2;
      border-radius: 8px;
      background: #ffffff;
      color: inherit;
      font: 14px/1.5 ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
      resize: vertical;
    }
    @media (prefers-color-scheme: dark) {
      :root {
        background: #111418;
        color: #eef1f5;
      }
      p {
        color: #a9b2bd;
      }
      form, textarea {
        background: #181d23;
        border-color: #313943;
      }
      .secondary {
        background: #2a323d;
        color: #eef1f5;
      }
      .secondary:hover {
        background: #354150;
      }
      input[type="file"] {
        background: #14191f;
        border-color: #4b5563;
      }
      input[type="text"], input[type="url"] {
        background: #14191f;
        border-color: #4b5563;
      }
      label, summary {
        color: #c6cdd6;
      }
      .message {
        background: #3b1518;
        color: #ffd7d4;
        border-color: #7a2d33;
      }
    }
  </style>
</head>
<body>
  <main>
    <header>
      <h1>MarkItDown</h1>
      <p>Convert a local document to Markdown.</p>
    </header>

    <form action="/convert" method="post" enctype="multipart/form-data">
      <input name="files" type="file" multiple required>
      <details>
        <summary>Advanced options</summary>
        <div class="advanced">
          <label>Extension hint<input name="extension" type="text" placeholder="pdf"></label>
          <label>Charset<input name="charset" type="text" placeholder="utf-8"></label>
          <label class="checkbox"><input name="use_plugins" type="checkbox" value="1"> Use plugins</label>
          <label class="checkbox"><input name="keep_data_uris" type="checkbox" value="1"> Keep data URIs</label>
        </div>
      </details>
      <div class="actions">
        <button type="submit">Convert</button>
        <button type="submit" class="secondary" formaction="/convert.zip">Download ZIP</button>
      </div>
    </form>

    <form action="/convert-url" method="post">
      <input name="url" type="url" placeholder="https://example.com/document.html" required>
      <details>
        <summary>Advanced options</summary>
        <div class="advanced">
          <label>Extension hint<input name="extension" type="text" placeholder="html"></label>
          <label>Charset<input name="charset" type="text" placeholder="utf-8"></label>
          <label class="checkbox"><input name="use_plugins" type="checkbox" value="1"> Use plugins</label>
          <label class="checkbox"><input name="keep_data_uris" type="checkbox" value="1"> Keep data URIs</label>
        </div>
      </details>
      <button type="submit">Convert URL</button>
    </form>

    {{if .Error}}<div class="message">{{.Error}}</div>{{end}}

    {{if .Results}}
    <section>
      <div class="result-head">
        <h2>Results</h2>
        <div class="actions">
          <button type="button" class="secondary" id="copy-all-button">Copy all</button>
        </div>
      </div>
      {{range .Results}}
      <div class="result">
        <div class="result-head">
          <h2>{{.FileName}}</h2>
          {{if .Markdown}}
          <div class="actions">
            <button type="button" class="copy-button" data-target="{{.ID}}">Copy</button>
            <button type="button" class="download-button" data-target="{{.ID}}" data-filename="{{.DownloadName}}">Save</button>
          </div>
          {{end}}
        </div>
        {{if .Error}}<div class="file-error">{{.Error}}</div>{{end}}
        {{if .Markdown}}<textarea id="{{.ID}}" readonly>{{.Markdown}}</textarea>{{end}}
      </div>
      {{end}}
    </section>
    {{end}}
  </main>
  <script>
    function outputFor(button) {
      return document.getElementById(button.dataset.target);
    }

    document.querySelectorAll(".copy-button").forEach((button) => {
      button.addEventListener("click", async () => {
        const output = outputFor(button);
        if (!output) return;
        await navigator.clipboard.writeText(output.value);
        button.textContent = "Copied";
        setTimeout(() => button.textContent = "Copy", 1200);
      });
    });

    document.querySelectorAll(".download-button").forEach((button) => {
      button.addEventListener("click", () => {
        const output = outputFor(button);
        if (!output) return;
        const file = new Blob([output.value], { type: "text/markdown;charset=utf-8" });
        const link = document.createElement("a");
        link.href = URL.createObjectURL(file);
        link.download = button.dataset.filename || "document.md";
        link.click();
        URL.revokeObjectURL(link.href);
      });
    });

    const copyAllButton = document.getElementById("copy-all-button");
    if (copyAllButton) {
      copyAllButton.addEventListener("click", async () => {
        const allMarkdown = Array.from(document.querySelectorAll("textarea"))
          .map((output) => output.value)
          .join("\n\n");
        await navigator.clipboard.writeText(allMarkdown);
        copyAllButton.textContent = "Copied";
        setTimeout(() => copyAllButton.textContent = "Copy all", 1200);
      });
    }
  </script>
</body>
</html>`))

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", render(pageData{}))
	mux.HandleFunc("POST /convert", convert)
	mux.HandleFunc("POST /convert.zip", convertZip)
	mux.HandleFunc("POST /convert-url", convertURL)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("frontend listening on http://0.0.0.0:8080")
	log.Fatal(server.ListenAndServe())
}

func render(data pageData) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := page.Execute(w, data); err != nil {
			log.Printf("render page: %v", err)
		}
	}
}

func convert(w http.ResponseWriter, r *http.Request) {
	headers, ok := uploadedFiles(w, r)
	if !ok {
		return
	}
	options := optionsFromRequest(r)

	results := make([]conversionResult, 0, len(headers))
	for index, header := range headers {
		results = append(results, convertFile(r.Context(), header, index, options))
	}

	render(pageData{Results: results})(w, r)
}

func convertZip(w http.ResponseWriter, r *http.Request) {
	headers, ok := uploadedFiles(w, r)
	if !ok {
		return
	}
	options := optionsFromRequest(r)

	var archive bytes.Buffer
	zipWriter := zip.NewWriter(&archive)
	usedNames := map[string]int{}
	var errors strings.Builder

	for index, header := range headers {
		result := convertFile(r.Context(), header, index, options)
		if result.Error != "" {
			fmt.Fprintf(&errors, "%s: %s\n", result.FileName, result.Error)
			continue
		}

		entry, err := zipWriter.Create(uniqueFileName(result.DownloadName, usedNames))
		if err != nil {
			http.Error(w, "Could not create ZIP archive.", http.StatusInternalServerError)
			return
		}
		if _, err := entry.Write([]byte(result.Markdown)); err != nil {
			http.Error(w, "Could not write ZIP archive.", http.StatusInternalServerError)
			return
		}
	}

	if errors.Len() > 0 {
		entry, err := zipWriter.Create("errors.txt")
		if err != nil {
			http.Error(w, "Could not create ZIP archive.", http.StatusInternalServerError)
			return
		}
		if _, err := entry.Write([]byte(errors.String())); err != nil {
			http.Error(w, "Could not write ZIP archive.", http.StatusInternalServerError)
			return
		}
	}

	if err := zipWriter.Close(); err != nil {
		http.Error(w, "Could not finish ZIP archive.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="markitdown-results.zip"`)
	w.Header().Set("Content-Length", strconv.Itoa(archive.Len()))
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(archive.Bytes()); err != nil {
		log.Printf("write ZIP response: %v", err)
	}
}

func convertURL(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		render(pageData{Error: "Could not read the URL."})(w, r)
		return
	}

	rawURL := strings.TrimSpace(r.FormValue("url"))
	if rawURL == "" {
		render(pageData{Error: "Enter a URL to convert."})(w, r)
		return
	}

	result := convertLocation(r.Context(), rawURL, 0, optionsFromRequest(r))
	render(pageData{Results: []conversionResult{result}})(w, r)
}

func uploadedFiles(w http.ResponseWriter, r *http.Request) ([]*multipart.FileHeader, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		render(pageData{Error: "Upload failed or files are larger than 50 MB."})(w, r)
		return nil, false
	}

	headers := r.MultipartForm.File["files"]
	if len(headers) == 0 {
		render(pageData{Error: "Choose at least one file to convert."})(w, r)
		return nil, false
	}

	return headers, true
}

func convertFile(requestContext context.Context, header *multipart.FileHeader, index int, options conversionOptions) conversionResult {
	result := conversionResult{
		ID:           "result-" + strconv.Itoa(index),
		FileName:     header.Filename,
		DownloadName: markdownFileName(header.Filename),
	}

	file, err := header.Open()
	if err != nil {
		result.Error = "Could not open the uploaded file."
		return result
	}
	defer file.Close()

	tmp, err := os.CreateTemp("", "markitdown-*"+filepath.Ext(header.Filename))
	if err != nil {
		result.Error = "Could not prepare the uploaded file."
		return result
	}
	defer os.Remove(tmp.Name())

	if _, err := io.Copy(tmp, file); err != nil {
		tmp.Close()
		result.Error = "Could not read the uploaded file."
		return result
	}
	if err := tmp.Close(); err != nil {
		result.Error = "Could not finish reading the uploaded file."
		return result
	}

	ctx, cancel := context.WithTimeout(requestContext, 2*time.Minute)
	defer cancel()

	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "markitdown", commandArgs(tmp.Name(), options)...)
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		result.Error = "Conversion failed."
		if stderr.Len() > 0 {
			result.Error += " " + stderr.String()
		}
		return result
	}

	result.Markdown = string(output)
	return result
}

func convertLocation(requestContext context.Context, location string, index int, options conversionOptions) conversionResult {
	result := conversionResult{
		ID:           "result-" + strconv.Itoa(index),
		FileName:     location,
		DownloadName: markdownURLFileName(location),
	}

	if isYouTubeURL(location) {
		return convertYouTube(requestContext, location, result)
	}

	ctx, cancel := context.WithTimeout(requestContext, 2*time.Minute)
	defer cancel()

	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "markitdown", commandArgs(location, options)...)
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		result.Error = "Conversion failed."
		if stderr.Len() > 0 {
			result.Error += " " + stderr.String()
		}
		return result
	}

	result.Markdown = string(output)
	return result
}

func convertYouTube(requestContext context.Context, location string, result conversionResult) conversionResult {
	ctx, cancel := context.WithTimeout(requestContext, 3*time.Minute)
	defer cancel()

	metadataOutput, err := exec.CommandContext(ctx, "yt-dlp", "--skip-download", "--no-warnings", "-J", location).Output()
	if err != nil {
		result.Error = "YouTube metadata conversion failed."
		return result
	}

	var metadata youtubeMetadata
	if err := json.Unmarshal(metadataOutput, &metadata); err != nil {
		result.Error = "Could not read YouTube metadata."
		return result
	}

	if metadata.Title != "" {
		result.FileName = metadata.Title
		result.DownloadName = markdownFileName(safeFileName(metadata.Title))
	}

	transcript, transcriptErr := fetchYouTubeTranscript(ctx, location)

	var markdown strings.Builder
	title := metadata.Title
	if title == "" {
		title = location
	}
	fmt.Fprintf(&markdown, "# %s\n\n", title)
	if metadata.WebpageURL != "" {
		fmt.Fprintf(&markdown, "- **Source:** %s\n", metadata.WebpageURL)
	} else {
		fmt.Fprintf(&markdown, "- **Source:** %s\n", location)
	}
	if metadata.Uploader != "" {
		fmt.Fprintf(&markdown, "- **Channel:** %s\n", metadata.Uploader)
	}
	if metadata.Duration > 0 {
		fmt.Fprintf(&markdown, "- **Duration:** %s\n", formatDuration(metadata.Duration))
	}
	if metadata.Description != "" {
		fmt.Fprintf(&markdown, "\n## Description\n\n%s\n", metadata.Description)
	}
	if transcript != "" {
		fmt.Fprintf(&markdown, "\n## Transcript\n\n%s\n", transcript)
	} else if transcriptErr != nil {
		fmt.Fprintf(&markdown, "\n## Transcript\n\nTranscript unavailable: %s\n", transcriptErr)
	}

	result.Markdown = markdown.String()
	return result
}

func fetchYouTubeTranscript(ctx context.Context, location string) (string, error) {
	dir, err := os.MkdirTemp("", "markitdown-youtube-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)

	outputTemplate := filepath.Join(dir, "subtitle.%(ext)s")
	cmd := exec.CommandContext(
		ctx,
		"yt-dlp",
		"--skip-download",
		"--no-warnings",
		"--write-subs",
		"--write-auto-subs",
		"--sub-langs",
		"en,en-orig",
		"--sub-format",
		"json3",
		"-o",
		outputTemplate,
		location,
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", errors.New(strings.TrimSpace(string(output)))
	}

	matches, err := filepath.Glob(filepath.Join(dir, "subtitle.*.json3"))
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no English subtitles found")
	}

	content, err := os.ReadFile(matches[0])
	if err != nil {
		return "", err
	}

	var subtitle youtubeSubtitle
	if err := json.Unmarshal(content, &subtitle); err != nil {
		return "", err
	}

	parts := []string{}
	for _, event := range subtitle.Events {
		for _, segment := range event.Segments {
			text := strings.TrimSpace(segment.Text)
			if text != "" {
				parts = append(parts, text)
			}
		}
	}

	return strings.Join(strings.Fields(strings.Join(parts, " ")), " "), nil
}

func isYouTubeURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.TrimPrefix(strings.ToLower(parsed.Hostname()), "www.")
	return host == "youtube.com" || host == "youtu.be"
}

func safeFileName(name string) string {
	var builder strings.Builder
	lastWasSeparator := false

	for _, char := range strings.TrimSpace(name) {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			builder.WriteRune(char)
			lastWasSeparator = false
			continue
		}

		if unicode.IsSpace(char) || char == '-' || char == '_' {
			if builder.Len() > 0 && !lastWasSeparator {
				builder.WriteRune('-')
				lastWasSeparator = true
			}
		}
	}

	cleaned := strings.Trim(builder.String(), "-")
	if cleaned == "" {
		return "youtube"
	}
	return cleaned
}

func formatDuration(seconds int) string {
	duration := time.Duration(seconds) * time.Second
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	secs := int(duration.Seconds()) % 60
	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, secs)
	}
	return fmt.Sprintf("%d:%02d", minutes, secs)
}

func optionsFromRequest(r *http.Request) conversionOptions {
	return conversionOptions{
		UsePlugins:   r.FormValue("use_plugins") == "1",
		Extension:    strings.TrimSpace(r.FormValue("extension")),
		Charset:      strings.TrimSpace(r.FormValue("charset")),
		KeepDataURIs: r.FormValue("keep_data_uris") == "1",
	}
}

func commandArgs(location string, options conversionOptions) []string {
	args := []string{}
	if options.UsePlugins {
		args = append(args, "--use-plugins")
	}
	if options.Extension != "" {
		args = append(args, "--extension", options.Extension)
	}
	if options.Charset != "" {
		args = append(args, "--charset", options.Charset)
	}
	if options.KeepDataURIs {
		args = append(args, "--keep-data-uris")
	}
	return append(args, location)
}

func markdownFileName(name string) string {
	base := filepath.Base(name)
	ext := filepath.Ext(base)
	if ext == "" {
		return base + ".md"
	}
	return base[:len(base)-len(ext)] + ".md"
}

func markdownURLFileName(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return "url.md"
	}

	name := strings.Trim(parsed.Host+parsed.EscapedPath(), "/")
	name = strings.NewReplacer("/", "-", "\\", "-", ":", "-").Replace(name)
	if name == "" {
		name = parsed.Host
	}
	return markdownFileName(name)
}

func uniqueFileName(name string, used map[string]int) string {
	used[name]++
	if used[name] == 1 {
		return name
	}

	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	return fmt.Sprintf("%s-%d%s", base, used[name], ext)
}
