package query

import tree_sitter "github.com/tree-sitter/go-tree-sitter"

func ExtractCurrentSymbol(root *tree_sitter.Node, source []byte, byteOffSet uint) string {
	for root != nil && root.Kind() != "identifier" {
		root = root.FirstChildForByte(byteOffSet)
	}

	if root == nil {
		return ""
	}

	return root.Utf8Text(source)
}

func ExtractCurrentStringLiteral(root *tree_sitter.Node, source []byte, byteOffSet uint) string {
	for root != nil && root.Kind() != "string_content" {
		root = root.FirstChildForByte(byteOffSet)
	}

	if root == nil {
		return ""
	}

	return root.Utf8Text(source)
}

func ExtractFunctionNameFromOffset(root *tree_sitter.Node, source []byte, byteOffSet uint) string {
	for root != nil && root.Kind() != "function_definition" {
		root = root.FirstChildForByte(byteOffSet)
	}

	if root == nil {
		return ""
	}

	root = root.ChildByFieldName("name")

	if root == nil {
		return ""
	}

	return root.Utf8Text(source)
}

