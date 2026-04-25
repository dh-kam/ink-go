# Neovate Code Ink Usage Patterns

Source repo: `https://github.com/neovateai/neovate-code`
Inspected clone: `.tmp/project-research/neovate-code`
Inspected branch/commit: `master` @ `0a24b363ecbe24eb87a0190a88fcb78c80593b4b`

Notes:
- This is a hand-curated set of 50 real Ink usage patterns pulled from actual source, not a bulk JSX dump.
- Line refs point at the upstream clone under `.tmp/project-research/neovate-code`.
- Tree sketch format is intentionally compact: `Component[prop] > child + child`, with `?` for conditional branches.

## 01. App Shell Stack
File: `src/ui/App.tsx:76-89`
Pattern: The root TUI keeps chat history, transient status strips, input, and footer/debug widgets in one vertical shell keyed by rerender, fork, and transcript state.
Ink: `TerminalSizeProvider`, `Box`, `Messages`, `BackgroundPrompt`, `ActivityIndicator`, `QueueDisplay`, `ChatInput`, `TranscriptModeIndicator`, `ExitHint`, `Debug`
Tree: `TerminalSizeProvider > Box[col key=force+fork+transcript] > Messages + BackgroundPrompt + ActivityIndicator + QueueDisplay + (transcriptMode ? TranscriptModeIndicator : ChatInput) + ExitHint + Debug`

## 02. App Overlay Stack
File: `src/ui/App.tsx:90-106`, `src/ui/App.tsx:19-21`
Pattern: Modal and slash-command JSX overlays are mounted as siblings outside the main conversation column so they can interrupt normal flow without entering the chat stack.
Ink: `ForkModal`, `ApprovalModal`, `Box`
Tree: `TerminalSizeProvider > mainShell ; sibling ?ForkModal + ApprovalModal + Box > slashCommandJSX`

## 03. Static + Pending Message Split
File: `src/ui/Messages.tsx:198-241`
Pattern: Completed messages render in Ink `Static` with a synthetic header item, while pending tool-related messages stay dynamic below.
Ink: `Box`, `Static`, `MessageGroup`, `useMemo`
Tree: `Box[col] > Static[key=session+fork items=['header', ...completed]] > Header|MessageGroup + pendingMessages.map(MessageGroup)`

## 04. Intro Header Composite
File: `src/ui/Messages.tsx:378-385`
Pattern: The first-screen header composes the ASCII art brand, product line, starter tips, and model warning into one padded intro block.
Ink: `Box`, `ProductASCIIArt`, `ProductInfo`, `GettingStartedTips`, `ModelConfigurationWarning`
Tree: `Box[col py=1] > ProductASCIIArt + ProductInfo + GettingStartedTips + ModelConfigurationWarning`

## 05. Model Configuration Warning Card
File: `src/ui/Messages.tsx:292-375`
Pattern: When no model is configured, the app shows a rounded yellow warning card listing providers, env/login status, and suggested slash commands.
Ink: `Box`, `Text`
Tree: `Box[col] > Box[col mt=1 border=round borderColor=yellow p=1] > Text[bold yellow title] + Box[col mt=1] > Text[intro] + Box[col mt=1] > providerRows + Box[col mt=1] > suggestedActions ; ?Box[mt=1] > Text[red initError]`

## 06. User Message Bubble
File: `src/ui/Messages.tsx:389-419`
Pattern: Regular user prompts render as a username label followed by a dark background bubble offset from the left edge.
Ink: `Box`, `Text`
Tree: `Box[col mt ml=user] > Text[bold userNameColor] + Box > Text[bg=#555 fg=#cdcdcd](message)`

## 07. User Cancel Notice
File: `src/ui/Messages.tsx:396-418`
Pattern: The same user-message component swaps the bubble for a single canceled-request line when the message is tagged as canceled.
Ink: `Box`, `Text`
Tree: `Box[col mt ml=user] > Text[bold userNameColor] + Text[color=canceled]("User canceled the request")`

## 08. Bash Command Message Row
File: `src/ui/Messages.tsx:42-62`
Pattern: User-entered bash commands are rendered as a minimal one-line row with a bold command marker and emphasized command text.
Ink: `Box`, `Text`, `useMemo`
Tree: `Box[col mt ml=user] > Box[row] > Text[bold chatBorderBash]("! ") + Text[bold toolColor](command)`

## 09. Bash Output Expandable Row
File: `src/ui/Messages.tsx:65-85`
Pattern: Bash stdout/stderr lines reuse the shared expandable-output component and inherit the user-side left offset.
Ink: `Box`, `ExpandableOutput`, `useMemo`
Tree: `Box[col ml=user] > ExpandableOutput[content=cleanedOutput isError?]`

## 10. Assistant Markdown Response Block
File: `src/ui/Messages.tsx:423-436`
Pattern: Plain assistant text responses render as a branded lowercase product label followed by markdown-rendered content.
Ink: `Box`, `Text`, `Markdown`
Tree: `Box[col mt] > Text[bold color=#FF3070](productNameLower) + Markdown(text)`

## 11. Tool Use Row With Optional Description
File: `src/ui/Messages.tsx:440-467`
Pattern: Tool invocations render inline as a bold tool label followed by a wrapped, optionally truncated description in parentheses.
Ink: `Box`, `Text`, `useMemo`
Tree: `Box[row wrap mt] > Text[bold tool](displayName|name) + ?Text[toolDescription]("(description)")`

## 12. Generic Tool Pair Block
File: `src/ui/Messages.tsx:488-499`
Pattern: Non-special tools render as a simple vertical pairing of the tool-use header and, if present, its result block underneath with extra top margin.
Ink: `Box`, `ToolUse`, `ToolResultItem`
Tree: `Box[col] > ToolUse + ?Box[mt=toolResult] > ToolResultItem`

## 13. Thinking Block With Truncation Hint
File: `src/ui/Messages.tsx:594-628`
Pattern: Reasoning text renders as a gray "thinking" header plus italic content, capped to two lines outside transcript mode with a hidden-lines hint.
Ink: `Box`, `Text`, `useMemo`
Tree: `Box[col mt] > Text[bold gray]("thinking") + Text[gray italic](displayText) + ?Text[gray dim]("... N more lines hidden ...")`

## 14. Tool Result Diff Viewer Branch
File: `src/ui/Messages.tsx:631-667`
Pattern: Structured tool results of type `diff_viewer` route into the dedicated diff renderer with transcript-sensitive height limits.
Ink: `DiffViewer`, `useAppStore`
Tree: `DiffViewer[originalContent newContent fileName maxHeight=transcript?Infinity:10 startLineNumber?]`

## 15. Tool Result Todo Write Branch
File: `src/ui/Messages.tsx:642-676`
Pattern: Structured todo-write results render the todo diff as a compact checklist rather than raw text.
Ink: `TodoList`
Tree: `TodoList[oldTodos newTodos verbose=false]`

## 16. Tool Result Bash Output Branch
File: `src/ui/Messages.tsx:682-687`
Pattern: Successful bash tool results reuse the expandable-output component instead of inline text.
Ink: `ExpandableOutput`
Tree: `ExpandableOutput[content=text]`

## 17. Tool Result Generic Text Branch
File: `src/ui/Messages.tsx:690-694`
Pattern: Generic successful tool results render as a single colored line prefixed by the downward result marker.
Ink: `Text`
Tree: `Text[color=toolResult]("arrowDown text")`

## 18. Chat Composer Chrome
File: `src/ui/ChatInput.tsx:174-253`
Pattern: The main composer uses top and bottom border rules, a mode-colored prompt glyph, multiline text input, optional reverse-search row, and a status line underneath.
Ink: `Box`, `Text`, `TextInput`, `ModeIndicator`, `StatusLine`, `DebugRandomNumber`, `useMemo`, `useCallback`
Tree: `Box[col mt=chatInput] > ModeIndicator + Box[col] > Text[color=border]('---') + Box[row gap=1] > Text[color=active?](promptSymbol) + TextInput[multiline showCursor? placeholder cursorOffset columns=terminal] + DebugRandomNumber + ?Box[row ml=2] > ReverseSearchInput + Text[color=border]('---') + StatusLine`

## 19. Reverse Search Row Under Composer
File: `src/ui/ChatInput.tsx:241-249`
Pattern: When reverse history search is active, the composer adds a dedicated inline subrow indented under the main prompt.
Ink: `Box`, `ReverseSearchInput`
Tree: `Box[row ml=2] > ReverseSearchInput[history onExit onCancel onMatchChange]`

## 20. Slash Command Suggestions Dropdown
File: `src/ui/ChatInput.tsx:254-277`
Pattern: Slash-command suggestions render in a scrolling suggestion window, with each row normalized into a two-column suggestion item.
Ink: `Suggestion`, `SuggestionItem`
Tree: `Suggestion[selectedIndex maxVisible=10] > SuggestionItem[name='/' + command description firstColumnWidth maxWidth]`

## 21. File Suggestions Dropdown
File: `src/ui/ChatInput.tsx:278-299`
Pattern: File path suggestions reuse the same suggestion list but compute column width from the longest visible file label.
Ink: `Suggestion`, `SuggestionItem`
Tree: `Suggestion[selectedIndex maxVisible=10] > SuggestionItem[name=fileDisplayText description firstColumnWidth=min(maxName+4, columns-10) maxWidth=columns]`

## 22. Mode Indicator Badge
File: `src/ui/ModeIndicator.tsx:11-46`
Pattern: A right-aligned top-row badge flips between plan, brainstorm, bash, or empty states using inline text fragments rather than separate panels.
Ink: `Box`, `Text`
Tree: `Box[row gap=1 mt] > Box[grow=1] + Box > (planFragment | brainstormFragment | bashFragment | Text[' '])`

## 23. Status Telemetry Strip
File: `src/ui/StatusLine.tsx:42-147`, `src/ui/StatusLine.tsx:181-186`
Pattern: The status line compresses model, thinking effort, current folder, token usage, context-left percentage, approval mode, and session id into one gray telemetry row.
Ink: `Box`, `Text`, `useMemo`
Tree: `Box[col px=2] > Box > Text[gray]("[model or warning][thinking] | folder | tokenUsed | contextPct | approval | sessionId") + StatusSide`

## 24. Processing Activity Indicator
File: `src/ui/ActivityIndicator.tsx:21-36`, `src/ui/ActivityIndicator.tsx:72-110`
Pattern: While processing, the app shows animated gradient status text followed by a gray status capsule containing cancel, token, and retry information.
Ink: `Box`, `Text`, `GradientText`, `useMemo`, `useEffect`, `useState`
Tree: `Box[row mt=activity] > Box[row] > GradientText[text='Processing...' highlightIndex] + Box[ml=1] > Text[activityText](statusText)`

## 25. Failed Activity Indicator
File: `src/ui/ActivityIndicator.tsx:21-31`, `src/ui/ActivityIndicator.tsx:93-110`
Pattern: Failed or exit-with-error states collapse the animated variant into a single red error line.
Ink: `Box`, `Text`, `useMemo`
Tree: `Box[row mt=activity] > Text[color=red]("Failed: error" | "Unknown status")`

## 26. Background Prompt Output Summary
File: `src/ui/BackgroundPrompt.tsx:32-45`
Pattern: Live background bash output is truncated into a compact result preview with an optional more-output line and a `ctrl+b` hint.
Ink: `Box`, `Text`
Tree: `Box[col] > Box[col px=1] > Text[color=toolResult]("resultPreview") + ?Text[gray dim]("... more output") + Text[cyan dim]("ctrl+b to run in background")`

## 27. Queued Messages Warning Card
File: `src/ui/QueueDisplay.tsx:11-38`
Pattern: Queued prompts render in a rounded warning box with a count header, numbered preview rows, and an execution notice.
Ink: `Box`, `Text`
Tree: `Box[col border=round borderColor=warning px=1] > Box[row justify=space-between] > Text[bold warning](count) + Box[col mt=1] > queuedRows + Box[mt=1] > Text[yellow](notice)`

## 28. Exit Summary Panel
File: `src/ui/ExitHint.tsx:25-48`
Pattern: On exit, the app replaces interaction with a gray summary panel listing cwd, model, total tokens, session id, approval mode, and log file.
Ink: `Box`, `Text`, `useMemo`
Tree: `Box[col mt=1] > Text[gray dim]('---') + Text[gray bold dim]('Session ended') + Box[col mt=1] > detailLines`

## 29. Transcript Mode Banner
File: `src/ui/TranscriptModeIndicator.tsx:7-24`
Pattern: Transcript mode swaps out the composer for a lightweight read-only banner with a top-only border and toggle hint.
Ink: `Box`, `Text`
Tree: `Box[col mt=1] > Box[border=single topOnly gray] + Box > Text[dim]('Showing detailed transcript ...')`

## 30. Standard SelectInput Option List
File: `src/ui/SelectInput.tsx:131-279`
Pattern: Text-only select options render as numbered rows with a cursor prefix, optional multi-select checkbox, answer coloring, and secondary descriptions.
Ink: `Box`, `Text`, `useInput`
Tree: `Box[col] > options.map(Box[col] > Box[row] > Text[prefix] + Text[number] + ?Text[checkbox] + Text[label] + ?Text[tick] ; ?Box[ml] > Text[dim](description))`

## 31. Inline Custom-Input Select Row
File: `src/ui/SelectInput.tsx:153-215`
Pattern: When the focused option is of type `input`, the choice row inlines a `TextInput` in place of the label while preserving numbering and checkbox chrome.
Ink: `Box`, `Text`, `TextInput`
Tree: `Box[row] > Text[prefix] + Text[number] + ?Text[checkbox] + TextInput[value placeholder columns=terminalPrefix onSubmit]`

## 32. Searchable Grouped Picker
File: `src/ui/PaginatedGroupSelectInput.tsx:593-663`
Pattern: The grouped selector composes searchable headers, paged result items, selected-item context, pagination metadata, and help text into one column.
Ink: `Box`, `Text`, `useInput`, custom search/pagination hooks
Tree: `Box[col] > SearchInputDisplay + Box[col] > currentPageItems.map(ListItemRenderer) + SelectedItemInfo + PaginationInfo + HelpText`

## 33. Question Navigation Strip
File: `src/ui/AskQuestionModal.tsx:161-230`
Pattern: Multi-question approval flows use a flat nav strip with left/right arrows, answered-state icons, and a submit tab that can become the active segment.
Ink: `Box`, `Text`
Tree: `Box[row mb=1 wrap] > ?Text[leftArrow] + questions.map(Text[activeBg?|dim?]('icon header')) + submitTab + ?Text[rightArrow]`

## 34. Question Page
File: `src/ui/AskQuestionModal.tsx:270-408`
Pattern: Each question page shows a full-width divider, the nav strip, a bold question title, a dynamic `SelectInput`, and one-line interaction hints.
Ink: `Box`, `Text`, `SelectInput`, `useMemo`, `useCallback`
Tree: `Box[col] > QuestionBorder + QuestionNav + Box[mb=1] > Text[bold askPrimary](question) + SelectInput[mode=single|multi defaultValue submitButtonText] + Box[mt=1] > Text[dim hint]`

## 35. Submit Review Page
File: `src/ui/AskQuestionModal.tsx:421-488`
Pattern: The final review step summarizes answered questions as a compact list, warns on incomplete input, and uses another `SelectInput` for submit vs cancel.
Ink: `Box`, `Text`, `SelectInput`
Tree: `Box[col] > QuestionBorder + QuestionNav + title + ?warning + ?Box[col mb=1] > answeredQARows + Box[mb=1] > Text[dim] + SelectInput[submit|cancel]`

## 36. Diff Approval Flow
File: `src/ui/ApprovalModal.tsx:37-68`, `src/ui/ApprovalModal.tsx:89-110`, `src/ui/ApprovalModal.tsx:341-364`
Pattern: Edit/write approvals show a title naming the target file, a dashed-divider diff preview, a short question, and a three-option select flow.
Ink: `Box`, `Text`, `DashedDivider`, `DiffViewer`, `SelectInput`, `useMemo`, `useCallback`
Tree: `Box[col] > TopDivider + titleRow(file) + Box[col mb=1] > DashedDivider + DiffViewer[maxHeight=transcript?Infinity:500] + DashedDivider + Box[mb=1] > Text(question) + SelectInput[single approve/always/denyInput] + Box[mt=1] > Text[dim]('Esc to exit')`

## 37. Bash Approval Flow
File: `src/ui/ApprovalModal.tsx:70-77`, `src/ui/ApprovalModal.tsx:113-125`, `src/ui/ApprovalModal.tsx:341-364`
Pattern: Bash approvals swap the diff preview for an indented command line plus optional description, then reuse the same approval question and selector.
Ink: `Box`, `Text`, `SelectInput`, `useMemo`, `useCallback`
Tree: `Box[col] > TopDivider + titleRow('Bash command') + Box[col mb=1] > Box[ml=2] > Text(command) + ?Box[ml=2] > Text[askSecondary](description) + question + SelectInput + footerHint`

## 38. Invalid Questions Modal
File: `src/ui/ApprovalModal.tsx:137-160`
Pattern: Broken ask-user-question payloads produce a standalone red rounded modal with dismiss instructions.
Ink: `Box`, `Text`, `useInput`
Tree: `Box[col p=1 border=round borderColor=red] > Text[red bold](title) + Text(body) + Box[mt=1] > Text[dim](dismissHint)`

## 39. Memory Rule Destination Picker
File: `src/ui/MemoryModal.tsx:15-71`
Pattern: Saving a memory rule opens a rounded warning box showing the rule text and a numbered destination picker backed by `ink-select-input`.
Ink: `Box`, `Text`, `SelectInput` from `ink-select-input`, `useInput`, `useMemo`
Tree: `Box[col p=1 border=round borderColor=warning] > Text[warning bold](title) + Box[my=1] > Text(rule) + Box[my=1] > Text[bold]('Select destination:') + SelectInput[items=project|global]`

## 40. Plan Approval View
File: `src/ui/PlanMode/PlanApprovalView.tsx:166-208`
Pattern: Exit-plan-mode approval shows a cyan divider, markdown plan preview, approval options, and an optional `ctrl+g` editor hint line.
Ink: `Box`, `Text`, `Markdown`, `SelectInput`, `useInput`, `useMemo`, `useCallback`
Tree: `Box[col] > Divider[cyan] + Box[col mb=1] > Text(intro) + Box[col mb=1 px=2] > Markdown(planPreview) + Divider[cyan] + Box[mb=1] > Text(question) + SelectInput[single text|input] + ?Box[mt=1] > Text[dim](editorHint)`

## 41. Plan Viewer Preview
File: `src/ui/PlanMode/PlanViewer.tsx:34-55`
Pattern: Approved plan results render file path metadata, markdown content, and a transcript-sensitive hidden-lines hint.
Ink: `Box`, `Text`, `Markdown`, `useMemo`
Tree: `Box[col] > Box[px=1] > Text[gray](filePath) + Box[px=1 mt=1] > Markdown(visibleContent) + ?Box[px=1 mt=1] > Text[gray dim](hiddenLinesHint)`

## 42. Exit Plan Mode Result
File: `src/ui/PlanMode/ExitPlanModeDisplay.tsx:12-37`, `src/ui/PlanMode/ExitPlanModeDisplay.tsx:52-67`
Pattern: Exit-plan-mode tool pairs render a tool-style header and, when the result carries approved plan payload, embed `PlanViewer` under it.
Ink: `Box`, `Text`, `PlanViewer`
Tree: `Box[col] > Box[mt] > Text[bold tool](displayName) + ?Box[mt=toolResult] > (PlanViewer | Text[result] | Text[error])`

## 43. Rewind Message List
File: `src/ui/ForkModal.tsx:279-367`
Pattern: The fork/rewind modal shows a full-width divider, instructions, a selectable list of prior user turns, per-turn code-change summaries, and a `(current)` sentinel row.
Ink: `Box`, `Text`
Tree: `Box[col] > Box > Text[askPrimary bold](divider) + header + intro + Box[col] > (emptyNotice | userMessageRows + currentRow) + Box[mt=1] > Text[dim](enter/esc hint)`

## 44. Rewind Restore Preview
File: `src/ui/ForkModal.tsx:382-539`
Pattern: After selecting a rewind point, the modal previews conversation/code impact, lists restore-mode options, and warns about manual/bash edits.
Ink: `Box`, `Text`, `SelectInput`, `useMemo`, `useState`
Tree: `Box[col] > divider + title + intro + Box[col mt=1 mb=1] > quotedMessage + ?relativeTime + Box > Text[dim](conversationImpact) + Box[mb=1] > fileChangeSummary + SelectInput[restoreMode] + ?Box[mt=1] > Text[dim](manualEditWarning) + Box[mt=1] > Text[dim](enter/esc hint)`

## 45. Sorted Todo Checklist
File: `src/ui/Todo.tsx:51-82`, `src/ui/Todo.tsx:118-133`
Pattern: Todo writes render as a sorted checklist where status drives checkbox glyph, color, bolding, and optional priority suffix.
Ink: `Box`, `Text`, `useMemo`
Tree: `Box[col] > newTodos.sort(compareTodos).map(TodoItem[row] > Box[minWidth=2] > Text[color bold?](checkbox) + Box > Text[color strike? bold?](content) + ?Text[dim](priority))`

## 46. Agent Progress Stream
File: `src/ui/AgentProgress/AgentProgressOverlay.tsx:124-204`
Pattern: Running subagents render a tool-style header, optional hidden-count notice, prompt preview, recent nested log items, and a stats footer.
Ink: `Box`, `Text`, `LogItemRenderer`, `useMemo`
Tree: `Box[col] > AgentToolUse[status=running model?] + Box[col] > ?Box[pl=1] > Text[dim](hiddenCount) + ?promptBlock + visibleItems.map(LogItemRenderer) + Box[pl=1 mt=0] > Text[gray dim](toolUses/tokens hint)`

## 47. Agent Completed Transcript Expansion
File: `src/ui/AgentProgress/AgentProgressOverlay.tsx:298-385`
Pattern: Finished subagents show header + stats, and in transcript mode expand into explicit Prompt and Response sections with markdown output.
Ink: `Box`, `Text`, `Markdown`, `useMemo`
Tree: `Box[col] > AgentToolUse[status=completed|failed model?] + statsLine + (transcriptMode ? Box[col ml=2 mt=1] > promptSection + responseSection(Markdown) : Box[ml=2] > Text[gray dim](ctrl+o hint))`

## 48. Workspace List Panel
File: `src/commands/workspace/components.tsx:55-112`
Pattern: The workspace command lists active worktrees as stacked detail cards with cleanliness status, branch metadata, path, and an optional created-at line.
Ink: `Box`, `Text`, `useApp`
Tree: `Box[col py=1] > Box > Text[bold]('Active Workspaces:') + Box[col mt=1] > worktreeCards[col mb=1] > nameRow + Box[col pl=2] > branch/original/path/+?createdAt + Box[mt=1] > Text[dim](completionHint)`

## 49. Commit Review + Action Menu
File: `src/commands/commit.tsx:1011-1023`, `src/ui/CommitResultCard.tsx:11-30`, `src/ui/CommitActionSelector.tsx:149-177`
Pattern: The interactive commit flow first shows a compact commit/branch summary card, then an action menu for copy, commit, push, branch, PR, edit, or cancel.
Ink: `Box`, `Text`, `CommitResultCard`, `CommitActionSelector`
Tree: `Box[col p=1] > Box[col] > CommitResultCard[commitMessage branchName breaking? summary] + Box[mt=1] > CommitActionSelector[actions disabled? default=push]`

## 50. Interactive Skill Preview Selector
File: `src/commands/skill.tsx:464-496`
Pattern: The interactive skill installer previews discovered skills as a cursorable multi-select list with radio-style markers and a selected-count footer.
Ink: `Box`, `Text`, `useInput`
Tree: `Box[col] > Text[bold](title) + Text[dim](controlsHint) + Box[col mt=1] > skills.map(Box[row] > Text[cursor?] + Text[selected?] + Text(name) + Text[dim](' - description')) + Box[mt=1] > Text[dim](selectedCount)`
