package components

import (
	"regexp"
	"strings"

	"github.com/dh-kam/goink.go/pkg/styles"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

// SyntaxLanguage selects the grammar used by Syntax.
type SyntaxLanguage string

const (
	// SyntaxGo highlights a Go source snippet.
	SyntaxGo SyntaxLanguage = "go"
	// SyntaxJSON highlights a JSON document.
	SyntaxJSON SyntaxLanguage = "json"
	// SyntaxYAML highlights a YAML document.
	SyntaxYAML SyntaxLanguage = "yaml"
	// SyntaxMarkdown highlights a Markdown document.
	SyntaxMarkdown SyntaxLanguage = "markdown"
	// SyntaxBash highlights a Bash shell snippet.
	SyntaxBash SyntaxLanguage = "bash"
	// SyntaxPython highlights a Python source snippet.
	SyntaxPython SyntaxLanguage = "python"
	// SyntaxRust highlights a Rust source snippet.
	SyntaxRust SyntaxLanguage = "rust"
	// SyntaxSQL highlights an SQL source snippet (case-insensitive
	// keywords such as SELECT / FROM / WHERE).
	SyntaxSQL SyntaxLanguage = "sql"
	// SyntaxJavaScript highlights a JavaScript / ECMAScript source
	// snippet — keywords, template literals, regex, and comments.
	SyntaxJavaScript SyntaxLanguage = "javascript"
	// SyntaxDiff highlights a unified diff — '+' adds in green, '-'
	// removes in red, '@@ ... @@' hunk headers in cyan, plain context
	// lines untouched.
	SyntaxDiff SyntaxLanguage = "diff"
)

// SyntaxProps configures Syntax.
type SyntaxProps struct {
	Code     string
	Language SyntaxLanguage
}

// Syntax renders a syntax-highlighted code block as a text element.
//
// It is a deliberately minimal port of ink-syntax-highlight that recognises
// only Go and JSON. Unsupported languages fall back to plain text. Tokens are
// classified with a single regexp using ordered alternation; the first match
// wins so longer / more specific patterns must appear earlier in the union.
func Syntax(props SyntaxProps) *vdom.Node {
	highlighted := highlightCode(props.Code, props.Language)
	return vdom.CreateElement("text", markPublicComponentProps(nil), vdom.CreateTextNode(highlighted))
}

func highlightCode(code string, language SyntaxLanguage) string {
	if code == "" {
		return ""
	}

	switch language {
	case SyntaxGo:
		return applyTokens(code, goTokenizer)
	case SyntaxJSON:
		return applyTokens(code, jsonTokenizer)
	case SyntaxYAML:
		return applyTokens(code, yamlTokenizer)
	case SyntaxMarkdown:
		return applyTokens(code, markdownTokenizer)
	case SyntaxBash:
		return applyTokens(code, bashTokenizer)
	case SyntaxPython:
		return applyTokens(code, pythonTokenizer)
	case SyntaxRust:
		return applyTokens(code, rustTokenizer)
	case SyntaxSQL:
		return applyTokens(code, sqlTokenizer)
	case SyntaxJavaScript:
		return applyTokens(code, javascriptTokenizer)
	case SyntaxDiff:
		return highlightDiff(code)
	default:
		return code
	}
}

// tokenColorizer maps a regexp submatch index to an ANSI-wrapped string.
type tokenColorizer func(submatchIndex int, raw string) string

type tokenizer struct {
	pattern    *regexp.Regexp
	colorizers []tokenColorizer
}

func applyTokens(code string, t tokenizer) string {
	matches := t.pattern.FindAllStringSubmatchIndex(code, -1)
	if len(matches) == 0 {
		return code
	}

	var out strings.Builder
	out.Grow(len(code) + 16*len(matches))

	cursor := 0
	for _, match := range matches {
		// match[0]/match[1] is the full match; subgroups follow.
		fullStart, fullEnd := match[0], match[1]
		if fullStart > cursor {
			out.WriteString(code[cursor:fullStart])
		}

		// Find which subgroup actually matched.
		colorized := ""
		handled := false
		for group := 1; group <= len(t.colorizers); group++ {
			start := match[2*group]
			end := match[2*group+1]
			if start < 0 || end < 0 {
				continue
			}

			raw := code[start:end]
			colorized = t.colorizers[group-1](group, raw)
			handled = true
			break
		}

		if !handled {
			colorized = code[fullStart:fullEnd]
		}

		out.WriteString(colorized)
		cursor = fullEnd
	}

	if cursor < len(code) {
		out.WriteString(code[cursor:])
	}

	return out.String()
}

// --- Go tokenizer ---------------------------------------------------------

var goKeywords = map[string]struct{}{
	"break": {}, "case": {}, "chan": {}, "const": {}, "continue": {},
	"default": {}, "defer": {}, "else": {}, "fallthrough": {}, "for": {},
	"func": {}, "go": {}, "goto": {}, "if": {}, "import": {},
	"interface": {}, "map": {}, "package": {}, "range": {}, "return": {},
	"select": {}, "struct": {}, "switch": {}, "type": {}, "var": {},
	"true": {}, "false": {}, "nil": {},
}

// goTokenPattern matches, in priority order:
//  1. line comments  (//...)
//  2. block comments (/* ... */)
//  3. double-quoted strings (with escape support)
//  4. backtick raw strings
//  5. numeric literals (int / float / hex)
//  6. identifiers (later filtered for keywords)
var goTokenPattern = regexp.MustCompile(
	`(//[^\n]*)` +
		`|(/\*[\s\S]*?\*/)` +
		`|("(?:\\.|[^"\\\n])*")` +
		"|(`[^`]*`)" +
		`|(\b(?:0[xX][0-9a-fA-F]+|\d+(?:\.\d+)?)\b)` +
		`|([A-Za-z_][A-Za-z0-9_]*)`,
)

var goTokenizer = tokenizer{
	pattern: goTokenPattern,
	colorizers: []tokenColorizer{
		colorizeComment, // // line comment
		colorizeComment, // /* block comment */
		colorizeString,  // "..."
		colorizeString,  // `...`
		colorizeNumber,  // 123 / 0x1f / 1.5
		colorizeGoIdent, // identifier (filter for keyword)
	},
}

func colorizeComment(_ int, raw string) string {
	// Use Dim() so legacy terminals without bright-black still render visibly.
	return styles.Dim(raw)
}

func colorizeString(_ int, raw string) string {
	return styles.Colorize(raw, styles.Green, styles.Foreground)
}

func colorizeNumber(_ int, raw string) string {
	return styles.Colorize(raw, styles.Yellow, styles.Foreground)
}

func colorizeGoIdent(_ int, raw string) string {
	if _, isKeyword := goKeywords[raw]; isKeyword {
		return styles.Colorize(raw, styles.Cyan, styles.Foreground)
	}
	return raw
}

// --- JSON tokenizer -------------------------------------------------------

// jsonTokenPattern matches:
//  1. object keys ("key" plus the trailing optional whitespace + ':' so the
//     colon can be re-emitted uncolored — Go's RE2 engine has no lookahead)
//  2. plain double-quoted strings
//  3. numeric literals (int / float / negative / exponent)
//  4. literal keywords true/false/null
var jsonTokenPattern = regexp.MustCompile(
	`("(?:\\.|[^"\\])*"\s*:)` +
		`|("(?:\\.|[^"\\])*")` +
		`|(-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?)` +
		`|\b(true|false|null)\b`,
)

var jsonTokenizer = tokenizer{
	pattern: jsonTokenPattern,
	colorizers: []tokenColorizer{
		colorizeJSONKey,
		colorizeString,
		colorizeNumber,
		colorizeJSONLiteral,
	},
}

// colorizeJSONKey expects raw of the form `"key"<ws>:` — colorize the
// quoted key portion and re-emit the trailing whitespace + colon untouched.
func colorizeJSONKey(_ int, raw string) string {
	closeQuote := strings.LastIndex(raw, `"`)
	if closeQuote < 0 {
		return raw
	}
	keyPart := raw[:closeQuote+1]
	tail := raw[closeQuote+1:]
	return styles.Colorize(keyPart, styles.Blue, styles.Foreground) + tail
}

func colorizeJSONLiteral(_ int, raw string) string {
	return styles.Colorize(raw, styles.Cyan, styles.Foreground)
}

// --- YAML tokenizer -------------------------------------------------------

// yamlTokenPattern matches, in priority order:
//  1. comments — '#' to end of line (preceded by start-of-line or whitespace)
//  2. block document markers '---' / '...'
//  3. mapping keys: bare-word or single/double quoted string immediately
//     followed by ':' and whitespace (or end of line)
//  4. double-quoted strings
//  5. single-quoted strings
//  6. numeric literals (int / float / scientific)
//  7. boolean / null literals (case-insensitive forms YAML 1.1 + 1.2)
//  8. list/sequence markers '-' at start of indented line
var yamlTokenPattern = regexp.MustCompile(
	`(#[^\n]*)` +
		`|(^---|^\.\.\.)` +
		`|((?:[A-Za-z_][\w.\-]*|"[^"\n]*"|'[^'\n]*')\s*:(?:\s|$))` +
		`|("(?:\\.|[^"\\\n])*")` +
		`|('[^'\n]*')` +
		`|(-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?)` +
		`|\b(true|false|True|False|TRUE|FALSE|null|Null|NULL|yes|no|Yes|No|YES|NO|~)\b` +
		`|((?m:^\s*-(?:\s|$)))`,
)

var yamlTokenizer = tokenizer{
	pattern: yamlTokenPattern,
	colorizers: []tokenColorizer{
		colorizeComment,     // # comment
		colorizeYAMLMarker,  // --- / ...
		colorizeYAMLKey,     // key:
		colorizeString,      // "..."
		colorizeString,      // '...'
		colorizeNumber,      // 42, 3.14
		colorizeJSONLiteral, // true / false / null / ~
		colorizeYAMLDash,    // - list marker
	},
}

func colorizeYAMLMarker(_ int, raw string) string {
	return styles.Colorize(raw, styles.Magenta, styles.Foreground)
}

// colorizeYAMLKey expects raw of the form `<key><ws>:` (with optional
// trailing whitespace) — colorize the key, leave colon and whitespace alone.
func colorizeYAMLKey(_ int, raw string) string {
	colon := strings.LastIndex(raw, ":")
	if colon < 0 {
		return raw
	}
	keyPart := strings.TrimRight(raw[:colon], " \t")
	tail := raw[len(keyPart):]
	return styles.Colorize(keyPart, styles.Blue, styles.Foreground) + tail
}

func colorizeYAMLDash(_ int, raw string) string {
	return styles.Colorize(raw, styles.Magenta, styles.Foreground)
}

// --- Markdown tokenizer ---------------------------------------------------

// markdownTokenPattern matches, in priority order:
//  1. fenced code blocks ```...```
//  2. ATX headings: leading '#' run plus the rest of the line
//  3. blockquote markers (`>` at start of line)
//  4. unordered list markers (`-`, `*`, `+` at start of indented line)
//  5. ordered list markers (digit + '.' at start of indented line)
//  6. inline code spans `...`
//  7. bold spans **...** / __...__
//  8. italic spans *...* / _..._  (lower priority than bold)
//  9. inline links [text](url)
var markdownTokenPattern = regexp.MustCompile(
	"(?ms)" +
		"(^```[\\s\\S]*?^```)" +
		`|(?m:^(#{1,6}\s[^\n]*))` +
		`|(?m:^(>\s[^\n]*))` +
		`|(?m:^(\s*[-*+]\s))` +
		`|(?m:^(\s*\d+\.\s))` +
		"|(`[^`\n]+`)" +
		`|(\*\*[^*\n]+\*\*|__[^_\n]+__)` +
		`|(\*[^*\n]+\*|_[^_\n]+_)` +
		`|(\[[^\]\n]+\]\([^)\n]+\))`,
)

var markdownTokenizer = tokenizer{
	pattern: markdownTokenPattern,
	colorizers: []tokenColorizer{
		colorizeMarkdownCodeBlock,  // ``` ... ```
		colorizeMarkdownHeading,    // # heading
		colorizeMarkdownBlockquote, // > quote
		colorizeMarkdownListMarker, // - / * / +
		colorizeMarkdownListMarker, // 1.
		colorizeMarkdownCodeSpan,   // `code`
		colorizeMarkdownBold,       // **bold**
		colorizeMarkdownItalic,     // *italic*
		colorizeMarkdownLink,       // [text](url)
	},
}

func colorizeMarkdownCodeBlock(_ int, raw string) string {
	return styles.Colorize(raw, styles.Green, styles.Foreground)
}

func colorizeMarkdownHeading(_ int, raw string) string {
	return styles.WrapWithANSI(raw, styles.Cyan.ToANSI(styles.Foreground), styles.BoldCode())
}

func colorizeMarkdownBlockquote(_ int, raw string) string {
	return styles.Dim(raw)
}

func colorizeMarkdownListMarker(_ int, raw string) string {
	return styles.Colorize(raw, styles.Magenta, styles.Foreground)
}

func colorizeMarkdownCodeSpan(_ int, raw string) string {
	return styles.Colorize(raw, styles.Green, styles.Foreground)
}

func colorizeMarkdownBold(_ int, raw string) string {
	return styles.WrapWithANSI(raw, styles.BoldCode())
}

func colorizeMarkdownItalic(_ int, raw string) string {
	return styles.WrapWithANSI(raw, styles.ItalicCode())
}

// colorizeMarkdownLink wraps the text portion in cyan and the URL portion
// in dimmed text so the structure stays readable in any terminal palette.
func colorizeMarkdownLink(_ int, raw string) string {
	close := strings.Index(raw, "](")
	if close < 0 {
		return raw
	}
	textPart := raw[:close+1]
	urlPart := raw[close+1:]
	return styles.Colorize(textPart, styles.Cyan, styles.Foreground) + styles.Dim(urlPart)
}

// --- Bash tokenizer -------------------------------------------------------

// bashKeywords are the reserved words highlighted by the Bash tokenizer.
var bashKeywords = map[string]struct{}{
	"if": {}, "then": {}, "else": {}, "elif": {}, "fi": {},
	"case": {}, "esac": {}, "in": {},
	"for": {}, "while": {}, "until": {}, "do": {}, "done": {},
	"function": {}, "return": {}, "select": {}, "time": {},
	"break": {}, "continue": {},
}

// bashTokenPattern matches:
//  1. comments — '#' to end of line
//  2. double-quoted strings (with escape support)
//  3. single-quoted strings
//  4. variable expansions $VAR / ${VAR}
//  5. numeric literals
//  6. identifiers (later filtered for keyword)
var bashTokenPattern = regexp.MustCompile(
	`(#[^\n]*)` +
		`|("(?:\\.|[^"\\\n])*")` +
		`|('[^'\n]*')` +
		`|(\$(?:\{[^}\n]+\}|[A-Za-z_][A-Za-z0-9_]*))` +
		`|(\b\d+(?:\.\d+)?\b)` +
		`|([A-Za-z_][A-Za-z0-9_]*)`,
)

var bashTokenizer = tokenizer{
	pattern: bashTokenPattern,
	colorizers: []tokenColorizer{
		colorizeComment,   // # comment
		colorizeString,    // "..."
		colorizeString,    // '...'
		colorizeBashVar,   // $VAR / ${VAR}
		colorizeNumber,    // 123
		colorizeBashIdent, // identifier (filter for keyword)
	},
}

func colorizeBashVar(_ int, raw string) string {
	return styles.Colorize(raw, styles.Yellow, styles.Foreground)
}

func colorizeBashIdent(_ int, raw string) string {
	if _, isKeyword := bashKeywords[raw]; isKeyword {
		return styles.Colorize(raw, styles.Cyan, styles.Foreground)
	}
	return raw
}

// --- Python tokenizer -----------------------------------------------------

// pythonKeywords is the set of reserved words highlighted as keywords.
// Common builtins (None / True / False / self / cls) are included so the
// rendered snippet matches the look of Ink's reference Python highlighter.
var pythonKeywords = map[string]struct{}{
	"False": {}, "None": {}, "True": {}, "and": {}, "as": {},
	"assert": {}, "async": {}, "await": {}, "break": {}, "class": {},
	"continue": {}, "def": {}, "del": {}, "elif": {}, "else": {},
	"except": {}, "finally": {}, "for": {}, "from": {}, "global": {},
	"if": {}, "import": {}, "in": {}, "is": {}, "lambda": {},
	"nonlocal": {}, "not": {}, "or": {}, "pass": {}, "raise": {},
	"return": {}, "try": {}, "while": {}, "with": {}, "yield": {},
	"self": {}, "cls": {},
}

// pythonTokenPattern matches, in priority order:
//  1. triple-quoted strings (""" ... """ and ''' ... ''')
//  2. line comments (# ...)
//  3. double-quoted strings — including optional f/r/b prefix
//  4. single-quoted strings — including optional f/r/b prefix
//  5. decorators (@name.path)
//  6. numeric literals
//  7. identifiers (later filtered for keyword)
var pythonTokenPattern = regexp.MustCompile(
	`("""[\s\S]*?"""|'''[\s\S]*?''')` +
		`|(#[^\n]*)` +
		`|([fFrRbB]?"(?:\\.|[^"\\\n])*")` +
		`|([fFrRbB]?'(?:\\.|[^'\\\n])*')` +
		`|(@[A-Za-z_][A-Za-z0-9_.]*)` +
		`|(\b\d+(?:\.\d+)?(?:[eE][+-]?\d+)?\b)` +
		`|([A-Za-z_][A-Za-z0-9_]*)`,
)

var pythonTokenizer = tokenizer{
	pattern: pythonTokenPattern,
	colorizers: []tokenColorizer{
		colorizeString,      // """ ... """
		colorizeComment,     // # comment
		colorizeString,      // "..."
		colorizeString,      // '...'
		colorizeDecorator,   // @decorator
		colorizeNumber,      // 123 / 1.5
		colorizePythonIdent, // identifier (filter for keyword)
	},
}

func colorizeDecorator(_ int, raw string) string {
	return styles.Colorize(raw, styles.Magenta, styles.Foreground)
}

func colorizePythonIdent(_ int, raw string) string {
	if _, isKeyword := pythonKeywords[raw]; isKeyword {
		return styles.Colorize(raw, styles.Cyan, styles.Foreground)
	}
	return raw
}

// --- Rust tokenizer -------------------------------------------------------

// rustKeywords is the set of reserved words highlighted as keywords.
var rustKeywords = map[string]struct{}{
	"as": {}, "async": {}, "await": {}, "break": {}, "const": {},
	"continue": {}, "crate": {}, "dyn": {}, "else": {}, "enum": {},
	"extern": {}, "false": {}, "fn": {}, "for": {}, "if": {},
	"impl": {}, "in": {}, "let": {}, "loop": {}, "match": {},
	"mod": {}, "move": {}, "mut": {}, "pub": {}, "ref": {},
	"return": {}, "self": {}, "Self": {}, "static": {}, "struct": {},
	"super": {}, "trait": {}, "true": {}, "type": {}, "unsafe": {},
	"use": {}, "where": {}, "while": {},
}

// rustTokenPattern matches, in priority order:
//  1. line comments (// ...)
//  2. block comments (/* ... */)
//  3. attributes (#[ ... ] or #![ ... ])
//  4. double-quoted strings (with escape support)
//  5. char literals 'x' or '\n' — kept compact so they don't collide with lifetimes
//  6. lifetimes ('a, 'static)
//  7. numeric literals (int / float / hex)
//  8. identifiers (later filtered for keyword)
var rustTokenPattern = regexp.MustCompile(
	`(//[^\n]*)` +
		`|(/\*[\s\S]*?\*/)` +
		`|(#!?\[[^\]\n]*\])` +
		`|("(?:\\.|[^"\\\n])*")` +
		`|('(?:\\[\\'nrt0"]|\\x[0-9a-fA-F]{2}|\\u\{[0-9a-fA-F]+\}|[^'\\])')` +
		`|('[A-Za-z_][A-Za-z0-9_]*)` +
		`|(\b(?:0[xX][0-9a-fA-F_]+|\d[\d_]*(?:\.\d[\d_]*)?)\b)` +
		`|([A-Za-z_][A-Za-z0-9_]*)`,
)

var rustTokenizer = tokenizer{
	pattern: rustTokenPattern,
	colorizers: []tokenColorizer{
		colorizeComment,    // // line comment
		colorizeComment,    // /* block comment */
		colorizeAttribute,  // #[derive(...)]
		colorizeString,     // "..."
		colorizeString,     // 'x' (char literal — green like a string)
		colorizeLifetime,   // 'a / 'static
		colorizeNumber,     // 123 / 0x1f / 1.5
		colorizeRustIdent,  // identifier (filter for keyword)
	},
}

func colorizeAttribute(_ int, raw string) string {
	return styles.Colorize(raw, styles.Magenta, styles.Foreground)
}

func colorizeLifetime(_ int, raw string) string {
	return styles.Colorize(raw, styles.Yellow, styles.Foreground)
}

func colorizeRustIdent(_ int, raw string) string {
	if _, isKeyword := rustKeywords[raw]; isKeyword {
		return styles.Colorize(raw, styles.Cyan, styles.Foreground)
	}
	return raw
}

// --- SQL tokenizer --------------------------------------------------------

// sqlKeywords is the set of reserved words highlighted as keywords.
// Membership is checked case-insensitively (the identifier is upper-cased
// before lookup) so SELECT / select / Select all colorize identically.
var sqlKeywords = map[string]struct{}{
	"SELECT": {}, "FROM": {}, "WHERE": {}, "AND": {}, "OR": {},
	"NOT": {}, "NULL": {}, "IS": {}, "IN": {}, "LIKE": {},
	"BETWEEN": {}, "INSERT": {}, "INTO": {}, "VALUES": {}, "UPDATE": {},
	"SET": {}, "DELETE": {}, "CREATE": {}, "TABLE": {}, "DROP": {},
	"ALTER": {}, "ADD": {}, "COLUMN": {}, "INDEX": {}, "VIEW": {},
	"PRIMARY": {}, "KEY": {}, "FOREIGN": {}, "REFERENCES": {}, "JOIN": {},
	"INNER": {}, "LEFT": {}, "RIGHT": {}, "OUTER": {}, "ON": {},
	"GROUP": {}, "BY": {}, "ORDER": {}, "HAVING": {}, "LIMIT": {},
	"OFFSET": {}, "DISTINCT": {}, "AS": {}, "UNION": {}, "ALL": {},
	"CASE": {}, "WHEN": {}, "THEN": {}, "ELSE": {}, "END": {},
	"IF": {}, "EXISTS": {}, "DEFAULT": {}, "TRUE": {}, "FALSE": {},
	"BEGIN": {}, "COMMIT": {}, "ROLLBACK": {}, "TRANSACTION": {},
	"WITH": {}, "RETURNING": {}, "USING": {},
}

// sqlTokenPattern matches, in priority order:
//  1. line comments (-- ...)
//  2. block comments (/* ... */)
//  3. single-quoted strings (with embedded '' escapes for SQL)
//  4. double-quoted identifiers (treated as strings — used by Postgres)
//  5. numeric literals
//  6. identifiers (later filtered for keyword, case-insensitive)
var sqlTokenPattern = regexp.MustCompile(
	`(--[^\n]*)` +
		`|(/\*[\s\S]*?\*/)` +
		`|('(?:''|[^'\n])*')` +
		`|("(?:[^"\\\n]|\\.)*")` +
		`|(\b\d+(?:\.\d+)?\b)` +
		`|([A-Za-z_][A-Za-z0-9_]*)`,
)

var sqlTokenizer = tokenizer{
	pattern: sqlTokenPattern,
	colorizers: []tokenColorizer{
		colorizeComment,  // -- line comment
		colorizeComment,  // /* block comment */
		colorizeString,   // '...'
		colorizeString,   // "..."
		colorizeNumber,   // 123 / 1.5
		colorizeSQLIdent, // identifier (filter for keyword, case-insensitive)
	},
}

func colorizeSQLIdent(_ int, raw string) string {
	if _, isKeyword := sqlKeywords[strings.ToUpper(raw)]; isKeyword {
		return styles.Colorize(raw, styles.Cyan, styles.Foreground)
	}
	return raw
}

// --- JavaScript tokenizer -------------------------------------------------

// javascriptKeywords is the set of reserved words highlighted as keywords.
// The set covers ES2015+ control flow, declarations, modules, async/await,
// the literal forms (true/false/null/undefined) and a handful of common
// pseudo-keywords (this/super) that ink-syntax-highlight also tints.
var javascriptKeywords = map[string]struct{}{
	"break": {}, "case": {}, "catch": {}, "class": {}, "const": {},
	"continue": {}, "debugger": {}, "default": {}, "delete": {}, "do": {},
	"else": {}, "export": {}, "extends": {}, "finally": {}, "for": {},
	"function": {}, "if": {}, "import": {}, "in": {}, "instanceof": {},
	"let": {}, "new": {}, "of": {}, "return": {}, "super": {},
	"switch": {}, "this": {}, "throw": {}, "try": {}, "typeof": {},
	"var": {}, "void": {}, "while": {}, "with": {}, "yield": {},
	"async": {}, "await": {}, "static": {}, "true": {}, "false": {},
	"null": {}, "undefined": {}, "NaN": {}, "Infinity": {},
}

// javascriptTokenPattern matches, in priority order:
//  1. line comments (// ...)
//  2. block comments (/* ... */)
//  3. template literals (`...`) — including interpolation as inner text
//  4. double-quoted strings (with escape support)
//  5. single-quoted strings (with escape support)
//  6. regex literals — anchored to a leading boundary character set so the
//     '/' division operator does not accidentally trigger one. The lookahead
//     for '=' in TS-style `// foo` is unnecessary because line comments are
//     matched before this group.
//  7. numeric literals (int / float / hex / scientific / bigint suffix)
//  8. identifiers (later filtered for keyword)
var javascriptTokenPattern = regexp.MustCompile(
	`(//[^\n]*)` +
		`|(/\*[\s\S]*?\*/)` +
		"|(`(?:\\\\.|[^`\\\\])*`)" +
		`|("(?:\\.|[^"\\\n])*")` +
		`|('(?:\\.|[^'\\\n])*')` +
		`|((?:^|[=({,;:!&|?+\-*/<>~^%\[])\s*/(?:\\.|\[(?:\\.|[^\]\\\n])*\]|[^/\\\n])+/[gimsuy]*)` +
		`|(\b(?:0[xX][0-9a-fA-F]+|\d+(?:\.\d+)?(?:[eE][+-]?\d+)?n?)\b)` +
		`|([A-Za-z_$][A-Za-z0-9_$]*)`,
)

var javascriptTokenizer = tokenizer{
	pattern: javascriptTokenPattern,
	colorizers: []tokenColorizer{
		colorizeComment,        // // line comment
		colorizeComment,        // /* block comment */
		colorizeString,         // template literal `...`
		colorizeString,         // "..."
		colorizeString,         // '...'
		colorizeJavaScriptRegex, // /pattern/flags (with leading boundary char)
		colorizeNumber,         // 123 / 0xff / 1e2 / 9n
		colorizeJavaScriptIdent, // identifier (filter for keyword)
	},
}

// colorizeJavaScriptRegex wraps the regex literal portion of the match
// (everything from the first '/' onward) in magenta and re-emits the
// leading boundary character — '=' / '(' / ',' / etc. — verbatim so the
// surrounding code keeps its original spacing.
func colorizeJavaScriptRegex(_ int, raw string) string {
	slash := strings.Index(raw, "/")
	if slash < 0 {
		return raw
	}
	prefix := raw[:slash]
	regex := raw[slash:]
	return prefix + styles.Colorize(regex, styles.Magenta, styles.Foreground)
}

func colorizeJavaScriptIdent(_ int, raw string) string {
	if _, isKeyword := javascriptKeywords[raw]; isKeyword {
		return styles.Colorize(raw, styles.Cyan, styles.Foreground)
	}
	return raw
}

// --- Diff highlighter -----------------------------------------------------

// highlightDiff tints a unified diff line-by-line. Lines starting with
// '+' (other than the file header '+++') become green, '-' (other than
// '---') become red, '@@' hunk headers become cyan, and 'diff --git'
// / 'index ' style metadata is dimmed. The transform is line-oriented
// because each diff line is independent — no need for a regex tokenizer.
func highlightDiff(code string) string {
	if code == "" {
		return ""
	}

	lines := strings.Split(code, "\n")
	for i, line := range lines {
		switch {
		case strings.HasPrefix(line, "+++"), strings.HasPrefix(line, "---"):
			lines[i] = styles.Dim(line)
		case strings.HasPrefix(line, "@@"):
			lines[i] = styles.Colorize(line, styles.Cyan, styles.Foreground)
		case strings.HasPrefix(line, "+"):
			lines[i] = styles.Colorize(line, styles.Green, styles.Foreground)
		case strings.HasPrefix(line, "-"):
			lines[i] = styles.Colorize(line, styles.Red, styles.Foreground)
		case strings.HasPrefix(line, "diff "), strings.HasPrefix(line, "index "):
			lines[i] = styles.Dim(line)
		}
	}
	return strings.Join(lines, "\n")
}
