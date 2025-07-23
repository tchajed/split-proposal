# split-proposal

[![CI](https://github.com/tchajed/split-proposal/actions/workflows/ci.yml/badge.svg)](https://github.com/tchajed/split-proposal/actions/workflows/ci.yml)

Split an NSF proposal into submission documents (summary, project description, references).

## Running

Add `\pdfbookmark` commands to demarcate your project description section to make this reliable. You'll need:

- `\pdfbookmark[0]{Project Description}{Project Description}` before the project description
- `\pdfbookmark[0]{References cited}{References cited}` before references
- (optional) `\pdfbookmark[0]{Data management plan}{Data management plan}`
- (optional) `\pdfbookmark[0]{Mentoring plan}{Mentoring plan}`

See [sample/main.tex](sample/main.tex) for a complete working example.

If you don't have bookmarks, it is assumed your summary is 1 page, description is 15 pages, and the rest of the document is references.

To split the combined proposal file run:

```sh
go run github.com/tchajed/split-proposal@latest -file main.pdf
```
