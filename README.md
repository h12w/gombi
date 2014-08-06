Gombi: Creating Your Own Parser is Easier than Regular Expressions
==================================================================

Gombi is a Go lexer/parser library that is inspired by but not limited to parser
combinators. It is practical, reasonably fast and extremely easy to use.

Unlike lex/yacc, Gombi uses an internal DSL (in ordinary Go code) to describe
the language syntax and construct the lexer/parser at runtime, just like regular
expressions or parser combinators.
