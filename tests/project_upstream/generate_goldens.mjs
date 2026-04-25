import fs from "node:fs/promises";
import path from "node:path";
import {EventEmitter} from "node:events";
import {fileURLToPath} from "node:url";

import React from "../../../ink/node_modules/react/index.js";
import chalk from "../../../ink/node_modules/chalk/source/index.js";
import {
	Box,
	Newline,
	Spacer,
	Static,
	Text,
	Transform,
	render,
} from "../../../ink/build/index.js";

import {buildCases} from "./cases.mjs";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const casesPath = path.join(__dirname, "cases.json");
const goldensPath = path.join(__dirname, "goldens.json");

class FakeStdout extends EventEmitter {
	constructor(columns) {
		super();
		this.columns = columns;
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

const buildGolden = spec => {
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

	const results = specs.map(buildGolden);

	await fs.writeFile(goldensPath, `${JSON.stringify(results, null, 2)}\n`);
};

await main();
