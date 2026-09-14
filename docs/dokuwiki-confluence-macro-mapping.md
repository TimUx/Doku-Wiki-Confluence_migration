# DokuWiki → Confluence syntax and macro mapping

This migration targets Confluence Data Center and uses Confluence storage format for exported pages.

## Core syntax

| DokuWiki | Confluence representation | Policy |
|---|---|---|
| `====== Heading ======` | `<h1>...` etc. | Native heading |
| `**bold**` | `<strong>` | Native |
| `//italic//` | `<em>` | Native |
| `__underline__` | `<u>` | Native HTML formatting |
| `''monospace''` | `<code>` | Native inline code |
| `~~deleted~~` | `<del>` | Native HTML formatting |
| `_{sub}` / `^{super}` | `<sub>` / `<sup>` | Native HTML formatting |
| `\\` | `<br/>` | Native line break |
| `----` | `<hr/>` | Native separator |
| `((footnote))` | footnote-compatible inline marker | Preserved |
| `[[url|label]]` external | `<a href="...">` | Real external link |
| `[[page|label]]` internal | explicit DokuWiki placeholder | Target-independent migration |
| `[[#section|label]]` | `#anchor` link | Same-page link |
| `{{image}}` | `<ac:image><ri:attachment .../>` | Exact attachment filename |
| tables | Confluence table | Colspan preserved |
| nested lists | nested list structure | Preserve indentation |
| `~~NOCACHE~~` | ignored control directive | No Confluence equivalent needed |
| `~~NOTOC~~` | suppress generated TOC | Preserves DokuWiki intent |

## DokuWiki Include plugin

| DokuWiki syntax | Confluence macro | Migration policy |
|---|---|---|
| `{{page>id}}` | **Include Page** (`include`) | **Placeholder by default** |
| `{{section>id#section}}` | **Excerpt Include** (`excerpt-include`) only when the source semantics match an excerpt | Placeholder by default |
| `{{namespace>...}}` | Page Tree / Content Report depending on intent | Placeholder; cannot infer the desired report safely |
| `{{tagtopic>...}}` | Content by Label / Content Report Table | Placeholder; requires label mapping |

The project deliberately does not silently turn DokuWiki includes into live Confluence dependencies. This prevents a migration from becoming dependent on pages that may not exist yet.

## Wrap plugin

| Wrap class | Confluence macro / representation |
|---|---|
| `info` | **Info** macro |
| `tip` | **Tip** macro |
| `warning`, `caution` | **Warning** macro |
| `important` | **Note** macro |
| `notice`, `help` | **Note** macro |
| `danger`, `safety`, `alert` | **Warning** macro |
| `box` | **Panel** macro |
| `spoiler` | **Expand** macro |
| `group` + `column` | **Section** + **Column** macros |
| `tabs` | **UI Tabs** where the source contains compatible tab structure; otherwise preserve as container |
| `button` | **UI Button** where the wrapped content is a link |
| `pagebreak` | **Seitenumbruch - PDF** where available |
| `clear` | structural no-op |
| `noprint` / `onlyprint` | preserve as metadata/placeholder; no safe generic Confluence equivalent |
| `hide` | **Expand** is the closest safe semantic equivalent; never treat it as security |
| `hi`, `lo`, `em` | inline emphasis / formatting |
| `indent`, `outdent` | layout/container representation |

Unknown Wrap classes are never discarded; the original classes remain available on the generated container.

## Numbered Headings plugin

The Numbered Headings plugin changes heading numbering rather than providing a Confluence content macro. The migration therefore keeps the resulting numbering in the heading text. Confluence's TOC macro can additionally provide outline numbering when appropriate, but that is not used to replace the source heading numbering.

## Common Confluence macros available in the target instance

The migration should prefer these native macros when the DokuWiki semantics match:

- `toc` — Table of Contents / **Inhalt**
- `anchor` — explicit anchor targets
- `info`, `tip`, `note`, `warning` — semantic callout boxes
- `panel` — generic styled box
- `expand` — collapsible/spoiler content
- `include` — full-page include
- `excerpt` / `excerpt-include` — reusable page excerpts
- `section` / `column` — multi-column layouts
- `children` / `pagetree` — page hierarchy navigation
- `attachments` — attachment lists
- `code` — syntax-highlighted code
- `status` — status lozenge

Third-party macros from the target instance are only used when the source syntax clearly expresses the same semantics. The migration does not guess a third-party macro merely because it exists.
