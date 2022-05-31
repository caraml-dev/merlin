Standard Transformer UX Improvement:

- [x] UI: Validate simulate input, especially when parsing JSON input
- [x] Rebase simulation branch to pull aria's change
- [x] Hide the old transformation graph
- [x] Change node trace's spec detail from JSON to YAML
- [x] Update node trace's spec detail panel:
  - [x] Use accordion
  - [x] Reorder: Input, Spec, Output
  - [x] If null, do not display
- [x] Fix creating output using raw_request and model_response
- [x] Pin sidebar location of standard transformer config
- [x] Form row direction from column to row to save more space

Stretch goals:

- [x] If output is table type, display it as table. use tabular to display JSON raw too
- [x] If input is table type, display it as table. use tabular to display JSON raw too
- Next/prev across node details modal
