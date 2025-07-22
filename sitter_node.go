package main

import (
	"log"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

func Depth(node *sitter.Node) int {
	depth := 0
	for node != nil && node.Parent() != nil {
		depth++
		node = node.Parent()
	}
	return depth
}

func NodeLeaf(node *sitter.Node, index int) *sitter.Node {
	if node == nil {
		log.Panicln("Cannot find leaf from nil node")
	}
	if node.ChildCount() == 0 {
		return node
	}
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		if int(child.StartByte()) <= index && int(child.EndByte()) > index {
			return NodeLeaf(child, index)
		}

	}
	return node
}

func MinimalNode(node *sitter.Node, a uint, b uint) *sitter.Node {
	if node.StartByte() <= a && node.EndByte() >= b {
		for i := range node.ChildCount() {
			if node := MinimalNode(node.Child(i), a, b); node != nil {
				return node
			}
		}
		return node
	}
	return nil
}

func MinimalNodeDepth(node *sitter.Node, a uint, b uint, depth int) *sitter.Node {
	node = MinimalNode(node, a, b)
	for Depth(node) > depth && NodeMatch(node.Parent(), a, b) {
		node = node.Parent()
	}
	return node
}

func NextCousinDepth(node *sitter.Node, depth int) *sitter.Node {
	cousin := NextCousin(node)
	for cousin != nil && cousin.ChildCount() != 0 && Depth(cousin) < depth {
		cousin = cousin.Child(0)
	}
	return cousin
}

func NextCousin(node *sitter.Node) *sitter.Node {
	sibling := node.NextSibling()
	if sibling != nil {
		return sibling
	}
	parent := node.Parent()
	if parent == nil {
		log.Println("No parent")
		return nil
	}
	ancle := parent.NextSibling()
	if ancle == nil {
		log.Println("No ancle")
		ancle = NextCousin(parent)
		if ancle == nil {
			return nil
		}
	}
	if ancle.ChildCount() == 0 {
		log.Println("No cousins")
		return ancle
	} else {
		log.Println("Cousin found")
		return ancle.Child(0)
	}
}

func PrevCousinDepth(node *sitter.Node, depth int) *sitter.Node {
	cousin := PrevCousin(node)
	for cousin != nil && cousin.ChildCount() != 0 && Depth(cousin) < depth {
		cousin = cousin.Child(cousin.ChildCount() - 1)
	}
	return cousin
}

func PrevCousin(node *sitter.Node) *sitter.Node {
	sibling := node.PrevSibling()
	if sibling != nil {
		return sibling
	}
	parent := node.Parent()
	if parent == nil {
		log.Println("No parent")
		return nil
	}
	ancle := parent.PrevSibling()
	if ancle == nil {
		log.Println("No ancle")
		ancle = PrevCousin(parent)
		if ancle == nil {
			return nil
		}
	}
	if ancle.ChildCount() == 0 {
		log.Println("No cousins")
		return ancle
	} else {
		log.Println("Cousin found")
		return ancle.Child(ancle.ChildCount() - 1)
	}
}

func FirstSibling(node *sitter.Node) *sitter.Node {
	if node == nil {
		return node
	}
	parent := node.Parent()
	if parent == nil {
		return node
	}
	return parent.Child(0)
}

func LastSibling(node *sitter.Node) *sitter.Node {
	if node == nil {
		return node
	}
	parent := node.Parent()
	if parent == nil {
		return node
	}
	return parent.Child(parent.ChildCount() - 1)
}

func NodeMatch(node *sitter.Node, start uint, end uint) bool {
	return node != nil && node.StartByte() == start && node.EndByte() == end
}
