package components_test

import (
	"strings"
	"testing"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/styles"
)

// syntaxText extracts the rendered text-node content from a Syntax node.
func syntaxText(t *testing.T, p components.SyntaxProps) string {
	t.Helper()
	node := components.Syntax(p)
	if node == nil {
		t.Fatalf("Syntax returned nil")
	}
	if len(node.Children) != 1 {
		t.Fatalf("Syntax produced %d children, want 1", len(node.Children))
	}
	return node.Children[0].Text
}

func TestSyntaxEmptyCode(t *testing.T) {
	got := syntaxText(t, components.SyntaxProps{Language: components.SyntaxGo})
	if got != "" {
		t.Fatalf("empty code should render empty, got %q", got)
	}
}

func TestSyntaxUnknownLanguagePassThrough(t *testing.T) {
	src := `func main() { return 42 }`
	got := syntaxText(t, components.SyntaxProps{Code: src, Language: "ruby"})
	if got != src {
		t.Fatalf("unknown language should pass through unchanged.\n got: %q\nwant: %q", got, src)
	}
}

func TestSyntaxGoColorsKeywordStringNumber(t *testing.T) {
	src := `package main
func add(a int) int { return a + 42 }
var s = "hello"
// trailing comment`

	got := syntaxText(t, components.SyntaxProps{Code: src, Language: components.SyntaxGo})

	// Each token kind must contribute at least one ANSI sequence.
	if !strings.Contains(got, styles.Cyan.ToANSI(styles.Foreground)) {
		t.Fatalf("expected cyan keyword color in Go output, got %q", got)
	}
	if !strings.Contains(got, styles.Green.ToANSI(styles.Foreground)) {
		t.Fatalf("expected green string color in Go output, got %q", got)
	}
	if !strings.Contains(got, styles.Yellow.ToANSI(styles.Foreground)) {
		t.Fatalf("expected yellow number color in Go output, got %q", got)
	}
	if !strings.Contains(got, styles.DimCode()) {
		t.Fatalf("expected dim comment ANSI code in Go output, got %q", got)
	}
	if !strings.Contains(got, "package") || !strings.Contains(got, "func") {
		t.Fatalf("Go keywords missing from output: %q", got)
	}
	if !strings.Contains(got, "trailing comment") {
		t.Fatalf("comment text missing from output: %q", got)
	}
}

func TestSyntaxGoMultiLine(t *testing.T) {
	src := "var x = 1\nvar y = 2"
	got := syntaxText(t, components.SyntaxProps{Code: src, Language: components.SyntaxGo})
	if !strings.Contains(got, "\n") {
		t.Fatalf("multi-line input should preserve newlines, got %q", got)
	}
	// Both keywords must be styled.
	cyanCount := strings.Count(got, styles.Cyan.ToANSI(styles.Foreground))
	if cyanCount < 2 {
		t.Fatalf("expected at least 2 keyword colorizations, got %d in %q", cyanCount, got)
	}
}

func TestSyntaxGoBlockComment(t *testing.T) {
	src := "/* doc */ var x = 1"
	got := syntaxText(t, components.SyntaxProps{Code: src, Language: components.SyntaxGo})
	if !strings.Contains(got, styles.DimCode()) {
		t.Fatalf("block comment should be dimmed, got %q", got)
	}
	if !strings.Contains(got, "doc") {
		t.Fatalf("comment text lost: %q", got)
	}
}

func TestSyntaxGoBacktickString(t *testing.T) {
	src := "var s = `raw`"
	got := syntaxText(t, components.SyntaxProps{Code: src, Language: components.SyntaxGo})
	if !strings.Contains(got, styles.Green.ToANSI(styles.Foreground)) {
		t.Fatalf("backtick string should be green, got %q", got)
	}
}

func TestSyntaxGoNonKeywordIdentifierUncolored(t *testing.T) {
	src := "myVariable"
	got := syntaxText(t, components.SyntaxProps{Code: src, Language: components.SyntaxGo})
	// Plain identifiers must not gain ANSI escapes.
	if strings.Contains(got, "\x1b[") {
		t.Fatalf("plain identifier should not be colorized, got %q", got)
	}
	if got != "myVariable" {
		t.Fatalf("unexpected transform of plain identifier: %q", got)
	}
}

func TestSyntaxJSONKeyValueNumberLiteral(t *testing.T) {
	src := `{"name": "ada", "age": 36, "active": true, "ref": null}`
	got := syntaxText(t, components.SyntaxProps{Code: src, Language: components.SyntaxJSON})

	if !strings.Contains(got, styles.Blue.ToANSI(styles.Foreground)) {
		t.Fatalf("expected blue key color in JSON output, got %q", got)
	}
	if !strings.Contains(got, styles.Green.ToANSI(styles.Foreground)) {
		t.Fatalf("expected green string color in JSON output, got %q", got)
	}
	if !strings.Contains(got, styles.Yellow.ToANSI(styles.Foreground)) {
		t.Fatalf("expected yellow number color in JSON output, got %q", got)
	}
	if !strings.Contains(got, styles.Cyan.ToANSI(styles.Foreground)) {
		t.Fatalf("expected cyan literal color in JSON output, got %q", got)
	}
	for _, tok := range []string{"name", "ada", "36", "true", "null"} {
		if !strings.Contains(got, tok) {
			t.Fatalf("token %q missing from JSON output: %q", tok, got)
		}
	}
}

func TestSyntaxJSONNegativeNumber(t *testing.T) {
	src := `{"v": -1.5e2}`
	got := syntaxText(t, components.SyntaxProps{Code: src, Language: components.SyntaxJSON})
	if !strings.Contains(got, "-1.5e2") {
		t.Fatalf("negative/exponent number lost: %q", got)
	}
	if !strings.Contains(got, styles.Yellow.ToANSI(styles.Foreground)) {
		t.Fatalf("number should be colorized: %q", got)
	}
}

func TestSyntaxYAMLColorizesKeysCommentsLiterals(t *testing.T) {
	src := "# top-level\n" +
		"name: ink-go\n" +
		"version: 1.5\n" +
		"published: true\n" +
		"description: \"Go port of Ink\"\n"

	got := syntaxText(t, components.SyntaxProps{Code: src, Language: components.SyntaxYAML})

	if !strings.Contains(got, styles.Blue.ToANSI(styles.Foreground)) {
		t.Fatalf("YAML keys should render with the blue key color: %q", got)
	}
	if !strings.Contains(got, styles.Green.ToANSI(styles.Foreground)) {
		t.Fatalf("YAML quoted strings should render with the green string color: %q", got)
	}
	if !strings.Contains(got, styles.Yellow.ToANSI(styles.Foreground)) {
		t.Fatalf("YAML numeric literals should render with the yellow number color: %q", got)
	}
	if !strings.Contains(got, styles.Cyan.ToANSI(styles.Foreground)) {
		t.Fatalf("YAML true/false/null should render with the cyan literal color: %q", got)
	}
	// Comments should be present somewhere — Dim() emits ESC[2m.
	if !strings.Contains(got, "\x1b[2m") {
		t.Fatalf("YAML comments should be dimmed: %q", got)
	}
}

func TestSyntaxYAMLListMarkerColorized(t *testing.T) {
	src := "- alpha\n- beta\n- gamma\n"
	got := syntaxText(t, components.SyntaxProps{Code: src, Language: components.SyntaxYAML})
	if !strings.Contains(got, styles.Magenta.ToANSI(styles.Foreground)) {
		t.Fatalf("YAML list markers should render with the magenta marker color: %q", got)
	}
}

func TestSyntaxMarkdownColorsHeadingsCodeAndBold(t *testing.T) {
	src := "# Heading\n" +
		"\n" +
		"Body text with `inline code` and **bold span** and _italic_.\n" +
		"\n" +
		"```go\n" +
		"fmt.Println(\"hi\")\n" +
		"```\n"

	got := syntaxText(t, components.SyntaxProps{Code: src, Language: components.SyntaxMarkdown})

	if !strings.Contains(got, styles.Cyan.ToANSI(styles.Foreground)) {
		t.Fatalf("Markdown headings should render with the cyan color: %q", got)
	}
	if !strings.Contains(got, styles.Green.ToANSI(styles.Foreground)) {
		t.Fatalf("Markdown code spans/blocks should render with the green color: %q", got)
	}
	if !strings.Contains(got, styles.BoldCode()) {
		t.Fatalf("Markdown bold span should emit BoldCode(): %q", got)
	}
	if !strings.Contains(got, styles.ItalicCode()) {
		t.Fatalf("Markdown italic span should emit ItalicCode(): %q", got)
	}
}

func TestSyntaxMarkdownLinkSplitsTextAndUrl(t *testing.T) {
	src := "See [Ink](https://github.com/vadimdemedes/ink) for context.\n"
	got := syntaxText(t, components.SyntaxProps{Code: src, Language: components.SyntaxMarkdown})

	if !strings.Contains(got, styles.Cyan.ToANSI(styles.Foreground)) {
		t.Fatalf("Markdown link text should render in cyan: %q", got)
	}
	// URL portion should be dimmed.
	if !strings.Contains(got, "\x1b[2m") {
		t.Fatalf("Markdown link url should be dimmed: %q", got)
	}
}

func TestSyntaxBashColorsKeywordsStringsAndVariables(t *testing.T) {
	// Variable expansions inside double-quoted strings are intentionally
	// consumed by the string match (RE2 alternation is left-to-right and
	// rewriting inside string contents would require a second pass);
	// keep the test variable bare so the variable colorizer can fire.
	src := "#!/bin/bash\n" +
		"NAME=world\n" +
		"if [ -n $NAME ]; then\n" +
		"  echo \"Hello, world!\"\n" +
		"fi\n"

	got := syntaxText(t, components.SyntaxProps{Code: src, Language: components.SyntaxBash})

	if !strings.Contains(got, styles.Cyan.ToANSI(styles.Foreground)) {
		t.Fatalf("Bash keywords should render in cyan: %q", got)
	}
	if !strings.Contains(got, styles.Green.ToANSI(styles.Foreground)) {
		t.Fatalf("Bash strings should render in green: %q", got)
	}
	if !strings.Contains(got, styles.Yellow.ToANSI(styles.Foreground)) {
		t.Fatalf("Bash $variables should render in yellow: %q", got)
	}
	// Comment line should be dimmed.
	if !strings.Contains(got, "\x1b[2m") {
		t.Fatalf("Bash comments should be dimmed: %q", got)
	}
}

func TestSyntaxNodeIsTextElement(t *testing.T) {
	node := components.Syntax(components.SyntaxProps{Code: "x", Language: components.SyntaxGo})
	if node.ElementType != "text" {
		t.Fatalf("Syntax should render as text element, got %q", node.ElementType)
	}
}

// --- Python -----------------------------------------------------------------

func TestSyntaxPythonColorsKeywordsStringsNumbersComments(t *testing.T) {
	src := "# greet\n" +
		"def greet(name: str) -> str:\n" +
		"    return \"hello, \" + name\n" +
		"x = 42\n"

	got := syntaxText(t, components.SyntaxProps{Code: src, Language: components.SyntaxPython})

	if !strings.Contains(got, styles.Cyan.ToANSI(styles.Foreground)) {
		t.Fatalf("Python keywords should render in cyan: %q", got)
	}
	if !strings.Contains(got, styles.Green.ToANSI(styles.Foreground)) {
		t.Fatalf("Python strings should render in green: %q", got)
	}
	if !strings.Contains(got, styles.Yellow.ToANSI(styles.Foreground)) {
		t.Fatalf("Python numbers should render in yellow: %q", got)
	}
	if !strings.Contains(got, styles.DimCode()) {
		t.Fatalf("Python comments should be dimmed: %q", got)
	}
	for _, tok := range []string{"def", "return", "greet", "hello", "42"} {
		if !strings.Contains(got, tok) {
			t.Fatalf("token %q missing from Python output: %q", tok, got)
		}
	}
}

func TestSyntaxPythonTripleQuotedString(t *testing.T) {
	src := "doc = \"\"\"multi\nline\"\"\""
	got := syntaxText(t, components.SyntaxProps{Code: src, Language: components.SyntaxPython})
	if !strings.Contains(got, styles.Green.ToANSI(styles.Foreground)) {
		t.Fatalf("Python triple-quoted string should be green: %q", got)
	}
	if !strings.Contains(got, "multi") || !strings.Contains(got, "line") {
		t.Fatalf("triple-quoted body lost: %q", got)
	}
}

func TestSyntaxPythonDecoratorColorized(t *testing.T) {
	src := "@staticmethod\ndef m(): pass"
	got := syntaxText(t, components.SyntaxProps{Code: src, Language: components.SyntaxPython})
	if !strings.Contains(got, styles.Magenta.ToANSI(styles.Foreground)) {
		t.Fatalf("Python decorator should render in magenta: %q", got)
	}
	if !strings.Contains(got, "@staticmethod") {
		t.Fatalf("decorator text lost: %q", got)
	}
}

// --- Rust -------------------------------------------------------------------

func TestSyntaxRustColorsKeywordsStringsNumbersComments(t *testing.T) {
	src := "// entry point\n" +
		"fn add(a: i32, b: i32) -> i32 {\n" +
		"    let s = \"hi\";\n" +
		"    return a + b + 7;\n" +
		"}\n"

	got := syntaxText(t, components.SyntaxProps{Code: src, Language: components.SyntaxRust})

	if !strings.Contains(got, styles.Cyan.ToANSI(styles.Foreground)) {
		t.Fatalf("Rust keywords should render in cyan: %q", got)
	}
	if !strings.Contains(got, styles.Green.ToANSI(styles.Foreground)) {
		t.Fatalf("Rust strings should render in green: %q", got)
	}
	if !strings.Contains(got, styles.Yellow.ToANSI(styles.Foreground)) {
		t.Fatalf("Rust numbers should render in yellow: %q", got)
	}
	if !strings.Contains(got, styles.DimCode()) {
		t.Fatalf("Rust comments should be dimmed: %q", got)
	}
	for _, kw := range []string{"fn", "let", "return"} {
		if !strings.Contains(got, kw) {
			t.Fatalf("Rust keyword %q missing from output: %q", kw, got)
		}
	}
}

func TestSyntaxRustLifetimeAndAttribute(t *testing.T) {
	src := "#[derive(Debug)]\nstruct Foo<'a> { name: &'a str }"
	got := syntaxText(t, components.SyntaxProps{Code: src, Language: components.SyntaxRust})

	// Attribute is magenta, lifetime is yellow.
	if !strings.Contains(got, styles.Magenta.ToANSI(styles.Foreground)) {
		t.Fatalf("Rust attribute should be magenta: %q", got)
	}
	if !strings.Contains(got, styles.Yellow.ToANSI(styles.Foreground)) {
		t.Fatalf("Rust lifetime should be yellow: %q", got)
	}
	if !strings.Contains(got, "#[derive(Debug)]") {
		t.Fatalf("attribute text lost: %q", got)
	}
	if !strings.Contains(got, "'a") {
		t.Fatalf("lifetime text lost: %q", got)
	}
}

func TestSyntaxRustBlockCommentDimmed(t *testing.T) {
	src := "/* doc */ let x = 1;"
	got := syntaxText(t, components.SyntaxProps{Code: src, Language: components.SyntaxRust})
	if !strings.Contains(got, styles.DimCode()) {
		t.Fatalf("Rust block comment should be dimmed: %q", got)
	}
	if !strings.Contains(got, "doc") {
		t.Fatalf("comment text lost: %q", got)
	}
}

// --- SQL --------------------------------------------------------------------

func TestSyntaxSQLColorsKeywordsStringsNumbersComments(t *testing.T) {
	src := "-- find adults\n" +
		"SELECT name, age FROM people WHERE age >= 18 AND active = 'yes';\n"

	got := syntaxText(t, components.SyntaxProps{Code: src, Language: components.SyntaxSQL})

	if !strings.Contains(got, styles.Cyan.ToANSI(styles.Foreground)) {
		t.Fatalf("SQL keywords should render in cyan: %q", got)
	}
	if !strings.Contains(got, styles.Green.ToANSI(styles.Foreground)) {
		t.Fatalf("SQL strings should render in green: %q", got)
	}
	if !strings.Contains(got, styles.Yellow.ToANSI(styles.Foreground)) {
		t.Fatalf("SQL numbers should render in yellow: %q", got)
	}
	if !strings.Contains(got, styles.DimCode()) {
		t.Fatalf("SQL line comment should be dimmed: %q", got)
	}
	for _, tok := range []string{"SELECT", "FROM", "WHERE", "AND", "yes"} {
		if !strings.Contains(got, tok) {
			t.Fatalf("token %q missing from SQL output: %q", tok, got)
		}
	}
}

func TestSyntaxSQLKeywordsAreCaseInsensitive(t *testing.T) {
	src := "select 1 from dual"
	got := syntaxText(t, components.SyntaxProps{Code: src, Language: components.SyntaxSQL})
	// Lowercase keywords must still be colorized.
	if !strings.Contains(got, styles.Cyan.ToANSI(styles.Foreground)) {
		t.Fatalf("SQL lowercase keywords should still colorize: %q", got)
	}
	// And the original casing must be preserved verbatim in the output.
	if !strings.Contains(got, "select") || !strings.Contains(got, "from") {
		t.Fatalf("SQL output dropped original casing: %q", got)
	}
}

func TestSyntaxSQLBlockComment(t *testing.T) {
	src := "/* note */ SELECT 1"
	got := syntaxText(t, components.SyntaxProps{Code: src, Language: components.SyntaxSQL})
	if !strings.Contains(got, styles.DimCode()) {
		t.Fatalf("SQL block comment should be dimmed: %q", got)
	}
	if !strings.Contains(got, "note") {
		t.Fatalf("comment text lost: %q", got)
	}
}

// --- JavaScript -------------------------------------------------------------

func TestSyntaxJavaScriptColorsKeywordsStringsNumbersComments(t *testing.T) {
	src := "// entry point\n" +
		"const greet = function (name) {\n" +
		"    return \"hello, \" + name;\n" +
		"};\n" +
		"let x = 42;\n"

	got := syntaxText(t, components.SyntaxProps{Code: src, Language: components.SyntaxJavaScript})

	if !strings.Contains(got, styles.Cyan.ToANSI(styles.Foreground)) {
		t.Fatalf("JS keywords should render in cyan: %q", got)
	}
	if !strings.Contains(got, styles.Green.ToANSI(styles.Foreground)) {
		t.Fatalf("JS strings should render in green: %q", got)
	}
	if !strings.Contains(got, styles.Yellow.ToANSI(styles.Foreground)) {
		t.Fatalf("JS numbers should render in yellow: %q", got)
	}
	if !strings.Contains(got, styles.DimCode()) {
		t.Fatalf("JS comments should be dimmed: %q", got)
	}
	for _, kw := range []string{"const", "function", "return", "let"} {
		if !strings.Contains(got, kw) {
			t.Fatalf("JS keyword %q missing from output: %q", kw, got)
		}
	}
}

func TestSyntaxJavaScriptTemplateLiteral(t *testing.T) {
	src := "let s = `hello ${name}!`"
	got := syntaxText(t, components.SyntaxProps{Code: src, Language: components.SyntaxJavaScript})
	if !strings.Contains(got, styles.Green.ToANSI(styles.Foreground)) {
		t.Fatalf("JS template literal should be green: %q", got)
	}
	if !strings.Contains(got, "hello ${name}!") {
		t.Fatalf("template body lost: %q", got)
	}
}

func TestSyntaxJavaScriptRegexLiteral(t *testing.T) {
	src := "const re = /[a-z]+/gi"
	got := syntaxText(t, components.SyntaxProps{Code: src, Language: components.SyntaxJavaScript})
	if !strings.Contains(got, styles.Magenta.ToANSI(styles.Foreground)) {
		t.Fatalf("JS regex literal should be magenta: %q", got)
	}
	if !strings.Contains(got, "/[a-z]+/gi") {
		t.Fatalf("regex literal text lost: %q", got)
	}
}

func TestSyntaxJavaScriptDivisionNotMistakenForRegex(t *testing.T) {
	// In `let q = x / 2` the '/' is division, not the start of a regex.
	// Our pattern requires a leading boundary char like '=' so the second
	// '/' (closing) won't be picked up as opening a regex either.
	src := "let q = x\nq = q / 2\nq = q / 4"
	got := syntaxText(t, components.SyntaxProps{Code: src, Language: components.SyntaxJavaScript})
	if strings.Contains(got, styles.Magenta.ToANSI(styles.Foreground)) {
		t.Fatalf("division operator must not trigger regex tinting: %q", got)
	}
}

func TestSyntaxJavaScriptBlockCommentDimmed(t *testing.T) {
	src := "/* doc */ var x = 1"
	got := syntaxText(t, components.SyntaxProps{Code: src, Language: components.SyntaxJavaScript})
	if !strings.Contains(got, styles.DimCode()) {
		t.Fatalf("JS block comment should be dimmed: %q", got)
	}
	if !strings.Contains(got, "doc") {
		t.Fatalf("comment text lost: %q", got)
	}
}

// --- Diff -------------------------------------------------------------------

func TestSyntaxDiffColorsAddRemoveHunkHeader(t *testing.T) {
	src := "diff --git a/x b/x\n" +
		"index 0000000..1111111 100644\n" +
		"--- a/x\n" +
		"+++ b/x\n" +
		"@@ -1,3 +1,3 @@\n" +
		" context line\n" +
		"-removed line\n" +
		"+added line\n"

	got := syntaxText(t, components.SyntaxProps{Code: src, Language: components.SyntaxDiff})

	if !strings.Contains(got, styles.Green.ToANSI(styles.Foreground)) {
		t.Fatalf("Diff add lines should be green: %q", got)
	}
	if !strings.Contains(got, styles.Red.ToANSI(styles.Foreground)) {
		t.Fatalf("Diff remove lines should be red: %q", got)
	}
	if !strings.Contains(got, styles.Cyan.ToANSI(styles.Foreground)) {
		t.Fatalf("Diff hunk header should be cyan: %q", got)
	}
	// File headers / metadata should be dimmed.
	if !strings.Contains(got, "\x1b[2m") {
		t.Fatalf("Diff file header should be dimmed: %q", got)
	}
	for _, tok := range []string{"context line", "removed line", "added line", "@@ -1,3 +1,3 @@"} {
		if !strings.Contains(got, tok) {
			t.Fatalf("token %q missing from Diff output: %q", tok, got)
		}
	}
}

func TestSyntaxDiffPlusPlusPlusNotTreatedAsAddLine(t *testing.T) {
	// The '+++ b/foo' file header must be dimmed, not green-painted as
	// an add line — otherwise file metadata leaks into the change set.
	src := "+++ b/foo\n+ real add\n"
	got := syntaxText(t, components.SyntaxProps{Code: src, Language: components.SyntaxDiff})

	if !strings.Contains(got, "+++ b/foo") {
		t.Fatalf("file header text lost: %q", got)
	}
	// Dim ANSI must be present for the file header.
	if !strings.Contains(got, "\x1b[2m") {
		t.Fatalf("'+++' file header should be dimmed: %q", got)
	}
	// And the real add line should be green.
	if !strings.Contains(got, styles.Green.ToANSI(styles.Foreground)) {
		t.Fatalf("real '+' add line should be green: %q", got)
	}
}

func TestSyntaxDiffEmptyInput(t *testing.T) {
	got := syntaxText(t, components.SyntaxProps{Code: "", Language: components.SyntaxDiff})
	if got != "" {
		t.Fatalf("empty diff should render empty, got %q", got)
	}
}

func TestSyntaxDiffContextLineUntouched(t *testing.T) {
	src := " plain context line\n"
	got := syntaxText(t, components.SyntaxProps{Code: src, Language: components.SyntaxDiff})
	if strings.Contains(got, "\x1b[") {
		t.Fatalf("context line must not gain ANSI codes: %q", got)
	}
}
