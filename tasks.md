## WIP

## TODO
11. Add more tree operatrions
12. Separate tree view from window view
13. Split into subpackages
16. Make erasing after insert continuous (single modification, signle undo)
17. Make lines calculations incremental
21. Add frame moving operations
22. Fix status line style
23. Add cursor movement in insert
24. Make insert forward delete operation continuous
25. 
## VALIDATION
4. Fix text_offset being off by 1 when going from node overflowing a line to a node at the beginning of the line
## DONE
20. Add clipboard support
10. Add more treesitter parsers
3. Add count operation to tree and visual modes
1. In tree mode make window shift to show as much of the node as possible. Example: when go couple of nodes forward and screen moves down only the first line of the function show instead of the whole function or its top part
15. Rewrite undo tree/list as separate from buffer
19. Update cursor node after undo/redo
5. Fix not being able to go the end of the line when in selection mode or moving horizontali at all if selecting over the end of the line.
6. Fix single char selection not being drawn
7. Fix selection to the right of window frame beging processed.
9. Fix window node not being updated on edits
2. Improve buffer.Lines() to not recalculate every time (recalculates after change)
15. Rewriting undo system
8. Migrate to official golang treesitter package
14. Add word end operator



# PRESENTATION

1. Nawigacja linie i znaki: jkhlwbeE
2. Nawigacja słowami
3. Nawigacja do początku linii, początku tekstu linii, końca linii
3. Usuwanie znaków
4. Usuwanie linii
5. Pasek statusu
6. Przesuwanie pół ekranu góra dół
7. Wyśrodkowanie linii
8. Przesuwanie ekranu linia góra dół
6. Wprowadzanie tekstu
7. Usuwanie tekstu w trakcie wprowadzania, usuwanie ostatniego słowa, usuwanie z przodu
8. Cofanie i ponowienie zmian
9. Wprowdzenie z usunięciem
9. Powtarzanie zmian n razy
10. Iść do linii, do ostatniej linii
11. Zaznaczanie tekstu
12. Kopiowanie tekstu
13. Wklejanie tekstu
14. Usuwanie tekstu
15. Od zaznaczania do wprowadzenia z usunięciem
16. Tryb drzewa
17. węzeł w górę/ dół, prawo/lewo
18. węzeł prawo/lewo do końca rodzeństwa
20. Do końca rodzeńctwa
21. do początku rodzeństwa
22. Usunięcie węzła
23. Zamiana miejscami z nastepnym i n nastepnym
24. Inne języki programowania
25. Ekran powitania


Tryb do zmiany plików i innych operacji?


