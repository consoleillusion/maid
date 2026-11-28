// Builds the design as one PDF: typst compile design/index.typ
// The chapter order comes from index.md, so this file does not change when
// chapters are added. maid writes it; edit the style settings below freely.
#import "@preview/cmarker:0.1.10"

#let sans = ("Inter", "Noto Sans", "Source Sans 3", "IBM Plex Sans", "DejaVu Sans", "Liberation Sans", "Arial")
#let mono = ("JetBrains Mono", "DejaVu Sans Mono")

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
  i.slice(0, i.len() - 1).join("/")
}

// Read index.md: the title, then "## Part" headings and numbered chapter links.
#let index-src = strip-frontmatter(read("index.md"))
#let title = {
  let m = index-src.match(regex("(?m)^# +(.+?)\s*$"))
  if m == none { "Design" } else { m.captures.at(0) }
}
#let intro = {
  let after = index-src.split(regex("(?m)^# +.+$")).at(1, default: "")
  after.split(regex("(?m)^## ")).at(0).trim()
}
#let parts = {
  let out = ()
  for line in index-src.split("\n") {
    let h = line.match(regex("^## +(.+?)\s*$"))
    let c = line.match(regex("^\s*\d+[.)]\s+\[(.+?)\]\((.+?\.md)\)"))
    if h != none {
      out.push((name: h.captures.at(0), chapters: ()))
    } else if c != none {
      if out.len() == 0 { out.push((name: none, chapters: ())) }
      let last = out.pop()
      last.chapters.push((title: c.captures.at(0), path: normalise(c.captures.at(1))))
      out.push(last)
    }
  }
  out
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
  let p = if src.starts-with("/") { src } else { "/design/" + normalise(if dir == "" { src } else { dir + "/" + src }) }
  image(p, ..args)
}

#set document(title: title)
#set text(font: sans, size: 10.5pt, lang: "en")
#show raw: set text(font: mono, size: 9pt)
#set par(justify: false, leading: 0.65em, spacing: 1.1em)
#set heading(numbering: "1.1")
#show heading: set text(weight: "semibold")
#show heading.where(level: 1): it => { pagebreak(weak: true); v(2em); text(26pt, it); v(1.5em) }
#show heading.where(level: 2): it => { v(1.2em); text(16pt, it); v(0.4em) }
#show heading.where(level: 4): set heading(numbering: none)
#show heading.where(level: 5): set heading(numbering: none)
#show table: set text(size: 9.5pt)
#set table(stroke: 0.5pt + luma(180), inset: 6pt)
#show table.cell.where(y: 0): set text(weight: "semibold")
#show link: set text(fill: rgb("#1f4e8c"))
#show raw.where(block: true): block.with(fill: luma(245), inset: 8pt, radius: 3pt, width: 100%)

// Title page: the title in the bottom left.
#page(margin: (x: 2.5cm, y: 3cm), numbering: none)[
  #align(bottom + left)[
    #text(36pt, weight: "bold", title)
    #v(0.6em)
    #text(11pt, fill: luma(90), datetime.today().display("[day] [month repr:long] [year]"))
  ]
]

#set page(paper: "a4", margin: (x: 2.3cm, y: 2.5cm), numbering: "1",
  footer: context [#text(8.5pt, fill: luma(110))[#title #h(1fr) #counter(page).display()]])

// Front page text from index.md, then the contents.
#if intro != "" {
  cmarker.render(intro, h1-level: 2, scope: (link: chapter-link("")))
  v(2em)
}
#text(20pt, weight: "semibold")[Contents]
#v(0.8em)
#outline(title: none, depth: 2, indent: auto)

#for part in parts {
  if part.name != none { heading(level: 1, part.name) }
  for ch in part.chapters {
    let dir = dir-of(ch.path)
    let (meta, body) = cmarker.render-with-metadata(
      read(ch.path),
      metadata-block: "frontmatter-yaml",
      h1-level: 2,
      label-prefix: ch.path + ":",
      scope: (link: chapter-link(dir), image: chapter-image(dir)),
    )
    [#metadata(ch.path)#label(ch.path)]
    body
    if meta != none and "status" in meta {
      align(right, text(8.5pt, fill: luma(110))[Status: #meta.status])
    }
  }
}
