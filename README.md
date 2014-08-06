Gombi: Creating Your Own Parser is Easier than Regular Expressions
==================================================================

Gombi is a Go lexer/parser library that is inspired by but not limited to parser
combinators. It is practical, reasonably fast and extremely easy to use.

Unlike lex/yacc, Gombi uses an internal DSL (in ordinary Go code) to describe
the language syntax and construct the lexer/parser at runtime, just like parser
combinators.

Unlike a combinator parser, Gombi neither limits its API to functional only, nor
limits its implementation to functional combinators. Go is not a pure functional
language as Haskell, so cloning a combinator parser like Parsec to Go will only
lead to an implementaion much worse than Parsec. Instead, Gombi should choose
any Go language structures that are suitable for a modular and convenient API,
and any algorithms that can be efficiently implemented with Go.

*Under development*
