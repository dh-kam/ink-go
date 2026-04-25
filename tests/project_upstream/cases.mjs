import {
	box,
	namedCase,
	raw,
	staticNode,
	text,
	textValue,
} from "./helpers.mjs";

const buildGeminiCliCases = () => [
	namedCase(
		"gemini-cli/checklist/completed-listitem",
		box({flexDirection: "row", columnGap: 1, "aria-role": "listitem"}, [
			text([raw("✓")], {color: "green", "aria-label": "Completed"}),
			box({flexShrink: 1}, [
				text([raw("Ship project-based golden suite")], {
					color: "gray",
					strikethrough: false,
				}),
			]),
		]),
		100,
		false,
		false,
		{
			project: "gemini-cli",
			sourceRepo: "google-gemini/gemini-cli",
			sourceFile: "packages/cli/src/ui/components/ChecklistItem.tsx",
		},
	),
	namedCase(
		"gemini-cli/checklist/in-progress-listitem",
		box({flexDirection: "row", columnGap: 1, "aria-role": "listitem"}, [
			text([raw("»")], {color: "cyan", "aria-label": "In Progress"}),
			box({flexShrink: 1}, [
				text([raw("Capture real Ink usage patterns")], {
					color: "cyan",
				}),
			]),
		]),
		100,
		false,
		false,
		{
			project: "gemini-cli",
			sourceRepo: "google-gemini/gemini-cli",
			sourceFile: "packages/cli/src/ui/components/ChecklistItem.tsx",
		},
	),
	namedCase(
		"gemini-cli/checklist/cancelled-listitem",
		box({flexDirection: "row", columnGap: 1, "aria-role": "listitem"}, [
			text([raw("✗")], {color: "red", "aria-label": "Cancelled"}),
			box({flexShrink: 1}, [
				text([raw("Discard outdated snapshot")], {
					color: "gray",
					strikethrough: true,
				}),
			]),
		]),
		100,
		false,
		false,
		{
			project: "gemini-cli",
			sourceRepo: "google-gemini/gemini-cli",
			sourceFile: "packages/cli/src/ui/components/ChecklistItem.tsx",
		},
	),
	namedCase(
		"gemini-cli/progress-bar/partial",
		box({flexDirection: "row"}, [
			text([raw("▬▬▬▬")], {color: "green"}),
			text([raw("▬▬▬▬▬▬")], {color: "gray"}),
		]),
		100,
		false,
		false,
		{
			project: "gemini-cli",
			sourceRepo: "google-gemini/gemini-cli",
			sourceFile: "packages/cli/src/ui/components/ProgressBar.tsx",
		},
	),
	namedCase(
		"gemini-cli/progress-bar/warning-threshold",
		box({flexDirection: "row"}, [
			text([raw("▬▬▬▬▬▬▬▬")], {color: "yellow"}),
			text([raw("▬▬")], {color: "gray"}),
		]),
		100,
		false,
		false,
		{
			project: "gemini-cli",
			sourceRepo: "google-gemini/gemini-cli",
			sourceFile: "packages/cli/src/ui/components/ProgressBar.tsx",
		},
	),
	namedCase(
		"gemini-cli/progress-bar/full-error",
		box({flexDirection: "row"}, [
			text([raw("▬▬▬▬▬▬▬▬▬▬")], {color: "red"}),
			text([raw("")], {color: "gray"}),
		]),
		100,
		false,
		false,
		{
			project: "gemini-cli",
			sourceRepo: "google-gemini/gemini-cli",
			sourceFile: "packages/cli/src/ui/components/ProgressBar.tsx",
		},
	),
	namedCase(
		"gemini-cli/loading-indicator/inline-thinking",
		box({}, [
			box({marginRight: 1}, [textValue("⠋")]),
			box({flexShrink: 1}, [
				text([raw("Thinking about the next shell step")], {
					color: "white",
					italic: true,
					wrap: "truncate-end",
				}),
			]),
			box({flexShrink: 0, width: 1}, []),
			textValue("(esc to cancel, 7s)"),
		]),
		100,
		false,
		false,
		{
			project: "gemini-cli",
			sourceRepo: "google-gemini/gemini-cli",
			sourceFile: "packages/cli/src/ui/components/LoadingIndicator.tsx",
		},
	),
	namedCase(
		"gemini-cli/loading-indicator/inline-waiting-for-confirmation",
		box({}, [
			box({marginRight: 1}, [textValue("⠏")]),
			box({flexShrink: 1}, [
				text([raw("Waiting for confirmation")], {
					color: "white",
					italic: true,
					wrap: "truncate-end",
				}),
				text([raw(" (press tab to focus)")], {
					color: "cyan",
					italic: true,
				}),
			]),
		]),
		100,
		false,
		false,
		{
			project: "gemini-cli",
			sourceRepo: "google-gemini/gemini-cli",
			sourceFile: "packages/cli/src/ui/components/LoadingIndicator.tsx",
		},
	),
];

const buildNeovateCodeCases = () => [
	namedCase(
		"neovate-code/messages/static-header-and-dynamic-pending",
		box({flexDirection: "column"}, [
			staticNode({
				items: ["header", "Completed item A", "Completed item B"],
				template: box({flexDirection: "column"}, [
					text([raw("{{item}}")]),
				]),
			}),
			box({flexDirection: "column"}, [
				textValue("Pending item A"),
				textValue("Pending item B"),
			]),
		]),
		100,
		false,
		false,
		{
			project: "neovate-code",
			sourceRepo: "neovateai/neovate-code",
			sourceFile: "src/ui/Messages.tsx",
		},
	),
	namedCase(
		"neovate-code/bash-command-message",
		box({flexDirection: "column", marginTop: 1, marginLeft: 2}, [
			box({}, [
				text([raw("! ")], {color: "yellow", bold: true}),
				text([raw("git status --short")], {bold: true, color: "cyan"}),
			]),
		]),
		100,
		false,
		false,
		{
			project: "neovate-code",
			sourceRepo: "neovateai/neovate-code",
			sourceFile: "src/ui/Messages.tsx",
		},
	),
	namedCase(
		"neovate-code/user-message-bubble",
		box({flexDirection: "column", marginTop: 1, marginLeft: 2}, [
			text([raw("dh.kam")], {bold: true, color: "cyan"}),
			box({}, [
				text([raw("Ship the Go port parity work.")], {
					color: "white",
					backgroundColor: "#555555",
				}),
			]),
		]),
		100,
		false,
		false,
		{
			project: "neovate-code",
			sourceRepo: "neovateai/neovate-code",
			sourceFile: "src/ui/Messages.tsx",
		},
	),
	namedCase(
		"neovate-code/user-cancel-notice",
		box({flexDirection: "column", marginTop: 1, marginLeft: 2}, [
			text([raw("dh.kam")], {bold: true, color: "cyan"}),
			text([raw("User canceled the request")], {color: "red"}),
		]),
		100,
		false,
		false,
		{
			project: "neovate-code",
			sourceRepo: "neovateai/neovate-code",
			sourceFile: "src/ui/Messages.tsx",
		},
	),
	namedCase(
		"neovate-code/approval-modal-invalid-questions",
		box({flexDirection: "column", padding: 1, borderStyle: "round", borderColor: "red"}, [
			text([raw("Invalid Questions")], {color: "red", bold: true}),
			textValue("No questions provided to askUserQuestion tool"),
			box({marginTop: 1}, [
				text([raw("Press enter or esc to cancel")], {dimColor: true}),
			]),
		]),
		100,
		false,
		false,
		{
			project: "neovate-code",
			sourceRepo: "neovateai/neovate-code",
			sourceFile: "src/ui/ApprovalModal.tsx",
		},
	),
	namedCase(
		"neovate-code/thinking-block-with-hidden-lines",
		box({flexDirection: "column", marginTop: 1}, [
			text([raw("thinking")], {bold: true, color: "gray"}),
			text([raw("Working through the shell plan line by line")], {
				color: "gray",
				italic: true,
			}),
			text([raw("... 3 more lines hidden ...")], {color: "gray", dimColor: true}),
		]),
		100,
		false,
		false,
		{
			project: "neovate-code",
			sourceRepo: "neovateai/neovate-code",
			sourceFile: "src/ui/Messages.tsx",
		},
	),
	namedCase(
		"neovate-code/queued-messages-warning-card",
		box({flexDirection: "column", borderStyle: "round", borderColor: "yellow", paddingX: 1}, [
			box({flexDirection: "row", justifyContent: "space-between"}, [
				text([raw("2 queued messages")], {bold: true, color: "yellow"}),
			]),
			box({flexDirection: "column", marginTop: 1}, [
				textValue("1. Summarize the repo state"),
				textValue("2. Create project-based goldens"),
			]),
			box({marginTop: 1}, [
				text([raw("They will run after the current task finishes")], {color: "yellow"}),
			]),
		]),
		100,
		false,
		false,
		{
			project: "neovate-code",
			sourceRepo: "neovateai/neovate-code",
			sourceFile: "src/ui/QueueDisplay.tsx",
		},
	),
	namedCase(
		"neovate-code/transcript-mode-banner",
		box({flexDirection: "column", marginTop: 1}, [
			box({borderStyle: "single", borderBottom: false, borderLeft: false, borderRight: false}, []),
			box({}, [
				text([raw("Showing detailed transcript. Press Esc or Ctrl+C to exit.")], {dimColor: true}),
			]),
		]),
		100,
		false,
		false,
		{
			project: "neovate-code",
			sourceRepo: "neovateai/neovate-code",
			sourceFile: "src/ui/TranscriptModeIndicator.tsx",
		},
	),
	namedCase(
		"neovate-code/approval-modal-edit-title",
		box({}, [
			text([raw("Edit file ")], {bold: true, color: "yellow"}),
			textValue("src/ui/App.tsx"),
		]),
		100,
		false,
		false,
		{
			project: "neovate-code",
			sourceRepo: "neovateai/neovate-code",
			sourceFile: "src/ui/ApprovalModal.tsx",
		},
	),
	namedCase(
		"neovate-code/approval-modal-bash-preview",
		box({flexDirection: "column", marginBottom: 1}, [
			box({marginLeft: 2}, [
				textValue("npm test -- --watch"),
			]),
			box({marginLeft: 2}, [
				text([raw("Run focused test file in watch mode")], {color: "gray"}),
			]),
		]),
		100,
		false,
		false,
		{
			project: "neovate-code",
			sourceRepo: "neovateai/neovate-code",
			sourceFile: "src/ui/ApprovalModal.tsx",
		},
	),
	namedCase(
		"neovate-code/app-main-column-and-footer-modals",
		box({flexDirection: "column"}, [
			box({flexDirection: "column"}, [
				textValue("Messages"),
				textValue("BackgroundPrompt"),
				textValue("ActivityIndicator"),
				textValue("QueueDisplay"),
				textValue("ChatInput"),
				textValue("ExitHint"),
				textValue("Debug"),
			]),
			textValue("ApprovalModal"),
			textValue("SlashCommandJSX"),
		]),
		100,
		false,
		false,
		{
			project: "neovate-code",
			sourceRepo: "neovateai/neovate-code",
			sourceFile: "src/ui/App.tsx",
		},
	),
];

const buildShopifyCliCases = () => [
	namedCase(
		"shopify-cli/prompt-layout/submitted-confirmation-row",
		box({flexDirection: "row"}, [
			box({width: 3}, [
				text([raw("✓")], {color: "cyan"}),
			]),
			text([raw("Apply changes")], {color: "cyan"}),
		]),
		100,
		false,
		false,
		{
			project: "shopify-cli",
			sourceRepo: "Shopify/cli",
			sourceFile: "packages/cli-kit/src/private/node/ui/components/Prompts/PromptLayout.tsx",
		},
	),
];

const buildTweakccCases = () => [
	namedCase(
		"tweakcc/main-view/left-border-notification-banner",
		box({
			borderLeft: true,
			borderStyle: "single",
			borderColor: "blue",
			paddingLeft: 1,
			flexDirection: "column",
		}, [
			text([raw("Theme changes staged but not yet applied")], {color: "blue"}),
		]),
		100,
		false,
		false,
		{
			project: "tweakcc",
			sourceRepo: "Piebald-AI/tweakcc",
			sourceFile: "src/ui/components/MainView.tsx",
		},
	),
];

const buildNanocoderCases = () => [
	namedCase(
		"nanocoder/assistant-message/response-card",
		box({flexDirection: "column"}, [
			box({marginBottom: 1}, [
				text([raw("gpt-5.4:")], {color: "cyan", bold: true}),
			]),
			box({
				flexDirection: "column",
				marginBottom: 1,
				padding: 1,
				borderLeft: true,
				borderStyle: "single",
				borderLeftColor: "gray",
				backgroundColor: "black",
			}, [
				textValue("Generated answer content"),
			]),
			box({marginBottom: 2}, [
				text([raw("~234 tokens")], {color: "gray"}),
			]),
		]),
		100,
		false,
		false,
		{
			project: "nanocoder",
			sourceRepo: "nano-collective/nanocoder",
			sourceFile: "source/components/assistant-message.tsx",
		},
	),
];

export const buildCases = () => [
	...buildGeminiCliCases(),
	...buildNeovateCodeCases(),
	...buildShopifyCliCases(),
	...buildTweakccCases(),
	...buildNanocoderCases(),
];
