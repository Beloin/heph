package runtime

import (
	protocol "github.com/tliron/glsp/protocol_3_16"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

type docTuple struct {
	*Document

	// Last reported version
	version protocol.Integer
}

type Manager struct {
	DocumentMap map[protocol.DocumentUri]*docTuple
	Parser      *tree_sitter.Parser
}

func NewManager(parser *tree_sitter.Parser) *Manager {
	return &Manager{DocumentMap: map[protocol.DocumentUri]*docTuple{}, Parser: parser}
}

// GetDocument queries and look for an existing document in Manager.
// returns nil if not present
func (m *Manager) GetDocument(uri protocol.DocumentUri) *Document {
	if tuple, found := m.DocumentMap[uri]; found {
		return tuple.Document
	}

	return nil
}

func (m *Manager) SetDocument(uri protocol.DocumentUri, version protocol.Integer, doc *Document) {
	if tuple, found := m.DocumentMap[uri]; found {
		tuple.Document = doc
		tuple.version = version
	} else {
		m.DocumentMap[uri] = &docTuple{
			Document: doc,
			version:  version,
		}
	}
}
