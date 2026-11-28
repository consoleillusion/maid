// Builds the design as one PDF: scripts/compile-pdf.sh
// The chapter order comes from index.md, so this file does not change when
// chapters are added. The script passes mode=dark for the dark build and
// cover=yes when design/cover_image.png exists.
#import "@preview/cmarker:0.1.10"
#import "@preview/merman:0.6.0": mermaid-profile, show-mermaid-blocks

#let sans = ("Inter", "Noto Sans", "Source Sans 3", "IBM Plex Sans", "DejaVu Sans", "Liberation Sans", "Arial")
#let mono = ("JetBrains Mono", "DejaVu Sans Mono")

// The build script chooses the mode. Direct compilation stays light.
#let dark = sys.inputs.at("mode", default: "light") == "dark"
#let cover = sys.inputs.at("cover", default: "") == "yes"
#let ink = if dark { rgb("#eeeeee") } else { rgb("#000000") }
#let muted = if dark { rgb("#aaaaaa") } else { luma(90) }
#let stroke-colour = if dark { rgb("#555555") } else { luma(180) }
#let code-fill = if dark { rgb("#333333") } else { luma(245) }
#let link-colour = if dark { rgb("#8ab4f8") } else { rgb("#1f4e8c") }
#let page-fill = if dark { rgb("#222222") } else { white }

// Remove a leading YAML frontmatter block.
#let strip-frontmatter(s) = {
  if s.starts-with("---\n") or s.starts-with("---\r\n") {
    let parts = s.split(regex("\r?\n---\r?\n"))
    parts.slice(1).join("\n---\n")
  } else { s }
}

// Normalise a relative path such as foundation/../method/phases.md.
#let normalise(p) = {
  let out = ()
  for s in p.split("/") {
    if s == "" or s == "." { continue }
    if s == ".." { if out.len() > 0 { let _ = out.pop() } } else { out.push(s) }
  }
  out.join("/")
}

#let dir-of(p) = {
  let i = p.split("/")
  if i.len() < 2 { "" } else { i.slice(0, i.len() - 1).join("/") }
}

// Read index.md: the title, then the files in reading order. A part starts
// at each new part folder; its name is the folder name.
#let index-src = strip-frontmatter(read("index.md"))
#let title = {
  let m = index-src.match(regex("(?m)^# +(.+?)\\s*$"))
  if m == none { "Design" } else { m.captures.at(0) }
}
#let part-name(dir) = {
  let w = dir.replace("-", " ")
  upper(w.first()) + w.slice(1)
}
#let parts = {
  let out = ()
  for line in index-src.split("\n") {
    let c = line.match(regex("^\\s*(?:[-*+]|\\d+[.)])\\s+(?:\\[.+?\\]\\((.+?\\.md)\\)|(\\S+\\.md))\\s*$"))
    if c == none { continue }
    let p = normalise(if c.captures.at(0) != none { c.captures.at(0) } else { c.captures.at(1) })
    let dir = dir-of(p)
    if out.len() == 0 or out.last().dir != dir {
      out.push((dir: dir, name: if dir == "" { none } else { part-name(dir) }, chapters: ()))
    }
    let last = out.pop()
    last.chapters.push(p)
    out.push(last)
  }
  out
}

// The breadcrumb names the design, the current part and the current chapter in lowercase.
#let breadcrumb-part = state("breadcrumb-part", "")
#let breadcrumb-chapter = state("breadcrumb-chapter", "")
#let chapter-title(path) = {
  let m = read(path).match(regex("(?m)^# +(.+?)\\s*$"))
  if m == none { path } else { m.captures.at(0) }
}

// Links to other chapters become links inside the PDF; web links stay.
#let chapter-link(dir) = (dest, ..args) => {
  let body = args.pos().at(0, default: none)
  if type(dest) == str and not dest.contains("://") and not dest.starts-with("mailto:") and not dest.starts-with("#") {
    let (file, ..anchor) = dest.split("#")
    let target = if file == "" { none } else { normalise(if dir == "" { file } else { dir + "/" + file }) }
    let name = if anchor.len() > 0 { target + ":" + anchor.at(0) } else { target }
    context {
      if target != none and query(label(name)).len() > 0 { link(label(name), body) } else { body }
    }
  } else {
    link(dest, ..args)
  }
}

#let chapter-image(dir) = (src, ..args) => {
  let p = if src.starts-with("/") { src } else { normalise(if dir == "" { src } else { dir + "/" + src }) }
  image(p, ..args)
}

#set document(title: title)
#set text(font: sans, size: 10.5pt, lang: "en", fill: ink)
#show raw: set text(font: mono, size: 9pt)
#set par(justify: false, leading: 0.85em, spacing: 1.4em)
// Only parts have numbers, in roman numerals; chapters and sections have none.
#set heading(numbering: none)
#show heading: set text(weight: "semibold")
#show heading.where(level: 1): it => { pagebreak(weak: true); v(1.2em); text(26pt, it); v(1.4em) }
#show heading.where(level: 2): it => { v(1.6em); text(16pt, it); v(0.8em) }
#show heading.where(level: 3): it => { v(1.2em); text(12pt, it); v(0.5em) }
#set list(indent: 1.2em, body-indent: 0.6em, spacing: 0.5em)
#set enum(indent: 1.2em, body-indent: 0.6em, spacing: 0.5em)
#show table: set text(size: 9.5pt)
#set table(stroke: 0.5pt + stroke-colour, inset: 8pt)
#show table.cell.where(y: 0): set text(weight: "semibold")
#show link: set text(fill: link-colour)
// Mermaid fences become diagrams; other code keeps a grey block.
#let diagrams = mermaid-profile(theme-preset: "editor", theme-appearance: if dark { "dark" } else { "light" })
#let diagram = show-mermaid-blocks(profile: diagrams, document-context: true, width: 100%)
#show raw.where(block: true): it => {
  if it.lang == "mermaid" {
    diagram(it)
  } else {
    block(fill: code-fill, inset: 10pt, radius: 3pt, width: 100%, spacing: 1em, it)
  }
}

// Title page: the cover image fills the page width behind the title.
#page(margin: 0cm, fill: page-fill, numbering: none)[
  #if cover [
    #place(top + center, image("./cover_image.png", width: 100%, height: auto))
  ]
  #place(bottom + left, dx: 2.3cm, dy: -2.5cm)[
    #text(11pt, fill: muted, datetime.today().display("[day] [month repr:long] [year]"))
    #v(0.6em)
    #text(36pt, weight: "bold", fill: ink, title)
  ]
]

#set page(paper: "a4", margin: (x: 2.3cm, y: 2.5cm), fill: page-fill, numbering: "1",
  footer: context [#text(8.5pt, fill: muted)[
    #lower(title) #h(0.4em) · #h(0.4em) #lower(breadcrumb-part.get()) #h(0.4em) · #h(0.4em) #lower(breadcrumb-chapter.get())
    #h(1fr) #counter(page).display()
  ]])

#text(20pt, weight: "semibold")[Contents]
#v(0.8em)
#outline(title: none, depth: 2, indent: auto)

#for part in parts {
  breadcrumb-part.update(if part.name != none { part.name } else { "" })
  if part.name != none { heading(level: 1, numbering: "I", part.name) }
  for path in part.chapters {
    let dir = dir-of(path)
    breadcrumb-chapter.update(chapter-title(path))
    let (meta, body) = cmarker.render-with-metadata(
      read(path),
      metadata-block: "frontmatter-yaml",
      h1-level: if dir == "" { 1 } else { 2 },
      label-prefix: path + ":",
      scope: (link: chapter-link(dir), image: chapter-image(dir)),
    )
    [#metadata(path)#label(path)]
    body
    if meta != none and "status" in meta {
      align(right, text(8.5pt, fill: muted)[Status: #meta.status])
    }
  }
}
