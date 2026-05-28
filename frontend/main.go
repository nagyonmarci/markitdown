package main

import (
	"bytes"
	"context"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const maxUploadSize = 50 << 20

type pageData struct {
	FileName     string
	DownloadName string
	Markdown     string
	Error        string
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
      input[type="file"] {
        background: #14191f;
        border-color: #4b5563;
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
      <input name="file" type="file" required>
      <button type="submit">Convert</button>
    </form>

    {{if .Error}}<div class="message">{{.Error}}</div>{{end}}

    {{if .Markdown}}
    <section>
      <div class="result-head">
        <h2>{{.FileName}}</h2>
        <div class="actions">
          <button type="button" id="copy-button">Copy</button>
          <button type="button" id="download-button" data-filename="{{.DownloadName}}">Save</button>
        </div>
      </div>
      <textarea id="markdown-output" readonly>{{.Markdown}}</textarea>
    </section>
    {{end}}
  </main>
  <script>
    const output = document.getElementById("markdown-output");
    const copyButton = document.getElementById("copy-button");
    const downloadButton = document.getElementById("download-button");

    if (output && copyButton) {
      copyButton.addEventListener("click", async () => {
        await navigator.clipboard.writeText(output.value);
        copyButton.textContent = "Copied";
        setTimeout(() => copyButton.textContent = "Copy", 1200);
      });
    }

    if (output && downloadButton) {
      downloadButton.addEventListener("click", () => {
        const file = new Blob([output.value], { type: "text/markdown;charset=utf-8" });
        const link = document.createElement("a");
        link.href = URL.createObjectURL(file);
        link.download = downloadButton.dataset.filename || "document.md";
        link.click();
        URL.revokeObjectURL(link.href);
      });
    }
  </script>
</body>
</html>`))

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", render(pageData{}))
	mux.HandleFunc("POST /convert", convert)

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
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		render(pageData{Error: "Upload failed or file is larger than 50 MB."})(w, r)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		render(pageData{Error: "Choose a file to convert."})(w, r)
		return
	}
	defer file.Close()

	tmp, err := os.CreateTemp("", "markitdown-*"+filepath.Ext(header.Filename))
	if err != nil {
		render(pageData{Error: "Could not prepare the uploaded file."})(w, r)
		return
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if _, err := io.Copy(tmp, file); err != nil {
		render(pageData{Error: "Could not read the uploaded file."})(w, r)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()

	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "markitdown", tmp.Name())
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		message := "Conversion failed."
		if stderr.Len() > 0 {
			message += " " + stderr.String()
		}
		render(pageData{FileName: header.Filename, Error: message})(w, r)
		return
	}

	render(pageData{
		FileName:     header.Filename,
		DownloadName: markdownFileName(header.Filename),
		Markdown:     string(output),
	})(w, r)
}

func markdownFileName(name string) string {
	base := filepath.Base(name)
	ext := filepath.Ext(base)
	if ext == "" {
		return base + ".md"
	}
	return base[:len(base)-len(ext)] + ".md"
}
