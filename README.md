# split-proposal

[![CI](https://github.com/tchajed/split-proposal/actions/workflows/ci.yml/badge.svg)](https://github.com/tchajed/split-proposal/actions/workflows/ci.yml)

Split an NSF proposal into submission documents.

Add `\pdfbookmark` commands to demarcate your project description section to make this reliable:

```tex
\documentclass{article}

% preamble

\begin{document}

\input{summary}

\pdfbookmark[0]{Project Description}{Project Description}

% your project description parts
\input{intro}
\input{research}

\newpage
\pdfbookmark[0]{References cited}{References cited}
% bibliography commands
\bibliography{refs}
\end{document}
```

If you don't have bookmarks, it is assumed your summary is 1 page, description is 15 pages, and the rest of the document is references.

Then run on your compiled PDF:

```sh
go run github.com/tchajed/split-proposal@latest - -file main.pdf
```
