## TODO's


### Lexer
- Signed numbers
- Floats
- Alternative number representation (binary, octal, etc)
- Template strings
- Chars


### Parsing
- [ ] Comments
- [ ] New lines
- [ ] Control flow
    - [x] if
    - [x] if else
    - [ ] switch/match
- [ ] Loops
    - [x] loop (forever) 
    - [x] for with condition
    - [ ] for i = 0; ...
    - [ ] for i in 0..10
- [ ] Statements
    - [ ] explicit type declarations on variable assignments
    - [ ] re-assign variable
    - [ ] assert
- [ ] Bitwise operators
- [ ] Call Expressions
    - [ ] labeled arguments
    - [ ] punned labeled arguments
- [ ] Literals
    - [ ] Composite Literals (e.g. slices)
    - [ ] Struct Literals
- [ ] Elements
- [ ] Visibility
- [ ] Modules
    - [ ] Document symbols map
    - [ ] Files vs packages


### Compiler

- [x] Unary expressions
- [x] Binary expressions
- [x] Identifiers
- [x] Integer literals
- [x] Functions
- [x] Block statements
- [x] Return statements
...
- [x] Resolve how to implement xml/element expressions


### Syntax Highlighter


## Future
- Type checker
- LSP 
- Linter
- Formatter
- Syntax Highlight (tree-sitter)


## Considerations

Support pattern matching vs exhaustive switches?    
    -> Optional return statements?

Investigate compiling to llvm ir and working towards a "real" language

How should union members be lifted to higher scopes, e.g. in rust:

```rust
enum Option {
    Some(T)
    None,
}

use Option::*;

let foo = Some(1);
```
