## WIP
15. Rewriting undo system
14. Add word end operator
## TODO
1. In tree mode make window shift to show as much of the node as possible. Example: when go couple of nodes forward and screen moves down only the first line of the function show instead of the whole function or its top part
3. Add count operation to tree and visual modes
8. Migrate to official golang treesitter package
10. Add more treesitter parsers
11. Add more tree operatrions
12. Separate tree view from window view
13. Split into subpackages
15. Rewrite undo tree/list as separate from buffer
16. Make erasing after insert continuous (single modification, signle undo)
17. Make lines calculations incremental
18. Find reason why tree movement blocks on some deletions (for example deleting new lines after import block in go)
19. Update cursor node after undo/redo
## VALIDATION
4. Fix text_offset being off by 1 when going from node overflowing a line to a node at the beginning of the line
## DONE
5. Fix not being able to go the end of the line when in selection mode or moving horizontali at all if selecting over the end of the line.
6. Fix single char selection not being drawn
7. Fix selection to the right of window frame beging processed.
9. Fix window node not being updated on edits
2. Improve buffer.Lines() to not recalculate every time (recalculates after change)
