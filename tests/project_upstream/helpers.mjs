export const raw = value => ({type: "raw", value});

export const compactChildren = children =>
	children.filter(child => child !== undefined && child !== null);

export const text = (children = [], props = undefined) => ({
	type: "text",
	...(props && Object.keys(props).length > 0 ? {props} : {}),
	children: compactChildren(children),
});

export const box = (props = undefined, children = []) => ({
	type: "box",
	...(props && Object.keys(props).length > 0 ? {props} : {}),
	children: compactChildren(children),
});

export const newline = count => ({
	type: "newline",
	...(count > 1 ? {count} : {}),
});

export const spacer = () => ({type: "spacer"});
export const empty = () => ({type: "empty"});

export const transform = (preset, children = [], props = undefined) => ({
	type: "transform",
	preset,
	...(props && Object.keys(props).length > 0 ? {props} : {}),
	children: compactChildren(children),
});

export const staticNode = ({items = [], template = undefined, props = undefined, children = []}) => ({
	type: "static",
	...(props && Object.keys(props).length > 0 ? {props} : {}),
	...(items.length > 0 ? {items} : {}),
	...(template ? {template} : {}),
	...(children.length > 0 ? {children: compactChildren(children)} : {}),
});

export const textValue = value => text([raw(value)]);

export const namedCase = (
	name,
	node,
	columns = 100,
	screenReader = false,
	ansi = false,
	metadata = undefined,
) => ({
	name,
	columns,
	screenReader,
	ansi,
	...(metadata ?? {}),
	node,
});

export const namedAnsiCase = (name, node, columns = 100, metadata = undefined) =>
	namedCase(name, node, columns, false, true, metadata);

export const namedErrorCase = (name, node, expectedError, columns = 100, metadata = undefined) => ({
	name,
	columns,
	mode: "error",
	...(metadata ?? {}),
	expectedError,
	node,
});

export const namedManagedFramesCase = (
	name,
	frames,
	columns = 100,
	env = undefined,
	metadata = undefined,
) => ({
	name,
	columns,
	mode: "managed-frames",
	...(env && Object.keys(env).length > 0 ? {env} : {}),
	...(metadata ?? {}),
	frames,
});
