package markdown

import (
	"testing"

	"github.com/webitel/im-delivery-service/internal/domain/model"
)

func TestRender_PlainTextNoEntitiesNoSpecialChars(t *testing.T) {
	t.Run("plain text without entities or special chars", func(t *testing.T) {
		plainText := "hello world"
		entities := []model.Entity{}

		got := Render(plainText, entities)
		want := "hello world"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestRender_PlainTextNoEntitiesWithSpecialChars(t *testing.T) {
	t.Run("plain text with literal special chars must be escaped", func(t *testing.T) {
		plainText := "5*3=15, email_user@domain, ~~strikethrough~~, [link](url), `code`, ```pre```"
		entities := []model.Entity{}

		got := Render(plainText, entities)
		want := "5\\*3=15, email\\_user@domain, \\~\\~strikethrough\\~\\~, \\[link\\](url), \\`code\\`, \\`\\`\\`pre\\`\\`\\`"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestRender_EntityTypes(t *testing.T) {
	t.Run("BOLD entity", func(t *testing.T) {
		plainText := "hello world"
		entities := []model.Entity{
			{
				Type:   model.EntityTypeBold,
				Offset: 0,
				Length: 5,
				Value:  nil,
			},
		}

		got := Render(plainText, entities)
		want := "**hello** world"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("ITALIC entity", func(t *testing.T) {
		plainText := "hello world"
		entities := []model.Entity{
			{
				Type:   model.EntityTypeItalic,
				Offset: 6,
				Length: 5,
				Value:  nil,
			},
		}

		got := Render(plainText, entities)
		want := "hello *world*"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("STRIKETHROUGH entity", func(t *testing.T) {
		plainText := "hello world"
		entities := []model.Entity{
			{
				Type:   model.EntityTypeStrikethrough,
				Offset: 0,
				Length: 11,
				Value:  nil,
			},
		}

		got := Render(plainText, entities)
		want := "~~hello world~~"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("CODE entity", func(t *testing.T) {
		plainText := "const x = 5"
		entities := []model.Entity{
			{
				Type:   model.EntityTypeCode,
				Offset: 0,
				Length: 11,
				Value:  nil,
			},
		}

		got := Render(plainText, entities)
		want := "`const x = 5`"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("PRE entity", func(t *testing.T) {
		plainText := "function foo() {\n  return 42;\n}"
		entities := []model.Entity{
			{
				Type:   model.EntityTypePre,
				Offset: 0,
				Length: int32(len("function foo() {\n  return 42;\n}")),
				Value:  nil,
			},
		}

		got := Render(plainText, entities)
		want := "```\nfunction foo() {\n  return 42;\n}\n```"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("LINK entity", func(t *testing.T) {
		plainText := "click here"
		url := "https://example.com"
		entities := []model.Entity{
			{
				Type:   model.EntityTypeLink,
				Offset: 0,
				Length: 10,
				Value:  &url,
			},
		}

		got := Render(plainText, entities)
		want := "[click here](<https://example.com>)"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestRender_LinkWithComplexURL(t *testing.T) {
	t.Run("LINK with URL containing special characters", func(t *testing.T) {
		plainText := "search results"
		url := "https://example.com/search?q=hello&lang=en&page=1"
		entities := []model.Entity{
			{
				Type:   model.EntityTypeLink,
				Offset: 0,
				Length: 14,
				Value:  &url,
			},
		}

		got := Render(plainText, entities)
		want := "[search results](<https://example.com/search?q=hello&lang=en&page=1>)"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("LINK with URL containing parentheses", func(t *testing.T) {
		plainText := "wiki"
		url := "https://en.wikipedia.org/wiki/Go_(language)"
		entities := []model.Entity{
			{
				Type:   model.EntityTypeLink,
				Offset: 0,
				Length: 4,
				Value:  &url,
			},
		}

		got := Render(plainText, entities)
		want := "[wiki](<https://en.wikipedia.org/wiki/Go_(language)>)"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("LINK with nil Value URL", func(t *testing.T) {
		plainText := "click here"
		entities := []model.Entity{
			{
				Type:   model.EntityTypeLink,
				Offset: 0,
				Length: 10,
				Value:  nil,
			},
		}

		got := Render(plainText, entities)
		want := "[click here](<>)"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestRender_NestedEntities(t *testing.T) {
	t.Run("italic fully inside bold", func(t *testing.T) {
		plainText := "bold italic"
		entities := []model.Entity{
			// ITALIC entity (child)
			{
				Type:   model.EntityTypeItalic,
				Offset: 5,
				Length: 6,
				Value:  nil,
			},
			// BOLD entity (parent)
			{
				Type:   model.EntityTypeBold,
				Offset: 0,
				Length: 11,
				Value:  nil,
			},
		}

		got := Render(plainText, entities)
		want := "**bold *italic***"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("bold inside LINK display text", func(t *testing.T) {
		plainText := "bold text link"
		url := "https://example.com"
		entities := []model.Entity{
			// BOLD entity (child)
			{
				Type:   model.EntityTypeBold,
				Offset: 0,
				Length: 4,
				Value:  nil,
			},
			// LINK entity (parent)
			{
				Type:   model.EntityTypeLink,
				Offset: 0,
				Length: 14,
				Value:  &url,
			},
		}

		got := Render(plainText, entities)
		want := "[**bold** text link](<https://example.com>)"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("italic inside strikethrough", func(t *testing.T) {
		plainText := "strike italic"
		entities := []model.Entity{
			// ITALIC entity (child)
			{
				Type:   model.EntityTypeItalic,
				Offset: 7,
				Length: 6,
				Value:  nil,
			},
			// STRIKETHROUGH entity (parent)
			{
				Type:   model.EntityTypeStrikethrough,
				Offset: 0,
				Length: 13,
				Value:  nil,
			},
		}

		got := Render(plainText, entities)
		want := "~~strike *italic*~~"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestRender_SiblingEntities(t *testing.T) {
	t.Run("multiple non-overlapping entities", func(t *testing.T) {
		plainText := "bold italic strikethrough"
		entities := []model.Entity{
			{
				Type:   model.EntityTypeBold,
				Offset: 0,
				Length: 4,
				Value:  nil,
			},
			{
				Type:   model.EntityTypeItalic,
				Offset: 5,
				Length: 6,
				Value:  nil,
			},
			{
				Type:   model.EntityTypeStrikethrough,
				Offset: 12,
				Length: 13,
				Value:  nil,
			},
		}

		got := Render(plainText, entities)
		want := "**bold** *italic* ~~strikethrough~~"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestRender_EdgeCases(t *testing.T) {
	t.Run("entity anchored at offset 0", func(t *testing.T) {
		plainText := "start end"
		entities := []model.Entity{
			{
				Type:   model.EntityTypeBold,
				Offset: 0,
				Length: 5,
				Value:  nil,
			},
		}

		got := Render(plainText, entities)
		want := "**start** end"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("entity ending exactly at len(plainText)", func(t *testing.T) {
		plainText := "start end"
		entities := []model.Entity{
			{
				Type:   model.EntityTypeItalic,
				Offset: 6,
				Length: 3,
				Value:  nil,
			},
		}

		got := Render(plainText, entities)
		want := "start *end*"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("two adjacent back-to-back entities", func(t *testing.T) {
		plainText := "bolditalic"
		entities := []model.Entity{
			{
				Type:   model.EntityTypeBold,
				Offset: 0,
				Length: 4,
				Value:  nil,
			},
			{
				Type:   model.EntityTypeItalic,
				Offset: 4,
				Length: 6,
				Value:  nil,
			},
		}

		got := Render(plainText, entities)
		want := "**bold***italic*"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestRender_MalformedEntities(t *testing.T) {
	t.Run("entity with negative offset is skipped", func(t *testing.T) {
		plainText := "hello world"
		entities := []model.Entity{
			{
				Type:   model.EntityTypeBold,
				Offset: -1,
				Length: 5,
				Value:  nil,
			},
		}

		got := Render(plainText, entities)
		want := "hello world"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("entity with zero length is skipped", func(t *testing.T) {
		plainText := "hello world"
		entities := []model.Entity{
			{
				Type:   model.EntityTypeBold,
				Offset: 0,
				Length: 0,
				Value:  nil,
			},
		}

		got := Render(plainText, entities)
		want := "hello world"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("entity out of bounds is skipped", func(t *testing.T) {
		plainText := "hello"
		entities := []model.Entity{
			{
				Type:   model.EntityTypeBold,
				Offset: 0,
				Length: 10,
				Value:  nil,
			},
		}

		got := Render(plainText, entities)
		want := "hello"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("mix of malformed and valid entities", func(t *testing.T) {
		plainText := "hello world test"
		entities := []model.Entity{
			// Malformed: out of bounds
			{
				Type:   model.EntityTypeBold,
				Offset: 0,
				Length: 100,
				Value:  nil,
			},
			// Valid: ITALIC
			{
				Type:   model.EntityTypeItalic,
				Offset: 6,
				Length: 5,
				Value:  nil,
			},
			// Malformed: negative offset
			{
				Type:   model.EntityTypeStrikethrough,
				Offset: -5,
				Length: 4,
				Value:  nil,
			},
			// Valid: BOLD
			{
				Type:   model.EntityTypeBold,
				Offset: 12,
				Length: 4,
				Value:  nil,
			},
		}

		got := Render(plainText, entities)
		want := "hello *world* **test**"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestRender_CodeAndPreSkipChildren(t *testing.T) {
	t.Run("CODE span does not recurse into children", func(t *testing.T) {
		plainText := "*not italic*"
		entities := []model.Entity{
			// ITALIC (spurious child, should be skipped)
			{
				Type:   model.EntityTypeItalic,
				Offset: 0,
				Length: 1,
				Value:  nil,
			},
			// CODE (parent, consumes the whole span)
			{
				Type:   model.EntityTypeCode,
				Offset: 0,
				Length: 12,
				Value:  nil,
			},
		}

		got := Render(plainText, entities)
		want := "`*not italic*`"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("PRE span does not recurse into children", func(t *testing.T) {
		plainText := "**not bold**"
		entities := []model.Entity{
			// BOLD (spurious child, should be skipped)
			{
				Type:   model.EntityTypeBold,
				Offset: 0,
				Length: 2,
				Value:  nil,
			},
			// PRE (parent, consumes the whole span)
			{
				Type:   model.EntityTypePre,
				Offset: 0,
				Length: 12,
				Value:  nil,
			},
		}

		got := Render(plainText, entities)
		want := "```\n**not bold**\n```"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestRender_EscapingInNestedContexts(t *testing.T) {
	t.Run("special chars in BOLD text are escaped", func(t *testing.T) {
		plainText := "a*b"
		entities := []model.Entity{
			{
				Type:   model.EntityTypeBold,
				Offset: 0,
				Length: 3,
				Value:  nil,
			},
		}

		got := Render(plainText, entities)
		want := "**a\\*b**"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("special chars in LINK display text are escaped", func(t *testing.T) {
		plainText := "text_with_underscores"
		url := "https://example.com"
		entities := []model.Entity{
			{
				Type:   model.EntityTypeLink,
				Offset: 0,
				Length: 21,
				Value:  &url,
			},
		}

		got := Render(plainText, entities)
		want := "[text\\_with\\_underscores](<https://example.com>)"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestRender_SortingDefensively(t *testing.T) {
	t.Run("unsorted entities are sorted defensively", func(t *testing.T) {
		plainText := "one two three"
		// Provide entities out of order: second entity before first
		entities := []model.Entity{
			{
				Type:   model.EntityTypeItalic,
				Offset: 8,
				Length: 5,
				Value:  nil,
			},
			{
				Type:   model.EntityTypeBold,
				Offset: 0,
				Length: 3,
				Value:  nil,
			},
		}

		got := Render(plainText, entities)
		want := "**one** two *three*"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("entities with same offset are sorted by length descending", func(t *testing.T) {
		plainText := "nested"
		entities := []model.Entity{
			// Shorter entity first (should be re-sorted to second)
			{
				Type:   model.EntityTypeItalic,
				Offset: 0,
				Length: 3,
				Value:  nil,
			},
			// Longer entity second (should be re-sorted to first)
			{
				Type:   model.EntityTypeBold,
				Offset: 0,
				Length: 6,
				Value:  nil,
			},
		}

		got := Render(plainText, entities)
		want := "***nes*ted**"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestRender_InvalidOverlap(t *testing.T) {
	t.Run("partially overlapping (not properly nested) entities must not panic", func(t *testing.T) {
		plainText := "abcdefghijklmnopqrst"
		entities := []model.Entity{
			{Type: model.EntityTypeBold, Offset: 0, Length: 5},
			{Type: model.EntityTypeItalic, Offset: 3, Length: 8},
		}

		got := Render(plainText, entities)
		want := "**abc*de***fghijklmnopqrst"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("second overlap shape must not panic", func(t *testing.T) {
		plainText := "abcdefghijklmnopqrst"
		entities := []model.Entity{
			{Type: model.EntityTypeBold, Offset: 3, Length: 4},
			{Type: model.EntityTypeBold, Offset: 5, Length: 5},
		}

		// Must complete without panicking; the exact clamped rendering is secondary
		// to the no-panic guarantee for genuinely malformed/overlapping input.
		_ = Render(plainText, entities)
	})
}

func TestRender_UnknownEntityType(t *testing.T) {
	t.Run("unrecognized entity type degrades to plain text instead of deleting it", func(t *testing.T) {
		plainText := "hello world"
		entities := []model.Entity{
			{Type: model.EntityType("UNDERLINE"), Offset: 0, Length: 5},
		}

		got := Render(plainText, entities)
		want := "hello world"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestRender_BacktickInsideFence(t *testing.T) {
	t.Run("literal backtick inside CODE content widens the fence", func(t *testing.T) {
		plainText := "a`b"
		entities := []model.Entity{
			{Type: model.EntityTypeCode, Offset: 0, Length: 3},
		}

		got := Render(plainText, entities)
		want := "``a`b``"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("literal triple backtick inside PRE content widens the fence", func(t *testing.T) {
		plainText := "a```b"
		entities := []model.Entity{
			{Type: model.EntityTypePre, Offset: 0, Length: 5},
		}

		got := Render(plainText, entities)
		want := "````\na```b\n````"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}
