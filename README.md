Gombi: Creating Your Own Parser is Easier than Regular Expressions
==================================================================

Gombi is a combinator-style scanner & parser library written in Go. It is
practical, reasonably fast and extremely easy to use.

Quick start
-----------
```bash
go get -u h12.io/gombi
```

Design
------

[Combinator parsers](http://en.wikipedia.org/wiki/Parser_combinator) are
straightforward to construct, modular and easily maintainable, compared to
parser generators like Lex/Yacc.

* Internal DSL
    * no additional code generation and compilation.
* Composable
    * a subset of the syntax tree is also a parser.
    * a language can be easily embedded into another one.

Gombi is inspired by but not limited to parser combinators. Unlike a combinator
parser, Gombi neither limits its API to functional only, nor limits its
implementation to functional combinators. Go is not a pure functional language
as Haskell, so cloning a combinator parser like Parsec to Go will only lead to
an implementaion much worse than Parsec. Instead, Gombi is free to choose any Go
language structures that are suitable for a modular and convenient API, and any
algorithms that can be efficiently implemented in Go.

*Under development*
