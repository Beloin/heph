package query

// TODO: bsena; Create targets here, or extract the data?
// Look for the already existing target extraction from other heph packages
// example: graph.Target


const targetQuery = `
(function_definition
  name: (identifier) @function.name
  parameters: (parameters) @function.params
  body: (block .
     (expression_statement
      (string (string_content) )) @function.docstring)?)
`


func QueryTarget() {
}
