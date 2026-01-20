# ADR Close処理の責任所在設計

# Status
- Draft: 記述中またはレビュー中

# Decision
read/writeのタイムアウトのためのTouchはctrlチャネルを通じて送らず、、直接sessionを操作することとする。

# Context
ctrlチャネルを通じてTouchを送る設計にした場合、
