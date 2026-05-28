# MarkItDown

[![PyPI](https://img.shields.io/pypi/v/markitdown.svg)](https://pypi.org/project/markitdown/)
![PyPI - Downloads](https://img.shields.io/pypi/dd/markitdown)
[![Built by AutoGen Team](https://img.shields.io/badge/Built%20by-AutoGen%20Team-blue)](https://github.com/microsoft/autogen)
[![CI](https://github.com/nagyonmarci/markitdown/actions/workflows/ci.yml/badge.svg)](https://github.com/nagyonmarci/markitdown/actions/workflows/ci.yml)
[![Release](https://github.com/nagyonmarci/markitdown/actions/workflows/release.yml/badge.svg)](https://github.com/nagyonmarci/markitdown/actions/workflows/release.yml)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/nagyonmarci/markitdown/badge)](https://securityscorecards.dev/viewer/?uri=github.com/nagyonmarci/markitdown)
[![Container](https://img.shields.io/badge/GHCR-ghcr.io%2Fnagyonmarci%2Fmarkitdown-blue?logo=docker)](https://github.com/nagyonmarci/markitdown/pkgs/container/markitdown)

> [!IMPORTANT]
> MarkItDown performs I/O with the privileges of the current process. Like open() or requests.get(), it will access resources that the process itself can access. Sanitize your inputs in untrusted environments, and call the narrowest `convert_*` function needed for your use case (e.g., `convert_stream()`, or `convert_local()`). See the [Security Considerations](#security-considerations) section of the documentation for more information.

MarkItDown is a lightweight Python utility for converting various files to Markdown for use with LLMs and related text analysis pipelines. To this end, it is most comparable to [textract](https://github.com/deanmalmgren/textract), but with a focus on preserving important document structure and content as Markdown (including: headings, lists, tables, links, etc.) While the output is often reasonably presentable and human-friendly, it is meant to be consumed by text analysis tools -- and may not be the best option for high-fidelity document conversions for human consumption.

MarkItDown currently supports the conversion from:

- PDF
- PowerPoint
- Word
- Excel
- Images (EXIF metadata and OCR)
- Audio (EXIF metadata and speech transcription)
- HTML
- Text-based formats (CSV, JSON, XML)
- ZIP files (iterates over contents)
- Youtube URLs
- EPubs
- ... and more!

## Why Markdown?

Markdown is extremely close to plain text, with minimal markup or formatting, but still
provides a way to represent important document structure. Mainstream LLMs, such as
OpenAI's GPT-4o, natively "_speak_" Markdown, and often incorporate Markdown into their
responses unprompted. This suggests that they have been trained on vast amounts of
Markdown-formatted text, and understand it well. As a side benefit, Markdown conventions
are also highly token-efficient.

## Prerequisites
MarkItDown requires Python 3.10 or higher. It is recommended to use a virtual environment to avoid dependency conflicts.

With the standard Python installation, you can create and activate a virtual environment using the following commands:

```bash
python -m venv .venv
source .venv/bin/activate
```

If using `uv`, you can create a virtual environment with:

```bash
uv venv --python=3.12 .venv
source .venv/bin/activate
# NOTE: Be sure to use 'uv pip install' rather than just 'pip install' to install packages in this virtual environment
```

If you are using Anaconda, you can create a virtual environment with:

```bash
conda create -n markitdown python=3.12
conda activate markitdown
```

## Installation

To install MarkItDown, use pip: `pip install 'markitdown[all]'`. Alternatively, you can install it from the source:

```bash
git clone git@github.com:microsoft/markitdown.git
cd markitdown
pip install -e 'packages/markitdown[all]'
```

## Usage

### Command-Line

```bash
markitdown path-to-file.pdf > document.md
```

Or use `-o` to specify the output file:

```bash
markitdown path-to-file.pdf -o document.md
```

You can also pipe content:

```bash
cat path-to-file.pdf | markitdown
```

### Optional Dependencies
MarkItDown has optional dependencies for activating various file formats. Earlier in this document, we installed all optional dependencies with the `[all]` option. However, you can also install them individually for more control. For example:

```bash
pip install 'markitdown[pdf, docx, pptx]'
```

will install only the dependencies for PDF, DOCX, and PPTX files.

At the moment, the following optional dependencies are available:

* `[all]` Installs all optional dependencies
* `[pptx]` Installs dependencies for PowerPoint files
* `[docx]` Installs dependencies for Word files
* `[xlsx]` Installs dependencies for Excel files
* `[xls]` Installs dependencies for older Excel files
* `[pdf]` Installs dependencies for PDF files
* `[outlook]` Installs dependencies for Outlook messages
* `[az-doc-intel]` Installs dependencies for Azure Document Intelligence
* `[az-content-understanding]` Installs dependencies for Azure Content Understanding
* `[audio-transcription]` Installs dependencies for audio transcription of wav and mp3 files
* `[youtube-transcription]` Installs dependencies for fetching YouTube video transcription

### Plugins

MarkItDown also supports 3rd-party plugins. Plugins are disabled by default. To list installed plugins:

```bash
markitdown --list-plugins
```

To enable plugins use:

```bash
markitdown --use-plugins path-to-file.pdf
```

To find available plugins, search GitHub for the hashtag `#markitdown-plugin`. To develop a plugin, see `packages/markitdown-sample-plugin`.

#### markitdown-ocr Plugin

The `markitdown-ocr` plugin adds OCR support to PDF, DOCX, PPTX, and XLSX converters, extracting text from embedded images using LLM Vision — the same `llm_client` / `llm_model` pattern that MarkItDown already uses for image descriptions. No new ML libraries or binary dependencies required.

**Installation:**

```bash
pip install markitdown-ocr
pip install openai  # or any OpenAI-compatible client
```

**Usage:**

Pass the same `llm_client` and `llm_model` you would use for image descriptions:

```python
from markitdown import MarkItDown
from openai import OpenAI

md = MarkItDown(
    enable_plugins=True,
    llm_client=OpenAI(),
    llm_model="gpt-4o",
)
result = md.convert("document_with_images.pdf")
print(result.text_content)
```

If no `llm_client` is provided the plugin still loads, but OCR is silently skipped and the standard built-in converter is used instead.

See [`packages/markitdown-ocr/README.md`](packages/markitdown-ocr/README.md) for detailed documentation.

### Azure Content Understanding

[Azure Content Understanding](https://learn.microsoft.com/azure/ai-services/content-understanding/) provides higher-quality conversion with structured field extraction (YAML front matter), multi-modal support (documents, images, audio, video), and configurable analyzers.

Install: `pip install 'markitdown[az-content-understanding]'`

#### When to use Content Understanding

Content Understanding is ideal when you need capabilities beyond what built-in or Document Intelligence converters provide:

- **Audio and video files** — CU is the only option for video, and the higher-quality cloud option for audio. Built-in converters have no video support and only basic audio transcription.
- **Structured field extraction** — [Prebuilt](https://learn.microsoft.com/azure/ai-services/content-understanding/concepts/prebuilt-analyzers) or [custom-built](https://learn.microsoft.com/azure/ai-services/content-understanding/how-to/customize-analyzer-content-understanding-studio?tabs=portal) analyzers extract domain-specific fields (invoice amounts, receipt dates, contract clauses) serialized as YAML front matter. Neither built-in nor Doc Intel integration exposes fields.
- **Higher-quality document extraction** — Cloud-based layout analysis and OCR for scanned PDFs, complex tables, and multi-page documents.
- **Single API for all modalities** — One `cu_endpoint` handles documents, images, audio, and video with automatic analyzer routing.

| Capability | Built-in converters | Azure Document Intelligence | Azure Content Understanding |
|------------|---------------------|-----------------------------|-----------------------------|
| Document conversion | Offline, format-specific extraction | Cloud layout extraction | Cloud multimodal extraction |
| Structured fields | Not available | Not exposed by this integration | YAML front matter from analyzer fields |
| Custom analyzers | Not available | Not configurable in this integration | Supported with `cu_analyzer_id` |
| Audio and video | Basic audio, no video | Not supported | Audio and video analyzers |
| Cost | Local compute only | Billable Azure API calls | Billable Azure API calls |

**CLI:**

```bash
markitdown path-to-file.pdf --use-cu --cu-endpoint "<content_understanding_endpoint>"
```

**Python API:**

```python
from markitdown import MarkItDown

# Zero-config — auto-selects analyzer per file type
md = MarkItDown(cu_endpoint="<content_understanding_endpoint>")
result = md.convert("report.pdf")   # documents → prebuilt-documentSearch
result = md.convert("meeting.mp4")  # video → prebuilt-videoSearch
result = md.convert("call.wav")     # audio → prebuilt-audioSearch
print(result.markdown)
```

**With a custom analyzer** (for domain-specific field extraction):

```python
md = MarkItDown(
    cu_endpoint="<content_understanding_endpoint>",
    cu_analyzer_id="my-invoice-analyzer",
)
result = md.convert("invoice.pdf")
print(result.markdown)
# Output includes YAML front matter with extracted fields:
# ---
# contentType: document
# fields:
#   VendorName: CONTOSO LTD.
#   InvoiceDate: '2019-11-15'
# ---
# <!-- page 1 -->
# ...
```

When `cu_analyzer_id` is set, the converter automatically scopes it to compatible file types based on the analyzer's modality. Incompatible types (e.g., audio files with a document analyzer) auto-route to default prebuilt analyzers.

**Cost note:** Each `convert()` call for a CU-routed format is a billable Azure API call. Use `cu_file_types` to restrict which formats route to CU:

```python
from markitdown.converters import ContentUnderstandingFileType

md = MarkItDown(
    cu_endpoint="<content_understanding_endpoint>",
    cu_file_types=[ContentUnderstandingFileType.PDF],  # only PDFs use CU
)
```

More information about Azure Content Understanding can be found [here](https://learn.microsoft.com/azure/ai-services/content-understanding/).

### Azure Document Intelligence

To use Microsoft Document Intelligence for conversion:

```bash
markitdown path-to-file.pdf -o document.md -d -e "<document_intelligence_endpoint>"
```

More information about how to set up an Azure Document Intelligence Resource can be found [here](https://learn.microsoft.com/en-us/azure/ai-services/document-intelligence/how-to-guides/create-document-intelligence-resource?view=doc-intel-4.0.0)

### Python API

Basic usage in Python:

```python
from markitdown import MarkItDown

md = MarkItDown(enable_plugins=False) # Set to True to enable plugins
result = md.convert("test.xlsx")
print(result.text_content)
```

Document Intelligence conversion in Python:

```python
from markitdown import MarkItDown

md = MarkItDown(docintel_endpoint="<document_intelligence_endpoint>")
result = md.convert("test.pdf")
print(result.text_content)
```

To use Large Language Models for image descriptions (currently only for pptx and image files), provide `llm_client` and `llm_model`:

```python
from markitdown import MarkItDown
from openai import OpenAI

client = OpenAI()
md = MarkItDown(llm_client=client, llm_model="gpt-4o", llm_prompt="optional custom prompt")
result = md.convert("example.jpg")
print(result.text_content)
```

### Docker

```sh
docker build -t markitdown:latest .
docker run --rm -i markitdown:latest < ~/your-file.pdf > output.md
```

## Web Frontend

The repository includes a lightweight web UI that exposes the full MarkItDown CLI over HTTP. It is written in Go and compiles to a self-contained static binary with zero runtime dependencies.

### Running the frontend

```sh
docker compose up
```

The UI is available at `http://localhost:8088`. Internally the container runs on port `8080`.

To build and run the frontend directly without Docker:

```sh
cd frontend
go build -o markitdown-frontend .
./markitdown-frontend
```

### Features

**File conversion**
- Upload one or more files (PDF, DOCX, PPTX, XLSX, images, audio, …) up to 50 MB total
- Results are displayed inline in editable text areas
- Per-result **Copy** and **Save** buttons, plus a **Copy all** button for batch workflows
- **Download ZIP** — packages all converted Markdown files into a single archive in one click

**URL conversion**
- Paste any URL to convert a remote HTML page, document, or feed directly
- **YouTube support** — YouTube URLs are handled via a dedicated path: `yt-dlp` fetches video metadata (title, channel, duration, description) and English subtitles, which are assembled into a structured Markdown document with a transcript section

**Advanced options** (available for both file and URL modes)
- Extension hint — override the file type detected by MarkItDown
- Charset — specify the input encoding
- Use plugins — enable installed third-party MarkItDown plugins
- Keep data URIs — retain embedded base64 images in the output

**UX details**
- Fully responsive layout, works on mobile
- System dark mode support via `prefers-color-scheme`
- Zero JavaScript dependencies — vanilla JS, no bundler, no framework

### Architecture

The frontend is a thin HTTP adapter over the `markitdown` CLI binary. It spawns a subprocess per request (`exec.CommandContext` with a 2-minute timeout), streams the output back to the browser, and handles errors gracefully. This design keeps the Go code simple and ensures the frontend stays in sync with the CLI automatically — no duplicated conversion logic.

The multi-stage `frontend/Dockerfile` compiles the Go binary in a `golang:1.22` builder stage, then copies it into the same Python runtime image used by the main `Dockerfile`, ensuring the `markitdown` CLI and all its dependencies are available at runtime.

---

## DevSecOps Pipeline

The repository ships a hardened, multi-stage CI/CD pipeline designed around **shift-left security**: every potential vulnerability is caught as early as possible — ideally before code lands on the main branch.

### Architecture overview

Three GitHub Actions workflows cover the full software delivery lifecycle:

| Workflow | Trigger | Purpose |
|---|---|---|
| [`ci.yml`](.github/workflows/ci.yml) | Every push to `main` + every PR | Quality gates and security scanning |
| [`release.yml`](.github/workflows/release.yml) | Push to `main` or version tag `v*.*.*` | Docker image build, signing, and publishing |
| [`scorecard.yml`](.github/workflows/scorecard.yml) | Weekly + every push to `main` | OSSF supply chain health score |

### Continuous Integration (`ci.yml`)

The CI workflow uses a **parallel fan-out** design — jobs are independent and run concurrently, minimising wall-clock time while maximising coverage.

#### Code quality

| Job | Tool | What it checks |
|---|---|---|
| `lint` | [pre-commit](https://pre-commit.com/) + [Black](https://github.com/psf/black) | Consistent code formatting across the entire codebase |
| `test` | [Hatch](https://hatch.pypa.io/) + pytest | Functional correctness on Python 3.10, 3.11, and 3.12 in parallel |

The test matrix runs each Python version as a separate job, so a version-specific regression is caught precisely instead of masking a passing version.

#### Security scanning

Five independent security lenses run in parallel on every push:

**1. Secret detection — Gitleaks**

[Gitleaks](https://github.com/gitleaks/gitleaks) performs a **full-history scan** (not just the HEAD diff) to detect accidentally committed credentials, API keys, and tokens before they can reach an attacker. Scanning the full history closes the gap where a secret committed and later deleted would otherwise go unnoticed.

**2. Static Application Security Testing — CodeQL + Semgrep**

Two independent SAST engines run in parallel:

- **[CodeQL](https://codeql.github.com/)** — GitHub's semantic analysis engine builds a code property graph and queries it for known vulnerability patterns. Both Python and Go are analysed in separate matrix jobs, so a finding in either language is surfaced independently.
- **[Semgrep](https://semgrep.dev/)** — rule-based scanner running the `p/python` and `p/owasp-top-ten` rulesets, covering injection flaws, insecure deserialization, broken access control, and the full OWASP Top 10. Findings are published to the Security tab as SARIF.

Running two independent SAST tools reduces false-negative rates — each engine uses different analysis techniques and may catch what the other misses.

**3. Container hardening — hadolint**

Both `Dockerfile` (Python runtime image) and `frontend/Dockerfile` (multi-stage Go build) are linted with [hadolint](https://github.com/hadolint/hadolint) against Dockerfile best practices: unnecessary packages, pinned base images, layer hygiene, and privilege escalation risks. The failure threshold is set to `error` so informational suggestions don't block CI while genuine misconfigurations do.

**4. Infrastructure-as-Code scanning — Checkov**

[Checkov](https://www.checkov.io/) (Prisma Cloud) scans `compose.yaml` for Docker Compose misconfigurations: containers running as root, missing health checks, exposed secrets, insecure bind mounts, and writable container filesystems. Results are uploaded as SARIF to the Security tab.

**5. Container image vulnerability scan — Trivy**

The Docker image is built from source and scanned with [Trivy](https://github.com/aquasecurity/trivy) for CVEs at `HIGH` and `CRITICAL` severity. The scan runs against the **built artifact**, not just the Dockerfile, which catches vulnerabilities introduced by transitive Python or system dependencies that static analysis cannot see.

#### PR gates

- **Dependency review** — flags newly introduced dependencies with known CVEs or incompatible licenses on every pull request, before merge.
- **Security summary** — a bot posts a **sticky comment** on every PR with a consolidated table of all security job results. The comment is updated in-place on each new push, so reviewers always see the current state without scrolling through history.

### Release & Supply Chain Security (`release.yml`)

The release pipeline automates the full supply chain from source commit to signed, attested, published artifact.

#### Multi-architecture builds

Images are built for both `linux/amd64` and `linux/arm64` using Docker Buildx and QEMU emulation. A single OCI image manifest covers both platforms — no user-side configuration required.

#### GitHub Container Registry (GHCR)

Images are published to `ghcr.io/nagyonmarci/markitdown` with a semantic tag strategy driven by [`docker/metadata-action`](https://github.com/docker/metadata-action):

| Tag | When | Meaning |
|---|---|---|
| `latest` | Every `main` push | Most recent main build |
| `<branch>` | Every branch push | Per-branch image for integration testing |
| `<major>.<minor>` | Git version tag | Floating minor version |
| `<major>.<minor>.<patch>` | Git version tag | Immutable patch release |

#### Keyless image signing — Cosign + Sigstore

Every image is signed using [Cosign](https://github.com/sigstore/cosign) in **keyless mode** via [Sigstore's](https://www.sigstore.dev/) Fulcio CA and Rekor transparency log. No long-lived private keys are stored anywhere in the repository or in CI secrets. The signature binds the image digest to the GitHub Actions OIDC identity, making it cryptographically verifiable that:

1. The image was produced by this specific workflow run.
2. It corresponds to a specific commit in this repository.
3. It has not been tampered with after publication.

Verification:
```sh
cosign verify ghcr.io/nagyonmarci/markitdown:latest \
  --certificate-identity-regexp "https://github.com/nagyonmarci/markitdown/.*" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com"
```

#### Software Bill of Materials (SBOM)

A [CycloneDX JSON SBOM](https://cyclonedx.org/) is generated for every released image using [`anchore/sbom-action`](https://github.com/anchore/sbom-action). The SBOM is attested and pushed to the registry alongside the image as an OCI referrer, making it queryable:

```sh
cosign verify-attestation ghcr.io/nagyonmarci/markitdown:latest \
  --type cyclonedx \
  --certificate-identity-regexp "https://github.com/nagyonmarci/markitdown/.*" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com"
```

This fulfils emerging regulatory requirements (e.g. US Executive Order 14028) that mandate a machine-readable inventory of all software components in distributed artifacts.

#### Post-publish vulnerability scan

After the image is pushed, Trivy scans the **published registry image** (not the local build cache) for `HIGH` and `CRITICAL` CVEs. This catches any discrepancy between the local build environment and what actually landed in the registry. Results are uploaded as SARIF to the Security tab.

#### GitHub Releases

On version tags (`v*.*.*`), a GitHub Release is created automatically with machine-generated release notes aggregated from merged PR titles and labels — no manual changelog entry required.

### OSSF Scorecard (`scorecard.yml`)

The pipeline participates in the [OpenSSF Scorecard](https://securityscorecards.dev/) program, which grades the repository's supply chain hygiene across 18 automated checks including:

- Branch protection and review requirements
- CI/CD pipeline security and pinned actions
- Dependency update automation
- Vulnerability disclosure policy
- Binary artifact and dangerous workflow checks

The score runs weekly and on every push to `main`, with SARIF results integrated into GitHub's Security tab. The public badge at the top of this README reflects the current live score.

### Automated dependency updates — Dependabot

[Dependabot](https://docs.github.com/en/code-security/dependabot) monitors four package ecosystems weekly and opens PRs automatically:

| Ecosystem | Directory | Covers |
|---|---|---|
| `github-actions` | `/` | Action version pins in all workflows |
| `pip` | `/packages/markitdown` | Python runtime dependencies |
| `gomod` | `/frontend` | Go module dependencies |
| `docker` | `/` | Base image tags in both Dockerfiles |

Dependabot PRs are validated by the full CI pipeline before any human reviews them, so a failing security scan on a dependency update is surfaced immediately.

---

## Contributing

This project welcomes contributions and suggestions. Most contributions require you to agree to a
Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us
the rights to use your contribution. For details, visit https://cla.opensource.microsoft.com.

When you submit a pull request, a CLA bot will automatically determine whether you need to provide
a CLA and decorate the PR appropriately (e.g., status check, comment). Simply follow the instructions
provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.

### How to Contribute

You can help by looking at issues or helping review PRs. Any issue or PR is welcome, but we have also marked some as 'open for contribution' and 'open for reviewing' to help facilitate community contributions. These are of course just suggestions and you are welcome to contribute in any way you like.

<div align="center">

|            | All                                                          | Especially Needs Help from Community                                                                                                      |
| ---------- | ------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------- |
| **Issues** | [All Issues](https://github.com/microsoft/markitdown/issues) | [Issues open for contribution](https://github.com/microsoft/markitdown/issues?q=is%3Aissue+is%3Aopen+label%3A%22open+for+contribution%22) |
| **PRs**    | [All PRs](https://github.com/microsoft/markitdown/pulls)     | [PRs open for reviewing](https://github.com/microsoft/markitdown/pulls?q=is%3Apr+is%3Aopen+label%3A%22open+for+reviewing%22)              |

</div>

### Running Tests and Checks

- Navigate to the MarkItDown package:

  ```sh
  cd packages/markitdown
  ```

- Install `hatch` in your environment and run tests:

  ```sh
  pip install hatch  # Other ways of installing hatch: https://hatch.pypa.io/dev/install/
  hatch shell
  hatch test
  ```

  (Alternative) Use the Devcontainer which has all the dependencies installed:

  ```sh
  # Reopen the project in Devcontainer and run:
  hatch test
  ```

- Run pre-commit checks before submitting a PR: `pre-commit run --all-files`

### Security Considerations

MarkItDown performs I/O with the privileges of the current process. Like `open()` or `requests.get()`, it will access resources that the process itself can access. 

**Sanitize your inputs:** Do not pass untrusted input directly to MarkItDown. If any part of the input may be controlled by an untrusted user or system, such as in hosted or server-side applications, it must be validated and restricted before calling MarkItDown. Depending on your environment, this may include restricting file paths, limiting URI schemes and network destinations, and blocking access to private, loopback, link-local, or metadata-service addresses. 

**Call only the conversion method you need:** Prefer the narrowest conversion API that fits your use case. MarkItDown's `convert()` method is intentionally permissive and can handle local files, remote URIs, and byte streams. If your application only needs to read local files, call `convert_local()` instead. If you need more control over URI fetching, call `requests.get()` yourself and pass the response object to `convert_response()`. For maximum control, open a stream to the input you want converted and call `convert_stream()`.

### Contributing 3rd-party Plugins

You can also contribute by creating and sharing 3rd party plugins. See `packages/markitdown-sample-plugin` for more details.

## Trademarks

This project may contain trademarks or logos for projects, products, or services. Authorized use of Microsoft
trademarks or logos is subject to and must follow
[Microsoft's Trademark & Brand Guidelines](https://www.microsoft.com/en-us/legal/intellectualproperty/trademarks/usage/general).
Use of Microsoft trademarks or logos in modified versions of this project must not cause confusion or imply Microsoft sponsorship.
Any use of third-party trademarks or logos are subject to those third-party's policies.
