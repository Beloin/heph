package runtime

import (
	// "github.com/hephbuild/heph/utils/queue"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

// NodeTraversalFunc defines a function to be called for each node in the tree traversal. It should return true to continue traversing or false to stop.
type NodeTraversalFunc = func(*tree_sitter.Node) bool

func Traverse(tree *tree_sitter.Tree, nodeFunc NodeTraversalFunc) {
	cu := tree.Walk()

	// Example:
	//           []
	//        [] [] []
	//       /       \
	//    [] []       [] []
	//   /    \           \
	// []      []          []
	//
	backToParent := false
	for {
		if backToParent {

			if cu.GotoNextSibling() {
				backToParent = false
				continue
			}

			if cu.GotoParent() {
				continue
			}

			break
		}

		child := cu.Node()
		// Callback
		if !nodeFunc(child) {
			break
		}

		// If change orders of First and Next changes to BFS
		if cu.GotoFirstChild() {
			continue
		}

		if cu.GotoNextSibling() {
			continue
		}

		if cu.GotoParent() {
			backToParent = true
		} else {
			// error?
			// break
		}

	}

	// TODO: bsena; check if we can use this implementation since is easier and simpler

	// Using a queue
	// currNode := cu.Node()
	// q := queue.Queue[*tree_sitter.Node]{Max: 100}
	// q.Enqueue(currNode)
	// for q.Len() > 0 {
	// 	v := q.Dequeue(1)[0]
	//
	// 	// Callback
	// 	if !nodeFunc(v) {
	// 		break
	// 	}
	//
	// 	cu := v.Walk()
	// 	for _, n := range v.Children(cu) {
	// 		q.Enqueue(&n)
	// 	}
	// }
}
