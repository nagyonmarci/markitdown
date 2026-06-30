import re
import bs4
from typing import Any, BinaryIO, List, Optional

from .._base_converter import DocumentConverter, DocumentConverterResult
from .._stream_info import StreamInfo
from ._markdownify import _CustomMarkdownify

ACCEPTED_MIME_TYPE_PREFIXES = [
    "text/html",
    "application/xhtml",
]

ACCEPTED_FILE_EXTENSIONS = [
    ".html",
    ".htm",
]

# www.linkedin.com, hu.linkedin.com, etc. -- always a /jobs/view/ posting
_LINKEDIN_JOB_URL_RE = re.compile(
    r"^https?://([a-zA-Z0-9-]+\.)?linkedin\.com/jobs/view/", re.IGNORECASE
)


def _clean_text(node: Optional[bs4.Tag]) -> Optional[str]:
    """Collapse whitespace in an element's text, or None if absent/empty."""
    if node is None:
        return None
    text = re.sub(r"\s+", " ", node.get_text(" ", strip=True)).strip()
    return text or None


class LinkedInJobConverter(DocumentConverter):
    """Handle LinkedIn job postings, keeping only the job content.

    LinkedIn's public ("guest") job view pages wrap the actual posting in a
    sea of navigation, sign-in prompts, cookie banners, "People also viewed"
    and "Similar jobs" widgets, and a footer with hundreds of links. The
    generic HtmlConverter dumps all of that into the Markdown. This converter
    extracts just the title, company, location, posting metadata, job criteria
    and description, and drops the rest.
    """

    def accepts(
        self,
        file_stream: BinaryIO,
        stream_info: StreamInfo,
        **kwargs: Any,  # Options to pass to the converter
    ) -> bool:
        """
        Make sure we're dealing with HTML content *from* a LinkedIn job view.
        """

        url = stream_info.url or ""
        mimetype = (stream_info.mimetype or "").lower()
        extension = (stream_info.extension or "").lower()

        if not _LINKEDIN_JOB_URL_RE.search(url):
            # Not a LinkedIn job URL
            return False

        if extension in ACCEPTED_FILE_EXTENSIONS:
            return True

        for prefix in ACCEPTED_MIME_TYPE_PREFIXES:
            if mimetype.startswith(prefix):
                return True

        # Not HTML content
        return False

    def convert(
        self,
        file_stream: BinaryIO,
        stream_info: StreamInfo,
        **kwargs: Any,  # Options to pass to the converter
    ) -> Optional[DocumentConverterResult]:
        # Parse the stream
        encoding = "utf-8" if stream_info.charset is None else stream_info.charset
        soup = bs4.BeautifulSoup(file_stream, "html.parser", from_encoding=encoding)

        # Remove javascript and style blocks
        for script in soup(["script", "style"]):
            script.extract()

        # Locate the core fields. Title and description are the anchors -- if we
        # can't find them the markup is not the guest job page we expect (e.g. a
        # login wall or a redesigned DOM), so we bail out and let the generic
        # HtmlConverter handle it (returning None falls through in the dispatcher).
        title = _clean_text(
            soup.select_one("h1.top-card-layout__title")
            or soup.select_one(".topcard__title")
        )
        description_elm = soup.select_one(
            ".show-more-less-html__markup"
        ) or soup.select_one(".description__text")
        if title is None or description_elm is None:
            return None

        company = _clean_text(
            soup.select_one("a.topcard__org-name-link")
            or soup.select_one(".topcard__flavor")
        )
        location = _clean_text(soup.select_one(".topcard__flavor--bullet"))
        posted = _clean_text(soup.select_one(".posted-time-ago__text"))
        applicants = _clean_text(soup.select_one(".num-applicants__caption"))

        # Build the header
        parts: List[str] = [f"# {title}"]

        meta_lines: List[str] = []
        if company:
            meta_lines.append(f"**Company:** {company}")
        if location:
            meta_lines.append(f"**Location:** {location}")
        if posted:
            meta_lines.append(f"**Posted:** {posted}")
        if applicants:
            meta_lines.append(f"**Applicants:** {applicants}")
        if meta_lines:
            parts.append("\n".join(meta_lines))

        # Job criteria (seniority level, employment type, job function, industries)
        criteria_lines: List[str] = []
        for item in soup.select("ul.description__job-criteria-list li"):
            header = _clean_text(
                item.select_one(".description__job-criteria-subheader")
            )
            value = _clean_text(item.select_one(".description__job-criteria-text"))
            if header and value:
                criteria_lines.append(f"- **{header}:** {value}")
        if criteria_lines:
            parts.append("## Job criteria\n\n" + "\n".join(criteria_lines))

        # Requirements added by the job poster (optional structured section)
        req_heading = soup.find(
            ["h2", "h3"],
            string=re.compile(r"requirements added by the job poster", re.I),
        )
        if req_heading:
            req_list = req_heading.find_next("ul")
            if req_list:
                items = [_clean_text(li) for li in req_list.find_all("li")]
                items = [i for i in items if i]
                if items:
                    parts.append(
                        "## Requirements added by the job poster\n\n"
                        + "\n".join(f"- {i}" for i in items)
                    )

        # Description body -- markdownify to preserve lists, bold, links, etc.
        description_md = (
            _CustomMarkdownify(**kwargs).convert_soup(description_elm).strip()
        )
        if description_md:
            parts.append("## Description\n\n" + description_md)

        markdown = "\n\n".join(parts).strip() + "\n"

        return DocumentConverterResult(
            markdown=markdown,
            title=title,
        )
