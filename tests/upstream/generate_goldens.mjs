import fs from "node:fs/promises";
import path from "node:path";
import {EventEmitter} from "node:events";
import {fileURLToPath} from "node:url";

import React, {useEffect, useRef, useState} from "../../../ink/node_modules/react/index.js";
import FakeTimers from "../../../ink/node_modules/@sinonjs/fake-timers/src/fake-timers-src.js";
import chalk from "../../../ink/node_modules/chalk/source/index.js";
import {
	Box,
	Newline,
	Spacer,
	Static,
	Text,
	Transform,
	measureElement,
	render,
	useCursor,
} from "../../../ink/build/index.js";

import {buildCases} from "./cases.mjs";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const originalStderrWrite = process.stderr.write.bind(process.stderr);
process.stderr.write = (chunk, ...args) => {
	if (String(chunk) === "\u001B[?25h") {
		return true;
	}

	return originalStderrWrite(chunk, ...args);
};

const casesPath = path.join(__dirname, "cases.json");
const goldensPath = path.join(__dirname, "goldens.json");

class FakeStdout extends EventEmitter {
	constructor(columns, {tty = false} = {}) {
		super();
		this.columns = columns;
		if (tty) {
			this.isTTY = true;
			this.rows = 24;
		}
		this._output = "";
		this._writes = [];
	}

	write(chunk) {
		const value = String(chunk);
		this._output = value;
		this._writes.push(value);
		return true;
	}

	get() {
		return this._output;
	}

	joined() {
		return this._writes.join("");
	}
}

const transformPreset = preset => {
	switch (preset) {
		case "identity":
			return children => children;
		case "bracket_index":
			return (children, index) => `[${index}: ${children}]`;
		case "brace_index":
			return (children, index) => `{${index}: ${children}}`;
		case "angle":
			return children => `<${children}>`;
		case "reverse":
			return children => [...children].reverse().join("");
		case "upper":
			return children => children.toUpperCase();
		default:
			throw new Error(`unknown transform preset: ${preset}`);
	}
};

const replaceTemplateString = (value, item, index) =>
	String(value ?? "")
		.replaceAll("{{item}}", item)
		.replaceAll("{{index}}", String(index));

const replaceTemplateValue = (value, item, index) => {
	if (typeof value === "string") {
		return replaceTemplateString(value, item, index);
	}

	if (Array.isArray(value)) {
		return value.map(entry => replaceTemplateValue(entry, item, index));
	}

	if (value && typeof value === "object") {
		return Object.fromEntries(
			Object.entries(value).map(([key, nestedValue]) => [
				key,
				replaceTemplateValue(nestedValue, item, index),
			]),
		);
	}

	return value;
};

const instantiateTemplate = (spec, item, index) => {
	const result = {
		...spec,
		type: replaceTemplateString(spec.type, item, index),
		value: replaceTemplateString(spec.value ?? "", item, index),
		preset: replaceTemplateString(spec.preset ?? "", item, index),
	};

	if (spec.props) {
		result.props = replaceTemplateValue(spec.props, item, index);
	}

	if (spec.children) {
		result.children = spec.children.map(child => instantiateTemplate(child, item, index));
	}

	if (spec.template) {
		result.template = instantiateTemplate(spec.template, item, index);
	}

	if (spec.items) {
		result.items = spec.items.map(childItem => replaceTemplateString(childItem, item, index));
	}

	return result;
};

const ensureKey = (spec, key) => {
	if (spec.type === "raw") {
		return spec;
	}

	return {
		...spec,
		props: {
			...(spec.props ?? {}),
			key,
		},
	};
};

const buildNode = spec => {
	switch (spec.type) {
		case "raw":
			return spec.value ?? "";
		case "empty":
			return null;
		case "text":
			return React.createElement(
				Text,
				spec.props ?? {},
				...(spec.children ?? []).map(buildNode),
			);
		case "box":
			return React.createElement(
				Box,
				spec.props ?? {},
				...(spec.children ?? []).map(buildNode),
			);
		case "newline":
			return React.createElement(Newline, {count: spec.count ?? 1});
		case "spacer":
			return React.createElement(Spacer, spec.props ?? {});
		case "static": {
			const template =
				spec.template ??
				{
					type: "text",
					children: [{type: "raw", value: "{{item}}"}],
				};

			return React.createElement(
				Static,
				{
					items: spec.items ?? [],
					style: spec.props ?? {},
				},
				(item, index) =>
					buildNode(ensureKey(instantiateTemplate(template, item, index), `static-${index}`)),
			);
		}
		case "transform":
			return React.createElement(
				Transform,
				{
					...(spec.props ?? {}),
					transform: transformPreset(spec.preset),
				},
				...(spec.children ?? []).map(buildNode),
			);
		default:
			throw new Error(`unknown node type: ${spec.type}`);
	}
};

const renderToString = (node, columns, screenReader = false, ansi = false) => {
	const stdout = new FakeStdout(columns ?? 100);
	const previousLevel = chalk.level;

	try {
		chalk.level = ansi ? 3 : 0;

		render(node, {
			stdout,
			debug: true,
			isScreenReaderEnabled: screenReader,
		});

		return stdout.get();
	} finally {
		chalk.level = previousLevel;
	}
};

const delay = milliseconds => new Promise(resolve => {
	setTimeout(resolve, milliseconds);
});

const measureElementFixture = () => {
	function Test() {
		const [width, setWidth] = useState(0);
		const ref = useRef(null);

		useEffect(() => {
			if (!ref.current) {
				return;
			}

			setWidth(measureElement(ref.current).width);
		}, []);

		return React.createElement(
			Box,
			{ref},
			React.createElement(Text, null, "Width: ", width),
		);
	}

	return React.createElement(Test);
};

const renderMeasureElementWrites = async (columns, throttled = false) => {
	const stdout = new FakeStdout(columns ?? 100, {tty: throttled});
	let instance;

	if (throttled) {
		instance = render(null, {stdout, patchConsole: false});
		instance.rerender(measureElementFixture());
		await delay(80);
	} else {
		instance = render(measureElementFixture(), {
			stdout,
			debug: true,
			patchConsole: false,
		});
		await delay(120);
	}

	const writes = [...stdout._writes];
	instance.unmount();
	return writes;
};

const throttleTextFixture = value => React.createElement(Text, null, value);

const throttleCursorFixture = value => {
	function Test() {
		const {setCursorPosition} = useCursor();
		setCursorPosition({x: 0, y: 0});

		return React.createElement(Text, null, value);
	}

	return React.createElement(Test);
};

const renderThrottleMaxFPSWrites = columns => {
	const clock = FakeTimers.install();
	const stdout = new FakeStdout(columns ?? 100);
	let instance;

	try {
		instance = render(throttleTextFixture("Hello"), {
			stdout,
			maxFps: 20,
			patchConsole: false,
		});
		instance.rerender(throttleTextFixture("World"));
		clock.tick(49);
		clock.tick(1);

		return [...stdout._writes];
	} finally {
		instance?.unmount();
		clock.uninstall();
	}
};

const renderTTYThrottleWrites = (columns, scenario) => {
	const clock = FakeTimers.install();
	const stdout = new FakeStdout(columns ?? 100, {tty: true});
	let instance;

	try {
		const fixture = scenario === "unchanged-cursor" ? throttleCursorFixture : throttleTextFixture;
		instance = render(fixture("Hello"), {
			stdout,
			maxFps: 20,
			patchConsole: false,
		});

		stdout._writes = [];
		instance.rerender(fixture(scenario === "trailing" ? "World" : "Hello"));
		clock.tick(49);
		clock.tick(1);

		return [...stdout._writes];
	} finally {
		instance?.unmount();
		clock.uninstall();
	}
};

const withTemporaryEnv = (env, fn) => {
	const previous = new Map();

	for (const [key, value] of Object.entries(env ?? {})) {
		const existing = Object.prototype.hasOwnProperty.call(process.env, key)
			? process.env[key]
			: undefined;
		previous.set(key, existing);

		process.env[key] = String(value);
	}

	try {
		return fn();
	} finally {
		for (const [key, value] of previous.entries()) {
			if (value === undefined) {
				delete process.env[key];
				continue;
			}

			process.env[key] = value;
		}
	}
};

const buildGolden = async spec => {
	const base = {
		name: spec.name,
		columns: spec.columns ?? 100,
		screenReader: spec.screenReader ?? false,
		ansi: spec.ansi ?? false,
		mode: spec.mode ?? "",
	};

	switch (spec.mode) {
		case "error":
			if (!spec.expectedError) {
				throw new Error(`missing expectedError for ${spec.name}`);
			}

			return {
				...base,
				error: spec.expectedError,
			};
		case "managed-frames":
			return {
				...base,
				contains: spec.expectedContains ?? [],
			};
		case "runtime-measure-element":
			return {
				...base,
				writes: await renderMeasureElementWrites(spec.columns ?? 100),
			};
		case "runtime-measure-element-throttled":
			return {
				...base,
				writes: await renderMeasureElementWrites(spec.columns ?? 100, true),
			};
		case "runtime-throttle-maxfps":
			return {
				...base,
				writes: renderThrottleMaxFPSWrites(spec.columns ?? 100),
			};
		case "runtime-throttle-tty-unchanged":
			return {
				...base,
				writes: renderTTYThrottleWrites(spec.columns ?? 100, "unchanged"),
			};
		case "runtime-throttle-tty-unchanged-cursor":
			return {
				...base,
				writes: renderTTYThrottleWrites(spec.columns ?? 100, "unchanged-cursor"),
			};
		case "runtime-throttle-tty-trailing":
			return {
				...base,
				writes: renderTTYThrottleWrites(spec.columns ?? 100, "trailing"),
			};
		default:
			return {
				...base,
				output: renderToString(
					buildNode(spec.node),
					spec.columns ?? 100,
					spec.screenReader ?? false,
					spec.ansi ?? false,
				),
			};
	}
};

const main = async () => {
	const specs = buildCases();
	await fs.writeFile(casesPath, `${JSON.stringify(specs, null, 2)}\n`);

	const results = [];
	for (const spec of specs) {
		results.push(await buildGolden(spec));
	}

	await fs.writeFile(goldensPath, `${JSON.stringify(results, null, 2)}\n`);
};

await main();
