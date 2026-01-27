## TODO's

### Lexer
- Signed numbers
- Floats
- Alternative number representation
- Template strings
- Chars


### Parsing
- Control flow
    - [x] if
    - [x] if else
    - [ ] switch
    - [ ] match
- Loops
    - [x] loop (forever) 
    - [x] for with condition
    - [ ] for i = 0; ...
    - [ ] for i in 0..10
    - [ ] 
- Statements
    - [ ] re-assign variable
    - [ ] asserts

- Elements
- Visibility
- Modules


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
