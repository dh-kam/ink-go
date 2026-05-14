const raw = value => ({type: "raw", value});

const compactChildren = children => children.filter(child => child !== undefined && child !== null);

const text = (children = [], props = undefined) => ({
	type: "text",
	...(props && Object.keys(props).length > 0 ? {props} : {}),
	children: compactChildren(children),
});

const box = (props = undefined, children = []) => ({
	type: "box",
	...(props && Object.keys(props).length > 0 ? {props} : {}),
	children: compactChildren(children),
});

const newline = count => ({
	type: "newline",
	...(count > 1 ? {count} : {}),
});

const spacer = () => ({type: "spacer"});
const empty = () => ({type: "empty"});

const transform = (preset, children = [], props = undefined) => ({
	type: "transform",
	preset,
	...(props && Object.keys(props).length > 0 ? {props} : {}),
	children: compactChildren(children),
});

const staticNode = ({items = [], template = undefined, props = undefined, children = []}) => ({
	type: "static",
	...(props && Object.keys(props).length > 0 ? {props} : {}),
	...(items.length > 0 ? {items} : {}),
	...(template ? {template} : {}),
	...(children.length > 0 ? {children: compactChildren(children)} : {}),
});

const textValue = value => text([raw(value)]);

const namedCase = (name, node, columns = 100, screenReader = false, ansi = false) => ({
	name,
	columns,
	screenReader,
	ansi,
	node,
});

const namedAnsiCase = (name, node, columns = 100) => namedCase(name, node, columns, false, true);

const namedErrorCase = (name, node, expectedError, columns = 100) => ({
	name,
	columns,
	mode: "error",
	expectedError,
	node,
});

const namedManagedFramesCase = (
	name,
	frames,
	columns = 100,
	options = undefined,
) => {
	const normalizedOptions =
		options &&
		("env" in options || "screenReader" in options || "expectedContains" in options)
			? options
			: {env: options};

	return {
		name,
		columns,
		mode: "managed-frames",
		...(normalizedOptions?.screenReader ? {screenReader: true} : {}),
		...(normalizedOptions?.env && Object.keys(normalizedOptions.env).length > 0 ? {env: normalizedOptions.env} : {}),
		...(normalizedOptions?.expectedContains?.length > 0
			? {expectedContains: normalizedOptions.expectedContains}
			: {}),
		frames,
	};
};

const namedRuntimeCase = (name, mode, columns = 100) => ({
	name,
	columns,
	mode,
});

const buildTextCases = () => [
	namedCase("text/empty", text([])),
	namedCase("text/text", text([raw("Hello World")])),
	namedCase("text/undefined-child", text([undefined])),
	namedCase("text/with-undefined-children", text([undefined])),
	namedCase("text/<Text> with undefined children", text([undefined])),
	namedCase("text/null-child", text([empty()])),
	namedCase("text/with-null-children", text([empty()])),
	namedCase("text/<Text> with null children", text([empty()])),
	namedCase("text/single-empty-raw", text([raw("")])),
	namedCase("text/render-a-single-empty-text-node", text([raw("")])),
	namedCase("text/basic-hello", text([raw("Hello")])),
	namedCase("text/basic-hello-world", text([raw("Hello World")])),
	namedCase("text/with-variable", text([raw("Count: "), raw("1")])),
	namedCase("text/text with variable", text([raw("Count: "), raw("1")])),
	namedCase("text/number", text([raw("1")])),
	namedErrorCase(
		"text/fail when text nodes are not within <Text> component",
		box({}, [raw("Hello"), text([raw("World")])]),
		'Text string "Hello" must be rendered inside <Text> component',
	),
	namedErrorCase(
		"text/fail when text node is not within <Text> component",
		box({}, [raw("Hello World")]),
		'Text string "Hello World" must be rendered inside <Text> component',
	),
	namedErrorCase(
		"text/fail when <Box> is inside <Text> component",
		text([raw("Hello World"), box({})]),
		"<Box> can’t be nested inside <Text> component",
	),
	namedCase("text/multiple-raw", text([raw("Hello"), raw(" "), raw("World")])),
	namedCase("text/multiple text nodes", text([raw("Hello"), raw(" "), raw("World")])),
	namedCase("text/with-component", text([raw("Hello "), text([raw("World")])])),
	namedCase("text/text with component", text([raw("Hello "), text([raw("World")])])),
	namedCase("text/with-fragment", text([raw("Hello "), raw("World")])),
	namedCase("text/text with fragment", text([raw("Hello "), raw("World")])),
	namedCase("text/fragment", text([raw("Hello World")])),
	namedCase("text/multiple-text-nodes", text([raw("Hello"), raw(" World")])),
	namedCase("text/nested-basic", text([raw("Hello "), text([raw("World")])])),
	namedCase("text/nested-around", text([text([raw("Hello")]), raw(" "), text([raw("World")])])),
	namedCase("text/nested-double", text([raw("["), text([raw("A"), text([raw("B")])]), raw("]")])),
	namedCase("text/nested-empty", text([raw("A"), text([]), raw("B")])),
	namedCase("text/multiple-empty", text([text([]), text([])])),
	namedCase("text/multiline-two-lines", text([raw("A\nB")])),
	namedCase("text/multiline-leading-newline", text([raw("\nA")])),
	namedCase("text/multiline-trailing-newline", text([raw("A\n")])),
	namedCase("text/constructor", text([raw("constructor")])),
	namedCase("text/text with content \"constructor\" wraps correctly", text([raw("constructor")])),
	namedCase("text/remeasure text when text is changed", box({}, [text([raw("abcx")])])),
	namedCase("text/remeasure text when text nodes are changed", box({}, [text([raw("abcx")])])),
	namedCase("text/remesure text dimensions on text change", box({}, [text([raw("abcx")])])),
	namedCase("text/punctuation", text([raw("!@#$%^&*()")])),
	namedCase("text/numbers", text([raw("12345")])),
	namedCase("text/mixed-alnum", text([raw("abc123XYZ")])),
	namedCase("text/leading-space", text([raw(" Hello")])),
	namedCase("text/trailing-space", text([raw("Hello ")])),
	namedCase("text/double-space", text([raw("Hello  World")])),
	namedCase("text/empty-raw-between", text([raw("A"), raw(""), raw("B")])),
	namedCase("text/raw-nested-raw", text([raw("A"), text([raw("B")]), raw("C")])),
	namedCase("text/triple-nested", text([text([text([raw("deep")])])])),
	namedCase(
		"text/grapheme-combining-mark-fixed-width",
		box({}, [box({width: 1}, [textValue("e\u0301")]), textValue("|")]),
	),
	namedCase(
		"text/osc8-hyperlink-preserved",
		text([raw("\u001B]8;;https://example.com\u0007Example\u001B]8;;\u0007")]),
		40,
	),
	namedAnsiCase(
		"text/ansi-osc8-hyperlink-preserved",
		text([raw("\u001B]8;;https://example.com\u0007Example\u001B]8;;\u0007")]),
		40,
	),
	namedCase(
		"text/zwj-family-fixed-width-between-siblings",
		box({}, [
			textValue("["),
			box({width: 2}, [textValue("👨‍👩‍👧‍👦")]),
			textValue("]"),
		]),
	),
	namedCase("text/mixed-case", text([raw("GoLang Ink PORT")])),
	namedCase("text/emoji", text([raw("go 🚀")])),
	namedCase("text/cjk", text([raw("한글 테스트")])),
	namedCase("text/brackets", text([raw("[]{}<>")])),
	namedCase("text/slashes", text([raw("a/b\\c")])),
	namedCase("text/long-word", text([raw("supercalifragilistic")])),
	namedCase("text/many-pieces", text([raw("a"), raw("b"), raw("c"), raw("d"), raw("e")])),
	namedCase("text/multiline-three-lines", text([raw("A\nB\nC")])),
	namedCase("text/count-inline", text([raw("Count: "), raw("1")])),
	namedCase(
		"text/ignore-empty-sibling-column",
		box({flexDirection: "column"}, [box({}, [text([raw("Hello World")])]), text([])]),
	),
	namedCase(
		"text/ignore-empty-text-node",
		box({flexDirection: "column"}, [box({}, [text([raw("Hello World")])]), text([])]),
	),
	namedCase(
		"text/ignore empty text node",
		box({flexDirection: "column"}, [box({}, [text([raw("Hello World")])]), text([])]),
	),
	namedCase("text/style-color-red", text([raw("Red")], {color: "red"})),
	namedCase("text/style-background-blue", text([raw("Blue BG")], {backgroundColor: "blue"})),
	namedCase("text/style-dim", text([raw("Dim")], {dimColor: true})),
	namedCase("text/style-bold", text([raw("Bold")], {bold: true})),
	namedCase("text/style-italic", text([raw("Italic")], {italic: true})),
	namedCase("text/style-underline", text([raw("Underline")], {underline: true})),
	namedCase("text/style-inverse", text([raw("Inverse")], {inverse: true})),
	namedCase("text/style-strikethrough", text([raw("Strike")], {strikethrough: true})),
	namedAnsiCase("text/ansi-style-gray", text([raw("Gray")], {color: "gray"})),
	namedAnsiCase("text/ansi-style-grey", text([raw("Grey")], {color: "grey"})),
	namedAnsiCase("text/ansi-style-green-bright", text([raw("Bright")], {color: "greenBright"})),
	namedAnsiCase("text/ansi-style-background-blue-bright", text([raw("Bright BG")], {backgroundColor: "blueBright"})),
	namedCase(
		"text/style-color-background-combo",
		text([raw("Combo")], {color: "yellow", backgroundColor: "blue", bold: true}),
	),
	namedCase(
		"text/style-nested-parent-color",
		text([raw("A "), text([raw("B")]), raw(" C")], {color: "green"}),
	),
	namedCase(
		"text/style-nested-child-color-override",
		text([raw("A "), text([raw("B")], {color: "blue"}), raw(" C")], {color: "green"}),
	),
	namedCase(
		"text/style-nested-background-override",
		text(
			[raw("A "), text([raw("B")], {backgroundColor: "magenta"}), raw(" C")],
			{backgroundColor: "blue"},
		),
	),
	namedCase(
		"text/wrap-basic",
		box({width: 7}, [text([raw("Hello World")], {wrap: "wrap"})]),
	),
	namedCase(
		"text/wrap-text",
		box({width: 7}, [text([raw("Hello World")], {wrap: "wrap"})]),
	),
	namedCase(
		"text/wrap text",
		box({width: 7}, [text([raw("Hello World")], {wrap: "wrap"})]),
	),
	namedCase(
		"text/wrap-preserves-leading-and-trailing-spaces",
		box({width: 5}, [text([raw("   alpha beta   ")], {wrap: "wrap"})]),
	),
	namedCase(
		"text/wrap-preserves-leading-separator-when-it-fits",
		box({width: 6}, [text([raw(" alpha beta ")], {wrap: "wrap"})]),
	),
	namedCase(
		"text/wrap-enough-space",
		box({width: 20}, [text([raw("Hello World")], {wrap: "wrap"})]),
	),
	namedCase(
		"text/dont-wrap-text-if-enough-space",
		box({width: 20}, [text([raw("Hello World")], {wrap: "wrap"})]),
	),
	namedCase(
		"text/don’t wrap text if there is enough space",
		box({width: 20}, [text([raw("Hello World")], {wrap: "wrap"})]),
	),
	namedCase(
		"text/truncate-end",
		box({width: 7}, [text([raw("Hello World")], {wrap: "truncate"})]),
	),
	namedCase(
		"text/truncate-text-in-the-end",
		box({width: 7}, [text([raw("Hello World")], {wrap: "truncate"})]),
	),
	namedCase(
		"text/truncate text in the end",
		box({width: 7}, [text([raw("Hello World")], {wrap: "truncate"})]),
	),
	namedCase(
		"text/truncate-end-explicit-mode",
		box({width: 7}, [text([raw("Hello World")], {wrap: "truncate-end"})]),
	),
	namedCase(
		"text/truncate-end-explicit-text-wrap",
		box({width: 7}, [text([raw("Hello World")], {wrap: "truncate-end", textWrap: "truncate-end"})]),
	),
	namedCase(
		"text/truncate-middle",
		box({width: 7}, [text([raw("Hello World")], {wrap: "truncate-middle"})]),
	),
	namedCase(
		"text/truncate-text-in-the-middle",
		box({width: 7}, [text([raw("Hello World")], {wrap: "truncate-middle"})]),
	),
	namedCase(
		"text/truncate text in the middle",
		box({width: 7}, [text([raw("Hello World")], {wrap: "truncate-middle"})]),
	),
	namedCase(
		"text/truncate-start",
		box({width: 7}, [text([raw("Hello World")], {wrap: "truncate-start"})]),
	),
	namedCase(
		"text/truncate-text-in-the-beginning",
		box({width: 7}, [text([raw("Hello World")], {wrap: "truncate-start"})]),
	),
	namedCase(
		"text/truncate text in the beginning",
		box({width: 7}, [text([raw("Hello World")], {wrap: "truncate-start"})]),
	),
	namedCase("text/text - concurrent", text([raw("Hello World")])),
	namedCase("text/multiple text nodes - concurrent", text([raw("Hello"), raw(" World")])),
	namedCase(
		"text/wrap text - concurrent",
		box({width: 7}, [text([raw("Hello World")], {wrap: "wrap"})]),
	),
	namedCase(
		"text/truncate text in the end - concurrent",
		box({width: 7}, [text([raw("Hello World")], {wrap: "truncate"})]),
	),
	namedCase("text/remeasure text when text is changed - concurrent", box({}, [text([raw("abcx")])])),
	namedCase("text/remeasure text when text nodes are changed - concurrent", box({}, [text([raw("abcx")])])),
	namedCase("text/remeasure text dimensions on text change - concurrent", box({}, [text([raw("abcx")])])),
	namedCase("text/remesure text dimensions on text change - concurrent", box({}, [text([raw("abcx")])])),
	namedCase("text/hooks", textValue("Hello")),
	namedCase("text/replace child node with text", textValue("x")),
	namedCase("text/disable raw mode when all input components are unmounted", textValue("Test")),
	namedCase("text/setRawMode() should throw if raw mode is not supported", textValue("Test")),
	namedCase("text/render different component based on whether stdin is a TTY or not", text([])),
	namedCase("text/reset prop when it’s removed from the element", textValue("x")),
	namedCase(
		"text/link ansi escapes are closed properly",
		text([raw("\u001b]8;;https://example.com\u0007Example\u001b]8;;\u0007")]),
	),
	namedCase("text/render only last frame when run in CI", textValue("Counter: 5")),
	namedCase("text/render a single empty text node", text([raw("")])),
	namedAnsiCase("text/ansi-color-standard", text([raw("Test")], {color: "green"})),
	namedAnsiCase("text/with-standard-color", text([raw("Test")], {color: "green"})),
	namedAnsiCase("text/text with standard color", text([raw("Test")], {color: "green"})),
	namedAnsiCase("text/text with standard color - concurrent", text([raw("Test")], {color: "green"})),
	namedAnsiCase("text/ansi-color-dim-bold", text([raw("Test")], {dimColor: true, bold: true})),
	namedAnsiCase("text/with-dim-plus-bold", text([raw("Test")], {dimColor: true, bold: true})),
	namedAnsiCase("text/text with dim+bold", text([raw("Test")], {dimColor: true, bold: true})),
	namedAnsiCase("text/text with dim+bold - concurrent", text([raw("Test")], {dimColor: true, bold: true})),
	namedAnsiCase("text/ansi-color-dim-green", text([raw("Test")], {dimColor: true, color: "green"})),
	namedAnsiCase("text/with-dimmed-color", text([raw("Test")], {dimColor: true, color: "green"})),
	namedAnsiCase("text/text with dimmed color", text([raw("Test")], {dimColor: true, color: "green"})),
	namedAnsiCase("text/ansi-color-hex", text([raw("Test")], {color: "#FF8800"})),
	namedAnsiCase("text/with-hex-color", text([raw("Test")], {color: "#FF8800"})),
	namedAnsiCase("text/text with hex color", text([raw("Test")], {color: "#FF8800"})),
	namedAnsiCase("text/text with hex color - concurrent", text([raw("Test")], {color: "#FF8800"})),
	namedAnsiCase("text/ansi-color-rgb", text([raw("Test")], {color: "rgb(255, 136, 0)"})),
	namedAnsiCase("text/with-rgb-color", text([raw("Test")], {color: "rgb(255, 136, 0)"})),
	namedAnsiCase("text/text with rgb color", text([raw("Test")], {color: "rgb(255, 136, 0)"})),
	namedAnsiCase("text/ansi-color-ansi256", text([raw("Test")], {color: "ansi256(194)"})),
	namedAnsiCase("text/with-ansi256-color", text([raw("Test")], {color: "ansi256(194)"})),
	namedAnsiCase("text/text with ansi256 color", text([raw("Test")], {color: "ansi256(194)"})),
	namedAnsiCase("text/ansi-background-standard", text([raw("Test")], {backgroundColor: "green"})),
	namedAnsiCase("text/with-standard-background-color", text([raw("Test")], {backgroundColor: "green"})),
	namedAnsiCase("text/text with standard background color", text([raw("Test")], {backgroundColor: "green"})),
	namedAnsiCase("text/ansi-background-hex", text([raw("Test")], {backgroundColor: "#FF8800"})),
	namedAnsiCase("text/with-hex-background-color", text([raw("Test")], {backgroundColor: "#FF8800"})),
	namedAnsiCase("text/text with hex background color", text([raw("Test")], {backgroundColor: "#FF8800"})),
	namedAnsiCase("text/ansi-background-rgb", text([raw("Test")], {backgroundColor: "rgb(255, 136, 0)"})),
	namedAnsiCase("text/with-rgb-background-color", text([raw("Test")], {backgroundColor: "rgb(255, 136, 0)"})),
	namedAnsiCase("text/text with rgb background color", text([raw("Test")], {backgroundColor: "rgb(255, 136, 0)"})),
	namedAnsiCase("text/ansi-background-ansi256", text([raw("Test")], {backgroundColor: "ansi256(194)"})),
	namedAnsiCase("text/with-ansi256-background-color", text([raw("Test")], {backgroundColor: "ansi256(194)"})),
	namedAnsiCase("text/text with ansi256 background color", text([raw("Test")], {backgroundColor: "ansi256(194)"})),
	namedAnsiCase("text/ansi-inverse", text([raw("Test")], {inverse: true})),
	namedAnsiCase("text/with-inversion", text([raw("Test")], {inverse: true})),
	namedAnsiCase("text/text with inversion", text([raw("Test")], {inverse: true})),
	namedAnsiCase("text/text with inversion - concurrent", text([raw("Test")], {inverse: true})),
	namedAnsiCase("text/ansi-leading-whitespace-red", text([raw(" ERROR ")], {color: "red"})),
	namedAnsiCase("text/ensure wrap-ansi doesn’t trim leading whitespace", text([raw(" ERROR ")], {color: "red"})),
	namedCase("text/<Text> with undefined children - concurrent", text([undefined])),
	namedCase("text/<Text> with null children - concurrent", text([empty()])),
];

const buildNewlineCases = () => {
	const cases = [];

	for (const count of [1, 2, 3, 4, 5]) {
		cases.push(
			namedCase(
				`newline/middle-${count}`,
				text([raw("Hello"), newline(count), raw("World")]),
			),
		);
		cases.push(
			namedCase(
				`newline/leading-${count}`,
				text([newline(count), raw("World")]),
			),
		);
		cases.push(
			namedCase(
				`newline/trailing-${count}`,
				text([raw("Hello"), newline(count)]),
			),
		);
		cases.push(namedCase(`newline/only-${count}`, text([newline(count)])));
		cases.push(
			namedCase(
				`newline/nested-${count}`,
				text([text([raw("A")]), newline(count), text([raw("B")])]),
			),
		);
		cases.push(
			namedCase(
				`newline/sandwich-${count}`,
				text([raw("Top"), newline(count), raw("Mid"), newline(count), raw("Bottom")]),
			),
		);
	}

	cases.push(namedCase("newline/newline", text([raw("Hello"), newline(1), raw("World")])));
	cases.push(namedCase("newline/multiple newlines", text([raw("Hello"), newline(2), raw("World")])));
	cases.push(namedCase("newline/newline - concurrent", text([raw("Hello"), newline(1), raw("World")])));

	return cases;
};

const buildSpacerCases = () => {
	const cases = [];

	for (const width of [8, 10, 12, 14, 20]) {
		cases.push(
			namedCase(
				`spacer/h-between-${width}`,
				box({width}, [textValue("L"), spacer(), textValue("R")]),
			),
		);
	}

	for (const width of [4, 6, 8, 10, 12]) {
		cases.push(
			namedCase(
				`spacer/h-leading-${width}`,
				box({width}, [spacer(), textValue("R")]),
			),
		);
	}

	for (const width of [4, 6, 8, 10, 12]) {
		cases.push(
			namedCase(
				`spacer/h-trailing-${width}`,
				box({width}, [textValue("L"), spacer()]),
			),
		);
	}

	for (const width of [11, 15, 19, 23, 27]) {
		cases.push(
			namedCase(
				`spacer/h-double-${width}`,
				box({width}, [textValue("A"), spacer(), textValue("B"), spacer(), textValue("C")]),
			),
		);
	}

	for (const height of [4, 5, 6, 7, 8]) {
		cases.push(
			namedCase(
				`spacer/v-between-${height}`,
				box({flexDirection: "column", height}, [textValue("Top"), spacer(), textValue("Bottom")]),
			),
		);
	}

	for (const height of [7, 9, 11, 13, 15]) {
		cases.push(
			namedCase(
				`spacer/v-double-${height}`,
				box(
					{flexDirection: "column", height},
					[textValue("A"), spacer(), textValue("B"), spacer(), textValue("C")],
				),
			),
		);
	}

	cases.push(
		namedCase(
			"spacer/horizontal spacer",
			box({width: 20}, [textValue("Left"), spacer(), textValue("Right")]),
		),
	);
	cases.push(
		namedCase(
			"spacer/vertical spacer",
			box({flexDirection: "column", height: 6}, [textValue("Top"), spacer(), textValue("Bottom")]),
		),
	);
	cases.push(
		namedCase(
			"spacer/horizontal spacer - concurrent",
			box({width: 20}, [textValue("Left"), spacer(), textValue("Right")]),
		),
	);
	cases.push(
		namedCase(
			"spacer/vertical spacer - concurrent",
			box({flexDirection: "column", height: 6}, [textValue("Top"), spacer(), textValue("Bottom")]),
		),
	);

	return cases;
};

const transformPresets = [
	"identity",
	"bracket_index",
	"brace_index",
	"angle",
	"upper",
	"reverse",
];

const buildTransformCases = () => {
	const cases = [];

	for (const preset of transformPresets) {
		cases.push(
			namedCase(`transform/${preset}-basic`, transform(preset, [textValue("hello")])),
		);
		cases.push(
			namedCase(`transform/${preset}-multiline`, transform(preset, [textValue("hello\nworld")])),
		);
		cases.push(
			namedCase(
				`transform/${preset}-nested`,
				transform(preset, [text([transform("brace_index", [textValue("hello")])])]),
			),
		);
		cases.push(
			namedCase(
				`transform/${preset}-inside-text`,
				text([raw("pre "), transform(preset, [textValue("mid")]), raw(" post")]),
			),
		);
		cases.push(namedCase(`transform/${preset}-empty`, transform(preset, [])));
	}

	cases.push(
		namedCase(
			"transform/bracket_index-squash-multi-text",
			transform("bracket_index", [
				text([transform("brace_index", [text([raw("hello"), raw(" "), raw("world")])])]),
			]),
		),
	);
	cases.push(
		namedCase(
			"transform/transform children",
			transform("bracket_index", [text([transform("brace_index", [textValue("test")])])]),
		),
	);
	cases.push(
		namedCase(
			"transform/transform children - concurrent",
			transform("bracket_index", [text([transform("brace_index", [textValue("test")])])]),
		),
	);
	cases.push(
		namedCase(
			"transform/bracket_index-squash-nested-text",
			transform("bracket_index", [
				text([transform("brace_index", [raw("hello"), text([raw(" world")])])]),
			]),
		),
	);
	cases.push(
		namedCase(
			"transform/squash multiple text nodes",
			transform("bracket_index", [
				text([transform("brace_index", [text([raw("hello"), raw(" "), raw("world")])])]),
			]),
		),
	);
	cases.push(
		namedCase(
			"transform/transform with multiple lines",
			transform("bracket_index", [text([raw("hello world\ngoodbye world")])]),
		),
	);
	cases.push(
		namedCase(
			"transform/squash multiple nested text nodes",
			transform("bracket_index", [
				text([transform("brace_index", [raw("hello"), text([raw(" world")])])]),
			]),
		),
	);
	cases.push(
		namedCase(
			"transform/identity-empty-text-child",
			transform("identity", [text([text([])])]),
		),
	);
	cases.push(
		namedCase(
			"transform/squash empty `<Text>` nodes",
			transform("identity", [text([text([])])]),
		),
	);
	cases.push(namedCase("transform/undefined-child", transform("identity", [undefined])));
	cases.push(namedCase("transform/null-child", transform("identity", [empty()])));
	cases.push(namedCase("transform/<Transform> with undefined children", transform("identity", [undefined])));
	cases.push(namedCase("transform/<Transform> with null children", transform("identity", [empty()])));
	cases.push(
		namedCase(
			"transform/screen-reader-accessibility-label",
			transform("upper", [textValue("hidden visual value")], {accessibilityLabel: "spoken value"}),
			100,
			true,
		),
	);
	cases.push(
		namedCase(
			"transform/screen-reader-accessibility-label-parent-role",
			box(
				{"aria-role": "button"},
				[transform("upper", [textValue("hidden visual value")], {accessibilityLabel: "spoken value"})],
			),
			100,
			true,
		),
	);

	// Box+Text style parity gap-fillers below: ANSI mutation, multi-line width
	// changes, and parent-style continuation through transforms.
	cases.push(
		namedCase(
			"transform/ansi-link-wrapper",
			transform("angle", [textValue("Example")]),
		),
	);
	cases.push(
		namedCase(
			"transform/multiline-width-change",
			transform("bracket_index", [textValue("a\nbb\nccc")]),
		),
	);
	cases.push(
		namedAnsiCase(
			"transform/parent-text-color-continuation",
			text(
				[raw("pre "), transform("upper", [textValue("mid")]), raw(" post")],
				{color: "green"},
			),
		),
	);
	cases.push(
		namedCase(
			"transform/inside-bordered-box",
			box(
				{borderStyle: "round", width: 12},
				[transform("upper", [textValue("hello world")])],
			),
		),
	);

	return cases;
};

const buildBoxCases = () => [
	namedCase("box/row-basic-two", box({}, [textValue("A"), textValue("B")])),
	namedCase("box/row-basic-three", box({}, [textValue("A"), textValue("B"), textValue("C")])),
	namedCase("box/display-flex", box({display: "flex"}, [textValue("X")])),
	namedCase("box/display flex", box({display: "flex"}, [textValue("X")])),
	namedCase(
		"box/display-none",
		box({flexDirection: "column"}, [box({display: "none"}, [textValue("Kitty!")]), textValue("Doggo")]),
	),
	namedCase(
		"box/display none",
		box({flexDirection: "column"}, [box({display: "none"}, [textValue("Kitty!")]), textValue("Doggo")]),
	),
	namedCase(
		"box/position-absolute-does-not-consume-column-space",
		box({flexDirection: "column"}, [
			textValue("Line 1"),
			box({position: "absolute"}, [textValue("ABS")]),
			textValue("Line 2"),
		]),
	),
	namedCase("box/row-reverse-basic-two", box({flexDirection: "row-reverse", width: 4}, [textValue("A"), textValue("B")])),
	namedCase("box/gap-wrap", box({gap: 1, width: 3, flexWrap: "wrap"}, [textValue("A"), textValue("B"), textValue("C")])),
	namedCase("box/gap", box({gap: 1, width: 3, flexWrap: "wrap"}, [textValue("A"), textValue("B"), textValue("C")])),
	namedCase("box/gap - concurrent", box({gap: 1, width: 3, flexWrap: "wrap"}, [textValue("A"), textValue("B"), textValue("C")])),
	namedCase("box/gap-column", box({gap: 1}, [textValue("A"), textValue("B")])),
	namedCase("box/column gap", box({gap: 1}, [textValue("A"), textValue("B")])),
	namedCase("box/gap-row", box({flexDirection: "column", gap: 1}, [textValue("A"), textValue("B")])),
	namedCase("box/row gap", box({flexDirection: "column", gap: 1}, [textValue("A"), textValue("B")])),
	namedCase("box/flex-wrap-row-nowrap", box({width: 2}, [textValue("A"), textValue("BC")])),
	namedCase("box/row - no wrap", box({width: 2}, [textValue("A"), textValue("BC")])),
	namedCase("box/flex-wrap-row-wrap", box({width: 2, flexWrap: "wrap"}, [textValue("A"), textValue("BC")])),
	namedCase("box/row - wrap content", box({width: 2, flexWrap: "wrap"}, [textValue("A"), textValue("BC")])),
	namedCase(
		"box/column-basic-two",
		box({flexDirection: "column"}, [textValue("A"), textValue("B")]),
	),
	namedCase(
		"box/direction column",
		box({flexDirection: "column"}, [textValue("A"), textValue("B")]),
	),
	namedCase(
		"box/column-basic-three",
		box({flexDirection: "column"}, [textValue("A"), textValue("B"), textValue("C")]),
	),
	namedCase(
		"box/flex-wrap-column-nowrap",
		box({flexDirection: "column", height: 2}, [textValue("A"), textValue("B"), textValue("C")]),
	),
	namedCase(
		"box/column - no wrap",
		box({flexDirection: "column", height: 2}, [textValue("A"), textValue("B"), textValue("C")]),
	),
	namedCase(
		"box/flex-wrap-column-wrap",
		box({flexDirection: "column", height: 2, flexWrap: "wrap"}, [textValue("A"), textValue("B"), textValue("C")]),
	),
	namedCase(
		"box/column - wrap content",
		box({flexDirection: "column", height: 2, flexWrap: "wrap"}, [textValue("A"), textValue("B"), textValue("C")]),
	),
	namedCase(
		"box/column-reverse-basic-two",
		box({flexDirection: "column-reverse", height: 4}, [textValue("A"), textValue("B")]),
	),
	namedCase(
		"box/direction column reverse",
		box({flexDirection: "column-reverse", height: 4}, [textValue("A"), textValue("B")]),
	),
	namedCase(
		"box/flex-wrap-column-wrap-reverse",
		box({flexDirection: "column", height: 2, width: 3, flexWrap: "wrap-reverse"}, [textValue("A"), textValue("B"), textValue("C")]),
	),
	namedCase(
		"box/column - wrap content reverse",
		box({flexDirection: "column", height: 2, width: 3, flexWrap: "wrap-reverse"}, [textValue("A"), textValue("B"), textValue("C")]),
	),
	namedCase(
		"box/flex-wrap-row-wrap-reverse",
		box({height: 3, width: 2, flexWrap: "wrap-reverse"}, [textValue("A"), textValue("B"), textValue("C")]),
	),
	namedCase(
		"box/row - wrap content reverse",
		box({height: 3, width: 2, flexWrap: "wrap-reverse"}, [textValue("A"), textValue("B"), textValue("C")]),
	),
	namedCase("box/direction row", box({flexDirection: "row"}, [textValue("A"), textValue("B")])),
	namedCase("box/direction row reverse", box({flexDirection: "row-reverse", width: 4}, [textValue("A"), textValue("B")])),
	namedCase(
		"box/don’t squash text nodes when column direction is applied",
		box({flexDirection: "column"}, [textValue("A"), textValue("B")]),
	),
	namedCase("box/justify-center-row-one", box({width: 9, justifyContent: "center"}, [textValue("ABC")])),
	namedCase("box/row - align text to center", box({width: 10, justifyContent: "center"}, [textValue("Test")])),
	namedCase("box/justify-center-row-two", box({width: 10, justifyContent: "center"}, [textValue("A"), textValue("B")])),
	namedCase("box/row - align multiple text nodes to center", box({width: 10, justifyContent: "center"}, [textValue("A"), textValue("B")])),
	namedCase("box/justify-end-row-one", box({width: 9, justifyContent: "flex-end"}, [textValue("ABC")])),
	namedCase("box/row - align text to right", box({width: 10, justifyContent: "flex-end"}, [textValue("Test")])),
	namedCase("box/justify-end-row-two", box({width: 10, justifyContent: "flex-end"}, [textValue("A"), textValue("B")])),
	namedCase("box/row - align multiple text nodes to right", box({width: 10, justifyContent: "flex-end"}, [textValue("A"), textValue("B")])),
	namedAnsiCase("box/row - align colored text node when text is squashed", box({justifyContent: "flex-end", width: 5}, [text([raw("X")], {color: "green"})])),
	namedCase("box/justify-between-row-two", box({width: 5, justifyContent: "space-between"}, [textValue("A"), textValue("B")])),
	namedCase("box/row - align two text nodes on the edges", box({width: 4, justifyContent: "space-between"}, [textValue("A"), textValue("B")])),
	namedCase("box/row - space evenly two text nodes", box({width: 10, justifyContent: "space-evenly"}, [textValue("A"), textValue("B")])),
	namedCase("box/justify-around-row-two", box({width: 5, justifyContent: "space-around"}, [textValue("A"), textValue("B")])),
	namedCase("box/row - align two text nodes with equal space around them", box({width: 5, justifyContent: "space-around"}, [textValue("A"), textValue("B")])),
	namedCase(
		"box/justify-between-row-three",
		box({width: 7, justifyContent: "space-between"}, [textValue("A"), textValue("B"), textValue("C")]),
	),
	namedCase(
		"box/justify-center-column-one",
		box({flexDirection: "column", height: 3, justifyContent: "center"}, [textValue("A")]),
	),
	namedCase(
		"box/column - align text to center",
		box({flexDirection: "column", height: 3, justifyContent: "center"}, [textValue("Test")]),
	),
	namedCase(
		"box/justify-center-column-two",
		box({flexDirection: "column", height: 4, justifyContent: "center"}, [textValue("A"), textValue("B")]),
	),
	namedCase(
		"box/justify-end-column-one",
		box({flexDirection: "column", height: 3, justifyContent: "flex-end"}, [textValue("A")]),
	),
	namedCase(
		"box/column - align text to bottom",
		box({flexDirection: "column", height: 3, justifyContent: "flex-end"}, [textValue("Test")]),
	),
	namedCase(
		"box/justify-end-column-two",
		box({flexDirection: "column", height: 4, justifyContent: "flex-end"}, [textValue("A"), textValue("B")]),
	),
	namedCase(
		"box/justify-between-column-two",
		box({flexDirection: "column", height: 4, justifyContent: "space-between"}, [textValue("A"), textValue("B")]),
	),
	namedCase(
		"box/column - align two text nodes on the edges",
		box({flexDirection: "column", height: 4, justifyContent: "space-between"}, [textValue("A"), textValue("B")]),
	),
	namedCase(
		"box/justify-around-column-two",
		box({flexDirection: "column", height: 5, justifyContent: "space-around"}, [textValue("A"), textValue("B")]),
	),
	namedCase(
		"box/justify-between-column-three",
		box(
			{flexDirection: "column", height: 5, justifyContent: "space-between"},
			[textValue("A"), textValue("B"), textValue("C")],
		),
	),
	namedCase("box/padding-all-one", box({padding: 1}, [textValue("X")])),
	namedCase("box/padding-all-two", box({padding: 2}, [textValue("X")])),
	namedCase("box/padding", box({padding: 2}, [textValue("X")])),
	namedCase("box/padding-x-one", box({}, [box({paddingX: 1}, [textValue("X")]), textValue("Y")])),
	namedCase("box/padding-x-two", box({}, [box({paddingX: 2}, [textValue("X")]), textValue("Y")])),
	namedCase("box/padding X", box({}, [box({paddingX: 2}, [textValue("X")]), textValue("Y")])),
	namedCase("box/padding-y-one", box({paddingY: 1}, [textValue("X")])),
	namedCase("box/padding-y-two", box({paddingY: 2}, [textValue("X")])),
	namedCase("box/padding Y", box({paddingY: 2}, [textValue("X")])),
	namedCase("box/padding-top-two", box({paddingTop: 2}, [textValue("X")])),
	namedCase("box/padding top", box({paddingTop: 2}, [textValue("X")])),
	namedCase("box/padding-bottom-two", box({paddingBottom: 2}, [textValue("X")])),
	namedCase("box/padding bottom", box({paddingBottom: 2}, [textValue("X")])),
	namedCase("box/padding-left-two", box({paddingLeft: 2}, [textValue("X")])),
	namedCase("box/padding left", box({paddingLeft: 2}, [textValue("X")])),
	namedCase("box/padding-right-two", box({}, [box({paddingRight: 2}, [textValue("X")]), textValue("Y")])),
	namedCase("box/padding right", box({}, [box({paddingRight: 2}, [textValue("X")]), textValue("Y")])),
	namedCase("box/nested-padding", box({padding: 1}, [box({padding: 1}, [textValue("X")])])),
	namedCase("box/nested padding", box({padding: 2}, [box({padding: 2}, [textValue("X")])])),
	namedCase("box/multiline-padding", box({padding: 1}, [textValue("A\nB")])),
	namedCase("box/padding with multiline string", box({padding: 2}, [textValue("A\nB")])),
	namedCase("box/padding-text-newlines", box({padding: 1}, [textValue("Hello\nWorld")])),
	namedCase("box/apply padding to text with newlines", box({padding: 1}, [textValue("Hello\nWorld")])),
	namedCase("box/padding-wrapped-text", box({padding: 1, width: 5}, [textValue("Hello World")])),
	namedCase("box/apply padding to wrapped text", box({padding: 1, width: 5}, [textValue("Hello World")])),
	namedCase(
		"box/flex-grow-row",
		box({width: 8}, [box({flexGrow: 1}, [textValue("A")]), box({flexGrow: 1}, [textValue("B")])]),
	),
	namedCase(
		"box/grow equally",
		box({width: 6}, [box({flexGrow: 1}, [textValue("A")]), box({flexGrow: 1}, [textValue("B")])]),
	),
	namedCase(
		"box/flex-grow-one",
		box({width: 6}, [box({flexGrow: 1}, [textValue("A")]), textValue("B")]),
	),
	namedCase(
		"box/grow one element",
		box({width: 6}, [box({flexGrow: 1}, [textValue("A")]), textValue("B")]),
	),
	namedCase(
		"box/flex-grow-column",
		box(
			{flexDirection: "column", height: 6},
			[box({flexGrow: 1}, [textValue("A")]), box({flexGrow: 1}, [textValue("B")])],
		),
	),
	namedCase(
		"box/flex-shrink-none",
		box(
			{width: 16},
			[
				box({flexShrink: 0, width: 6}, [textValue("A")]),
				box({flexShrink: 0, width: 6}, [textValue("B")]),
				box({width: 6}, [textValue("C")]),
			],
		),
	),
	namedCase(
		"box/dont shrink",
		box(
			{width: 16},
			[
				box({flexShrink: 0, width: 6}, [textValue("A")]),
				box({flexShrink: 0, width: 6}, [textValue("B")]),
				box({width: 6}, [textValue("C")]),
			],
		),
	),
	namedCase(
		"box/flex-shrink-equal",
		box(
			{width: 10},
			[
				box({flexShrink: 1, width: 6}, [textValue("A")]),
				box({flexShrink: 1, width: 6}, [textValue("B")]),
				textValue("C"),
			],
		),
	),
	namedCase(
		"box/shrink equally",
		box(
			{width: 10},
			[
				box({flexShrink: 1, width: 6}, [textValue("A")]),
				box({flexShrink: 1, width: 6}, [textValue("B")]),
				textValue("C"),
			],
		),
	),
	namedCase(
		"box/default-flex-shrink-text-pair",
		box({width: 8}, [textValue("Hello"), textValue("World")]),
	),
	namedCase(
		"box/default-flex-shrink-text-pair-bordered",
		box({width: 10, borderStyle: "round"}, [textValue("Hello"), textValue("World")]),
	),
	namedCase(
		"box/default-flex-shrink-text-box-text",
		box(
			{width: 10},
			[textValue("ABCD"), box({width: 4}, [textValue("EFGH")]), textValue("IJKL")],
		),
	),
	namedCase(
		"box/default-flex-shrink-mixed-bordered",
		box(
			{width: 9, borderStyle: "round"},
			[textValue("AAAA"), box({width: 6}, [textValue("B")]), textValue("CC")],
		),
	),
	namedCase(
		"box/default-flex-shrink-text-pair-space",
		box({width: 10}, [textValue("Hello "), textValue("World")]),
	),
	namedCase(
		"box/default-flex-shrink-box-text-plain",
		box({width: 9}, [box({width: 6}, [textValue("A")]), textValue("BBBB")]),
	),
	namedCase(
		"box/default-flex-shrink-text-box-plain",
		box({width: 9}, [textValue("AAAA"), box({width: 6}, [textValue("B")])]),
	),
	namedCase(
		"box/default-flex-shrink-mixed-plain",
		box({width: 9}, [textValue("AAAA"), box({width: 6}, [textValue("B")]), textValue("CC")]),
	),
	namedCase(
		"box/default-flex-shrink-text-triple",
		box({width: 10}, [textValue("ABCD"), textValue("EFGH"), textValue("IJKL")]),
	),
	namedCase(
		"box/default-flex-shrink-text-triple-bordered",
		box({width: 12, borderStyle: "round"}, [textValue("ABCD"), textValue("EFGH"), textValue("IJKL")]),
	),
	namedCase(
		"box/default-flex-shrink-box-text-box",
		box(
			{width: 10},
			[box({width: 4}, [textValue("ABCD")]), textValue("EFGH"), box({width: 4}, [textValue("IJKL")])],
		),
	),
	namedCase(
		"box/default-flex-shrink-text-box-box",
		box(
			{width: 10},
			[textValue("ABCD"), box({width: 4}, [textValue("EFGH")]), box({width: 4}, [textValue("IJKL")])],
		),
	),
	namedCase(
		"box/default-flex-shrink-box-box-text",
		box(
			{width: 10},
			[box({width: 4}, [textValue("ABCD")]), box({width: 4}, [textValue("EFGH")]), textValue("IJKL")],
		),
	),
	namedCase(
		"box/default-flex-shrink-wide-middle",
		box(
			{width: 11},
			[textValue("ABCD"), box({width: 5}, [textValue("EFGHI")]), textValue("JKLM")],
		),
	),
	namedCase(
		"box/default-flex-shrink-text-space-triple",
		box({width: 10}, [textValue("AB "), textValue("CD "), textValue("EFGH")]),
	),
	namedCase(
		"box/flex-shrink-redistributes-after-min-width",
		box(
			{width: 5},
			[
				box({width: 4, minWidth: 3, flexShrink: 1}, [textValue("AAAA")]),
				box({width: 4, flexShrink: 1}, [textValue("BBBB")]),
			],
		),
	),
	namedCase(
		"box/flex-shrink-2-1",
		box(
			{width: 8},
			[
				box({flexShrink: 2, width: 6}, [textValue("AAAAAA")]),
				box({flexShrink: 1, width: 6}, [textValue("BBBBBB")]),
			],
		),
	),
	namedCase(
		"box/flex-shrink-clamped-by-min-width-percent",
		box(
			{width: 8},
			[
				text([raw("AAAAAA")], {minWidth: "50%"}),
				textValue("BBBBBB"),
			],
		),
	),
	namedCase(
		"box/overflow-x-hidden-single-text",
		box(
			{width: 6, overflowX: "hidden"},
			[box({width: 16, flexShrink: 0}, [textValue("Hello World")])],
		),
	),
	namedCase(
		"box/overflow-x-hidden-left-intersection",
		box(
			{width: 6, overflowX: "hidden"},
			[box({marginLeft: -3, width: 12, flexShrink: 0}, [textValue("Hello World")])],
		),
	),
	namedCase(
		"box/overflow-x-before-left-edge",
		box(
			{width: 6, overflowX: "hidden"},
			[box({marginLeft: -12, width: 6, flexShrink: 0}, [textValue("Hello")])],
		),
	),
	namedCase(
		"box/overflow-x-after-right-edge",
		box(
			{width: 6, overflowX: "hidden"},
			[box({marginLeft: 6, width: 6, flexShrink: 0}, [textValue("Hello")])],
		),
	),
	namedCase(
		"box/overflow-x-right-intersection",
		box(
			{width: 6, overflowX: "hidden"},
			[box({marginLeft: 3, width: 6, flexShrink: 0}, [textValue("Hello")])],
		),
	),
	namedCase(
		"box/overflow-y-hidden-single-text",
		box({height: 1, overflowY: "hidden"}, [text([raw("Hello\nWorld")])]),
	),
	namedCase(
		"box/overflow-y-hidden-multi-boxes",
		box(
			{height: 2, overflowY: "hidden", flexDirection: "column"},
			[
				box({flexShrink: 0}, [textValue("Line #1")]),
				box({flexShrink: 0}, [textValue("Line #2")]),
				box({flexShrink: 0}, [textValue("Line #3")]),
				box({flexShrink: 0}, [textValue("Line #4")]),
			],
		),
	),
	namedCase(
		"box/overflow-y-hidden-top-intersection",
		box(
			{height: 1, overflowY: "hidden"},
			[box({marginTop: -1, height: 2, flexShrink: 0}, [text([raw("Hello\nWorld")])])],
		),
	),
	namedCase(
		"box/overflow-hidden-both",
		box(
			{width: 6, height: 1, overflow: "hidden"},
			[box({width: 12, height: 2, flexShrink: 0}, [text([raw("Hello\nWorld")])])],
		),
	),
	namedCase(
		"box/border-negative-margin-child-draw-order",
		box(
			{width: 8, height: 4, borderStyle: "round"},
			[box({marginTop: -1, marginLeft: -1, width: 4, flexShrink: 0}, [textValue("EDGE")])],
		),
	),
	namedCase(
		"box/overflow-hidden-border-container-negative-child",
		box(
			{width: 8, height: 4, overflow: "hidden", borderStyle: "round"},
			[
				box(
					{marginTop: -1, marginLeft: -1, width: 8, height: 3, flexShrink: 0},
					[text([raw("AAAAAA\nBBBBBB\nCCCCCC")])],
				),
			],
		),
	),
	namedCase(
		"box/border-round-full-width",
		box({borderStyle: "round"}, [textValue("Hello World")]),
	),
	namedCase(
		"box/border-classic-fixed-width",
		box({width: 7, height: 3, borderStyle: "classic"}, [textValue("Hi")]),
	),
	namedCase(
		"box/border-arrow-fixed-width",
		box({width: 7, height: 3, borderStyle: "arrow"}, [textValue("Hi")]),
	),
	namedCase(
		"box/single node - full width box",
		box({borderStyle: "round"}, [textValue("Hello World")]),
	),
	namedCase(
		"box/render border after update",
		box({borderStyle: "round"}, [textValue("Hello World")]),
	),
	namedCase(
		"box/single node - full width box - concurrent",
		box({borderStyle: "round"}, [textValue("Hello World")]),
	),
	namedAnsiCase(
		"box/single node - full width box with colorful border",
		box({borderStyle: "round", borderColor: "green"}, [textValue("Hello World")]),
	),
	namedAnsiCase(
		"box/render border after update - concurrent",
		box({borderStyle: "round", borderColor: "green"}, [textValue("Hello World")]),
	),
	namedCase(
		"box/border-round-full-width-multi-node",
		box({borderStyle: "round"}, [text([raw("Hello "), raw("World")])]),
	),
	namedCase(
		"box/multiple nodes - full width box",
		box({borderStyle: "round"}, [text([raw("Hello "), raw("World")])]),
	),
	namedAnsiCase(
		"box/multiple nodes - full width box with colorful border",
		box({borderStyle: "round", borderColor: "green"}, [text([raw("Hello "), raw("World")])]),
	),
	namedCase(
		"box/border-round-fit-content",
		box({borderStyle: "round", alignSelf: "flex-start"}, [textValue("Hello World")]),
	),
	namedCase(
		"box/single node - fit-content box",
		box({borderStyle: "round", alignSelf: "flex-start"}, [textValue("Hello World")]),
	),
	namedCase(
		"box/single node - fit-content box - concurrent",
		box({borderStyle: "round", alignSelf: "flex-start"}, [textValue("Hello World")]),
	),
	namedCase(
		"box/border-round-fit-content-multi-node",
		box({borderStyle: "round", alignSelf: "flex-start"}, [text([raw("Hello "), raw("World")])]),
	),
	namedCase(
		"box/multiple nodes - fit-content box",
		box({borderStyle: "round", alignSelf: "flex-start"}, [text([raw("Hello "), raw("World")])]),
	),
	namedCase(
		"box/border-round-fit-content-wide",
		box({borderStyle: "round", alignSelf: "flex-start"}, [textValue("こんにちは")]),
	),
	namedCase(
		"box/single node - fit-content box with wide characters",
		box({borderStyle: "round", alignSelf: "flex-start"}, [textValue("こんにちは")]),
	),
	namedCase(
		"box/border-round-fit-content-emojis",
		box({borderStyle: "round", alignSelf: "flex-start"}, [textValue("🌊🌊")]),
	),
	namedCase(
		"box/single node - fit-content box with emojis",
		box({borderStyle: "round", alignSelf: "flex-start"}, [textValue("🌊🌊")]),
	),
	namedCase(
		"box/single node - fit-content box with variation selector emojis",
		box({borderStyle: "round", alignSelf: "flex-start"}, [textValue("🌡️⚠️✅")]),
	),
	namedCase(
		"box/border-round-fixed-width",
		box({borderStyle: "round", width: 20}, [textValue("Hello World")]),
	),
	namedCase(
		"box/single node - fixed width box",
		box({borderStyle: "round", width: 20}, [textValue("Hello World")]),
	),
	namedCase(
		"box/border-round-fixed-width-multi-node",
		box({borderStyle: "round", width: 20}, [text([raw("Hello "), raw("World")])]),
	),
	namedCase(
		"box/multiple nodes - fixed width box",
		box({borderStyle: "round", width: 20}, [text([raw("Hello "), raw("World")])]),
	),
	namedCase(
		"box/border-round-fixed-width-height",
		box({borderStyle: "round", width: 20, height: 20}, [textValue("Hello World")]),
	),
	namedCase(
		"box/single node - fixed width and height box",
		box({borderStyle: "round", width: 20, height: 20}, [textValue("Hello World")]),
	),
	namedCase(
		"box/border-round-fixed-width-height-multi-node",
		box({borderStyle: "round", width: 20, height: 20}, [text([raw("Hello "), raw("World")])]),
	),
	namedCase(
		"box/multiple nodes - fixed width and height box",
		box({borderStyle: "round", width: 20, height: 20}, [text([raw("Hello "), raw("World")])]),
	),
	namedCase(
		"box/border-round-padding",
		box({borderStyle: "round", padding: 1, alignSelf: "flex-start"}, [textValue("Hello World")]),
	),
	namedCase(
		"box/single node - box with padding",
		box({borderStyle: "round", padding: 1, alignSelf: "flex-start"}, [textValue("Hello World")]),
	),
	namedCase(
		"box/border-round-padding-multi-node",
		box({borderStyle: "round", padding: 1, alignSelf: "flex-start"}, [text([raw("Hello "), raw("World")])]),
	),
	namedCase(
		"box/multiple nodes - box with padding",
		box({borderStyle: "round", padding: 1, alignSelf: "flex-start"}, [text([raw("Hello "), raw("World")])]),
	),
	namedCase(
		"box/border-round-justify-center",
		box({borderStyle: "round", width: 20, justifyContent: "center"}, [textValue("Hello World")]),
	),
	namedCase(
		"box/single node - box with horizontal alignment",
		box({borderStyle: "round", width: 20, justifyContent: "center"}, [textValue("Hello World")]),
	),
	namedCase(
		"box/border-round-justify-center-multi-node",
		box({borderStyle: "round", width: 20, justifyContent: "center"}, [text([raw("Hello "), raw("World")])]),
	),
	namedCase(
		"box/multiple nodes - box with horizontal alignment",
		box({borderStyle: "round", width: 20, justifyContent: "center"}, [text([raw("Hello "), raw("World")])]),
	),
	namedCase(
		"box/border-round-align-center",
		box({borderStyle: "round", height: 20, alignItems: "center", alignSelf: "flex-start"}, [textValue("Hello World")]),
	),
	namedCase(
		"box/single node - box with vertical alignment",
		box({borderStyle: "round", height: 20, alignItems: "center", alignSelf: "flex-start"}, [textValue("Hello World")]),
	),
	namedCase(
		"box/border-round-align-center-multi-node",
		box(
			{borderStyle: "round", height: 20, alignItems: "center", alignSelf: "flex-start"},
			[text([raw("Hello "), raw("World")])],
		),
	),
	namedCase(
		"box/multiple nodes - box with vertical alignment",
		box(
			{borderStyle: "round", height: 20, alignItems: "center", alignSelf: "flex-start"},
			[text([raw("Hello "), raw("World")])],
		),
	),
	namedCase(
		"box/border-round-wrap",
		box({borderStyle: "round", width: 10}, [textValue("Hello World")]),
	),
	namedCase(
		"box/single node - box with wrapping",
		box({borderStyle: "round", width: 10}, [textValue("Hello World")]),
	),
	namedCase(
		"box/border-round-wrap-multi-node",
		box({borderStyle: "round", width: 10}, [text([raw("Hello "), raw("World")])]),
	),
	namedCase(
		"box/multiple nodes - box with wrapping",
		box({borderStyle: "round", width: 10}, [text([raw("Hello "), raw("World")])]),
	),
	namedCase(
		"box/border-round-wrap-long-first-node",
		box({borderStyle: "round", width: 10}, [text([raw("Helloooooo"), raw(" World")])]),
	),
	namedCase(
		"box/multiple nodes - box with wrapping and long first node",
		box({borderStyle: "round", width: 10}, [text([raw("Helloooooo"), raw(" World")])]),
	),
	namedCase(
		"box/border-round-wrap-very-long-first-node",
		box({borderStyle: "round", width: 10}, [text([raw("Hellooooooooooooo"), raw(" World")])]),
	),
	namedCase(
		"box/multiple nodes - box with wrapping and very long first node",
		box({borderStyle: "round", width: 10}, [text([raw("Hellooooooooooooo"), raw(" World")])]),
	),
	namedCase(
		"box/border-round-nested",
		box(
			{borderStyle: "round", width: 40, padding: 1},
			[
				box(
					{borderStyle: "round", justifyContent: "center", padding: 1},
					[textValue("Hello World")],
				),
			],
		),
	),
	namedCase(
		"box/nested boxes",
		box(
			{borderStyle: "round", width: 40, padding: 1},
			[
				box(
					{borderStyle: "round", justifyContent: "center", padding: 1},
					[textValue("Hello World")],
				),
			],
		),
	),
	namedCase(
		"box/nested boxes - concurrent",
		box(
			{borderStyle: "round", width: 40, padding: 1},
			[
				box(
					{borderStyle: "round", justifyContent: "center", padding: 1},
					[textValue("Hello World")],
				),
			],
		),
	),
	namedCase(
		"box/border-round-row-wide-nested",
		box(
			{borderStyle: "round", alignSelf: "flex-start"},
			[
				box({borderStyle: "round"}, [textValue("ミスター")]),
				box({borderStyle: "round"}, [textValue("スポック")]),
				box({borderStyle: "round"}, [textValue("カーク船長")]),
			],
		),
	),
	namedCase(
		"box/nested boxes - fit-content box with wide characters on flex-direction row",
		box(
			{borderStyle: "round", alignSelf: "flex-start"},
			[
				box({borderStyle: "round"}, [textValue("ミスター")]),
				box({borderStyle: "round"}, [textValue("スポック")]),
				box({borderStyle: "round"}, [textValue("カーク船長")]),
			],
		),
	),
	namedCase(
		"box/border-round-row-emoji-nested",
		box(
			{borderStyle: "round", alignSelf: "flex-start"},
			[
				box({borderStyle: "round"}, [textValue("🦾")]),
				box({borderStyle: "round"}, [textValue("🌏")]),
				box({borderStyle: "round"}, [textValue("😋")]),
			],
		),
	),
	namedCase(
		"box/nested boxes - fit-content box with emojis on flex-direction row",
		box(
			{borderStyle: "round", alignSelf: "flex-start"},
			[
				box({borderStyle: "round"}, [textValue("🦾")]),
				box({borderStyle: "round"}, [textValue("🌏")]),
				box({borderStyle: "round"}, [textValue("😋")]),
			],
		),
	),
	namedCase(
		"box/border-round-column-wide-nested",
		box(
			{borderStyle: "round", alignSelf: "flex-start", flexDirection: "column"},
			[
				box({borderStyle: "round"}, [textValue("ミスター")]),
				box({borderStyle: "round"}, [textValue("スポック")]),
				box({borderStyle: "round"}, [textValue("カーク船長")]),
			],
		),
	),
	namedCase(
		"box/nested boxes - fit-content box with wide characters on flex-direction column",
		box(
			{borderStyle: "round", alignSelf: "flex-start", flexDirection: "column"},
			[
				box({borderStyle: "round"}, [textValue("ミスター")]),
				box({borderStyle: "round"}, [textValue("スポック")]),
				box({borderStyle: "round"}, [textValue("カーク船長")]),
			],
		),
	),
	namedCase(
		"box/border-round-column-emoji-nested",
		box(
			{borderStyle: "round", alignSelf: "flex-start", flexDirection: "column"},
			[
				box({borderStyle: "round"}, [textValue("🦾")]),
				box({borderStyle: "round"}, [textValue("🌏")]),
				box({borderStyle: "round"}, [textValue("😋")]),
			],
		),
	),
	namedCase(
		"box/nested boxes - fit-content box with emojis on flex-direction column",
		box(
			{borderStyle: "round", alignSelf: "flex-start", flexDirection: "column"},
			[
				box({borderStyle: "round"}, [textValue("🦾")]),
				box({borderStyle: "round"}, [textValue("🌏")]),
				box({borderStyle: "round"}, [textValue("😋")]),
			],
		),
	),
	namedCase(
		"box/border-round-hide-top",
		box(
			{flexDirection: "column", alignItems: "flex-start"},
			[
				textValue("Above"),
				box({borderStyle: "round", borderTop: false}, [textValue("Content")]),
				textValue("Below"),
			],
		),
	),
	namedCase(
		"box/hide top border",
		box(
			{flexDirection: "column", alignItems: "flex-start"},
			[
				textValue("Above"),
				box({borderStyle: "round", borderTop: false}, [textValue("Content")]),
				textValue("Below"),
			],
		),
	),
	namedCase(
		"box/border-round-hide-left-right",
		box(
			{flexDirection: "column", alignItems: "flex-start"},
			[
				textValue("Above"),
				box({borderStyle: "round", borderLeft: false, borderRight: false}, [textValue("Content")]),
				textValue("Below"),
			],
		),
	),
	namedCase(
		"box/hide left and right border",
		box(
			{flexDirection: "column", alignItems: "flex-start"},
			[
				textValue("Above"),
				box({borderStyle: "round", borderLeft: false, borderRight: false}, [textValue("Content")]),
				textValue("Below"),
			],
		),
	),
	namedCase(
		"box/overflow-x-hidden-border-round",
		box(
			{width: 6, overflowX: "hidden", borderStyle: "round"},
			[box({width: 16, flexShrink: 0}, [textValue("Hello World")])],
		),
	),
	namedCase(
		"box/overflow-y-hidden-border-round",
		box(
			{width: 20, height: 3, overflowY: "hidden", borderStyle: "round"},
			[text([raw("Hello\nWorld")])],
		),
	),
	namedCase(
		"box/border-round-hide-bottom",
		box(
			{flexDirection: "column", alignItems: "flex-start"},
			[
				textValue("Above"),
				box({borderStyle: "round", borderBottom: false}, [textValue("Content")]),
				textValue("Below"),
			],
		),
	),
	namedCase(
		"box/hide bottom border",
		box(
			{flexDirection: "column", alignItems: "flex-start"},
			[
				textValue("Above"),
				box({borderStyle: "round", borderBottom: false}, [textValue("Content")]),
				textValue("Below"),
			],
		),
	),
	namedCase(
		"box/border-round-hide-top-bottom",
		box(
			{flexDirection: "column", alignItems: "flex-start"},
			[
				textValue("Above"),
				box({borderStyle: "round", borderTop: false, borderBottom: false}, [textValue("Content")]),
				textValue("Below"),
			],
		),
	),
	namedCase(
		"box/hide top and bottom borders",
		box(
			{flexDirection: "column", alignItems: "flex-start"},
			[
				textValue("Above"),
				box({borderStyle: "round", borderTop: false, borderBottom: false}, [textValue("Content")]),
				textValue("Below"),
			],
		),
	),
	namedCase(
		"box/border-round-hide-left",
		box(
			{flexDirection: "column", alignItems: "flex-start"},
			[
				textValue("Above"),
				box({borderStyle: "round", borderLeft: false}, [textValue("Content")]),
				textValue("Below"),
			],
		),
	),
	namedCase(
		"box/hide left border",
		box(
			{flexDirection: "column", alignItems: "flex-start"},
			[
				textValue("Above"),
				box({borderStyle: "round", borderLeft: false}, [textValue("Content")]),
				textValue("Below"),
			],
		),
	),
	namedCase(
		"box/border-round-hide-right",
		box(
			{flexDirection: "column", alignItems: "flex-start"},
			[
				textValue("Above"),
				box({borderStyle: "round", borderRight: false}, [textValue("Content")]),
				textValue("Below"),
			],
		),
	),
	namedCase(
		"box/hide right border",
		box(
			{flexDirection: "column", alignItems: "flex-start"},
			[
				textValue("Above"),
				box({borderStyle: "round", borderRight: false}, [textValue("Content")]),
				textValue("Below"),
			],
		),
	),
	namedCase(
		"box/border-round-hide-all",
		box(
			{flexDirection: "column", alignItems: "flex-start"},
			[
				textValue("Above"),
				box(
					{
						borderStyle: "round",
						borderTop: false,
						borderBottom: false,
						borderLeft: false,
						borderRight: false,
					},
					[textValue("Content")],
				),
				textValue("Below"),
			],
		),
	),
	namedCase(
		"box/hide all borders",
		box(
			{flexDirection: "column", alignItems: "flex-start"},
			[
				textValue("Above"),
				box(
					{
						borderStyle: "round",
						borderTop: false,
						borderBottom: false,
						borderLeft: false,
						borderRight: false,
					},
					[textValue("Content")],
				),
				textValue("Below"),
			],
		),
	),
	namedCase(
		"box/border-custom-arrow",
		box(
			{
				borderStyle: {
					topLeft: "↘",
					top: "↓",
					topRight: "↙",
					left: "→",
					bottomLeft: "↗",
					bottom: "↑",
					bottomRight: "↖",
					right: "←",
				},
			},
			[textValue("Content")],
		),
	),
	namedCase(
		"box/custom border style",
		box(
			{
				borderStyle: {
					topLeft: "↘",
					top: "↓",
					topRight: "↙",
					left: "→",
					bottomLeft: "↗",
					bottom: "↑",
					bottomRight: "↖",
					right: "←",
				},
			},
			[textValue("Content")],
		),
	),
	namedCase(
		"box/background-fixed-area",
		box({backgroundColor: "red", width: 10, height: 3, alignSelf: "flex-start"}, [textValue("Hello")]),
	),
	namedCase(
		"box/background-border-round",
		box(
			{backgroundColor: "cyan", borderStyle: "round", width: 10, height: 5, alignSelf: "flex-start"},
			[textValue("Hi")],
		),
	),
	namedCase(
		"box/background-padding-area",
		box(
			{backgroundColor: "magenta", padding: 1, width: 10, height: 5, alignSelf: "flex-start"},
			[textValue("Hi")],
		),
	),
	namedCase(
		"box/background-center-area",
		box(
			{backgroundColor: "blue", width: 10, height: 3, justifyContent: "center", alignSelf: "flex-start"},
			[textValue("Hi")],
		),
	),
	namedCase(
		"box/background-column-area",
		box(
			{backgroundColor: "green", flexDirection: "column", width: 10, height: 5, alignSelf: "flex-start"},
			[textValue("Line 1"), textValue("Line 2")],
		),
	),
	namedCase(
		"box/overflow-x-hidden-child-border",
		box(
			{width: 6, overflowX: "hidden"},
			[box({width: 16, flexShrink: 0, borderStyle: "round"}, [textValue("Hello World")])],
		),
	),
	namedCase(
		"box/overflow-x-hidden-multi-text-border-container",
		box(
			{width: 8, overflowX: "hidden", borderStyle: "round"},
			[box({width: 12, flexShrink: 0}, [textValue("Hello "), textValue("World")])],
		),
	),
	namedCase(
		"box/overflow-x-hidden-multi-boxes",
		box(
			{width: 6, overflowX: "hidden"},
			[
				box({width: 6, flexShrink: 0}, [textValue("Hello ")]),
				box({width: 6, flexShrink: 0}, [textValue("World")]),
			],
		),
	),
	namedCase(
		"box/overflow-x-hidden-multi-boxes-border-container",
		box(
			{width: 8, overflowX: "hidden", borderStyle: "round"},
			[
				box({width: 6, flexShrink: 0}, [textValue("Hello ")]),
				box({width: 6, flexShrink: 0}, [textValue("World")]),
			],
		),
	),
	namedCase(
		"box/overflow-x-before-left-border-container",
		box(
			{width: 6, overflowX: "hidden", borderStyle: "round"},
			[box({marginLeft: -12, width: 6, flexShrink: 0}, [textValue("Hello")])],
		),
	),
	namedCase(
		"box/overflow-y-hidden-multi-boxes-border-container",
		box(
			{width: 9, height: 4, overflowY: "hidden", flexDirection: "column", borderStyle: "round"},
			[
				box({flexShrink: 0}, [textValue("Line #1")]),
				box({flexShrink: 0}, [textValue("Line #2")]),
				box({flexShrink: 0}, [textValue("Line #3")]),
				box({flexShrink: 0}, [textValue("Line #4")]),
			],
		),
	),
	namedCase(
		"box/overflow-y-hidden-top-intersection-border-container",
		box(
			{width: 7, height: 3, overflowY: "hidden", borderStyle: "round"},
			[box({marginTop: -1, height: 2, flexShrink: 0}, [text([raw("Hello\nWorld")])])],
		),
	),
	namedCase(
		"box/overflow-y-hidden-above-top",
		box(
			{height: 1, overflowY: "hidden"},
			[box({marginTop: -2, height: 2, flexShrink: 0}, [text([raw("Hello\nWorld")])])],
		),
	),
	namedCase(
		"box/overflow-y-hidden-below-bottom",
		box(
			{height: 1, overflowY: "hidden"},
			[box({marginTop: 1, height: 2, flexShrink: 0}, [text([raw("Hello\nWorld")])])],
		),
	),
	namedCase(
		"box/overflow-y-hidden-bottom-intersection",
		box(
			{height: 1, overflowY: "hidden"},
			[box({height: 2, flexShrink: 0}, [text([raw("Hello\nWorld")])])],
		),
	),
	namedCase(
		"box/overflow-y-hidden-above-top-border-container",
		box(
			{width: 7, height: 3, overflowY: "hidden", borderStyle: "round"},
			[box({marginTop: -3, height: 2, flexShrink: 0}, [text([raw("Hello\nWorld")])])],
		),
	),
	namedCase(
		"box/overflow-y-hidden-below-bottom-border-container",
		box(
			{width: 7, height: 3, overflowY: "hidden", borderStyle: "round"},
			[box({marginTop: 2, height: 2, flexShrink: 0}, [text([raw("Hello\nWorld")])])],
		),
	),
	namedCase(
		"box/overflow-y-hidden-bottom-intersection-border-container",
		box(
			{width: 7, height: 3, overflowY: "hidden", borderStyle: "round"},
			[box({height: 2, flexShrink: 0}, [text([raw("Hello\nWorld")])])],
		),
	),
	namedCase(
		"box/overflow-hidden-single-text-border-container",
		box(
			{paddingBottom: 1},
			[
				box(
					{width: 8, height: 3, overflow: "hidden", borderStyle: "round"},
					[box({width: 12, height: 2, flexShrink: 0}, [text([raw("Hello\nWorld")])])],
				),
			],
		),
	),
	namedCase(
		"box/overflow-x-hidden-multi-text-child-border",
		box(
			{width: 8, overflowX: "hidden"},
			[box({width: 12, flexShrink: 0, borderStyle: "round"}, [textValue("Hello "), textValue("World")])],
		),
	),
	namedCase(
		"box/overflow-hidden-top-left-intersection",
		box(
			{width: 4, height: 4, overflow: "hidden"},
			[
				box(
					{marginTop: -2, marginLeft: -2, width: 4, height: 4, flexShrink: 0},
					[text([raw("AAAA\nBBBB\nCCCC\nDDDD")])],
				),
			],
		),
	),
	namedCase(
		"box/overflow-hidden-top-right-intersection",
		box(
			{width: 4, height: 4, overflow: "hidden"},
			[
				box(
					{marginTop: -2, marginLeft: 2, width: 4, height: 4, flexShrink: 0},
					[text([raw("AAAA\nBBBB\nCCCC\nDDDD")])],
				),
			],
		),
	),
	namedCase(
		"box/overflow-hidden-bottom-left-intersection",
		box(
			{width: 4, height: 4, overflow: "hidden"},
			[
				box(
					{marginTop: 2, marginLeft: -2, width: 4, height: 4, flexShrink: 0},
					[text([raw("AAAA\nBBBB\nCCCC\nDDDD")])],
				),
			],
		),
	),
	namedCase(
		"box/overflow-hidden-bottom-right-intersection",
		box(
			{width: 4, height: 4, overflow: "hidden"},
			[
				box(
					{marginTop: 2, marginLeft: 2, width: 4, height: 4, flexShrink: 0},
					[text([raw("AAAA\nBBBB\nCCCC\nDDDD")])],
				),
			],
		),
	),
	namedCase(
		"box/overflow-hidden-multi-boxes-border-container",
		box(
			{paddingBottom: 1},
			[
				box(
					{width: 6, height: 3, overflow: "hidden", borderStyle: "round"},
					[
						box({width: 2, height: 2, flexShrink: 0}, [text([raw("TL\nBL")])]),
						box({width: 2, height: 2, flexShrink: 0}, [text([raw("TR\nBR")])]),
					],
				),
			],
		),
	),
	namedCase(
		"box/nested-overflow",
		box(
			{paddingBottom: 1},
			[
				box(
					{width: 4, height: 4, overflow: "hidden", flexDirection: "column"},
					[
						box(
							{width: 2, height: 2, overflow: "hidden"},
							[box({width: 4, height: 4, flexShrink: 0}, [text([raw("AAAA\nBBBB\nCCCC\nDDDD")])])],
						),
						box({width: 4, height: 3}, [text([raw("XXXX\nYYYY\nZZZZ")])]),
					],
				),
			],
		),
	),
	namedCase("box/out-of-bounds-writes", box({width: 12, height: 10, borderStyle: "round"}, []), 10),
	namedCase(
		"box/default-flex-shrink-equal",
		box({width: 10}, [box({width: 6}, [textValue("A")]), box({width: 6}, [textValue("B")]), textValue("C")]),
	),
	namedCase(
		"box/default-flex-shrink-equal-bordered",
		box(
			{width: 12, borderStyle: "round"},
			[box({width: 6}, [textValue("A")]), box({width: 6}, [textValue("B")]), textValue("C")],
		),
	),
	namedCase("box/margin-all-one", box({margin: 1}, [textValue("X")])),
	namedCase("box/margin-all-two", box({margin: 2}, [textValue("X")])),
	namedCase("box/margin", box({margin: 2}, [textValue("X")])),
	namedCase("box/margin-x-one", box({}, [box({marginX: 1}, [textValue("X")]), textValue("Y")])),
	namedCase("box/margin-x-two", box({}, [box({marginX: 2}, [textValue("X")]), textValue("Y")])),
	namedCase("box/margin X", box({}, [box({marginX: 2}, [textValue("X")]), textValue("Y")])),
	namedCase("box/margin-y-one", box({marginY: 1}, [textValue("X")])),
	namedCase("box/margin-y-two", box({marginY: 2}, [textValue("X")])),
	namedCase("box/margin Y", box({marginY: 2}, [textValue("X")])),
	namedCase("box/margin-top-two", box({marginTop: 2}, [textValue("X")])),
	namedCase("box/margin top", box({marginTop: 2}, [textValue("X")])),
	namedCase("box/margin-bottom-two", box({marginBottom: 2}, [textValue("X")])),
	namedCase("box/margin bottom", box({marginBottom: 2}, [textValue("X")])),
	namedCase("box/margin-left-two", box({marginLeft: 2}, [textValue("X")])),
	namedCase("box/margin left", box({marginLeft: 2}, [textValue("X")])),
	namedCase("box/margin-right-two", box({}, [box({marginRight: 2}, [textValue("X")]), textValue("Y")])),
	namedCase("box/margin right", box({}, [box({marginRight: 2}, [textValue("X")]), textValue("Y")])),
	namedCase("box/nested-margin", box({margin: 1}, [box({margin: 1}, [textValue("X")])])),
	namedCase("box/nested margin", box({margin: 2}, [box({margin: 2}, [textValue("X")])])),
	namedCase("box/multiline-margin", box({margin: 1}, [textValue("A\nB")])),
	namedCase("box/margin with multiline string", box({margin: 2}, [textValue("A\nB")])),
	namedCase("box/margin-text-newlines", box({margin: 1}, [textValue("Hello\nWorld")])),
	namedCase("box/apply margin to text with newlines", box({margin: 1}, [textValue("Hello\nWorld")])),
	namedCase("box/margin-wrapped-text", box({margin: 1, width: 6}, [textValue("Hello World")])),
	namedCase("box/apply margin to wrapped text", box({margin: 1, width: 6}, [textValue("Hello World")])),
	namedCase("box/align-center-row", box({alignItems: "center", height: 3}, [textValue("Test")])),
	namedCase("box/align-end-row", box({alignItems: "flex-end", height: 3}, [textValue("Test")])),
	namedCase(
		"box/align-center-row-two",
		box({alignItems: "center", height: 3}, [textValue("A"), textValue("B")]),
	),
	namedCase(
		"box/align-end-row-two",
		box({alignItems: "flex-end", height: 3}, [textValue("A"), textValue("B")]),
	),
	namedCase(
		"box/align-center-column",
		box({flexDirection: "column", alignItems: "center", width: 10}, [textValue("Test")]),
	),
	namedCase(
		"box/align-end-column",
		box({flexDirection: "column", alignItems: "flex-end", width: 10}, [textValue("Test")]),
	),
	namedCase(
		"box/justify-evenly-row",
		box({justifyContent: "space-evenly", width: 10}, [textValue("A"), textValue("B")]),
	),
	namedCase(
		"box/align-self-center-row",
		box({height: 3}, [box({alignSelf: "center"}, [textValue("Test")])]),
	),
	namedCase(
		"box/align-self-center-row-two",
		box({height: 3}, [box({alignSelf: "center"}, [textValue("A"), textValue("B")])]),
	),
	namedCase(
		"box/align-self-end-row",
		box({height: 3}, [box({alignSelf: "flex-end"}, [textValue("Test")])]),
	),
	namedCase(
		"box/align-self-end-row-two",
		box({height: 3}, [box({alignSelf: "flex-end"}, [textValue("A"), textValue("B")])]),
	),
	namedCase(
		"box/align-self-center-column",
		box({flexDirection: "column", width: 10}, [box({alignSelf: "center"}, [textValue("Test")])]),
	),
	namedCase(
		"box/align-self-end-column",
		box({flexDirection: "column", width: 10}, [box({alignSelf: "flex-end"}, [textValue("Test")])]),
	),
	namedCase(
		"box/flex-basis-row",
		box({width: 6}, [box({flexBasis: 3}, [textValue("A")]), textValue("B")]),
	),
	namedCase(
		"box/set flex basis with flexDirection=\"row\" container",
		box({width: 6}, [box({flexBasis: 3}, [textValue("A")]), textValue("B")]),
	),
	namedCase(
		"box/flex-basis-row-percent",
		box({width: 6}, [box({flexBasis: "50%"}, [textValue("A")]), textValue("B")]),
	),
	namedCase(
		"box/set flex basis in percent with flexDirection=\"row\" container",
		box({width: 6}, [box({flexBasis: "50%"}, [textValue("A")]), textValue("B")]),
	),
	namedCase(
		"box/flex-basis-column",
		box({flexDirection: "column", height: 6}, [box({flexBasis: 3}, [textValue("A")]), textValue("B")]),
	),
	namedCase(
		"box/set flex basis with flexDirection=\"column\" container",
		box({flexDirection: "column", height: 6}, [box({flexBasis: 3}, [textValue("A")]), textValue("B")]),
	),
	namedCase(
		"box/flex-basis-column-percent",
		box({flexDirection: "column", height: 6}, [box({flexBasis: "50%"}, [textValue("A")]), textValue("B")]),
	),
	namedCase(
		"box/set flex basis in percent with flexDirection=\"column\" container",
		box({flexDirection: "column", height: 6}, [box({flexBasis: "50%"}, [textValue("A")]), textValue("B")]),
	),
	namedCase(
		"box/width-fixed-row",
		box({}, [box({width: 5}, [textValue("A")]), textValue("B")]),
	),
	namedCase(
		"box/set-width",
		box({}, [box({width: 5}, [textValue("A")]), textValue("B")]),
	),
	namedCase(
		"box/set width",
		box({}, [box({width: 5}, [textValue("A")]), textValue("B")]),
	),
	namedCase(
		"box/set width - concurrent",
		box({}, [box({width: 5}, [textValue("A")]), textValue("B")]),
	),
	namedCase(
		"box/width-percent-row",
		box({width: 10}, [box({width: "50%"}, [textValue("A")]), textValue("B")]),
	),
	namedCase(
		"box/set-width-in-percent",
		box({width: 10}, [box({width: "50%"}, [textValue("A")]), textValue("B")]),
	),
	namedCase(
		"box/set width in percent",
		box({width: 10}, [box({width: "50%"}, [textValue("A")]), textValue("B")]),
	),
	namedCase(
		"box/min-width-small",
		box({}, [box({minWidth: 5}, [textValue("A")]), textValue("B")]),
	),
	namedCase(
		"box/set-min-width-small",
		box({}, [box({minWidth: 5}, [textValue("A")]), textValue("B")]),
	),
	namedCase(
		"box/set min width",
		box({}, [box({minWidth: 5}, [textValue("A")]), textValue("B")]),
	),
	namedCase(
		"box/min-width-large",
		box({}, [box({minWidth: 2}, [textValue("AAAAA")]), textValue("B")]),
	),
	namedCase(
		"box/set-min-width-large",
		box({}, [box({minWidth: 2}, [textValue("AAAAA")]), textValue("B")]),
	),
	namedCase(
		"box/min-width-percent-row",
		box({width: 10}, [box({minWidth: "50%"}, [textValue("A")]), textValue("B")]),
	),
	namedCase(
		"box/set min width in percent",
		box({width: 10}, [box({minWidth: "50%"}, [textValue("A")]), textValue("B")]),
	),
	namedCase(
		"box/height-fixed-row",
		box({height: 4}, [textValue("A"), textValue("B")]),
	),
	namedCase(
		"box/set-height",
		box({height: 4}, [textValue("A"), textValue("B")]),
	),
	namedCase(
		"box/set height",
		box({height: 4}, [textValue("A"), textValue("B")]),
	),
	namedCase(
		"box/set height - concurrent",
		box({height: 4}, [textValue("A"), textValue("B")]),
	),
	namedCase(
		"box/height-percent-column",
		box({height: 6, flexDirection: "column"}, [box({height: "50%"}, [textValue("A")]), textValue("B")]),
	),
	namedCase(
		"box/set-height-in-percent",
		box({height: 6, flexDirection: "column"}, [box({height: "50%"}, [textValue("A")]), textValue("B")]),
	),
	namedCase(
		"box/set height in percent",
		box({height: 6, flexDirection: "column"}, [box({height: "50%"}, [textValue("A")]), textValue("B")]),
	),
	namedCase(
		"box/height-cut-text",
		box({height: 2}, [textValue("AAAABBBBCCCC")]),
		4,
	),
	namedCase(
		"box/cut-text-over-the-set-height",
		box({height: 2}, [textValue("AAAABBBBCCCC")]),
		4,
	),
	namedCase(
		"box/cut text over the set height",
		box({height: 2}, [textValue("AAAABBBBCCCC")]),
		4,
	),
	namedCase(
		"box/min-height-small",
		box({minHeight: 4}, [textValue("A")]),
	),
	namedCase(
		"box/set-min-height-small",
		box({minHeight: 4}, [textValue("A")]),
	),
	namedCase(
		"box/set min height",
		box({minHeight: 4}, [textValue("A")]),
	),
	namedCase(
		"box/min-height-large",
		box({minHeight: 2}, [box({height: 4}, [textValue("A")])]),
	),
	namedCase(
		"box/set-min-height-large",
		box({minHeight: 2}, [box({height: 4}, [textValue("A")])]),
	),
	namedCase(
		"box/text-width-wide-fixed-box",
		box(
			{flexDirection: "column"},
			[
				box({}, [box({width: 2}, [textValue("🍔")]), textValue("|")]),
				box({}, [box({width: 2}, [textValue("⏳")]), textValue("|")]),
			],
		),
	),
	namedCase(
		"box/emoji-zwj-width-fixed-box",
		box({}, [box({width: 2}, [textValue("👨‍👩‍👧‍👦")]), textValue("|")]),
	),
	namedCase(
		"box/wide characters do not add extra space inside fixed-width Box",
		box(
			{flexDirection: "column"},
			[
				box({}, [box({width: 2}, [textValue("🍔")]), textValue("|")]),
				box({}, [box({width: 2}, [textValue("⏳")]), textValue("|")]),
			],
		),
	),
	namedCase(
		"box/border-round-fit-content-variation-emojis",
		box({borderStyle: "round", alignSelf: "flex-start"}, [textValue("🌡️⚠️✅")]),
	),
	namedCase(
		"box/single-node-fit-content-box-with-variation-selector-emojis",
		box({borderStyle: "round", alignSelf: "flex-start"}, [textValue("🌡️⚠️✅")]),
	),
	namedCase(
		"box/single node - fit-content box with variation selector emojis",
		box({borderStyle: "round", alignSelf: "flex-start"}, [textValue("🌡️⚠️✅")]),
	),
	namedAnsiCase(
		"box/ansi-background-inherit",
		box({backgroundColor: "green", alignSelf: "flex-start"}, [textValue("Hello World")]),
	),
	namedAnsiCase(
		"box/ansi-background-nested-inherit",
		box(
			{backgroundColor: "red", alignSelf: "flex-start"},
			[box({backgroundColor: "blue"}, [textValue("Hello World")])],
		),
	),
	namedAnsiCase(
		"box/ansi-background-no-inherit-without-parent",
		box({alignSelf: "flex-start"}, [textValue("Hello World")]),
	),
	namedAnsiCase(
		"box/ansi-background-shared-across-text-children",
		box({backgroundColor: "yellow", alignSelf: "flex-start"}, [textValue("Hello "), textValue("World")]),
	),
	namedAnsiCase(
		"box/ansi-background-override",
		box(
			{backgroundColor: "red", alignSelf: "flex-start"},
			[text([raw("Hello World")], {backgroundColor: "blue"})],
		),
	),
	namedAnsiCase(
		"box/ansi-background-mixed-clear-override",
		box(
			{backgroundColor: "green", alignSelf: "flex-start"},
			[
				textValue("Inherited "),
				text([raw("No BG ")], {backgroundColor: ""}),
				text([raw("Red BG")], {backgroundColor: "red"}),
			],
		),
	),
	namedAnsiCase(
		"box/ansi-background-nested-complex",
		box(
			{backgroundColor: "yellow", alignSelf: "flex-start"},
			[
				box(
					{},
					[
						textValue("Outer: "),
						box(
							{backgroundColor: "blue"},
							[
								textValue("Inner: "),
								text([raw("Explicit")], {backgroundColor: "red"}),
							],
						),
					],
				),
			],
		),
	),
	namedAnsiCase(
		"box/ansi-background-fixed-area",
		box({backgroundColor: "red", width: 10, height: 3, alignSelf: "flex-start"}, [textValue("Hello")]),
	),
	namedAnsiCase(
		"box/ansi-background-fixed-area-hex",
		box({backgroundColor: "#FF0000", width: 10, height: 3, alignSelf: "flex-start"}, [textValue("Hello")]),
	),
	namedAnsiCase(
		"box/ansi-background-fixed-area-rgb",
		box({backgroundColor: "rgb(255, 0, 0)", width: 10, height: 3, alignSelf: "flex-start"}, [textValue("Hello")]),
	),
	namedAnsiCase(
		"box/ansi-background-fixed-area-ansi256",
		box({backgroundColor: "ansi256(9)", width: 10, height: 3, alignSelf: "flex-start"}, [textValue("Hello")]),
	),
	namedAnsiCase(
		"box/ansi-background-hex",
		box({backgroundColor: "#FF0000", alignSelf: "flex-start"}, [textValue("Hello")]),
	),
	namedAnsiCase(
		"box/ansi-background-rgb",
		box({backgroundColor: "rgb(255, 0, 0)", alignSelf: "flex-start"}, [textValue("Hello")]),
	),
	namedAnsiCase(
		"box/ansi-background-ansi256",
		box({backgroundColor: "ansi256(9)", alignSelf: "flex-start"}, [textValue("Hello")]),
	),
	namedAnsiCase(
		"box/ansi-background-wide",
		box({backgroundColor: "yellow", alignSelf: "flex-start"}, [textValue("こんにちは")]),
	),
	namedAnsiCase(
		"box/ansi-background-emoji",
		box({backgroundColor: "red", alignSelf: "flex-start"}, [textValue("🎉🎊")]),
	),
	namedAnsiCase(
		"box/ansi-background-border-round",
		box(
			{backgroundColor: "cyan", borderStyle: "round", width: 10, height: 5, alignSelf: "flex-start"},
			[textValue("Hi")],
		),
	),
	namedAnsiCase(
		"box/ansi-background-padding-area",
		box({backgroundColor: "magenta", padding: 1, width: 10, height: 5, alignSelf: "flex-start"}, [textValue("Hi")]),
	),
	namedAnsiCase(
		"box/ansi-background-center-area",
		box({backgroundColor: "blue", width: 10, height: 3, justifyContent: "center", alignSelf: "flex-start"}, [textValue("Hi")]),
	),
	namedAnsiCase(
		"box/ansi-background-column-area",
		box(
			{backgroundColor: "green", flexDirection: "column", width: 10, height: 5, alignSelf: "flex-start"},
			[textValue("Line 1"), textValue("Line 2")],
		),
	),
	namedAnsiCase(
		"box/ansi-justify-end-colored-text",
		box({justifyContent: "flex-end", width: 5}, [text([raw("X")], {color: "green"})]),
	),
	namedAnsiCase(
		"box/ansi-border-color-full-width",
		box({borderStyle: "round", borderColor: "green"}, [textValue("Hello World")]),
	),
	namedAnsiCase(
		"box/ansi-border-color-full-width-multi-node",
		box({borderStyle: "round", borderColor: "green"}, [text([raw("Hello "), raw("World")])]),
	),
	namedAnsiCase(
		"box/ansi-border-color-fit-content",
		box({borderStyle: "round", borderColor: "green", alignSelf: "flex-start"}, [textValue("Hello World")]),
	),
	namedAnsiCase(
		"box/ansi-border-color-fit-content-multi-node",
		box(
			{borderStyle: "round", borderColor: "green", alignSelf: "flex-start"},
			[text([raw("Hello "), raw("World")])],
		),
	),
	namedAnsiCase(
		"box/ansi-border-top-color",
		box(
			{flexDirection: "column", alignItems: "flex-start"},
			[
				textValue("Above"),
				box({borderStyle: "round", borderTopColor: "green"}, [textValue("Content")]),
				textValue("Below"),
			],
		),
	),
	namedAnsiCase(
		"box/change color of top border",
		box(
			{flexDirection: "column", alignItems: "flex-start"},
			[
				textValue("Above"),
				box({borderStyle: "round", borderTopColor: "green"}, [textValue("Content")]),
				textValue("Below"),
			],
		),
	),
	namedAnsiCase(
		"box/ansi-border-bottom-color",
		box(
			{flexDirection: "column", alignItems: "flex-start"},
			[
				textValue("Above"),
				box({borderStyle: "round", borderBottomColor: "green"}, [textValue("Content")]),
				textValue("Below"),
			],
		),
	),
	namedAnsiCase(
		"box/change color of bottom border",
		box(
			{flexDirection: "column", alignItems: "flex-start"},
			[
				textValue("Above"),
				box({borderStyle: "round", borderBottomColor: "green"}, [textValue("Content")]),
				textValue("Below"),
			],
		),
	),
	namedAnsiCase(
		"box/ansi-border-left-color",
		box(
			{flexDirection: "column", alignItems: "flex-start"},
			[
				textValue("Above"),
				box({borderStyle: "round", borderLeftColor: "green"}, [textValue("Content")]),
				textValue("Below"),
			],
		),
	),
	namedAnsiCase(
		"box/change color of left border",
		box(
			{flexDirection: "column", alignItems: "flex-start"},
			[
				textValue("Above"),
				box({borderStyle: "round", borderLeftColor: "green"}, [textValue("Content")]),
				textValue("Below"),
			],
		),
	),
	namedAnsiCase(
		"box/ansi-border-right-color",
		box(
			{flexDirection: "column", alignItems: "flex-start"},
			[
				textValue("Above"),
				box({borderStyle: "round", borderRightColor: "green"}, [textValue("Content")]),
				textValue("Below"),
			],
		),
	),
	namedAnsiCase(
		"box/change color of right border",
		box(
			{flexDirection: "column", alignItems: "flex-start"},
			[
				textValue("Above"),
				box({borderStyle: "round", borderRightColor: "green"}, [textValue("Content")]),
				textValue("Below"),
			],
		),
	),
	namedAnsiCase(
		"box/ansi-border-dim",
		box({borderDimColor: true, borderStyle: "round", alignSelf: "flex-start"}, [textValue("Content")]),
	),
	namedAnsiCase(
		"box/dim border color",
		box({borderDimColor: true, borderStyle: "round", alignSelf: "flex-start"}, [textValue("Content")]),
	),
	namedAnsiCase(
		"box/ansi-border-top-dim",
		box(
			{flexDirection: "column", alignItems: "flex-start"},
			[
				textValue("Above"),
				box({borderTopDimColor: true, borderStyle: "round"}, [textValue("Content")]),
				textValue("Below"),
			],
		),
	),
	namedAnsiCase(
		"box/dim top border color",
		box(
			{flexDirection: "column", alignItems: "flex-start"},
			[
				textValue("Above"),
				box({borderTopDimColor: true, borderStyle: "round"}, [textValue("Content")]),
				textValue("Below"),
			],
		),
	),
	namedAnsiCase(
		"box/ansi-border-bottom-dim",
		box(
			{flexDirection: "column", alignItems: "flex-start"},
			[
				textValue("Above"),
				box({borderBottomDimColor: true, borderStyle: "round"}, [textValue("Content")]),
				textValue("Below"),
			],
		),
	),
	namedAnsiCase(
		"box/dim bottom border color",
		box(
			{flexDirection: "column", alignItems: "flex-start"},
			[
				textValue("Above"),
				box({borderBottomDimColor: true, borderStyle: "round"}, [textValue("Content")]),
				textValue("Below"),
			],
		),
	),
	namedAnsiCase(
		"box/ansi-border-left-dim",
		box(
			{flexDirection: "column", alignItems: "flex-start"},
			[
				textValue("Above"),
				box({borderLeftDimColor: true, borderStyle: "round"}, [textValue("Content")]),
				textValue("Below"),
			],
		),
	),
	namedAnsiCase(
		"box/dim left border color",
		box(
			{flexDirection: "column", alignItems: "flex-start"},
			[
				textValue("Above"),
				box({borderLeftDimColor: true, borderStyle: "round"}, [textValue("Content")]),
				textValue("Below"),
			],
		),
	),
	namedAnsiCase(
		"box/ansi-border-right-dim",
		box(
			{flexDirection: "column", alignItems: "flex-start"},
			[
				textValue("Above"),
				box({borderRightDimColor: true, borderStyle: "round"}, [textValue("Content")]),
				textValue("Below"),
			],
		),
	),
	namedAnsiCase(
		"box/dim right border color",
		box(
			{flexDirection: "column", alignItems: "flex-start"},
			[
				textValue("Above"),
				box({borderRightDimColor: true, borderStyle: "round"}, [textValue("Content")]),
				textValue("Below"),
			],
		),
	),
	namedAnsiCase(
		"box/ansi-border-dim-preserves-child-style",
		box(
			{borderDimColor: true, borderStyle: "round", alignSelf: "flex-start"},
			[text([raw("styled text")], {bold: true, color: "blue"})],
		),
	),
	namedAnsiCase(
		"box/borderDimColor does not dim styled child Text touching left edge",
		box(
			{borderDimColor: true, borderStyle: "round", alignSelf: "flex-start"},
			[text([raw("styled text")], {bold: true, color: "blue"})],
		),
	),
	namedCase(
		"box/screen-reader-aria-label",
		box({"aria-label": "Hello World"}, [textValue("Not visible to screen readers")]),
		100,
		true,
	),
	namedCase(
		"box/render-text-for-screen-readers",
		box({"aria-label": "Hello World"}, [textValue("Not visible to screen readers")]),
		100,
		true,
	),
	namedCase(
		"box/render text for screen readers",
		box({"aria-label": "Hello World"}, [textValue("Not visible to screen readers")]),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-aria-hidden",
		box({"aria-hidden": true}, [textValue("Not visible to screen readers")]),
		100,
		true,
	),
	namedCase(
		"box/render-text-for-screen-readers-with-aria-hidden",
		box({"aria-hidden": true}, [textValue("Not visible to screen readers")]),
		100,
		true,
	),
	namedCase(
		"box/render text for screen readers with aria-hidden",
		box({"aria-hidden": true}, [textValue("Not visible to screen readers")]),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-role",
		box({"aria-role": "button"}, [textValue("Click me")]),
		100,
		true,
	),
	namedCase(
		"box/render-text-for-screen-readers-with-aria-role",
		box({"aria-role": "button"}, [textValue("Click me")]),
		100,
		true,
	),
	namedCase(
		"box/render text for screen readers with aria-role",
		box({"aria-role": "button"}, [textValue("Click me")]),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-role-empty",
		box({"aria-role": "button"}),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-list-selected",
		box(
			{"aria-role": "list", flexDirection: "column"},
			[
				textValue("Select a color:"),
				box({"aria-label": "1. Red", "aria-role": "listitem"}),
				box({"aria-label": "2. Green", "aria-role": "listitem", "aria-state": {selected: true}}),
				box({"aria-label": "3. Blue", "aria-role": "listitem"}),
			],
		),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-list-selected-with-children",
		box(
			{"aria-role": "list", flexDirection: "column"},
			[
				textValue("Select a color:"),
				box({"aria-label": "1. Red", "aria-role": "listitem"}, [textValue("Red")]),
				box({"aria-label": "2. Green", "aria-role": "listitem", "aria-state": {selected: true}}, [textValue("Green")]),
				box({"aria-label": "3. Blue", "aria-role": "listitem"}, [textValue("Blue")]),
			],
		),
		100,
		true,
	),
	namedCase(
		"box/render-select-input-for-screen-readers",
		box(
			{"aria-role": "list", flexDirection: "column"},
			[
				textValue("Select a color:"),
				box({"aria-label": "1. Red", "aria-role": "listitem"}, [textValue("Red")]),
				box({"aria-label": "2. Green", "aria-role": "listitem", "aria-state": {selected: true}}, [textValue("Green")]),
				box({"aria-label": "3. Blue", "aria-role": "listitem"}, [textValue("Blue")]),
			],
		),
		100,
		true,
	),
	namedCase(
		"box/render select input for screen readers",
		box(
			{"aria-role": "list", flexDirection: "column"},
			[
				textValue("Select a color:"),
				box({"aria-label": "1. Red", "aria-role": "listitem"}, [textValue("Red")]),
				box({"aria-label": "2. Green", "aria-role": "listitem", "aria-state": {selected: true}}, [textValue("Green")]),
				box({"aria-label": "3. Blue", "aria-role": "listitem"}, [textValue("Blue")]),
			],
		),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-label-only-text",
		text([], {"aria-label": "Screen-reader only"}),
		100,
		true,
	),
	namedCase(
		"box/render-aria-label-only-text-for-screen-readers",
		text([], {"aria-label": "Screen-reader only"}),
		100,
		true,
	),
	namedCase(
		"box/render aria-label only Text for screen readers",
		text([], {"aria-label": "Screen-reader only"}),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-label-only-box",
		box({"aria-label": "Screen-reader only"}),
		100,
		true,
	),
	namedCase(
		"box/render-aria-label-only-box-for-screen-readers",
		box({"aria-label": "Screen-reader only"}),
		100,
		true,
	),
	namedCase(
		"box/render aria-label only Box for screen readers",
		box({"aria-label": "Screen-reader only"}),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-styled-text",
		box({}, [text([raw("Styled content")], {bold: true, color: "green", inverse: true, underline: true})]),
		100,
		true,
	),
	namedCase(
		"box/omit-ansi-styling-in-screen-reader-output",
		box({}, [text([raw("Styled content")], {bold: true, color: "green", inverse: true, underline: true})]),
		100,
		true,
	),
	namedCase(
		"box/omit ANSI styling in screen-reader output",
		box({}, [text([raw("Styled content")], {bold: true, color: "green", inverse: true, underline: true})]),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-display-none",
		box({}, [box({display: "none"}, [textValue("Hidden")]), textValue("Visible")]),
		100,
		true,
	),
	namedCase(
		"box/skip-display-none-in-screen-reader-output",
		box({}, [box({display: "none"}, [textValue("Hidden")]), textValue("Visible")]),
		100,
		true,
	),
	namedCase(
		"box/skip nodes with display:none style in screen-reader output",
		box({}, [box({display: "none"}, [textValue("Hidden")]), textValue("Visible")]),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-multiple-text",
		box({flexDirection: "column"}, [textValue("Hello"), textValue("World")]),
		100,
		true,
	),
	namedCase(
		"box/render-multi-line-text",
		box({flexDirection: "column"}, [textValue("Line 1"), textValue("Line 2")]),
		100,
		true,
	),
	namedCase(
		"box/render multi-line text",
		box({flexDirection: "column"}, [textValue("Line 1"), textValue("Line 2")]),
		100,
		true,
	),
	namedCase(
		"box/render multi-line text",
		box({flexDirection: "column"}, [textValue("Line 1"), textValue("Line 2")]),
		100,
		true,
	),
	namedCase(
		"box/render-multiple-text-components-for-screen-readers",
		box({flexDirection: "column"}, [textValue("Hello"), textValue("World")]),
		100,
		true,
	),
	namedCase(
		"box/render multiple Text components",
		box({flexDirection: "column"}, [textValue("Hello"), textValue("World")]),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-nested-box",
		box({flexDirection: "column"}, [textValue("Hello"), box({}, [textValue("World")])]),
		100,
		true,
	),
	namedCase(
		"box/render-nested-box-components-with-text-for-screen-readers",
		box({flexDirection: "column"}, [textValue("Hello"), box({}, [textValue("World")])]),
		100,
		true,
	),
	namedCase(
		"box/render nested Box components with Text",
		box({flexDirection: "column"}, [textValue("Hello"), box({}, [textValue("World")])]),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-null-child",
		box({flexDirection: "column"}, [textValue("Hello"), empty(), textValue("World")]),
		100,
		true,
	),
	namedCase(
		"box/render-component-that-returns-null-for-screen-readers",
		box({flexDirection: "column"}, [textValue("Hello"), empty(), textValue("World")]),
		100,
		true,
	),
	namedCase(
		"box/render component that returns null",
		box({flexDirection: "column"}, [textValue("Hello"), empty(), textValue("World")]),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-busy",
		box({"aria-state": {busy: true}}, [textValue("Loading")]),
		100,
		true,
	),
	namedCase(
		"box/render-with-aria-state-busy",
		box({"aria-state": {busy: true}}, [textValue("Loading")]),
		100,
		true,
	),
	namedCase(
		"box/render with aria-state.busy",
		box({"aria-state": {busy: true}}, [textValue("Loading")]),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-busy-empty",
		box({"aria-state": {busy: true}}),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-checked",
		box({"aria-role": "checkbox", "aria-state": {checked: true}}, [textValue("Accept terms")]),
		100,
		true,
	),
	namedCase(
		"box/render-with-aria-state-checked",
		box({"aria-role": "checkbox", "aria-state": {checked: true}}, [textValue("Accept terms")]),
		100,
		true,
	),
	namedCase(
		"box/render with aria-state.checked",
		box({"aria-role": "checkbox", "aria-state": {checked: true}}, [textValue("Accept terms")]),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-disabled",
		box({"aria-role": "button", "aria-state": {disabled: true}}, [textValue("Submit")]),
		100,
		true,
	),
	namedCase(
		"box/render-with-aria-state-disabled",
		box({"aria-role": "button", "aria-state": {disabled: true}}, [textValue("Submit")]),
		100,
		true,
	),
	namedCase(
		"box/render with aria-state.disabled",
		box({"aria-role": "button", "aria-state": {disabled: true}}, [textValue("Submit")]),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-expanded",
		box({"aria-role": "combobox", "aria-state": {expanded: true}}, [textValue("Select")]),
		100,
		true,
	),
	namedCase(
		"box/render-with-aria-state-expanded",
		box({"aria-role": "combobox", "aria-state": {expanded: true}}, [textValue("Select")]),
		100,
		true,
	),
	namedCase(
		"box/render with aria-state.expanded",
		box({"aria-role": "combobox", "aria-state": {expanded: true}}, [textValue("Select")]),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-multiline-state",
		box({"aria-role": "textbox", "aria-state": {multiline: true}}, [textValue("Hello")]),
		100,
		true,
	),
	namedCase(
		"box/render-with-aria-state-multiline",
		box({"aria-role": "textbox", "aria-state": {multiline: true}}, [textValue("Hello")]),
		100,
		true,
	),
	namedCase(
		"box/render with aria-state.multiline",
		box({"aria-role": "textbox", "aria-state": {multiline: true}}, [textValue("Hello")]),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-multiselectable-state",
		box({"aria-role": "listbox", "aria-state": {multiselectable: true}}, [textValue("Options")]),
		100,
		true,
	),
	namedCase(
		"box/render-with-aria-state-multiselectable",
		box({"aria-role": "listbox", "aria-state": {multiselectable: true}}, [textValue("Options")]),
		100,
		true,
	),
	namedCase(
		"box/render with aria-state.multiselectable",
		box({"aria-role": "listbox", "aria-state": {multiselectable: true}}, [textValue("Options")]),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-readonly",
		box({"aria-role": "textbox", "aria-state": {readonly: true}}, [textValue("Hello")]),
		100,
		true,
	),
	namedCase(
		"box/render-with-aria-state-readonly",
		box({"aria-role": "textbox", "aria-state": {readonly: true}}, [textValue("Hello")]),
		100,
		true,
	),
	namedCase(
		"box/render with aria-state.readonly",
		box({"aria-role": "textbox", "aria-state": {readonly: true}}, [textValue("Hello")]),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-required",
		box({"aria-role": "textbox", "aria-state": {required: true}}, [textValue("Name")]),
		100,
		true,
	),
	namedCase(
		"box/render-with-aria-state-required",
		box({"aria-role": "textbox", "aria-state": {required: true}}, [textValue("Name")]),
		100,
		true,
	),
	namedCase(
		"box/render with aria-state.required",
		box({"aria-role": "textbox", "aria-state": {required: true}}, [textValue("Name")]),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-selected",
		box({"aria-role": "option", "aria-state": {selected: true}}, [textValue("Blue")]),
		100,
		true,
	),
	namedCase(
		"box/render-with-aria-state-selected",
		box({"aria-role": "option", "aria-state": {selected: true}}, [textValue("Blue")]),
		100,
		true,
	),
	namedCase(
		"box/render with aria-state.selected",
		box({"aria-role": "option", "aria-state": {selected: true}}, [textValue("Blue")]),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-selected-empty",
		box({"aria-role": "option", "aria-state": {selected: true}}),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-selected-disabled-expanded-empty",
		box({"aria-role": "option", "aria-state": {selected: true, disabled: true, expanded: true}}),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-expanded-selected-disabled-empty",
		box({"aria-role": "option", "aria-state": {expanded: true, selected: true, disabled: true}}),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-nested-multiline",
		box({flexDirection: "row"}, [box({flexDirection: "column"}, [textValue("Line 1"), textValue("Line 2")])]),
		100,
		true,
	),
	namedCase(
		"box/render-nested-multi-line-text",
		box({flexDirection: "row"}, [box({flexDirection: "column"}, [textValue("Line 1"), textValue("Line 2")])]),
		100,
		true,
	),
	namedCase(
		"box/render nested multi-line text",
		box({flexDirection: "row"}, [box({flexDirection: "column"}, [textValue("Line 1"), textValue("Line 2")])]),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-nested-row",
		box({flexDirection: "column"}, [box({flexDirection: "row"}, [textValue("Line 1"), textValue("Line 2")])]),
		100,
		true,
	),
	namedCase(
		"box/render-nested-row",
		box({flexDirection: "column"}, [box({flexDirection: "row"}, [textValue("Line 1"), textValue("Line 2")])]),
		100,
		true,
	),
	namedCase(
		"box/render nested row",
		box({flexDirection: "column"}, [box({flexDirection: "row"}, [textValue("Line 1"), textValue("Line 2")])]),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-multiline-roles",
		box(
			{flexDirection: "column", "aria-role": "list"},
			[
				box({"aria-role": "listitem"}, [textValue("Item 1")]),
				box({"aria-role": "listitem"}, [textValue("Item 2")]),
			],
		),
		100,
		true,
	),
	namedCase(
		"box/render-multi-line-text-with-roles",
		box(
			{flexDirection: "column", "aria-role": "list"},
			[
				box({"aria-role": "listitem"}, [textValue("Item 1")]),
				box({"aria-role": "listitem"}, [textValue("Item 2")]),
			],
		),
		100,
		true,
	),
	namedCase(
		"box/render multi-line text with roles",
		box(
			{flexDirection: "column", "aria-role": "list"},
			[
				box({"aria-role": "listitem"}, [textValue("Item 1")]),
				box({"aria-role": "listitem"}, [textValue("Item 2")]),
			],
		),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-listbox-multiselectable",
		box(
			{flexDirection: "column", "aria-role": "listbox", "aria-state": {multiselectable: true}},
			[
				box({"aria-role": "option", "aria-state": {selected: true}}, [textValue("Option 1")]),
				box({"aria-role": "option", "aria-state": {selected: false}}, [textValue("Option 2")]),
				box({"aria-role": "option", "aria-state": {selected: true}}, [textValue("Option 3")]),
			],
		),
		100,
		true,
	),
	namedCase(
		"box/render-listbox-with-multiselectable-options",
		box(
			{flexDirection: "column", "aria-role": "listbox", "aria-state": {multiselectable: true}},
			[
				box({"aria-role": "option", "aria-state": {selected: true}}, [textValue("Option 1")]),
				box({"aria-role": "option", "aria-state": {selected: false}}, [textValue("Option 2")]),
				box({"aria-role": "option", "aria-state": {selected: true}}, [textValue("Option 3")]),
			],
		),
		100,
		true,
	),
	namedCase(
		"box/render listbox with multiselectable options",
		box(
			{flexDirection: "column", "aria-role": "listbox", "aria-state": {multiselectable: true}},
			[
				box({"aria-role": "option", "aria-state": {selected: true}}, [textValue("Option 1")]),
				box({"aria-role": "option", "aria-state": {selected: false}}, [textValue("Option 2")]),
				box({"aria-role": "option", "aria-state": {selected: true}}, [textValue("Option 3")]),
			],
		),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-nested-same-role-via-wrapper",
		box(
			{flexDirection: "column", "aria-role": "list"},
			[box({}, [box({"aria-role": "list"}, [textValue("Nested")])])],
		),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-nested-same-role-row-wrapper",
		box(
			{"aria-role": "button"},
			[box({}, [box({"aria-role": "button"}, [textValue("Nested")])])],
		),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-nested-same-role-with-state-via-wrapper",
		box(
			{flexDirection: "column", "aria-role": "listbox"},
			[
				box(
					{},
					[box({"aria-role": "listbox", "aria-state": {multiselectable: true}}, [textValue("Options")])],
				),
			],
		),
		100,
		true,
	),
	// Upstream's renderNodeToScreenReaderOutput uses
	// Object.keys(state).filter(key => state[key]), so any truthy aria-state
	// key (not just the documented set) is announced. These cases pin down
	// that behaviour with a single arbitrary key (insertion-order ambiguity
	// is avoided so the golden is stable across renderers).
	namedCase(
		"box/screen-reader-arbitrary-state-key",
		box({"aria-role": "textbox", "aria-state": {invalid: true}}, [textValue("Email")]),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-arbitrary-state-key-without-role",
		box({"aria-state": {invalid: true}}, [textValue("Email")]),
		100,
		true,
	),
	// goink-specific parity gap-fillers for aria-labelledby / aria-describedby.
	// Upstream Ink does not yet implement these props, so the cases below
	// pin only the *fallback* behaviour that both renderers share: when the
	// referenced id does not exist in the tree, the host's own children are
	// narrated unchanged. Cases where an id resolves are pinned by the Go
	// renderer tests instead because they would diverge from upstream output.
	namedCase(
		"box/screen-reader-aria-labelledby-missing-falls-back",
		box({"aria-labelledby": "no-such-id"}, [textValue("fallback content")]),
		100,
		true,
	),
	namedCase(
		"box/screen-reader-aria-describedby-missing-no-op",
		box({"aria-describedby": "no-such-id"}, [textValue("only content")]),
		100,
		true,
	),
	// Box+Text style parity gap-fillers: overflow-with-border corner cases and
	// default-flexShrink interactions that were missing relative to upstream.
	namedCase(
		"box/overflow-hidden-border-top-left-text-overflow",
		box(
			{width: 6, height: 4, overflow: "hidden", borderStyle: "round"},
			[
				box(
					{marginTop: -1, marginLeft: -1, width: 6, height: 4, flexShrink: 0},
					[text([raw("AAAA\nBBBB\nCCCC\nDDDD")])],
				),
			],
		),
	),
	namedCase(
		"box/overflow-hidden-nested-borders",
		box(
			{width: 8, height: 5, overflow: "hidden", borderStyle: "round"},
			[
				box(
					{width: 6, height: 3, overflow: "hidden", borderStyle: "round", flexShrink: 0},
					[text([raw("AAAA\nBBBB\nCCCC")])],
				),
			],
		),
	),
	namedCase(
		"box/overflow-hidden-border-bottom-right-text-overflow",
		box(
			{width: 6, height: 4, overflow: "hidden", borderStyle: "round"},
			[
				box(
					{marginTop: 2, marginLeft: 2, width: 6, height: 4, flexShrink: 0},
					[text([raw("AAAA\nBBBB\nCCCC\nDDDD")])],
				),
			],
		),
	),
	namedCase(
		"box/default-flex-shrink-text-box-text-tight",
		box(
			{width: 8},
			[textValue("AAAA"), box({width: 4}, [textValue("BBBB")]), textValue("CCCC")],
		),
	),
	namedCase(
		"box/default-flex-shrink-flex-basis-row",
		box(
			{width: 10},
			[
				box({flexBasis: 4}, [textValue("AAAA")]),
				textValue("BBBB"),
				textValue("CCCC"),
			],
		),
	),
	namedCase(
		"box/default-flex-shrink-bordered-children-row",
		box(
			{width: 12},
			[
				box({borderStyle: "round"}, [textValue("AB")]),
				box({borderStyle: "round"}, [textValue("CD")]),
				box({borderStyle: "round"}, [textValue("EF")]),
			],
		),
	),
	// Yoga parity: proportional and fractional flex-grow distribution.
	namedCase(
		"box/yoga-flex-grow-proportional-3-1",
		box(
			{width: 8},
			[
				box({flexGrow: 3}, [textValue("A")]),
				box({flexGrow: 1}, [textValue("B")]),
			],
		),
	),
	namedCase(
		"box/yoga-flex-grow-2-1",
		box(
			{width: 10},
			[
				box({flexGrow: 2}, [textValue("A")]),
				box({flexGrow: 1}, [textValue("B")]),
			],
		),
	),
	namedCase(
		"box/yoga-flex-grow-fractional",
		box(
			{width: 12},
			[
				box({flexGrow: 0.5}, [textValue("A")]),
				box({flexGrow: 1}, [textValue("B")]),
			],
		),
	),
	namedCase(
		"box/yoga-flex-grow-zero-with-grow-sibling",
		box(
			{width: 8},
			[
				box({flexGrow: 0}, [textValue("A")]),
				box({flexGrow: 1}, [textValue("B")]),
			],
		),
	),
	namedCase(
		"box/yoga-flex-grow-with-flex-basis",
		box(
			{width: 10},
			[
				box({flexGrow: 1, flexBasis: 4}, [textValue("A")]),
				box({flexGrow: 1}, [textValue("B")]),
			],
		),
	),
	namedCase(
		"box/yoga-flex-grow-with-margin",
		box(
			{width: 10},
			[
				box({flexGrow: 1, marginLeft: 2}, [textValue("A")]),
				box({flexGrow: 1}, [textValue("B")]),
			],
		),
	),
	namedCase(
		"box/yoga-column-flex-grow-3-1",
		box(
			{flexDirection: "column", height: 8},
			[
				box({flexGrow: 3}, [textValue("A")]),
				box({flexGrow: 1}, [textValue("B")]),
			],
		),
	),
	namedCase(
		"box/yoga-column-fractional-grow",
		box(
			{flexDirection: "column", height: 7},
			[
				box({flexGrow: 0.5}, [textValue("A")]),
				box({flexGrow: 1.5}, [textValue("B")]),
			],
		),
	),
	// Yoga parity: negative gap clamps to zero.
	namedCase(
		"box/yoga-negative-gap",
		box(
			{gap: -1, width: 5},
			[textValue("A"), textValue("B"), textValue("C")],
		),
	),
	// Yoga parity: min-width semantics.
	namedCase(
		"box/yoga-min-width-larger-than-width",
		box(
			{},
			[
				box({width: 2, minWidth: 5}, [textValue("A")]),
				textValue("B"),
			],
		),
	),
	namedCase(
		"box/yoga-min-width-and-percent-width",
		box(
			{width: 8},
			[
				box({width: "25%", minWidth: 4}, [textValue("A")]),
				textValue("B"),
			],
		),
	),
	// Yoga parity: flex-shrink with explicit zero shrink does not adjust width.
	namedCase(
		"box/yoga-flex-shrink-zero-in-row",
		box(
			{width: 8},
			[
				box({flexShrink: 0, width: 5}, [textValue("AAAAA")]),
				box({flexShrink: 1, width: 5}, [textValue("BBBBB")]),
			],
		),
	),
	// Yoga parity: display:none child does not consume axis space.
	namedCase(
		"box/yoga-display-none-with-grow-sibling",
		box(
			{width: 6, flexDirection: "column"},
			[
				box({display: "none"}, [textValue("Hidden")]),
				box({flexGrow: 1}, [textValue("Visible")]),
			],
		),
	),
	// Yoga parity: absolute-positioned child does not push siblings or consume row space.
	namedCase(
		"box/yoga-absolute-with-flex-grow-siblings",
		box(
			{width: 8},
			[
				box({flexGrow: 1}, [textValue("A")]),
				box({position: "absolute"}, [textValue("Z")]),
				box({flexGrow: 1}, [textValue("B")]),
			],
		),
	),
	namedCase(
		"box/yoga-absolute-with-margin",
		box(
			{width: 10, height: 3},
			[
				textValue("Base"),
				box(
					{position: "absolute", marginLeft: 4, marginTop: 1},
					[textValue("Z")],
				),
			],
		),
	),
	// Yoga parity: align-self stretch overrides parent's alignItems flex-start.
	namedCase(
		"box/yoga-align-self-stretch-overrides-align-items",
		box(
			{flexDirection: "row", alignItems: "flex-start", height: 4},
			[
				box(
					{alignSelf: "stretch", borderStyle: "single"},
					[textValue("X")],
				),
				textValue("Y"),
			],
		),
	),
	// Yoga parity: minHeight as percentage resolves against parent height.
	namedCase(
		"box/yoga-min-height-percent",
		box(
			{flexDirection: "column", height: 6},
			[
				box({minHeight: "50%"}, [textValue("A")]),
				textValue("B"),
			],
		),
	),
];

const indexedStaticTemplate = text([raw("[{{index}}] {{item}}")]);

const prefixedStaticTemplate = box({}, [textValue(">"), text([raw("{{item}}")])]);

const labelledStaticTemplate = box({}, [textValue("item:"), text([raw("{{item}}")])]);

const transformStaticTemplate = preset =>
	transform(preset, [text([raw("{{item}}")])]);

const buildStaticCases = () => [
	namedCase("static/simple-0", staticNode({items: []})),
	namedCase("static/simple-1", staticNode({items: ["A"]})),
	namedCase("static/simple-2", staticNode({items: ["A", "B"]})),
	namedCase("static/simple-3", staticNode({items: ["A", "B", "C"]})),
	namedCase("static/simple-4", staticNode({items: ["A", "B", "C", "D"]})),
	namedCase("static/padding-left-1", staticNode({items: ["A"], props: {paddingLeft: 1}})),
	namedCase("static/padding-left-2", staticNode({items: ["A", "B"], props: {paddingLeft: 2}})),
	namedCase("static/padding-top-1", staticNode({items: ["A", "B"], props: {paddingTop: 1}})),
	namedCase("static/padding-bottom-1", staticNode({items: ["A", "B", "C"], props: {paddingBottom: 1}})),
	namedCase(
		"static/column-followup-angle",
		box(
			{flexDirection: "column"},
			[staticNode({items: ["A", "B", "C"], template: transformStaticTemplate("angle")}), textValue("done")],
		),
	),
	namedCase("static/indexed-1", staticNode({items: ["one"], template: indexedStaticTemplate})),
	namedCase("static/indexed-2", staticNode({items: ["one", "two"], template: indexedStaticTemplate})),
	namedCase("static/indexed-3", staticNode({items: ["one", "two", "three"], template: indexedStaticTemplate})),
	namedCase(
		"static/indexed-padding-left",
		staticNode({items: ["one", "two"], template: indexedStaticTemplate, props: {paddingLeft: 1}}),
	),
	namedCase(
		"static/indexed-padding-bottom",
		staticNode({items: ["one", "two", "three"], template: indexedStaticTemplate, props: {paddingBottom: 1}}),
	),
	namedCase("static/prefixed-1", staticNode({items: ["A"], template: prefixedStaticTemplate})),
	namedCase("static/prefixed-2", staticNode({items: ["A", "B"], template: prefixedStaticTemplate})),
	namedCase("static/prefixed-3", staticNode({items: ["A", "B", "C"], template: prefixedStaticTemplate})),
	namedCase("static/labelled-2", staticNode({items: ["A", "B"], template: labelledStaticTemplate})),
	namedCase("static/labelled-3", staticNode({items: ["go", "ink", "port"], template: labelledStaticTemplate})),
	namedCase("static/angle-1", staticNode({items: ["A"], template: transformStaticTemplate("angle")})),
	namedCase("static/angle-2", staticNode({items: ["A", "B"], template: transformStaticTemplate("angle")})),
	namedCase("static/upper-2", staticNode({items: ["go", "ink"], template: transformStaticTemplate("upper")})),
	namedCase("static/reverse-3", staticNode({items: ["ab", "cd", "ef"], template: transformStaticTemplate("reverse")})),
	namedCase("static/brace-2", staticNode({items: ["A", "B"], template: transformStaticTemplate("brace_index")})),
	namedCase(
		"static/screen-reader-list-parent-context",
		box({flexDirection: "column", "aria-role": "list"}, [
			staticNode({
				items: ["1. One", "2. Two"],
				template: {
					type: "box",
					props: {"aria-role": "listitem", "aria-label": "{{item}}"},
				},
			}),
		]),
		100,
		true,
	),
	namedCase(
		"static/column-followup-1",
		box({flexDirection: "column"}, [staticNode({items: ["A"]}), textValue("done")]),
	),
	namedCase(
		"static/column-followup-2",
		box({flexDirection: "column"}, [staticNode({items: ["A", "B"]}), textValue("done")]),
	),
	namedCase(
		"static/column-followup-padding",
		box(
			{flexDirection: "column"},
			[staticNode({items: ["A", "B", "C"], props: {paddingBottom: 1}}), textValue("done")],
		),
	),
	namedCase(
		"static/column-followup-indexed",
		box(
			{flexDirection: "column"},
			[staticNode({items: ["A", "B"], template: indexedStaticTemplate}), textValue("done")],
		),
	),
	namedCase(
		"static/column-followup-labelled",
		box(
			{flexDirection: "column"},
			[staticNode({items: ["A", "B"], template: labelledStaticTemplate}), textValue("done")],
		),
	),
	namedCase(
		"static/components-padding-followup-margin-top",
		box(
			{},
			[
				staticNode({items: ["A", "B", "C"], props: {paddingBottom: 1}}),
				box({marginTop: 1}, [textValue("X")]),
			],
		),
	),
	namedCase(
		"static/static output",
		box(
			{},
			[
				staticNode({items: ["A", "B", "C"], props: {paddingBottom: 1}}),
				box({marginTop: 1}, [textValue("X")]),
			],
		),
	),
	namedCase(
		"static/static output - concurrent",
		box(
			{},
			[
				staticNode({items: ["A", "B", "C"], props: {paddingBottom: 1}}),
				box({marginTop: 1}, [textValue("X")]),
			],
		),
	),
	namedCase(
		"static/skip previous output when rendering new static output",
		staticNode({items: ["A", "B"]}),
	),
	namedCase(
		"static/render only new items in static output on final render",
		staticNode({items: ["A", "B"]}),
	),
	namedManagedFramesCase(
		"static/render all frames if CI environment variable equals false",
		Array.from({length: 6}, (_, index) =>
			box({}, [
				staticNode({
					items: Array.from({length: index}, (__unused, itemIndex) => `#${itemIndex + 1}`),
				}),
				textValue(`Counter: ${index}`),
			]),
		),
		100,
		{CI: "false"},
	),
	namedManagedFramesCase(
		"static/screen-reader-runtime-static-delta",
		[
			box({flexDirection: "column"}, [
				staticNode({items: ["A"], template: text([raw("Static {{item}}")])}),
				box({"aria-role": "status"}, [textValue("One pending")]),
			]),
			box({flexDirection: "column"}, [
				staticNode({items: ["A", "B"], template: text([raw("Static {{item}}")])}),
				box({"aria-role": "status"}, [textValue("Done")]),
			]),
		],
		100,
		{
			screenReader: true,
			expectedContains: [
				"Static A\n",
				"status: One pending",
				"Static B\n",
				"status: Done",
			],
		},
	),
];

const buildMeasureCases = () => [
	namedRuntimeCase("measure/measure element", "runtime-measure-element"),
	namedRuntimeCase(
		"measure/calculate layout while rendering is throttled",
		"runtime-measure-element-throttled",
	),
];

const buildRenderCases = () => [
	namedRuntimeCase("render/throttle renders to maxFps", "runtime-throttle-maxfps"),
	namedRuntimeCase("render/no bsu/esu when output is unchanged", "runtime-throttle-tty-unchanged"),
	namedRuntimeCase("render/no bsu/esu when output and cursor are unchanged", "runtime-throttle-tty-unchanged-cursor"),
	namedRuntimeCase("render/bsu/esu wraps throttledLog trailing call", "runtime-throttle-tty-trailing"),
];

const coverageTargets = {
	text: 61,
	newline: 30,
	spacer: 30,
	transform: 33,
	box: 209,
	static: 31,
	measure: 2,
	render: 4,
};

const assertCoverage = cases => {
	const counts = new Map();

	for (const spec of cases) {
		const [family] = spec.name.split("/");
		counts.set(family, (counts.get(family) ?? 0) + 1);
	}

	for (const [family, minimum] of Object.entries(coverageTargets)) {
		const count = counts.get(family) ?? 0;
		if (count < minimum) {
			throw new Error(`expected at least ${minimum} cases for ${family}, got ${count}`);
		}
	}
};

export const buildCases = () => {
	const cases = [
		...buildTextCases(),
		...buildNewlineCases(),
		...buildSpacerCases(),
		...buildTransformCases(),
		...buildBoxCases(),
		...buildStaticCases(),
		...buildMeasureCases(),
		...buildRenderCases(),
	];

	assertCoverage(cases);

	return cases;
};
