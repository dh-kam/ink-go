# Gemini CLI Ink Patterns

Inspected upstream repo: `https://github.com/google-gemini/gemini-cli`
Inspected commit: `8573650253888a252500856e240385ff8a2553c8`
Path base below: upstream repo root under `.tmp/project-research/gemini-cli`

Sketch conventions:
- `>` means parent to child.
- `+` means sibling.
- `()` means conditional branch.
- Prop shorthands: `col`, `row`, `w`, `h`, `mt`, `mb`, `ml`, `mr`, `pad`, `padX`, `padY`, `border`.

## App Shell

01. Source: `packages/cli/src/ui/App.tsx:21-27`
Pattern: On quit in alternate-buffer mode, the app keeps streaming context alive and renders the alternate-buffer-specific quit frame instead of the normal app shell.
Ink: `useIsScreenReaderEnabled`
Tree: `StreamingContext.Provider[value=streamingState] > AlternateBufferQuittingDisplay`

02. Source: `packages/cli/src/ui/App.tsx:21-30`
Pattern: On quit outside alternate-buffer mode, the app bypasses the normal layout entirely and renders only the terminal-width-aware quit transcript.
Ink: `useIsScreenReaderEnabled`
Tree: `QuittingDisplay`

03. Source: `packages/cli/src/ui/App.tsx:33-36`
Pattern: Normal runtime chooses between the accessibility-first screen-reader layout and the default layout, both wrapped in a streaming-state provider.
Ink: `useIsScreenReaderEnabled`
Tree: `StreamingContext.Provider[value=streamingState] > (ScreenReaderAppLayout | DefaultAppLayout)`

04. Source: `packages/cli/src/ui/layouts/DefaultAppLayout.tsx:31-85`
Pattern: Default shell is a fixed-width column with scrollable main history on top and a non-growing controls column at the bottom.
Ink: `Box`
Tree: `Box[col w=terminalWidth h?=terminalHeight padBottom?=1 ref=root] > MainContent + Box[col ref=mainControls w=terminalWidth h?=stableControls] > Notifications + CopyModeWarning + (customDialog | DialogManager | Composer[focused]) + ExitWarning`

05. Source: `packages/cli/src/ui/layouts/DefaultAppLayout.tsx:72-79`
Pattern: When dialogs are active, the default layout swaps the composer out for a centralized dialog manager inside the bottom controls column.
Ink: `Box`
Tree: `Box[col ref=mainControls] > Notifications + CopyModeWarning + DialogManager[terminalWidth addItem] + ExitWarning`

06. Source: `packages/cli/src/ui/layouts/DefaultAppLayout.tsx:79-81`
Pattern: In the normal interactive state, the bottom controls column resolves to the focused composer instead of a dialog.
Ink: `Box`
Tree: `Box[col ref=mainControls] > Notifications + CopyModeWarning + Composer[isFocused=true] + ExitWarning`

07. Source: `packages/cli/src/ui/layouts/ScreenReaderAppLayout.tsx:23-45`
Pattern: The screen-reader layout front-loads notifications and footer metadata, keeps the transcript in a flex-growing hidden-overflow box, then places the active dialog or composer after it.
Ink: `Box`
Tree: `Box[col w=90% h=100% ref=root] > Notifications + Footer + Box[flexGrow=1 overflow=hidden] > MainContent + (DialogManager | Composer) + ExitWarning`

08. Source: `packages/cli/src/ui/components/MainContent.tsx:256-305`
Pattern: Alternate-buffer and terminal-buffer history is virtualized through a `ScrollableList` that treats the header, history items, and pending block as one scrollable dataset.
Ink: `Box`
Tree: `ScrollableList[ref hasFocus width data=[header,*history,pending] initialScroll=end renderStatic? overflowToBackbuffer? scrollbar=mouseMode] > (AppHeader | HistoryItemDisplay | pendingGroup)`

09. Source: `packages/cli/src/ui/components/MainContent.tsx:308-321`
Pattern: Standard terminal mode renders historical items through Ink `Static` so old output stays anchored in scrollback, then appends pending items after the static block.
Ink: `Static`
Tree: `<> > Static[items=[AppHeader,*staticHistory,*lastResponse]] + pendingItems[Box[col] > *HistoryItemDisplay + ToolConfirmationQueue?]`

10. Source: `packages/cli/src/ui/components/AppHeader.tsx:97-107,143-157`
Pattern: Wide terminals use a row header with the gradient Gemini icon on the left and metadata on the right.
Ink: `Box`, `Text`
Tree: `Box[col] > Box[row mt=1 mb=1 padLeft=1] > Box[row] > Box[shrink=0] > ThemedGradient[ICON] + Box[ml=2 col] > Box[row] > Text[bold]{Gemini CLI} + Text{version} + updatingBox? + blankLine? + UserIdentity?`

11. Source: `packages/cli/src/ui/components/AppHeader.tsx:83-105,141-157`
Pattern: Narrow terminals or logged-out wide terminals switch the header to a column layout and may show the extended ASCII wordmark under the icon.
Ink: `Box`, `Text`
Tree: `Box[col] > Box[col mt=1 mb=1 padLeft=1] > Box[row] > ThemedGradient[ICON] + Box[ml=3] > Text{longAsciiLogo}? + Box[mt=1] > metadataColumn`

12. Source: `packages/cli/src/ui/components/AppHeader.tsx:110-125`
Pattern: The first metadata row embeds a spinner inline beside the version string when self-update is in progress.
Ink: `Box`, `Text`
Tree: `Box[row] > Text[bold]{Gemini CLI} + Text{vX} + Box[ml=2] > Text[secondary] > CliSpinner + " Updating"`

13. Source: `packages/cli/src/ui/components/UserIdentity.tsx:46-61`
Pattern: Signed-in identity is rendered as a two-part row: bold Google-auth copy or auth-type string, followed by a dim `/auth` command hint.
Ink: `Box`, `Text`
Tree: `Box[col] > Box[row] > Text[primary wrap=truncate-end] > (Text > Text[bold]{"Signed in with Google:"?} + email? | "Authenticated with <authType>") + Text[secondary]{" /auth"}`

14. Source: `packages/cli/src/ui/components/messages/ToolGroupMessage.tsx:389-482`
Pattern: Tool groups are routed into specialized child renderers: grouped subagents, compact dense tools, topic messages, shell-tool cards, or standard tool cards, with shared closing-border logic between groups.
Ink: `Box`, `Text`
Tree: `Box[col w=terminalWidth padRight=4] > (*group => (SubagentGroupDisplay + closingBorder?) | DenseToolMessage | Box[mb=1] > TopicMessage | ShellToolMessage | ToolMessage + outputFileNotice? + closingBorder?)`

## Status And Workflow

15. Source: `packages/cli/src/ui/components/Notifications.tsx:149-161`
Pattern: Startup warnings render as a vertical stack of rows with a fixed-width warning-glyph gutter and a growable message cell.
Ink: `Box`, `Text`
Tree: `Box[col my=1] > *warning => Box[row] > Box[w=3] > Text[warning]{"⚠ "} + Box[grow=1] > Text[warning]{message}`

16. Source: `packages/cli/src/ui/components/Notifications.tsx:163-177`
Pattern: Initialization failures render inside a rounded red border with the main error and a follow-up remediation sentence on the same row group.
Ink: `Box`, `Text`
Tree: `Box[border=round borderColor=error padX=1 mb=1] > Text[error]{"Initialization Error: ..."} + Text[error]{" Please check API key and configuration."}`

17. Source: `packages/cli/src/ui/components/UpdateNotification.tsx:14-22`
Pattern: Update notices are a single warning-colored message inside a rounded bordered bar with horizontal padding and vertical margin.
Ink: `Box`, `Text`
Tree: `Box[border=round borderColor=warning padX=1 my=1] > Text[warning]{message}`

18. Source: `packages/cli/src/ui/components/Composer.tsx:82-84`
Pattern: When approval is pending and the composer should collapse, the component returns nothing instead of leaving an empty frame.
Ink: none
Tree: `null`

19. Source: `packages/cli/src/ui/components/Composer.tsx:93-179`
Pattern: The live composer stack can include resume status, queued messages, todo tray, shortcuts help, toast rail, status rows, debug console, input prompt, and footer in one vertical bundle.
Ink: `Box`, `useIsScreenReaderEnabled`
Tree: `Box[col w=terminalWidth] > ConfigInitDisplay? + QueuedMessageDisplay? + TodoTray? + ShortcutsHelp? + Box[minH=1 ml?] > ToastDisplay? + Box[w=100% col] > StatusRow + OverflowProvider? > Box[col] > DetailedMessagesDisplay + ShowMoreLines + InputPrompt? + Footer?`

20. Source: `packages/cli/src/ui/components/ConfigInitDisplay.tsx:25-74`
Pattern: Startup progress is a spinner-prefixed single line whose message is rewritten to include MCP connection counts and waiting server names.
Ink: `Box`, `Text`
Tree: `Box[mt=1] > Text > GeminiSpinner + Text[primary]{message}`

21. Source: `packages/cli/src/ui/components/StatusRow.tsx:108-152`
Pattern: When hooks are active, the status node synthesizes a single loading line from one or more hook names and sends it through the generic loading-indicator component.
Ink: `Box`, `ResizeObserver`
Tree: `Box[ref=ResizeObserver] > LoadingIndicator[inline currentLoadingPhrase="Executing Hook(s): ..." thought=null]`

22. Source: `packages/cli/src/ui/components/StatusRow.tsx:110-152`
Pattern: When no hooks are active but the model is thinking, the same status node routes sanitized thought text and witty-phrase state into the inline loading indicator.
Ink: `Box`, `ResizeObserver`
Tree: `Box[ref=ResizeObserver] > LoadingIndicator[inline thought=sanitizedThought currentLoadingPhrase=undefined wittyPhrase elapsedTime]`

23. Source: `packages/cli/src/ui/components/StatusRow.tsx:304-305`
Pattern: If all detail and minimal-status conditions are false, the status row still reserves a one-line spacer so the composer height stays stable.
Ink: `Box`
Tree: `Box[h=1]`

24. Source: `packages/cli/src/ui/components/StatusRow.tsx:243-245,247-330`
Pattern: Minimal mode can compress row 1 into a tight horizontal cluster containing the loading/status node and a colored mode bullet.
Ink: `Box`, `Text`
Tree: `Box[col w=100%] > Box[row justify=space-between minH=1] > Box[row grow=1 shrink=1] > Box[row columnGap=1] > StatusNode + Box > Text[color=modeColor]{"● mode"}`

25. Source: `packages/cli/src/ui/components/StatusRow.tsx:331-336`
Pattern: Interactive shell wait state replaces the normal loader with a single warning-colored line spanning the row.
Ink: `Box`, `Text`
Tree: `Box[w=100% ml=1] > Text[warning]{interactiveShellWaitingPhrase}`

26. Source: `packages/cli/src/ui/components/StatusRow.tsx:383-450`
Pattern: Detailed row 2 shows approval-mode, shell-mode, and raw-markdown indicators on the left and the current status/context cluster on the right, switching to a column layout on narrow terminals.
Ink: `Box`, `Text`
Tree: `Box[(row|col) justify=space-between] > Box[row ml=1] > ApprovalModeIndicator? + Box[ml=1] > ShellModeIndicator? + Box[ml=1] > RawMarkdownIndicator? + Box[row mt? ml?] > StatusDisplay + ContextUsageDisplay?`

27. Source: `packages/cli/src/ui/components/StatusRow.tsx:434-449`
Pattern: In minimal mode with context visible, `ContextUsageDisplay` is appended as a right-side supplement after the compact `StatusDisplay`.
Ink: `Box`
Tree: `Box[row align=center] > StatusDisplay + Box[ml=1] > ContextUsageDisplay[promptTokenCount model terminalWidth]`

28. Source: `packages/cli/src/ui/components/BackgroundTaskDisplay.tsx:228-300,397-488`
Pattern: Background shell viewing mode is a bordered panel with a tab strip, active PID/focus badge, help legend, scrollable output body, and a truncated log-path footer.
Ink: `Box`, `Text`
Tree: `Box[col h=100% w=100% border=single borderColor?=focus] > Box[row justify=space-between borderBottomOnly padX=1] > Box[row] > renderTabs + Text[bold]{(PID...) (Focused)?} + helpText + Box[grow=1 overflow=hidden padX=1] > renderOutput + Box[padX=1] > Text[secondary]{"Log: <path>"}`

29. Source: `packages/cli/src/ui/components/BackgroundTaskDisplay.tsx:302-395,454-486`
Pattern: Process-selection mode swaps the output pane for a full-height `RadioButtonSelect` list under an instructional header, while preserving the surrounding shell frame.
Ink: `Box`, `Text`
Tree: `Box[col border=single] > headerRow + Box[grow=1 overflow=hidden padX=1] > Box[col h=100% w=100%] > Box[shrink=0 mb=1 pt=1] > Text[bold]{select/kill/cancel hints} + Box[grow=1 w=100%] > RadioButtonSelect[maxItemsToShow initialIndex onHighlight] + footer`

30. Source: `packages/cli/src/ui/components/BackgroundTaskDisplay.tsx:417-450`
Pattern: ANSI shell output is virtualized line-by-line; each line can itself be a wrapped `Text` containing nested token-level `Text` spans with fg/bg/inverse/dim/bold/italic/underline props.
Ink: `Box`, `Text`
Tree: `ScrollableList[data=AnsiLine[]] > Box[h=1 overflow=hidden] > Text[wrap=truncate] > *token => Text[color bg inverse dim bold italic underline]{token.text}`

31. Source: `packages/cli/src/ui/components/NewAgentsNotification.tsx:60-131`
Pattern: Newly discovered agents show in a rounded warning card with an inset single-border review list, optional MCP-server detail rows, and a radio-button action selector.
Ink: `Box`, `Text`
Tree: `Box[col w=100%] > Box[col border=round borderColor=warning pad=1 mx=1] > Box[col mb=1] > title + body + Box[col mt=1 border=single pad=1] > *agent => Box[col] > Box[row] > Box[shrink=0] > Text[bold]{"- name:"} + Text[secondary]{description} + Box[ml=2] > Text[secondary]{mcpServers}? + RadioButtonSelect`

32. Source: `packages/cli/src/ui/components/NewAgentsNotification.tsx:120-124`
Pattern: While processing the user’s choice, the action selector is replaced by a minimal spinner-plus-label row.
Ink: `Box`, `Text`
Tree: `Box > CliSpinner + Text{" Processing..."}`

## Dialogs And Selection

33. Source: `packages/cli/src/ui/components/EmptyWalletDialog.tsx:68-107`
Pattern: Quota exhaustion opens a rounded vertical dialog that mixes warning copy, slash-command hints, a muted credit-delay note, and a `RadioButtonSelect` of recovery actions.
Ink: `Box`, `Text`
Tree: `Box[col border=round pad=1] > Box[col mb=1] > warningLine + resetLine? + slashHint[/stats] + slashHint[/model] + slashHint[/auth] + Box[mb=1] > Text{purchase credits} + Box[mb=1] > Text[dim]{delay note} + Box[mb=1] > Text{proceed?} + Box[mt=1 mb=1] > RadioButtonSelect`

34. Source: `packages/cli/src/ui/components/ValidationDialog.tsx:162-179`
Pattern: The default validation state is a simple rounded box with an action sentence, radio options, and an optional dim learn-more footer with an accented URL.
Ink: `Box`, `Text`
Tree: `Box[col border=round pad=1] > Box[mb=1] > Text{Further action is required...} + Box[mt=1 mb=1] > RadioButtonSelect + Box[mt=1] > Text[dim]{"Learn more: "} + Text[accent]{url}?`

35. Source: `packages/cli/src/ui/components/ValidationDialog.tsx:132-151`
Pattern: After launching or displaying the verification link, the dialog becomes a waiting panel with a spinner row, optional URL/error text, and a dim Enter-to-confirm hint.
Ink: `Box`, `Text`
Tree: `Box[col border=round pad=1] > Box > CliSpinner + Text{" Waiting for verification..."} + Box[mt=1] > Text{errorOrUrl}? + Box[mt=1] > Text[dim]{"Press Enter when verification is complete."}`

36. Source: `packages/cli/src/ui/components/ValidationDialog.tsx:115-129`
Pattern: Browser-launch failures reuse the same rounded dialog shell but swap the body to an error line plus the original action selector.
Ink: `Box`, `Text`
Tree: `Box[col border=round pad=1] > Text[error]{launchError} + Box[mt=1] > RadioButtonSelect`

37. Source: `packages/cli/src/ui/components/ExitPlanModeDialog.tsx:242-281`
Pattern: Exiting plan mode delegates the actual questionnaire UI to `AskUserDialog`, feeding it the full plan text as the question and an extra external-editor shortcut in the footer.
Ink: `Box`, `useStdin`
Tree: `Box[col w=width] > AskUserDialog[questions=[CHOICE Approval(question=planContent,...)] extraParts=["<editHint> to edit plan"]]`

38. Source: `packages/cli/src/ui/components/AskUserDialog.tsx:223-283,1220-1230`
Pattern: Multi-question flows end on a review tab that shows answered vs unanswered rows, a warning if any answers are missing, and a footer for Enter submit plus next/prev edit navigation.
Ink: `Box`, `Text`
Tree: `Box[aria=Review] > Box[col] > progressHeader? + Box[mb=1] > Text[bold]{"Review your answers:"} + Box[mb=1] > Text[warning]{unansweredCount}? + Box[col] > *question => Box > Text[secondary]{header} + Text[secondary]{" → "} + Text[(primary|warning)]{answer|"(not answered)"} + DialogFooter`

39. Source: `packages/cli/src/ui/components/AskUserDialog.tsx:380-407,1256-1269`
Pattern: Text questions render a markdown prompt above a single input row with a green `> ` prefix and a live `TextInput`.
Ink: `Box`, `Text`, `type DOMElement`
Tree: `Box[col] > progressHeader? + Box[mb=1] > MaxSizedBox > MarkdownDisplay[autoBoldIfPlain(question)] + Box[row mb=1] > Text[success]{"> "} + TextInput[buffer placeholder onSubmit] + DialogFooter`

40. Source: `packages/cli/src/ui/components/AskUserDialog.tsx:890-1020,1270-1282`
Pattern: Choice questions use a scrollable base selection list that can synthesize “All of the above”, inline custom-option editing, clickable checkboxes, and a Done row for multi-select mode.
Ink: `Box`, `Text`, `type DOMElement`
Tree: `Box[col] > progressHeader? + Box[mb=1] > MaxSizedBox > Box[col] > MarkdownDisplay + Text[italic]{"(Select all that apply)"}? + BaseSelectionList[showScrollArrows maxItemsToShow focusKey?] > (optionRow[checkbox? + label + checkmark? + description?] | otherRow[checkbox? + TextInput + checkmark?] | doneRow[bold]) + DialogFooter`

41. Source: `packages/cli/src/ui/components/shared/BaseSelectionList.tsx:114-151,226-274`
Pattern: Shared selection lists use a dedicated indicator gutter, optional numbering column, selection background highlight, and conditional up/down arrows when the list is windowed.
Ink: `Box`, `Text`, `type DOMElement`
Tree: `Box[col] > Text{"▲"}? + *visibleItem => Box[bg?=focus] > Box[minW=2] > Text{indicator|" "} + Box[minW=numberLen mr=1] > Text{n.}? + Box[grow=1] > renderItem + Text{"▼"}?`

42. Source: `packages/cli/src/ui/components/shared/SearchableList.tsx:195-261`
Pattern: The shared searchable-list shell stacks an optional title, a rounded search field, optional header, a windowed item region with scroll arrows, and an optional footer callback.
Ink: `Box`, `Text`
Tree: `Box[col w=100% h=100% padX=1] > titleBox? + Box[border=round borderColor=default padX=1 mb=1] > TextInput? + headerBox? + Box[col grow=1] > upArrow? + *visibleItem => Box[mb=1 mx=1] > renderItem + downArrow? + footerBox?`

## Messages And History

43. Source: `packages/cli/src/ui/components/messages/GeminiMessage.tsx:31-50`
Pattern: Gemini responses use a fixed-width accent prefix column (`✦ `) and a growable markdown body whose width and available height are reduced by the prefix.
Ink: `Box`, `Text`
Tree: `Box[row] > Box[w=2] > Text[accent aria=modelPrefix]{"✦ "} + Box[grow=1 col] > MarkdownDisplay[text isPending terminalWidth=terminalWidth-2 availableHeight?=h-1 renderMarkdown]`

44. Source: `packages/cli/src/ui/components/messages/UserMessage.tsx:52-80`
Pattern: User prompts render inside a padded message background, with a fixed `> ` gutter and primary or accent text depending on whether the line is a slash command.
Ink: `Box`, `Text`
Tree: `HalfLinePaddedBox[bg=message useBackgroundColor] > Box[row alignSelf=start w=width padX? ml?] > Box[w=2 shrink=0] > Text[accent aria=userPrefix]{"> "} + Box[grow=1] > Text[wrap color=(accent|primary)]{displayText}`

45. Source: `packages/cli/src/ui/components/messages/Todo.tsx:17-62`
Pattern: The todo tray is not rendered inline by tools; instead it mines the most recent todo-producing tool group and projects it into a `Checklist` with a toggle command hint.
Ink: none
Tree: `Checklist[title="Todo" items=map(todos=>{status,label}) isExpanded=showFullTodos toggleHint="<cmd> to toggle"]`

46. Source: `packages/cli/src/ui/components/Checklist.tsx:89-107`
Pattern: Expanded checklist mode renders a top-only border, a title/score row, and a full column of checklist items.
Ink: `Box`, `Text`
Tree: `Box[border=single noBottom noLeft noRight padX=1] > Box[col rowGap=1] > ChecklistTitleDisplay + ChecklistListDisplay > *ChecklistItem`

47. Source: `packages/cli/src/ui/components/Checklist.tsx:108-123`
Pattern: Collapsed checklist mode compresses the same data into a single row: title/score block on the left and the current in-progress item truncated on the right.
Ink: `Box`
Tree: `Box[border=single noBottom noLeft noRight padX=1] > Box[row columnGap=1 h=1] > Box[shrink=0 grow=0] > ChecklistTitleDisplay + Box[shrink=1 grow=1] > ChecklistItem[wrap=truncate]?`

48. Source: `packages/cli/src/ui/components/messages/ToolMessage.tsx:82-153`
Pattern: Standard tool-call history is a sticky-header card: status icon, tool info, optional focus hint and trailing indicator above a bordered content well, with optional MCP progress and embedded shell prompt.
Ink: `Box`
Tree: `<> > StickyHeader[w=terminalWidth first borderColor] > ToolStatusIndicator + ToolInfo + FocusHint + TrailingIndicator? + Box[col w=terminalWidth border=round noTop/noBottom padX=1] > McpProgressIndicator? + ToolResultDisplay[hasFocus maxLines? overflowDirection] + Box[padLeft=statusIndicatorWidth mt=1] > ShellInputPrompt?`

49. Source: `packages/cli/src/ui/components/messages/ToolResultDisplay.tsx:145-161,292-306`
Pattern: File-diff tool results in standard mode are wrapped in a column box and delegated into `DiffRenderer`, with the outer `SlicingMaxSizedBox` handling truncation and hidden-line accounting.
Ink: `Box`, `Text`
Tree: `Box[col w=childWidth] > SlicingMaxSizedBox[maxLines maxHeight overflowDirection] > DiffRenderer[diffContent filename availableTerminalHeight terminalWidth=childWidth]`

50. Source: `packages/cli/src/ui/components/messages/ToolResultDisplay.tsx:211-241`
Pattern: ANSI tool results in alternate-buffer mode are virtualized with fixed-height rows and start at top or bottom depending on overflow direction.
Ink: `Box`
Tree: `Box[col w=childWidth maxH=listHeight] > ScrollableList[containerHeight=listHeight data=AnsiOutput fixedItemHeight hasFocus initialScroll=(0|end)] > Box[h=1 overflow=hidden] > AnsiLineText`
