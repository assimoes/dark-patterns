// highlighting shared by the adjudication (LLM-cited) and blind (self-cited) screens

export type Segment = { text: string; code?: string; color?: string };

const PALETTE = [
    "#d97706", // amber
    "#0284c7", // sky
    "#7c3aed", // violet
    "#db2777", // pink
    "#059669", // emerald
    "#4f46e5", // indigo
    "#dc2626", // red
    "#0d9488", // teal
];

// one colour per code, sorted so it's deterministic
export function assignColors(codes: string[]): Record<string, string> {
    const map: Record<string, string> = {};
    Array.from(new Set(codes))
        .sort()
        .forEach((code, i) => (map[code] = PALETTE[i % PALETTE.length]));
    return map;
}

// split body into plain + highlighted segments at the spans.
// spans matched by first occurrence and deduped by text, so a repeated phrase only highlights once
export function splitBySpans(
    body: string,
    spans: { text: string; code: string; color: string }[],
): Segment[] {
    const ranges = [] as { start: number; end: number; code: string; color: string }[];
    const seen = new Set<string>();
    for (const s of spans) {
        if (!s.text || seen.has(s.text)) continue;
        const start = body.indexOf(s.text);
        if (start === -1) continue;
        seen.add(s.text);
        ranges.push({ start, end: start + s.text.length, code: s.code, color: s.color });
    }
    ranges.sort((a, b) => a.start - b.start);

    const clean: typeof ranges = [];
    let lastEnd = -1;
    for (const r of ranges) {
        if (r.start >= lastEnd) {
            clean.push(r);
            lastEnd = r.end;
        }
    }

    const segments: Segment[] = [];
    let pos = 0;
    for (const r of clean) {
        if (r.start > pos) segments.push({ text: body.slice(pos, r.start) });
        segments.push({ text: body.slice(r.start, r.end), code: r.code, color: r.color });
        pos = r.end;
    }
    if (pos < body.length) segments.push({ text: body.slice(pos) });
    return segments;
}