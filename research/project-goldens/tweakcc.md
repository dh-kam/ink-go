# tweakcc Ink usage patterns

Source repo: `https://github.com/Piebald-AI/tweakcc`  
Commit inspected: `2e1d03e7e22f5a993a03df9b4f2b921b8c229d07`  
Checkout used: `.tmp/project-research/tweakcc`

Notes:
- These are hand-picked from real render sites in the repo, not generated from a bulk grep dump.
- I intentionally include state branches inside the same component when the rendered tree meaningfully changes.
- Tree sketch DSL:
  - `Box(col)` / `Box(row)` = `Box flexDirection="column"|"row"`
  - `Text(...)` = `Text` with notable props only
  - `Comp(...)` = another component node used as a subtree
  - `map[...]` = repeated children
  - `?` = conditional branch
  - `Fragment[...]` = inline sibling group without wrapper

## Curated patterns

1. `src/ui/components/SelectInput.tsx:34-58`
   - UI: Generic vertical selector with one active row, active-row styling, and a dim inline description that only appears for the active item.
   - Ink / hooks: `Box`, `Text`, `useInput`
   - Tree: `Box(col)[ map item -> Box[ Text[ Text(active ? bold+cyan+selectedStyles : styles)[prefix + name], active && desc ? Text(dim)[" - " + desc] : null ] ] ]`

2. `src/ui/components/InstallationPicker.tsx:37-72`
   - UI: Full-screen ambiguous-installation picker with a bold warning header, selectable rows, saved-config note, and footer help text.
   - Ink / hooks: `Box`, `Text`, `useInput`, `useState`
   - Tree: `Box(col)[ Text(bold,yellow)[header], Text[" "], map candidate -> Box[ Text(active ? bold+cyan)[prefix + path], Text(dim)[" (" + kind + ", v" + version + ")"] ], Text[" "], Text[ "Your choice will be saved to " + inline blue fields ], Text[" "], Text(dim)[help] ]`

3. `src/ui/components/MainView.tsx:99-108`
   - UI: Conditional migration notice banner shown above the main menu when config schema changed.
   - Ink / hooks: `Box`, `Text`
   - Tree: `Box(col,gap=1)[ configMigrated ? Box[ Text(bold,blue)[migration notice] ] : null, ... ]`

4. `src/ui/components/MainView.tsx:40-77`
   - UI: Left-border notification banner with color chosen from success/error/info/warning type.
   - Ink / hooks: `Box`, `Text`
   - Tree: `Box(borderLeft,borderStyle=bold,borderColor=typeColor,paddingLeft=1,col)[ Text(color=typeColor)[message] ]`

5. `src/ui/components/MainView.tsx:124-133`
   - UI: Auto-generated "changes configured but not applied" reminder injected as an info notification only when there is no competing notification.
   - Ink / hooks: `Box`, `Text`
   - Tree: `!changesApplied && !notification ? NotificationBanner(type=info,message=multiline apply reminder) : null`

6. `src/ui/components/PiebaldAnnouncement.tsx:24-67`
   - UI: Split marketing hero with orange side rule, stacked promo copy on the left, and a fixed-width ANSI-art logo column on the right.
   - Ink / hooks: `Box`, `Text`, `Link`
   - Tree: `Box(row,height=10,alignItems=center,borderLeft=bold orange)[ Box(col,flexGrow=1,paddingLeft=1)[ Text(bold,orange)[title], Text[" "], Text[body with bold inline brand], Text[copy + Link[url] > Text(orange)], Text[" "], Text(dim,gray)["press 'h' to hide"] ], Box(col,width=22,alignItems=center)[ map ansiLine -> Text[line] ] ]`

7. `src/ui/components/ThemesView.tsx:152-198`
   - UI: Split two-pane theme browser with commands on the left and a live `ThemePreview` on the right.
   - Ink / hooks: `Box`, `Text`, `useInput`, `useState`, `useContext`
   - Tree: `Box(row)[ Box(col,width=50%)[ Header["Themes"], helpTextCol, Box(col)[ map theme -> Text(active ? yellow)[prefix + label], communityRow ] ], Box(width=50%)[ ThemePreview(theme=selectedTheme) ] ]`

8. `src/ui/components/ThemesView.tsx:184-193`
   - UI: Right-pane placeholder when the synthetic "Browse community themes..." row is selected instead of a local theme.
   - Ink / hooks: `Box`, `Text`
   - Tree: `Box(width=50%)[ Box(col,justifyContent=center,alignItems=center,height=10)[ Text(dim)["Press enter to browse community themes."] ] ]`

9. `src/ui/components/CommunityThemesView.tsx:143-149`
   - UI: Minimal loading state for community themes.
   - Ink / hooks: `Box`, `Text`, `useEffect`, `useState`
   - Tree: `Box(col)[ Header["Community Themes"], Text["Loading community themes..."] ]`

10. `src/ui/components/CommunityThemesView.tsx:162-208`
   - UI: Loaded community-theme browser with a selectable list on the left and a centered placeholder on the right before any preview is fetched.
   - Ink / hooks: `Box`, `Text`, `useInput`, `useEffect`, `useState`, `useRef`, `useCallback`
   - Tree: `Box(row)[ Box(col,width=50%)[ Header, helpTextCol, Box(col)[ map entry -> Text(active ? yellow)[prefix + name + " (" + id + ") by @" + author] ], statusMessage ? Box[Text(cyan)] : null, previewLoading ? Box[Text(dim)] : null ], Box(width=50%)[ Box(col,justifyContent=center,alignItems=center,height=10)[ Text(dim)["Press space to preview the selected theme."] ] ] ]`

11. `src/ui/components/CommunityThemesView.tsx:197-200`
   - UI: Same browser shell, but the right pane swaps the placeholder for a real `ThemePreview` once a community theme loads.
   - Ink / hooks: `Box`, `Text`, `useInput`, `useEffect`, `useState`, `useRef`, `useCallback`
   - Tree: `Box(row)[ leftListPane, Box(width=50%)[ ThemePreview(theme=previewTheme) ] ]`

12. `src/ui/components/ThemeEditView.tsx:319-343`
   - UI: Exit confirmation dialog asking whether the edited theme should become the active Claude Code theme, implemented as a nested `SelectInput`.
   - Ink / hooks: `Box`, `Text`, `SelectInput`, `useInput`, `useState`
   - Tree: `Box(col,marginTop=1)[ Box(marginBottom=1)[ Text(bold)[question] ], SelectInput(items=["Yes","No"]), Text(dim)[enter/esc help] ]`

13. `src/ui/components/ThemeEditView.tsx:374-390`
   - UI: Dynamic callout for the currently selected color key, with the border tinted by the selected theme color and the title rendered via `ColoredColorName`.
   - Ink / hooks: `Box`, `Text`
   - Tree: `Box(borderLeft=single,borderColor=currentTheme.colors[selectedKey],paddingLeft=1,col)[ ColoredColorName(colorKey=selectedKey,bold), Text[color description] ]`

14. `src/ui/components/ThemeEditView.tsx:458-519`
   - UI: Scroll-windowed color list that shows only a centered slice around the selected color and adds "more above/below" markers.
   - Ink / hooks: `Box`, `Text`
   - Tree: `Fragment[ adjustedStartIndex>0 ? Text(dim)["up more above"] : null, map visibleColor -> Box[ Text(active ? yellow : white)[prefix], Box(width=maxKeyLength+2)[ Text[ColoredColorName(...)] ], ColoredText(color=currentColor)[formattedValue] ], endIndex<total ? Text(dim)["down more below"] : null ]`

15. `src/ui/components/ThemeEditView.tsx:523-532`
   - UI: Inline text editor for theme name or ID with a round border and save/cancel hint.
   - Ink / hooks: `Box`, `Text`
   - Tree: `Box(col,marginTop=1)[ Text["Editing theme name/id:"], Box(borderStyle=round,borderColor=yellow,paddingX=1)[ Text[editingValue] ], Text(dim)[enter to save, esc to cancel] ]`

16. `src/ui/components/ThemeEditView.tsx:534-554`
   - UI: Full takeover into `ColorPicker` while keeping edits live-synced back into the current theme.
   - Ink / hooks: `ColorPicker`
   - Tree: `ColorPicker(initialValue=originalValue,colorKey=selectedColorKey,theme=currentTheme,onColorChange=liveThemeUpdate,onExit=commitAndReturn)`

17. `src/ui/components/ColoredColorName.tsx:17-28`
   - UI: `inverseText` sample rendered as text-colored label on top of the theme's permission color.
   - Ink / hooks: `ColoredText`
   - Tree: `ColoredText(color=inverseText,backgroundColor=permission,bold?)[ "inverseText" ]`

18. `src/ui/components/ColoredColorName.tsx:30-41`
   - UI: Any `diff*` color rendered as normal text placed on top of its own diff background, instead of tinting the glyphs.
   - Ink / hooks: `ColoredText`
   - Tree: `ColoredText(backgroundColor=diffColor,color=text,bold?)[ diffKeyName ]`

19. `src/ui/components/ThemePreview.tsx:112-167`
   - UI: Animated rainbow shimmer word built out of one outer `Text` plus per-character child `Text` nodes with cycling colors.
   - Ink / hooks: `Text`; React `useState`, `useEffect`
   - Tree: `Text[ map char(i) -> Text(color=isInShimmer ? shimmerColors[i % n] : baseColors[i % n])[char] ]`

20. `src/ui/components/ThemePreview.tsx:182-223`
   - UI: "Clawd" identity/header block with nested text color/background combinations, version display, cwd line, and login-success message.
   - Ink / hooks: `Box`, `Text`
   - Tree: `Box(col)[ Text[ ColoredText(clawd_body)[" block"], Text(color=clawd_body,bg=clawd_background)["mascot"], Text[ ColoredText(clawd_body)["tail"], Text(bold)["Claude Code"], version ] ], Text[ mascot tail + subscription/model ], ColoredText(clawd_body)[cwd line], Text[ ColoredText(success)["Login successful. Press "], ColoredText(success,bold)["Enter"], ColoredText(success)[" to continue..."] ] ]`

21. `src/ui/components/ThemePreview.tsx:226-262`
   - UI: Code-diff sample with removed and added lines, including nested word-level highlight spans inside broader line-level backgrounds.
   - Ink / hooks: `Text`
   - Tree: `Fragment[ Text["top border"], Text[line1], Text[ lineNo + ColoredText(bg=diffRemoved,color=text)["- console.log("] + ColoredText(bg=diffRemovedWord)["Hello, World!"] + ColoredText(bg=diffRemoved)[");"] ], Text[ lineNo + ColoredText(bg=diffAdded,color=text)["+ console.log("] + ColoredText(bg=diffAddedWord)["Hello, Claude!"] + ColoredText(bg=diffAdded)[");"] ] ]`

22. `src/ui/components/ThemePreview.tsx:263-275`
   - UI: Warning-colored confirmation box rendered as literal border lines with a mixed-color middle row and a dim keyboard hint.
   - Ink / hooks: `Text`
   - Tree: `Fragment[ ColoredText(warning)["top border"], ColoredText(warning)["question row"], Text[ ColoredText(warning)["| "], Text(dim)["Enter to confirm · Esc to exit"], ColoredText(warning)[" |"] ] ]`

23. `src/ui/components/ThemePreview.tsx:283-314`
   - UI: Prompt sample line showing a user command followed by the animated `ultrathink` rainbow word inline.
   - Ink / hooks: `Text`, `UltrathinkRainbowShimmer`
   - Tree: `Text[ Text["> list the dir " + UltrathinkRainbowShimmer(text="ultrathink",...) ] ]`

24. `src/ui/components/ThemePreview.tsx:315-343`
   - UI: Plan/permission prompt row that embeds an inverse-text "Allow" pill inside a sentence.
   - Ink / hooks: `Text`, `ColoredText`
   - Tree: `Fragment[ ColoredText(planMode)["plan border"], Text[ ColoredText(planMode)["| "], ColoredText(permission)["Ready to code? "], Text["Here is Claude's plan:"], ColoredText(planMode)[" |"] ], ColoredText(permission)["permission border"], Text[ ColoredText(permission)["| "], Text(bold)["Permissions:"], ColoredText(bg=permission,color=inverseText,bold)[" Allow "], Text["  Deny   Workspace"], ColoredText(permission)["|"] ] ]`

25. `src/ui/components/ThemePreview.tsx:344-374`
   - UI: Rejected file-update sample followed by dimmed removed/added diff lines.
   - Ink / hooks: `Text`, `ColoredText`
   - Tree: `Fragment[ Text["> list the dir"], Text[ ColoredText(error)["dot"], Text[" Update(__init__.py)"] ], Text[ Text[" subprompt"], ColoredText(error)["User rejected update..."] ], Text[line1 with diffRemovedDimmed + diffRemovedWordDimmed], Text[line2 with diffAddedDimmed + diffAddedWordDimmed] ]`

26. `src/ui/components/ThemePreview.tsx:375-410`
   - UI: Stacked runtime status rows showing success, generic result text, thinking state, auto-accept mode, plan mode, and IDE connection.
   - Ink / hooks: `Text`, `ColoredText`
   - Tree: `Box(col)[ Text[ ColoredText(success)["dot"], Text[" List(.)"] ], Text[ dot + normal text + ColoredText(permission)["path"] + bold count ], Text[ ColoredText(claude)["Th"], ColoredText(claudeShimmer)["ink"], ColoredText(claude)["ing..."], Text["(esc to interrupt)"] ], Text[ ColoredText(autoAccept)["accept edits on..."] ], Text[ ColoredText(planMode)["plan mode on..."] ], Text[ ColoredText(ide)["IDE connected ..."] ] ]`

27. `src/ui/components/ColorPicker.tsx:405-494`
   - UI: HSL slider mode with three labeled gradient bars and a literal `|` marker inserted at the current value.
   - Ink / hooks: `Box`, `Text`, `useInput`, `useState`, `useEffect`, `Fragment`
   - Tree: `Box(col,border=single,padding=1)[ header/help, Box[row][ labelCol(Text(active ? yellow)["Hue (n):"]), Box[ map segment(i) -> Fragment[ i==marker ? Text["|"] : Text(bg=color)[" "] ] ] ], repeat for Saturation and Lightness ]`

28. `src/ui/components/ColorPicker.tsx:495-553`
   - UI: RGB slider mode mirroring the HSL layout but for red/green/blue channels.
   - Ink / hooks: `Box`, `Text`, `useInput`, `useState`, `useEffect`, `Fragment`
   - Tree: `Box(col,border=single,padding=1)[ header/help, Box[row][ labelCol(Text(active ? yellow)["Red (n):"]), gradientBarWithMarker ], repeat for Green and Blue ]`

29. `src/ui/components/ColorPicker.tsx:561-637`
   - UI: Three-column Hex/RGB/HSL readout with branch-specific swatch rendering for diff colors (text on colored background) and `inverseText` (text on permission background).
   - Ink / hooks: `Box`, `Text`
   - Tree: `Box(row,justifyContent=space-between)[ col["Hex", sample], col["RGB", sample], col["HSL", sample] ] where sample = diff? Text(bg=current,color=theme.text,bold)[value] : inverseText? Text(color=current,bg=theme.permission,bold)[value] : Text(color=current,bold)[value]`

30. `src/ui/components/ThinkingStyleView.tsx:298-340`
   - UI: Split header/preview layout where the preview shows a fixed-width spinner cell and Claude-colored "Thinking..." text.
   - Ink / hooks: `Box`, `Text`, `useInput`, `useState`, `useEffect`, `useContext`
   - Tree: `Box(row,width=100%)[ leftHeaderPane, Box(col,width=previewWidth)[ Text(bold)["Preview"], Box(border=single,paddingX=1,row)[ Box(width=currentPhaseLen+1)[ Text(color=claudeColor)[currentPhase] ], Text[ Text(color=claudeColor)["Thinking... "], Text["(esc to interrupt)"] ] ] ] ]`

31. `src/ui/components/ThinkingStyleView.tsx:410-499`
   - UI: Scroll-windowed phases editor with active-row highlighting plus separate inline add/edit round-box rows.
   - Ink / hooks: `Box`, `Text`
   - Tree: `Box(col)[ phaseHeader + help, Box(col)[ maybe Text(dim)["up more above"], map phase -> Text(active ? cyan)[prefix + phase], maybe Text(dim)["down more below"], addingNew ? Box[ Text(yellow)[prefix], Box(border=round,yellow)[Text[input]] ] : null, editingPhase ? Box[ Text["Editing: "], Box(border=round,yellow)[Text[input]] ] : null ] ]`

32. `src/ui/components/ThinkingStyleView.tsx:502-581`
   - UI: Scrollable preset list where each row shows the preset name plus a compact concatenated frame preview.
   - Ink / hooks: `Box`, `Text`
   - Tree: `Box(col)[ presetHeader + warning, Box(col)[ maybe Text(dim)["up more above"], map preset -> Text(active ? cyan)[prefix + preset.name + " " + previewFrames], maybe Text(dim)["down more below"] ] ]`

33. `src/ui/components/ThinkingVerbsView.tsx:167-194`
   - UI: Format-string editor with selected-section arrow, contextual helper text, and a round-bordered current value box.
   - Ink / hooks: `Box`, `Text`, `useInput`, `useState`, `useContext`
   - Tree: `Box(col)[ Text[ selector + Text(bold)["Format"] ], selected ? Text(dim)[edit hint] : null, Box(marginLeft=2)[ Box(border=round,borderColor=editing ? yellow : gray)[ Text[currentFormatOrInput] ] ] ]`

34. `src/ui/components/ThinkingVerbsView.tsx:210-284`
   - UI: Scroll-windowed verbs list with selection, add-new row, and edit row.
   - Ink / hooks: `Box`, `Text`
   - Tree: `Box(col)[ selectedSectionHelp, Box(col)[ maybe Text(dim)["up more above"], map verb -> Text(active ? cyan)[prefix + verb], maybe Text(dim)["down more below"], addingNewVerb ? Box(row)[ Text(yellow)[prefix], Box(border=round,yellow)[Text[input]] ] : null, editingVerb ? Box(row)[ Text["Editing: "], Box(border=round,yellow)[Text[input]] ] : null ] ]`

35. `src/ui/components/UserMessageDisplayView.tsx:425-450`
   - UI: Preview branch for custom top/bottom-only borders, rendered manually as border lines above and below padded content.
   - Ink / hooks: `Box`, `Text`
   - Tree: `Box(col)[ Text(color=borderColor)[borderLine], paddingY>0 ? Box(height=paddingY) : null, Box(paddingX)[ styledText ], paddingY>0 ? Box(height=paddingY) : null, Text(color=borderColor)[borderLine] ]`

36. `src/ui/components/UserMessageDisplayView.tsx:451-481`
   - UI: Preview branch for normal Ink border styles and optional fit-to-content behavior.
   - Ink / hooks: `Box`, `Text`
   - Tree: `borderStyle!="none" || padding || fit ? Box(borderStyle=?,borderColor=?,alignSelf=fit?flex-start:null,flexGrow=fit?null:1)[ maybe Box(paddingX,paddingY)[styledText] : styledText ] : ...`

37. `src/ui/components/UserMessageDisplayView.tsx:411-423,482-484`
   - UI: Base preview branch that returns only a styled `Text` node when no border, padding, or fit-box wrapper is needed.
   - Ink / hooks: `Text`
   - Tree: `Text(bold?,italic?,underline?,strikethrough?,inverse?,color=fg,bg=bg)[formattedText]`

38. `src/ui/components/UserMessageDisplayView.tsx:868-928`
   - UI: Before/after comparison panel with two equal-width columns, each showing a sample user message and a follow-up status line.
   - Ink / hooks: `Box`, `Text`
   - Tree: `Box(col,marginTop=1,paddingX=1)[ Text(bold)["Preview"], Box(row,gap=2)[ Box(col,width=50%)[ Text(underline)["Before"], sampleMessageDefault, sampleFollowup ], Box(col,width=50%)[ Text(underline)["After"], createPreview(), sampleFollowup ] ] ]`

39. `src/ui/components/MiscView.tsx:712-747`
   - UI: Modal-like double-bordered security warning used before enabling sudo permission bypass.
   - Ink / hooks: `Box`, `Text`, `useInput`, `useState`, `useMemo`, `useContext`
   - Tree: `Box(col,paddingX=2)[ Box(border=double,borderColor=yellow,padding=2,col)[ Box(marginBottom=1)[ Text(bold,yellow)["SECURITY WARNING"] ], Box(marginBottom=1)[ Text[body] ], Box(marginBottom=1)[ Text(red)[danger line] ], Box(marginBottom=1)[ Text(bold)["Use with extreme caution."] ], Box[ Text["Press " + inline red Enter + inline green Escape + "..."] ] ] ]`

40. `src/ui/components/MiscView.tsx:748-845`
   - UI: Paginated miscellaneous-settings browser showing only four items, scroll indicators, per-item description, and indicator glyphs for booleans, numerics, and multivalue settings.
   - Ink / hooks: `Box`, `Text`
   - Tree: `Box(col)[ Header, Text(dim)[nav help], hasMoreAbove ? Text(dim)["up more above"] : null, map visibleItem -> Box(col)[ Box[ Text[ Text(active ? cyan)[prefix], Text(bold,active?cyan)[title] ] ], Box[ Text(dim)["  " + description] ], Box(marginLeft=4,marginBottom=1)[ Text[indicator + statusText + maybe dim arrow hint] ] ], hasMoreBelow ? Text(dim)["down more below"] : null, Box(marginTop=1)[ Text(dim)["Item i of n"] ] ]`

41. `src/ui/components/ToolsetsView.tsx:148-180`
   - UI: Toolset list where the selected row turns yellow and badges for default auto-accept / plan-mode toolsets append themed labels on the same line.
   - Ink / hooks: `Box`, `Text`, `useInput`, `useState`, `useContext`
   - Tree: `Box(col)[ map toolset -> Box(row)[ Text(active ? yellow)[prefix + name + " "], Text(active ? yellow)["(" + toolCountLabel + ")"], isDefault ? Text(color=autoAcceptColor)[" accept edits"] : null, isPlanMode ? Text(color=planModeColor)[" plan mode"] : null ] ]`

42. `src/ui/components/ToolsetEditView.tsx:181-210`
   - UI: Allowed-tools matrix with special `All` and `None` rows above individual tools, each rendered as a selectable bullet row.
   - Ink / hooks: `Box`, `Text`, `useInput`, `useState`, `useEffect`, `useContext`
   - Tree: `Box(col)[ Text(bold)["Allowed Tools:"], Box(marginLeft=2)[ Text(active?cyan)[prefix + selectedDot + " All"] ], Box(marginLeft=2)[ Text(active?cyan)[prefix + selectedDot + " None"] ], map tool -> Box(marginLeft=2)[ Text(active?cyan)[prefix + chosenDot + " " + tool] ] ]`

43. `src/ui/components/InputPatternHighlightersView.tsx:177-223`
   - UI: Highlighter list rows that combine name, regex summary, optional styling summary, and enabled/disabled status color.
   - Ink / hooks: `Box`, `Text`, `useInput`, `useState`, `useContext`
   - Tree: `Box(col)[ map highlighter -> Box(row)[ Text(active ? yellow)[prefix + enabledDot + name + " "], Text(dim)["("], regex ? Text(dim)["/" + regex + "/" + flags] : Text(dim)["(no regex)"], styling.length ? Text(dim or lineColor)[" · " + styling.join(", ")] : null, Text(dim)[")"], enabled ? Text(color=success)[" enabled"] : Text(color=error)[" disabled"] ] ]`

44. `src/ui/components/InputPatternHighlighterEditView.tsx:436-456`
   - UI: Regex field with a round border that turns red on invalid regex and shows the validation error directly beneath.
   - Ink / hooks: `Box`, `Text`, `useEffect`, `useInput`, `useState`, `useContext`
   - Tree: `Box(col,width=45%)[ Text(selected ? yellow+bold)[prefix + "Regex Pattern"], Box(marginLeft=2,col)[ Box(border=round,borderColor=selected ? (editing?green:yellow) : regexError?red:gray)[ Text[regex or "(empty)"] ], regexError ? Text(red)[errorMessage] : null ] ]`

45. `src/ui/components/InputPatternHighlighterEditView.tsx:480-527`
   - UI: Side-by-side format-string and styling controls, with the formatting field on the left and a vertically stacked styling checklist on the right.
   - Ink / hooks: `Box`, `Text`
   - Tree: `Box(row,gap=2)[ Box(col,width=50%)[ Text(selected ? yellow+bold)[prefix + "Format String"], Box(marginLeft=2)[ Box(border=round,borderColor=selected ? editingState : gray)[ Text[format] ] ], selected ? Text(dim)["use {MATCH} as placeholder"] : null ], Box(col,width=45%)[ Text(selected ? yellow+bold)[prefix + "Styling"], Box(marginLeft=2,col)[ map option -> Text(highlighted ? cyan)[prefix + activeDot + label] ] ] ]`

46. `src/ui/components/InputPatternHighlighterEditView.tsx:530-570`
   - UI: Foreground/background mode selectors that show `[dot] none` vs `[dot] custom`, plus an inline swatch box when custom is active.
   - Ink / hooks: `Box`, `Text`
   - Tree: `Box(row,gap=2)[ Box(col,width=48%)[ Text(selected ? yellow+bold)[prefix + "Foreground Color"], Box(marginLeft=2,row,gap=1)[ Text(selected?yellow)["[dot] none  [dot] custom"], foregroundMode==custom ? Box(border=round,borderColor=selected?yellow:gray)[ Text(color=foreground)[" swatch "] ] : null ] ], Box(col,width=48%)[ same structure for background, but swatch uses backgroundColor ] ]`

47. `src/ui/components/InputPatternHighlighterEditView.tsx:604-628`
   - UI: Live preview panel that conditionally shows either a dim placeholder prompt or a highlighted rendering of the test text.
   - Ink / hooks: `Box`, `Text`, `HighlightedTestText`
   - Tree: `Box(border=round,padding=1)[ Box(col)[ Text(bold)["Live Preview:"], regex ? Box(marginTop=1)[ HighlightedTestText(...) ] : Text(dim)["Enter a regex pattern to see the preview"] ] ]`

48. `src/ui/components/ClaudeMdAltNamesView.tsx:147-213`
   - UI: Scroll-windowed alternative-filenames list with optional inline add-new row and inline editing row.
   - Ink / hooks: `Box`, `Text`, `useInput`, `useState`, `useContext`
   - Tree: `Box(col)[ altNames.length==0 ? Text["No alternative names configured..."] : Fragment[ maybe Text(dim)["up more above"], map visibleName -> Text(active ? cyan)[prefix + name], maybe Text(dim)["down more below"] ], addingNew ? Box(row,alignItems=center)[ Text(yellow)[prefix], Box(border=round,yellow)[Text[input]] ] : null, editing ? Box(row,marginTop=1,alignItems=center)[ Text["Editing: "], Box(border=round,yellow)[Text[input]] ] : null ]`

49. `src/ui/components/SubagentModelsView.tsx:126-165`
   - UI: Summary list of plan/explore/general-purpose agents, each rendered as a stacked block with title, dim description, and current model label.
   - Ink / hooks: `Box`, `Text`, `useInput`, `useState`, `useContext`
   - Tree: `Box(col)[ Header, helpText, map subagent -> Box(col,marginBottom=1)[ Box[ Text(active ? cyan)[prefix + Text(bold)[title]] ], Box(marginLeft=4)[ Text(dim)[description] ], Box(marginLeft=4)[ Text["Current: " + Text(green)[modelLabel]] ] ] ]`

50. `src/ui/components/SubagentModelsView.tsx:101-123`
   - UI: Full takeover model-picker overlay for one subagent, with a selectable list of label/value rows.
   - Ink / hooks: `Box`, `Text`, `useInput`, `useState`
   - Tree: `Box(col)[ Box(marginBottom=1)[ Header["Select Model for " + activeSubagentTitle] ], map option -> Box[ Text(active ? cyan)[prefix + option.label + maybe Text(dim)[" (" + value + ")"]] ] ]`
