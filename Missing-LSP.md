# Missing stuff from LSP


- Diagnostics
- BUILD structs 
- BUILD modules
- functions like heph.is_target requires modules
- Live heph server -> Add this as another ticket


- Use custom parser
// TODO: Why don't use starlark parser?
// e.g. syntax.Parse(filename string, src interface{}, mode syntax.Mode)
// - We would need a parser that parses blocks of code
// - We would need in-memory parse
// - We won't have custom queries
// - Parsers generate AST not an CST
// - In CST I can look direct into nodes an query specified node in position offset

## Future ideas

- Could use a DAG later
- Implement a trie to add symbols

