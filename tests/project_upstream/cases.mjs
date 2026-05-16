import {
	box,
	namedCase,
	raw,
	staticNode,
	text,
	textValue,
} from "./helpers.mjs";

const geminiCliMeta = sourceFile => ({
	project: "gemini-cli",
	sourceRepo: "google-gemini/gemini-cli",
	sourceFile,
});

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
	namedCase(
		"gemini-cli/layout/default-app-normal-composer",
		box({flexDirection: "column", width: 72, height: 12, paddingBottom: 1}, [
			box({flexDirection: "column", flexGrow: 1, overflow: "hidden"}, [
				textValue("Gemini CLI"),
				textValue("Static history item"),
				textValue("Pending assistant response"),
			]),
			box({flexDirection: "column", width: 72}, [
				box({flexDirection: "column", marginY: 1}, [
					text([raw("⚠ "), raw("Sandbox is disabled for this session")], {color: "yellow"}),
				]),
				textValue("> Summarize the current workspace"),
				text([raw("esc to cancel")], {dimColor: true}),
			]),
		]),
		80,
		false,
		false,
		geminiCliMeta("packages/cli/src/ui/layouts/DefaultAppLayout.tsx"),
	),
	namedCase(
		"gemini-cli/layout/default-app-dialog-manager",
		box({flexDirection: "column", width: 72, height: 10, paddingBottom: 1}, [
			box({flexDirection: "column", flexGrow: 1, overflow: "hidden"}, [
				textValue("MainContent"),
				textValue("Last response"),
			]),
			box({flexDirection: "column", width: 72}, [
				textValue("Notifications"),
				box({borderStyle: "round", borderColor: "cyan", paddingX: 1}, [
					textValue("DialogManager: approve tool call?"),
				]),
				text([raw("ctrl+c twice to exit")], {dimColor: true}),
			]),
		]),
		80,
		false,
		false,
		geminiCliMeta("packages/cli/src/ui/layouts/DefaultAppLayout.tsx"),
	),
	namedCase(
		"gemini-cli/layout/screen-reader-shell",
		box({flexDirection: "column", width: "90%", height: "100%"}, [
			textValue("Notifications"),
			textValue("Footer metadata"),
			box({flexDirection: "column", flexGrow: 1, overflow: "hidden"}, [
				textValue("MainContent"),
				textValue("Scrollable transcript"),
			]),
			textValue("Composer"),
			text([raw("Exit warning")], {dimColor: true}),
		]),
		80,
		false,
		false,
		geminiCliMeta("packages/cli/src/ui/layouts/ScreenReaderAppLayout.tsx"),
	),
	namedCase(
		"gemini-cli/header/wide-metadata-row",
		box({flexDirection: "column"}, [
			box({flexDirection: "row", marginTop: 1, marginBottom: 1, paddingLeft: 1}, [
				box({flexShrink: 0}, [
					text([raw("✦")], {color: "cyan", bold: true}),
				]),
				box({marginLeft: 2, flexDirection: "column"}, [
					box({flexDirection: "row"}, [
						text([raw("Gemini CLI")], {bold: true}),
						textValue(" v0.1.0"),
						box({marginLeft: 2}, [
							text([raw("⠋ Updating")], {color: "gray"}),
						]),
					]),
					text([raw("Signed in with Google: user@example.com")], {dimColor: true}),
				]),
			]),
		]),
		100,
		false,
		false,
		geminiCliMeta("packages/cli/src/ui/components/AppHeader.tsx"),
	),
	namedCase(
		"gemini-cli/header/narrow-wordmark-column",
		box({flexDirection: "column"}, [
			box({flexDirection: "column", marginTop: 1, marginBottom: 1, paddingLeft: 1}, [
				box({flexDirection: "row"}, [
					text([raw("✦")], {color: "cyan", bold: true}),
					box({marginLeft: 3}, [
						textValue("  ____                _       _\n / ___| ___ _ __ ___ (_)_ __ (_)")
					]),
				]),
				box({marginTop: 1}, [
					textValue("Gemini CLI v0.1.0"),
				]),
			]),
		]),
		48,
		false,
		false,
		geminiCliMeta("packages/cli/src/ui/components/AppHeader.tsx"),
	),
	namedCase(
		"gemini-cli/user-identity/google-auth-row",
		box({flexDirection: "column"}, [
			box({flexDirection: "row"}, [
				text([
					text([raw("Signed in with Google:")], {bold: true}),
					raw(" user@example.com"),
				], {wrap: "truncate-end"}),
				text([raw(" /auth")], {dimColor: true}),
			]),
		]),
		80,
		false,
		false,
		geminiCliMeta("packages/cli/src/ui/components/UserIdentity.tsx"),
	),
	namedCase(
		"gemini-cli/notifications/startup-warning-stack",
		box({flexDirection: "column", marginY: 1}, [
			box({flexDirection: "row"}, [
				box({width: 3}, [text([raw("⚠ ")], {color: "yellow"})]),
				box({flexGrow: 1}, [text([raw("MCP server alpha failed to start")], {color: "yellow"})]),
			]),
			box({flexDirection: "row"}, [
				box({width: 3}, [text([raw("⚠ ")], {color: "yellow"})]),
				box({flexGrow: 1}, [text([raw("Extension beta is disabled")], {color: "yellow"})]),
			]),
		]),
		100,
		false,
		false,
		geminiCliMeta("packages/cli/src/ui/components/Notifications.tsx"),
	),
	namedCase(
		"gemini-cli/notifications/initialization-error-card",
		box({borderStyle: "round", borderColor: "red", paddingX: 1, marginBottom: 1}, [
			text([raw("Initialization Error: invalid API key")], {color: "red"}),
			text([raw(" Please check API key and configuration.")], {color: "red"}),
		]),
		100,
		false,
		false,
		geminiCliMeta("packages/cli/src/ui/components/Notifications.tsx"),
	),
	namedCase(
		"gemini-cli/update-notification/warning-bar",
		box({borderStyle: "round", borderColor: "yellow", paddingX: 1, marginY: 1}, [
			text([raw("Update available: run npm install -g @google/gemini-cli")], {color: "yellow"}),
		]),
		100,
		false,
		false,
		geminiCliMeta("packages/cli/src/ui/components/UpdateNotification.tsx"),
	),
	namedCase(
		"gemini-cli/config-init/mcp-server-counts",
		box({marginTop: 1}, [
			text([
				raw("⠙ "),
				text([raw("Connecting to MCP servers: 2 connected, waiting for alpha")], {color: "white"}),
			]),
		]),
		100,
		false,
		false,
		geminiCliMeta("packages/cli/src/ui/components/ConfigInitDisplay.tsx"),
	),
	namedCase(
		"gemini-cli/status-row/minimal-mode",
		box({flexDirection: "column", width: "100%"}, [
			box({flexDirection: "row", justifyContent: "space-between", minHeight: 1}, [
				box({flexDirection: "row", flexGrow: 1, flexShrink: 1}, [
					box({flexDirection: "row", columnGap: 1}, [
						text([raw("⠋ Thinking")], {italic: true}),
						box({}, [text([raw("● auto")], {color: "green"})]),
					]),
				]),
				text([raw("gpt-5.4")], {dimColor: true}),
			]),
		]),
		100,
		false,
		false,
		geminiCliMeta("packages/cli/src/ui/components/StatusRow.tsx"),
	),
	namedCase(
		"gemini-cli/status-row/interactive-shell-wait",
		box({width: "100%", marginLeft: 1}, [
			text([raw("Waiting for interactive shell to finish")], {color: "yellow"}),
		]),
		100,
		false,
		false,
		geminiCliMeta("packages/cli/src/ui/components/StatusRow.tsx"),
	),
	namedCase(
		"gemini-cli/background-task/output-panel-shell",
		box({flexDirection: "column", height: 8, width: 70, borderStyle: "single", borderColor: "cyan"}, [
			box({flexDirection: "row", justifyContent: "space-between", borderStyle: "single", borderTop: false, borderLeft: false, borderRight: false, paddingX: 1}, [
				textValue("1 bash"),
				text([raw("PID 4242 Focused")], {bold: true}),
				text([raw("q close")], {dimColor: true}),
			]),
			box({flexGrow: 1, overflow: "hidden", paddingX: 1, flexDirection: "column"}, [
				textValue("$ npm test"),
				textValue("PASS tests/project_upstream"),
			]),
			box({paddingX: 1}, [
				text([raw("Log: /tmp/gemini-cli/task.log")], {dimColor: true, wrap: "truncate-end"}),
			]),
		]),
		100,
		false,
		false,
		geminiCliMeta("packages/cli/src/ui/components/BackgroundTaskDisplay.tsx"),
	),
	namedCase(
		"gemini-cli/background-task/process-selection-shell",
		box({flexDirection: "column", height: 8, width: 70, borderStyle: "single"}, [
			box({flexDirection: "row", justifyContent: "space-between", borderStyle: "single", borderTop: false, borderLeft: false, borderRight: false, paddingX: 1}, [
				textValue("Background Shells"),
				text([raw("enter view, x kill, esc cancel")], {dimColor: true}),
			]),
			box({flexGrow: 1, overflow: "hidden", paddingX: 1, flexDirection: "column"}, [
				box({flexShrink: 0, marginBottom: 1, paddingTop: 1}, [
					text([raw("Select a process")], {bold: true}),
				]),
				textValue("● 4242 npm test"),
				textValue("○ 7777 git status"),
			]),
		]),
		100,
		false,
		false,
		geminiCliMeta("packages/cli/src/ui/components/BackgroundTaskDisplay.tsx"),
	),
	namedCase(
		"gemini-cli/new-agents/review-card",
		box({flexDirection: "column", width: "100%"}, [
			box({flexDirection: "column", borderStyle: "round", borderColor: "yellow", padding: 1, marginX: 1}, [
				box({flexDirection: "column", marginBottom: 1}, [
					text([raw("New agents discovered")], {bold: true}),
					textValue("Review these agents before enabling them."),
				]),
				box({flexDirection: "column", marginTop: 1, borderStyle: "single", padding: 1}, [
					box({flexDirection: "row"}, [
						box({flexShrink: 0}, [text([raw("- code-reviewer:")], {bold: true})]),
						text([raw(" Reviews source changes")], {dimColor: true}),
					]),
					box({marginLeft: 2}, [
						text([raw("MCP: github")], {dimColor: true}),
					]),
				]),
				textValue("○ Enable all"),
				textValue("● Review individually"),
			]),
		]),
		100,
		false,
		false,
		geminiCliMeta("packages/cli/src/ui/components/NewAgentsNotification.tsx"),
	),
	namedCase(
		"gemini-cli/empty-wallet/quota-dialog",
		box({flexDirection: "column", borderStyle: "round", padding: 1}, [
			box({flexDirection: "column", marginBottom: 1}, [
				text([raw("Daily quota exhausted")], {color: "yellow", bold: true}),
				textValue("Use /stats to inspect usage."),
				textValue("Use /model to switch models."),
				textValue("Use /auth to change account."),
			]),
			box({marginBottom: 1}, [textValue("Purchase credits to continue immediately.")]),
			box({marginBottom: 1}, [text([raw("Credit updates may be delayed.")], {dimColor: true})]),
			textValue("Proceed?"),
			box({marginTop: 1, marginBottom: 1, flexDirection: "column"}, [
				textValue("● Continue with current model"),
				textValue("○ Switch model"),
			]),
		]),
		100,
		false,
		false,
		geminiCliMeta("packages/cli/src/ui/components/EmptyWalletDialog.tsx"),
	),
	namedCase(
		"gemini-cli/validation/default-action-dialog",
		box({flexDirection: "column", borderStyle: "round", padding: 1}, [
			box({marginBottom: 1}, [
				textValue("Further action is required to validate this account."),
			]),
			box({marginTop: 1, marginBottom: 1, flexDirection: "column"}, [
				textValue("● Open verification link"),
				textValue("○ Paste verification URL"),
			]),
			box({marginTop: 1}, [
				text([raw("Learn more: ")], {dimColor: true}),
				text([raw("https://example.com/validation")], {color: "cyan"}),
			]),
		]),
		100,
		false,
		false,
		geminiCliMeta("packages/cli/src/ui/components/ValidationDialog.tsx"),
	),
	namedCase(
		"gemini-cli/validation/waiting-panel",
		box({flexDirection: "column", borderStyle: "round", padding: 1}, [
			box({}, [
				textValue("⠋"),
				textValue(" Waiting for verification..."),
			]),
			box({marginTop: 1}, [textValue("https://example.com/device")]),
			box({marginTop: 1}, [
				text([raw("Press Enter when verification is complete.")], {dimColor: true}),
			]),
		]),
		100,
		false,
		false,
		geminiCliMeta("packages/cli/src/ui/components/ValidationDialog.tsx"),
	),
	namedCase(
		"gemini-cli/ask-user/review-unanswered",
		box({"aria-label": "Review", flexDirection: "column"}, [
			box({flexDirection: "column"}, [
				box({marginBottom: 1}, [text([raw("Review your answers:")], {bold: true})]),
				box({marginBottom: 1}, [text([raw("1 unanswered question")], {color: "yellow"})]),
				box({flexDirection: "column"}, [
					box({}, [
						text([raw("1. Path")], {dimColor: true}),
						text([raw(" → ")], {dimColor: true}),
						text([raw("src/app.tsx")]),
					]),
					box({}, [
						text([raw("2. Reason")], {dimColor: true}),
						text([raw(" → ")], {dimColor: true}),
						text([raw("(not answered)")], {color: "yellow"}),
					]),
				]),
				text([raw("Enter submit, tab edit next")], {dimColor: true}),
			]),
		]),
		100,
		false,
		false,
		geminiCliMeta("packages/cli/src/ui/components/AskUserDialog.tsx"),
	),
	namedCase(
		"gemini-cli/searchable-list/shell",
		box({flexDirection: "column", width: "100%", height: "100%", paddingX: 1}, [
			box({marginBottom: 1}, [text([raw("Choose a command")], {bold: true})]),
			box({borderStyle: "round", borderColor: "white", paddingX: 1, marginBottom: 1}, [
				textValue("> git"),
			]),
			box({marginBottom: 1}, [text([raw("Recent commands")], {dimColor: true})]),
			box({flexDirection: "column", flexGrow: 1}, [
				textValue("▲"),
				box({marginBottom: 1, marginX: 1}, [textValue("git status")]),
				box({marginBottom: 1, marginX: 1}, [textValue("git diff")]),
				textValue("▼"),
			]),
			text([raw("tab to change focus")], {dimColor: true}),
		]),
		100,
		false,
		false,
		geminiCliMeta("packages/cli/src/ui/components/shared/SearchableList.tsx"),
	),
	namedCase(
		"gemini-cli/messages/gemini-message",
		box({flexDirection: "row"}, [
			box({width: 2}, [text([raw("✦ ")], {color: "cyan", "aria-label": "Gemini"})]),
			box({flexGrow: 1, flexDirection: "column"}, [
				textValue("The workspace contains a Go port of Ink."),
				textValue("Project-based goldens compare against upstream Ink."),
			]),
		]),
		80,
		false,
		false,
		geminiCliMeta("packages/cli/src/ui/components/messages/GeminiMessage.tsx"),
	),
	namedCase(
		"gemini-cli/messages/user-message-background",
		box({backgroundColor: "#222222", paddingX: 1}, [
			box({flexDirection: "row", alignSelf: "flex-start", width: 70}, [
				box({width: 2, flexShrink: 0}, [
					text([raw("> ")], {color: "cyan", "aria-label": "User"}),
				]),
				box({flexGrow: 1}, [
					text([raw("Run parity tests for project goldens")], {wrap: "wrap"}),
				]),
			]),
		]),
		80,
		false,
		false,
		geminiCliMeta("packages/cli/src/ui/components/messages/UserMessage.tsx"),
	),
	namedCase(
		"gemini-cli/checklist/expanded",
		box({borderStyle: "single", borderBottom: false, borderLeft: false, borderRight: false, paddingX: 1}, [
			box({flexDirection: "column", rowGap: 1}, [
				text([raw("Todo 2/3")], {bold: true}),
				box({flexDirection: "row", columnGap: 1}, [
					text([raw("✓")], {color: "green"}),
					textValue("Capture upstream output"),
				]),
				box({flexDirection: "row", columnGap: 1}, [
					text([raw("»")], {color: "cyan"}),
					text([raw("Compare Go renderer")], {color: "cyan"}),
				]),
				box({flexDirection: "row", columnGap: 1}, [
					text([raw(" ")], {color: "gray"}),
					textValue("Document result"),
				]),
			]),
		]),
		100,
		false,
		false,
		geminiCliMeta("packages/cli/src/ui/components/Checklist.tsx"),
	),
	namedCase(
		"gemini-cli/checklist/collapsed",
		box({borderStyle: "single", borderBottom: false, borderLeft: false, borderRight: false, paddingX: 1}, [
			box({flexDirection: "row", columnGap: 1, height: 1}, [
				box({flexShrink: 0, flexGrow: 0}, [
					text([raw("Todo 2/3")], {bold: true}),
				]),
				box({flexShrink: 1, flexGrow: 1}, [
					text([raw("» Compare Go renderer")], {color: "cyan", wrap: "truncate-end"}),
				]),
			]),
		]),
		100,
		false,
		false,
		geminiCliMeta("packages/cli/src/ui/components/Checklist.tsx"),
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
