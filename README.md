# Lambda Calculus Implementations in Go

- `master` branch: simply-typed lambda calculus, one base type named `o`
- `untyped` branch: untyped lambda calculus with global assignment statements
  - there are some example expressions in the `examples/` directory of this
    branch that you can pipe into the program through standard input

Note that the parsers I've written require a lot of parentheses.

## Other features I might add

* show steps
* add more reductions
* use de Bruijn indices
* data encodings (there are some Church encodings in the `untyped` branch)
* global names (implemented in `untyped` branch)
* builtins (implemented in `untyped` branch)
* use an iterative algorithm for reductions
* repl (started work in `untyped` branch)
* process calculi???

I should really combine the code in the two branches into one program.
