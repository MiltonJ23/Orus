# File Extraction

The extractor adapter (`internal/adapters/extractor/`) handles PDF and EPUB file parsing for both metadata extraction and text content reading.

## LocalFileExtractor

Implements two port interfaces:
- `port.MetadataExtractor` — extracts title, author, page count, format
- `port.ContentReader` — extracts full text content split into readable pages

### Supported Formats

| Format | Metadata Library | Text Library |
|--------|-----------------|--------------|
| PDF | `ledongthuc/pdf` | `ledongthuc/pdf` |
| EPUB | `kapmahc/epub` | `kapmahc/epub` |

### Metadata Extraction

**PDF:** Title is derived from the filename (minus extension). Author defaults to `"Unknown"` as the PDF library does not reliably extract author metadata. Page count comes from `reader.NumPage()`.

**EPUB:** Title and author are extracted from OPF metadata. Page count is the number of spine items (chapters).

### Text Extraction

Text is extracted page-by-page and split into chunks of `linesPerChunk` (35 lines per "page" in the reader).

**PDF flow:**
1. Open file with `pdf.Open()`
2. Iterate over each page
3. Extract plain text with `page.GetPlainText()`
4. Split into lines, trim whitespace
5. Append page separator (`── Page N ──`)
6. Chunk into reader pages

**EPUB flow:**
1. Open file with `epub.Open()`
2. Iterate over spine items
3. Match each spine item to its manifest entry
4. Read raw HTML content
5. Strip HTML tags with `stripHTML()`
6. Clean HTML entities (`&amp;`, `&nbsp;`, etc.)
7. Prepend chapter separator (`═══ Chapitre N ═══`)
8. Chunk into reader pages

### HTML Stripping

The `stripHTML()` function performs basic HTML tag removal using a character-by-character state machine (inside/outside tag). It also replaces common HTML entities.

### Error Handling

- Unsupported formats return `ErrUnsupportedFileFormat`
- Corrupted or unreadable files return `ErrCorruptFile`
- If no text is extractable from a PDF, a placeholder message is returned
