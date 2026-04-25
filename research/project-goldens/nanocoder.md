# nanocoder Ink patterns

Repo: `nano-collective/nanocoder`  
Commit inspected: `c01b160`  
Scope: 50 hand-curated, source-backed Ink UI patterns/states from the repo. Tree sketches are normalized abstractions, not literal JSX.

Legend:
- `>` parent/child
- `+` sibling
- `?` conditional child/branch
- Props use `name=value`
- Text payloads appear in `[...]`

1. Source: `source/app/App.tsx:668-675`
   Pattern: Directory-trust loading gate shown before the rest of the app mounts.
   Ink/components/hooks: `Box`, `Text`, `Spinner`
   Tree: `ThemeProvider > Box(dir=column,p=1) > Text(color=secondary)[Spinner(type=dots), " Checking directory trust..."]`

2. Source: `source/app/App.tsx:686-696`
   Pattern: Full-screen trust-check failure state with error text and recovery hint.
   Ink/components/hooks: `Box`, `Text`
   Tree: `ThemeProvider > Box(dir=column,p=1) > Text(color=error)["Error checking directory trust: ..."] + Text(color=secondary)["Please restart the application or check your permissions."]`

3. Source: `source/app/App.tsx:704-712`
   Pattern: App-level security gate that swaps the main shell for a trust disclaimer modal.
   Ink/components/hooks: `SecurityDisclaimer`, `ThemeContext.Provider`, `TitleShapeContext.Provider`
   Tree: `ThemeProvider > TitleShapeProvider > SecurityDisclaimer(onConfirm,onExit)`

4. Source: `source/app/App.tsx:726-744`
   Pattern: Onboarding composition that stacks the welcome banner over the VS Code extension prompt.
   Ink/components/hooks: `Box`, `WelcomeMessage`, `VSCodeExtensionPrompt`
   Tree: `ThemeProvider > TitleShapeProvider > Box(dir=column,p=1) > WelcomeMessage + VSCodeExtensionPrompt`

5. Source: `source/app/App.tsx:750-901`
   Pattern: Main application shell composing frozen chat history, modal overlays, scheduler view, and input area.
   Ink/components/hooks: `Box`, `ChatHistory`, `StreamingMessage`, `FileExplorer`, `IdeSelector`, `ModalSelectors`, `SchedulerView`, `ChatInput`
   Tree: `ThemeProvider > TitleShapeProvider > UIStateProvider > Box(dir=column,p=1,w=100%) > ChatHistory(...) + ?Box(ml=-1)>FileExplorer + ?Box(ml=-1)>IdeSelector + ?Box(ml=-1)>ModalSelectors + ?SchedulerView + ?ChatInput`

6. Source: `source/app/components/chat-history.tsx:32-45`
   Pattern: Chat transcript container that keeps a frozen `Static` queue and a separate live-updating region.
   Ink/components/hooks: `Box`, `ChatQueue`
   Tree: `Box(flexGrow=1,dir=column,minH=0) > ?ChatQueue(staticComponents,queuedComponents) + ?Box(ml=-1,dir=column)>liveComponent`

7. Source: `source/components/welcome-message.tsx:33-58`
   Pattern: Narrow-terminal welcome layout with gradient big-text logo plus a compact bordered tips box.
   Ink/components/hooks: `Box`, `Text`, `BigText`, `Gradient`, `useResponsiveTerminal`
   Tree: `Fragment > Gradient(colors=[primary,tool])>BigText(text="NC") + Box(dir=column,mb=1,border=round,borderColor=primary,px=2,py=1) > Box(mb=1)>Text(primary,bold)[version] + Text(text)["Quick tips:"] + Text*(secondary)[tips]`

8. Source: `source/components/welcome-message.tsx:59-97`
   Pattern: Wide welcome layout with full `Nanocoder` logo and a titled onboarding card.
   Ink/components/hooks: `Box`, `Text`, `BigText`, `Gradient`, `TitledBoxWithPreferences`, `useResponsiveTerminal`
   Tree: `Fragment > Gradient(colors=[primary,tool])>BigText(text="Nanocoder") + TitledBox(title=welcome,width=boxWidth,border=primary,dir=column) > Box(pb=1)>Text(text)["Tips for getting started:"] + Box(dir=column,pb=1)>Text*(secondary)[numbered tips] + Text(text)["/help for help"]`

9. Source: `source/components/security-disclaimer.tsx:47-71`
   Pattern: Security warning card that asks whether the current directory is trusted.
   Ink/components/hooks: `Box`, `Text`, `SelectInput`, `TitledBoxWithPreferences`, `useTerminalWidth`
   Tree: `Box(dir=column,p=1) > TitledBox(title="Security Warning",width=boxWidth,border=error,dir=column) > Text(warning,bold)[trust question] + Text[cwd] + Box(mt=1)>Text[risk copy] + SelectInput(items=["Yes, proceed","No, exit"])`

10. Source: `source/components/vscode-extension-prompt.tsx:156-183`
    Pattern: No-CLI fallback screen with step-by-step manual editor setup instructions.
    Ink/components/hooks: `Box`, `Text`
    Tree: `Box(dir=column,py=1) > Text(warning)["No supported VS Code flavor ... found in PATH."] + Box(mt=1)>Text(text)["To enable VS Code integration:"] + Box(ml=2,dir=column,mt=1)>Text*(secondary)[steps 1-3] + Box(mt=1)>Text(secondary)["Continuing without VS Code integration..."]`

11. Source: `source/components/vscode-extension-prompt.tsx:225-255`
    Pattern: Extension install prompt that lists detected editors and asks for confirmation.
    Ink/components/hooks: `Box`, `Text`, `SelectInput`
    Tree: `Box(dir=column,py=1) > Text(primary,bold)["VS Code Extension"] + Box(mt=1)>Text(text)[diff-preview explainer] + ?Box(mt=1,dir=column)>Text(secondary)["Detected editors:"] + Text*(gray|white)[status rows] + Box(mt=1)>Text(text)["Install the extension now?"] + Box(mt=1)>SelectInput(items)`

12. Source: `source/components/vscode-extension-prompt.tsx:193-221`
    Pattern: Multi-editor selection mode using checkbox-like labels inside `SelectInput`.
    Ink/components/hooks: `Box`, `Text`, `SelectInput`
    Tree: `Box(dir=column,py=1) > Text(primary,bold)["Select Editors"] + Box(mt=1)>SelectInput(items=["[x]|[ ] cli", "--- Confirm ---", "--- Back ---"]) + Box(mt=1)>Text["Space/Enter to toggle, select Confirm to proceed"]`

13. Source: `source/components/vscode-extension-prompt.tsx:267-275`
    Pattern: Success acknowledgment with a single continuation prompt.
    Ink/components/hooks: `Box`, `Text`
    Tree: `Box(dir=column,py=1) > Text(success)["✓ " + message] + Box(mt=1)>Text(secondary)["Press Enter to continue..."]`

14. Source: `source/components/status.tsx:199-347`
    Pattern: Wide status dashboard card aggregating config, provider/model, MCP/LSP health, context usage, auto-compact, and update notice.
    Ink/components/hooks: `Text`, `Box`, `TitledBoxWithPreferences`, `useResponsiveTerminal`
    Tree: `TitledBox(title="Status",width=boxWidth,border=info,dir=column) > Text*[cwd,config,provider/model,theme] + ?Text[agents/preferences/custom commands/vscode] + ?Box(dir=column)[MCP summary + ?Box(ml=2,dir=column)>Text*(error)[failed servers]] + ?Box(dir=column)[LSP summary + ?Box(ml=2,dir=column)>Text*(error)[failed servers]] + ?Box(dir=column)[context line + ?warning hint] + ?Box(dir=column)[auto-compact lines] + ?Text*[update notice]`

15. Source: `source/components/development-mode-indicator.tsx:55-88`
    Pattern: Inline status chip row for development mode, optional tune badge, and context percentage.
    Ink/components/hooks: `Box`, `Text`, `useResponsiveTerminal`
    Tree: `Box(mt=1) > Text(color=modeColor)[Text(bold)[modeLabel], ?Text[" (Shift+Tab to cycle)"]] + ?Text(secondary)[" · "] + ?Text(info)[tuneLabel] + ?Text(secondary)[" · "] + ?Text(color=contextColor)[ctxPercent]`

16. Source: `source/components/assistant-message.tsx:33-57`
    Pattern: Assistant response card with standalone model label, left-side border accent, and token footer.
    Ink/components/hooks: `Box`, `Text`, `useTerminalWidth`
    Tree: `Fragment > Box(mt=1,mb=1)>Text(info,bold)[model + ":"] + Box(dir=column,mb=1,bg=base,w=boxWidth,p=1,borderLeft=bold,borderLeftColor=secondary)>Text(renderedMarkdown) + Box(mb=2)>Text(secondary)[~tokens]`

17. Source: `source/components/streaming-message.tsx:40-65`
    Pattern: Live streaming assistant output that truncates to the tail, shows a spinner, and reports token rate.
    Ink/components/hooks: `Box`, `Text`, `Spinner`, `useTerminalWidth`
    Tree: `Fragment > Box(mt=1,mb=1)>Text(info,bold)[Spinner(type=dots), " model"] + Text["  ~tokens · tok/s"] + Box(dir=column,mb=1,bg=base,w=boxWidth,p=1,borderLeft=bold,borderLeftColor=secondary)>?Text["…"] + Text(displayTail)`

18. Source: `source/components/user-message.tsx:72-124`
    Pattern: User message card that preserves paragraph spacing and highlights file placeholders inline.
    Ink/components/hooks: `Box`, `Text`, `useTerminalWidth`
    Tree: `Fragment > Box(mb=1)>Text(primary,bold)["You:"] + Box(dir=column,mb=1,bg=base,w=boxWidth,p=1,borderLeft=bold,borderLeftColor=primary)>Box(dir=column)>Box*(mb=paragraphGap?)>Text>Text*(placeholder? color=info,bold : color=text) + Box(mb=2)>Text(secondary)[~tokens]`

19. Source: `source/components/question-prompt.tsx:75-119`
    Pattern: Tool-issued question prompt with a bordered question header and multiple-choice answer list.
    Ink/components/hooks: `Box`, `Text`, `SelectInput`, `useInput`
    Tree: `Box(dir=column,mb=1) > Box(dir=row,mb=1,bg=base,w=boxWidth,p=1,borderLeft=bold,borderLeftColor=secondary)>Text(tool,bold)["?"] + Text(text)[" " + question] + Box(dir=column)>SelectInput(items=options) + Box(mt=1)>Text(secondary)["Press Escape to cancel"]`

20. Source: `source/components/question-prompt.tsx:96-111`
    Pattern: Freeform answer sub-mode for `ask_question`, with inline prompt and back hint.
    Ink/components/hooks: `Box`, `Text`, `TextInput`, `useInput`
    Tree: `Box(dir=column,mb=1) > questionHeader + Box(dir=column)>Box>Text(secondary)["> "] + TextInput(value,onSubmit) + Box(mt=1)>Text(secondary)["Press Enter to submit, Escape to go back"]`

21. Source: `source/components/user-input.tsx:476-495`
    Pattern: Disabled input footer reduced to a spinner row plus the persistent mode indicator.
    Ink/components/hooks: `Box`, `Text`, `Spinner`, `DevelopmentModeIndicator`
    Tree: `Box(dir=column,py=1,w=100%,mt=1) > Text(secondary)[Spinner(type=dots), " Press Esc to cancel", ?compactToggleHint] + DevelopmentModeIndicator(...)`

22. Source: `source/components/user-input.tsx:499-542`
    Pattern: Main interactive prompt box with a left accent border and optional bash-mode heading.
    Ink/components/hooks: `Box`, `Text`, `TextInput`
    Tree: `Fragment > Text(primary|tool,bold)["What would you like me to help with?"|"Bash mode"] + Box(dir=column,mt=1,bg=base,w=boxWidth,p=1,borderLeft=bold,borderLeftColor=primary|tool)>Box>?Text(promptColor)["> "] + TextInput(placeholder="/ commands, ! bash, ↑/↓ history",wrapWidth=boxWidth-3) + ?Text(secondary)["Press escape again to clear"]`

23. Source: `source/components/user-input.tsx:544-555`
    Pattern: Command autocompletion list rendered below the main input box.
    Ink/components/hooks: `Box`, `Text`
    Tree: `Box(dir=column,mt=1) > Text(secondary)["Available commands:"] + Text*(color=info|primary)["/" + commandName]`

24. Source: `source/components/user-input.tsx:557-572`
    Pattern: File mention autocomplete list with highlighted current selection.
    Ink/components/hooks: `Box`, `Text`
    Tree: `Box(dir=column,mt=1) > Text(secondary)["File suggestions (↑/↓ to navigate, Tab to select):"] + Text*(selected? color=info,bold : color=primary)[("▸ " | "  ") + file.path]`

25. Source: `source/components/tool-confirmation.tsx:137-177`
    Pattern: Tool approval modal that shows a formatter-generated preview above a yes/no `SelectInput`.
    Ink/components/hooks: `Box`, `Text`, `SelectInput`, `useInput`
    Tree: `Box(w=boxWidth,mb=1) > Box(dir=column) > ?Box(mb=1)>Text(secondary)["Loading preview..."] + ?Box(mb=1,dir=column)>Box>[formatterPreview] + Fragment[Box(mb=1)>Text(tool)["Do you want to execute ...?"] + SelectInput(items=[yes,no]) + Box(mt=1)>Text(secondary)["Press Escape to cancel"]]`

26. Source: `source/components/tool-confirmation.tsx:180-187`
    Pattern: Formatter-crash fallback that cancels execution and leaves a terse continuation notice.
    Ink/components/hooks: `Box`, `Text`
    Tree: `Box(w=boxWidth,mb=1) > Box(dir=column) > Box(mt=1)>Text(error)["Tool execution cancelled due to formatter error."] + Text(secondary)["Press Escape to continue"]`

27. Source: `source/components/bash-progress.tsx:91-127`
    Pattern: Live `execute_bash` progress card rendered as structured tool output with command and output preview.
    Ink/components/hooks: `Box`, `Text`, `ToolMessage`, `useInput`
    Tree: `ToolMessage(hideBox=true,isLive?) > Box(dir=column) > Text(tool)["⚒ execute_bash"] + Box>Text(secondary)["Command: "] + Box(ml=1,flexShrink=1)>Text(primary,wrap=truncate-end)[command] + ?Box(dir=column)>Text(secondary)["Output: "] + Text(text)[outputPreview]`

28. Source: `source/components/bash-progress.tsx:103-121`
    Pattern: Completed `execute_bash` summary that replaces preview text with a status dot and token estimate.
    Ink/components/hooks: `Box`, `Text`, `ToolMessage`
    Tree: `ToolMessage(hideBox=true) > Box(dir=column) > Text(tool)["⚒ execute_bash"] + commandRow + Box>Text(secondary)["Status: "] + Text(dotColor)["●"] + Box>Text(secondary)["Tokens: "] + Text(text)[~estimatedTokens]`

29. Source: `source/components/agent-progress.tsx:68-107`
    Pattern: Live subagent progress card with description truncation and inline tool-call/token counters.
    Ink/components/hooks: `Box`, `Text`, `ToolMessage`
    Tree: `ToolMessage(hideBox=true,isLive?) > Box(dir=column) > Text(tool)["⚒ agent: " + subagentName] + Box(flexShrink=1)>Text(primary,wrap=truncate-end)[shortDesc] + Box>Text(secondary)[toolCallCount/tokens summary]`

30. Source: `source/components/agent-progress.tsx:131-149`
    Pattern: Parallel subagent display that simply stacks one progress card per agent.
    Ink/components/hooks: `Box`, `AgentProgress`
    Tree: `Box(dir=column) > Box*(mb=1)>AgentProgress(subagentName,description,isLive/completed)`

31. Source: `source/components/file-explorer/index.tsx:364-425`
    Pattern: File preview mode with selection state, line-number gutter, scroll range, and action legend.
    Ink/components/hooks: `Box`, `Text`, `StyledTitle`, `useInput`
    Tree: `Box(dir=column,px=1) > StyledTitle("/explorer - path") + Box(mt=1)>Text(success|secondary)[selected?] + ?Text(secondary)[" | n file(s) (~tokens)"] + Box(dir=column,mt=1)>(Text(warning)[previewError] | Text*(wrap=truncate)[Text(secondary)[lineNo], Text(secondary)[" | "], line]) + ?Box(mt=1)>Text(secondary)[lineRange] + Box(mt=1)>Text(secondary)[scroll/select/back help]`

32. Source: `source/components/file-explorer/index.tsx:429-507`
    Pattern: Tree view with optional search header, selected-context budget notice, item list, and status/help footer.
    Ink/components/hooks: `Box`, `Text`, `StyledTitle`, `TreeItem`, `useInput`
    Tree: `Box(dir=column,px=1) > StyledTitle("/explorer") + ?Box(mt=1)>Text(primary)["Search: ", Text(bold)[query||"_"]] + ?Box(mt=1,dir=column)>Text(success)[selectedCount/tokens] + ?Text(warning)["That's a lot of context!"] + Box(dir=column,mt=1)[TreeItem* | Text(secondary)["No matches found"|"Empty directory"]] + Box(mt=1,dir=column)[?Box>Text(path + ?size), Box(mt=1)>Text(secondary)[help]]`

33. Source: `source/components/file-explorer/tree-item.tsx:18-58`
    Pattern: Directory row that reuses the prefix slot for expansion, partial-selection, or full-selection state.
    Ink/components/hooks: `Box`, `Text`
    Tree: `Box > Text(color=success|text)[indent + ("✓ " | "◐ " | "v " | "> " | "  ")] + Text(color=primary|success|text,bold?,inverse?)[dirName + "/"]`

34. Source: `source/components/session-selector.tsx:142-155`
    Pattern: Recent-session picker with truncated labels and a bottom interaction legend.
    Ink/components/hooks: `Box`, `Text`, `SelectInput`, `useInput`
    Tree: `Box(dir=column,my=1) > Text(bold)["Recent Sessions:"] + Box(mt=1)>SelectInput(limit=min(items,10),items="[n] title (messages) - age") + Box(mt=1)>Text["↑/↓ to navigate • Enter to select • Esc to cancel"]`

35. Source: `source/components/checkpoint-selector.tsx:66-121`
    Pattern: Backup-confirmation dialog shown before loading a checkpoint over an active session.
    Ink/components/hooks: `Box`, `Text`, `TitledBoxWithPreferences`, `useInput`
    Tree: `TitledBox(title="Checkpoint Load - Backup Confirmation",width=boxWidth,border=warning,dir=column) > Box(mb=1)>Text[text][currentMessageCount] + ?Box(dir=column,mb=1)>Text*[checkpoint details] + Box(mb=1)>Text(warning,bold)[backup question] + Box(mb=1)>Text[text]["[Y] Yes ... [N] No ... [Esc] Cancel"] + Box>Text(secondary)[key help]`

36. Source: `source/components/checkpoint-display.tsx:42-125`
    Pattern: Bordered checkpoint table with fixed-width columns and a text separator row.
    Ink/components/hooks: `Box`, `Text`
    Tree: `Box(dir=column,my=1) > Box(border=round,borderColor=primary,px=2,py=1) > Box(dir=column) > Text(primary,bold)[title] + Box(mt=1,dir=column) > Box(dir=row)[header columns] + Box>Text(secondary)["─"...] + Box(dir=row)*(name,created,messages,files,size columns)`

37. Source: `source/components/task-list-display.tsx:45-80`
    Pattern: Compact task list rows with dedicated icon, ordinal, and truncated title columns.
    Ink/components/hooks: `Box`, `Text`, `useTerminalWidth`
    Tree: `Box(dir=column,mt=1,mb=1) > Box(dir=column) > Box>Text(primary,bold)[title] + Box(dir=column)>Box(dir=row,w=boxWidth)*(Box(w=2)>Text(statusColor)[icon] + Box(w=3)>Text(secondary)[index] + Box(flexShrink=1)>Text(color=text|secondary,wrap=truncate-end)[task.title])`

38. Source: `source/wizards/steps/summary-step.tsx:135-230`
    Pattern: Combined provider/MCP summary screen with section dividers, nested detail blocks, and final action picker.
    Ink/components/hooks: `Box`, `Text`, `SelectInput`, `useInput`
    Tree: `Box(dir=column) > Box(mb=1)>Text(primary,bold)["Configuration Summary"] + Box(mb=1)>Text(secondary)["─"...] + Box(dir=column,mb=1)[config files] + Box(dir=column,mb=1)[providers list | Text(warning)["None"]] + Box(dir=column,mb=1)[MCP servers list with transport icon/details | Text(warning)["None"]] + ?Box(mb=1)>Text(warning)[noProvidersWarning] + SelectInput(options)`

39. Source: `source/wizards/steps/location-step.tsx:96-112`
    Pattern: Existing-config discovery screen that presents the found path and a next-action menu.
    Ink/components/hooks: `Box`, `Text`, `SelectInput`, `useInput`
    Tree: `Box(dir=column) > Box(mb=1,dir=column)>Text(primary,bold)["Configuration found at:"] + Text(secondary)[path|truncatedPath] + SelectInput(existingConfigOptions)`

40. Source: `source/wizards/steps/location-step.tsx:115-146`
    Pattern: Config-location chooser with an optional warning when only a global config already exists.
    Ink/components/hooks: `Box`, `Text`, `SelectInput`, `useInput`
    Tree: `Box(dir=column) > Box(mb=1)>Text(primary,bold)[locationQuestion] + ?Box(mb=1,dir=column)>Text(warning)[global note] + ?Text(secondary)[globalPath] + SelectInput(locationOptions) + ?Box(mt=1)>Text(secondary)[tip]`

41. Source: `source/wizards/steps/provider-step.tsx:602-621`
    Pattern: Provider wizard entry menu that asks whether to start from a template or custom setup.
    Ink/components/hooks: `Box`, `Text`, `SelectInput`, `useInput`
    Tree: `Box(dir=column) > Box(mb=1)>Text(primary,bold)["Let's add AI providers. Would you like to use a template?"] + ?Box(mb=1)>Text(success)[providerCount + " already added"] + SelectInput(initialOptions)`

42. Source: `source/wizards/steps/provider-step.tsx:625-645`
    Pattern: Provider template picker that also echoes already-added provider names.
    Ink/components/hooks: `Box`, `Text`, `SelectInput`
    Tree: `Box(dir=column) > Box(mb=1)>Text(primary,bold)["Choose a provider template:"] + ?Box(mb=1)>Text(success)["Added: providerA, providerB"] + SelectInput(templateOptions)`

43. Source: `source/wizards/steps/provider-step.tsx:671-682`
    Pattern: Small branching menu for editing versus deleting a specific provider.
    Ink/components/hooks: `Box`, `Text`, `SelectInput`
    Tree: `Box(dir=column) > Box(mb=1)>Text(primary,bold)["<provider> - What would you like to do?"] + SelectInput(items=["Edit this provider","Delete this provider"])`

44. Source: `source/wizards/steps/provider-step.tsx:686-759`
    Pattern: Field-entry form screen with field counter, required-marker prompt, rounded input box, optional masking, and responsive footer help.
    Ink/components/hooks: `Box`, `Text`, `TextInput`, `useInput`
    Tree: `Box(dir=column) > Box(mb=1)[Text(primary,bold)[template + " Configuration"], Text[" (Field i/n)"]] + Box>Text[prompt + ?requiredMarker + ?maskedHint] + Box(border=round,borderColor=secondary,mb=1)>TextInput(mask="*"? ifSensitive) + ?Box(mb=1)>Text(error) + (isNarrow ? Box(dir=column)>Text*(secondary)[help lines] : Box>Text(secondary)[inline help])`

45. Source: `source/wizards/steps/provider-step.tsx:763-778`
    Pattern: Interim loading state while the wizard fetches models from a provider endpoint.
    Ink/components/hooks: `Box`, `Text`, `Spinner`
    Tree: `Box(dir=column) > Box(mb=1)>Text(primary,bold)[template + " Configuration"] + Box>Text(info)[Spinner(type=dots), " Fetching models from " + baseUrlOrTemplate + "..."]`

46. Source: `source/wizards/steps/provider-step.tsx:798-838`
    Pattern: Multi-select model checklist built on `SelectInput`, with select-all and done sentinel rows.
    Ink/components/hooks: `Box`, `Text`, `SelectInput`
    Tree: `Box(dir=column) > Box(mb=1)>Text(primary,bold)[template + " Configuration"] + Box(mb=1)>Text["Select models to use (n selected):"] + SelectInput(items=["[ ] Select All"|"[✓] All selected", "[ ]|[✓] model"...,"Done - Continue with selected models"]) + ?Box(mt=1)>Text(error) + footerHelp`

47. Source: `source/wizards/steps/mcp-step.tsx:418-442`
    Pattern: MCP wizard landing menu that summarizes already-configured servers before offering add/edit/done actions.
    Ink/components/hooks: `Box`, `Text`, `SelectInput`, `useInput`
    Tree: `Box(dir=column) > Box(mb=1)>Text(primary,bold)["Configure MCP Servers"] + ?Box(dir=column,mb=1)>Text(success)[serverCount + " configured:"] + Text*(secondary)["• name (transport)"] + SelectInput(initialOptions)`

48. Source: `source/wizards/steps/mcp-step.tsx:449-493`
    Pattern: Tabbed server-template picker splitting local STDIO and remote HTTP/WebSocket options.
    Ink/components/hooks: `Box`, `Text`, `Tabs`, `Tab`, `SelectInput`
    Tree: `Box(dir=column) > Box(mb=1)>Text(primary,bold)["Add MCP Servers:"] + ?Box(mb=1)>Text(success)["Added: ..."] + Tabs(activeTabColor=success)>Tab("Local Servers (STDIO)") + Tab("Remote Servers (HTTP/WebSocket)") + Box(mt=1,mb=1)>Text[tab-specific prompt] + SelectInput(templateOptions) + Box(mt=1)>Text(secondary)["Arrow keys: Navigate | Tab: Switch tabs"]`

49. Source: `source/wizards/steps/mcp-step.tsx:541-618`
    Pattern: MCP field-entry form that supports both masked single-line input and multiline env-var capture.
    Ink/components/hooks: `Box`, `Text`, `TextInput`, `useInput`
    Tree: `Box(dir=column) > Box(mb=1)[Text(primary,bold)[template + " Configuration"], Text[" (Field i/n)"]] + Box>Text[prompt + ?requiredMarker + ?default + ?maskedHint] + (isMultiline ? Box(dir=column,mb=1)>Box(border=round,borderColor=secondary,px=1)>Text(buffer || "(empty)") + Box(mt=1)>Text(secondary)["Type to add lines. Press Esc when done to submit."] : Box(border=round,borderColor=secondary,mb=1)>TextInput(mask="*"? ifSensitive)) + ?Box(mb=1)>Text(error) + Box>Text(secondary)["Press Esc to submit"|"Press Enter to continue"]`

50. Source: `source/wizards/provider-wizard.tsx:417-453`
    Pattern: Provider-wizard completion screen that layers success state, saved path, auth next steps, and local-server reminder.
    Ink/components/hooks: `Box`, `Text`
    Tree: `Box(dir=column) > Box(mb=1)>Text(success,bold)["✓ Configuration saved!"] + Box(mb=1)>Text["Saved to: " + providerConfigPath] + ?Box(dir=column,mb=1)>Text(primary)["Run /copilot-login ..."] + ?Text(primary)["Run /codex-login ..."] + ?Box(mb=1)>Text["Ensure your local server(s) are running before use."] + Box>Text(secondary)["Press Enter to continue"]`
