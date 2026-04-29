export type MdSpan = { text: string; bold: boolean }

export type MdNode =
  | { type: 'h2'; spans: MdSpan[] }
  | { type: 'h3'; spans: MdSpan[] }
  | { type: 'p'; spans: MdSpan[] }
  | { type: 'ul'; items: MdSpan[][] }
  | { type: 'hr' }

export function parseSpans(text: string): MdSpan[] {
  const spans: MdSpan[] = []
  let i = 0
  while (i < text.length) {
    const boldStart = text.indexOf('**', i)
    if (boldStart === -1) {
      if (i < text.length) {
        spans.push({ text: text.slice(i), bold: false })
      }
      break
    }
    if (boldStart > i) {
      spans.push({ text: text.slice(i, boldStart), bold: false })
    }
    const boldEnd = text.indexOf('**', boldStart + 2)
    if (boldEnd === -1) {
      spans.push({ text: text.slice(boldStart), bold: false })
      break
    }
    spans.push({ text: text.slice(boldStart + 2, boldEnd), bold: true })
    i = boldEnd + 2
  }
  return spans
}

function isListItem(line: string): boolean {
  const trimmed = line.trim()
  return trimmed.startsWith('- ') || trimmed.startsWith('* ')
}

function stripListMarker(line: string): string {
  const trimmed = line.trim()
  if (trimmed.startsWith('- ')) return trimmed.slice(2)
  if (trimmed.startsWith('* ')) return trimmed.slice(2)
  return trimmed
}

export function parseMarkdown(source: string): MdNode[] {
  const nodes: MdNode[] = []
  const blocks = source.split(/\n\n+/)

  for (let b = 0; b < blocks.length; b++) {
    const block = blocks[b]!.trim()
    if (!block) continue

    if (block === '---' || block === '***' || block === '___') {
      nodes.push({ type: 'hr' })
      continue
    }

    const lines = block.split('\n')

    // Heading
    if (lines.length === 1) {
      if (block.startsWith('### ')) {
        nodes.push({ type: 'h3', spans: parseSpans(block.slice(4).trim()) })
        continue
      }
      if (block.startsWith('## ')) {
        nodes.push({ type: 'h2', spans: parseSpans(block.slice(3).trim()) })
        continue
      }
    }

    // List: every non-empty line starts with list marker
    const nonEmptyLines = lines.filter((l) => l.trim().length > 0)
    if (nonEmptyLines.length > 0 && nonEmptyLines.every(isListItem)) {
      nodes.push({
        type: 'ul',
        items: nonEmptyLines.map((line) => parseSpans(stripListMarker(line))),
      })
      continue
    }

    // Paragraph
    const flattened = lines.join(' ').trim()
    nodes.push({ type: 'p', spans: parseSpans(flattened) })
  }

  return nodes
}
