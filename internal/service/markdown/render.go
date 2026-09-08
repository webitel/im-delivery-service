package markdown

import (
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// Render reconstructs markdown from plain text and a slice of formatting entities.
// It is the semantic inverse of im-gateway-service's internal/service/markdown/parser.go
// Parse function. Entities are expected to be pre-sorted by (Offset asc, Length desc),
// but Render defensively re-sorts them to ensure correctness even if the input is unsorted.
//
// Render never trusts entity bounds: overlapping/malformed spans are clamped or dropped
// rather than causing a panic, and an entity type this package doesn't recognize (e.g. one
// added to the cluster-wide contract after this code was written) degrades to escaped
// plain text instead of silently deleting the text it covers.
//
// Known lossy edge cases:
//   - PRE (code block) entity language/info-string cannot be reconstructed; this is an
//     accepted loss since the parser (stripCodeFenceInfoString) discards it at parse time.
//   - LINK destinations are wrapped in angle brackets (`[text](<url>)`) so parentheses in
//     the URL don't break the link syntax; a literal `>` inside the URL is not solved in v1.
func Render(plainText string, entities []model.Entity) string {
	filtered := filterEntities(plainText, entities)

	slices.SortStableFunc(filtered, func(a, b model.Entity) int {
		if a.Offset != b.Offset {
			return int(a.Offset - b.Offset)
		}

		return int(b.Length - a.Length)
	})

	var out strings.Builder
	out.Grow(len(plainText))

	i := 0 // shared cursor into filtered, consumed by walk across the whole call tree
	walk(plainText, filtered, &i, 0, len(plainText), &out)

	return out.String()
}

// filterEntities drops entities that can never be valid: negative offset, non-positive
// length, a span past the end of plainText, or an offset/end that splits a multi-byte
// UTF-8 rune. This is defense in depth against a buggy upstream service -- Render must
// never panic, no matter what it's handed.
func filterEntities(plainText string, entities []model.Entity) []model.Entity {
	filtered := make([]model.Entity, 0, len(entities))

	for _, e := range entities {
		if e.Offset < 0 || e.Length <= 0 {
			continue
		}

		end := int(e.Offset) + int(e.Length)
		if end > len(plainText) {
			continue
		}

		if !utf8.RuneStart(plainText[e.Offset]) {
			continue
		}

		if end < len(plainText) && !utf8.RuneStart(plainText[end]) {
			continue
		}

		filtered = append(filtered, e)
	}

	return filtered
}

// walk emits plainText[from:limit] into out, wrapping any entity in [from, limit) with
// its markdown delimiters and recursing into its nested children. i is a cursor shared
// across the whole call tree: entities arrive pre-order sorted (offset asc, length
// desc), so once an entity is consumed, every later entity with Offset < that entity's
// end is one of its descendants.
//
// An entity whose declared end exceeds the caller's limit -- invalid overlap, not proper
// nesting -- is clamped to limit instead of being trusted. That clamp is what keeps a
// malformed/overlapping entity list from ever producing a reversed plainText[from:limit]
// slice further up the call stack.
func walk(plainText string, entities []model.Entity, i *int, from, limit int, out *strings.Builder) {
	for *i < len(entities) && int(entities[*i].Offset) < limit {
		e := entities[*i]
		eOffset := int(e.Offset)

		end := eOffset + int(e.Length)
		if end > limit {
			end = limit
		}

		out.WriteString(escapePlain(plainText[from:eOffset]))

		*i++

		switch e.Type {
		case model.EntityTypeBold:
			out.WriteString("**")
			walk(plainText, entities, i, eOffset, end, out)
			out.WriteString("**")

		case model.EntityTypeItalic:
			out.WriteString("*")
			walk(plainText, entities, i, eOffset, end, out)
			out.WriteString("*")

		case model.EntityTypeStrikethrough:
			out.WriteString("~~")
			walk(plainText, entities, i, eOffset, end, out)
			out.WriteString("~~")

		case model.EntityTypeCode:
			// CODE never recurses or escapes -- content is an opaque leaf, matching
			// im-gateway-service's parser (which never re-parses code-span content).
			skipChildren(entities, i, end)
			writeFenced(out, plainText[eOffset:end], 1)

		case model.EntityTypePre:
			// Same as CODE, but a multi-line fence: a single-line ```text``` is
			// parsed by CommonMark as inline code, not a code block.
			skipChildren(entities, i, end)
			writeFenced(out, plainText[eOffset:end], 3)

		case model.EntityTypeLink:
			out.WriteString("[")
			walk(plainText, entities, i, eOffset, end, out)
			out.WriteString("](<")

			if e.Value != nil {
				out.WriteString(*e.Value)
			}

			out.WriteString(">)")

		default:
			// Unrecognized entity type (e.g. a new span kind added upstream that this
			// service doesn't know how to render yet): never drop the text it covers --
			// degrade to escaped plain text instead of silently deleting it.
			skipChildren(entities, i, end)
			out.WriteString(escapePlain(plainText[eOffset:end]))
		}

		from = end
	}

	out.WriteString(escapePlain(plainText[from:limit]))
}

// skipChildren advances the shared cursor past every entity nested inside [.., limit)
// without rendering it -- used by CODE/PRE and unrecognized entity types, whose content
// is treated as an opaque leaf.
func skipChildren(entities []model.Entity, i *int, limit int) {
	for *i < len(entities) && int(entities[*i].Offset) < limit {
		*i++
	}
}

// writeFenced wraps content in a run of backticks at least minTicks long, on its own
// line when minTicks > 1 (PRE), extending the fence beyond the longest run of
// consecutive backticks already present in content -- otherwise a literal backtick
// inside CODE/PRE content could terminate the fence early (standard CommonMark practice).
func writeFenced(out *strings.Builder, content string, minTicks int) {
	fence := backtickFence(content, minTicks)

	out.WriteString(fence)

	if minTicks > 1 {
		out.WriteString("\n")
	}

	out.WriteString(content)

	if minTicks > 1 {
		out.WriteString("\n")
	}

	out.WriteString(fence)
}

// backtickFence returns a run of backticks at least minTicks long, extended if needed so
// it exceeds the longest run of consecutive backticks already present in content.
func backtickFence(content string, minTicks int) string {
	maxRun, current := 0, 0

	for i := range len(content) {
		if content[i] == '`' {
			current++
			if current > maxRun {
				maxRun = current
			}
		} else {
			current = 0
		}
	}

	n := minTicks
	if maxRun+1 > n {
		n = maxRun + 1
	}

	return strings.Repeat("`", n)
}

// escapePlain inserts a backslash before special markdown characters:
// *, _, ~, `, [, ], and \ itself.
func escapePlain(s string) string {
	if s == "" {
		return ""
	}

	var out strings.Builder
	out.Grow(len(s) + 8)

	for i := range len(s) {
		ch := s[i]
		switch ch {
		case '*', '_', '~', '`', '[', ']', '\\':
			out.WriteByte('\\')
			out.WriteByte(ch)
		default:
			out.WriteByte(ch)
		}
	}

	return out.String()
}
