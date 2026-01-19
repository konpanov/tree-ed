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
	if !NodeContains(node, a, b) {
		return nil
	}
	for searching := true; searching; {
		searching = false
		for i := range node.ChildCount() {
			if NodeContains(node.Child(i), a, b) {
				node = node.Child(i)
				searching = true
				break
			}
		}
	}
	return node
}

func NodeContains(node *sitter.Node, a uint, b uint) bool {
	return node != nil && node.StartByte() <= a && b <= node.EndByte()
}

func MinimalNodeDepth(node *sitter.Node, a uint, b uint, depth int) *sitter.Node {
	node = MinimalNode(node, a, b)
	for Depth(node) > depth && NodeMatch(node.Parent(), a, b) {
		node = node.Parent()
	}
	return node
}

func NextSiblingOrCousinDepth(node *sitter.Node, depth int) *sitter.Node {
	for node_depth := Depth(node); node_depth > depth; node_depth-- {
		node = node.Parent()
	}
	cousin := NextSiblingOrCousin(node)
	for cousin != nil && cousin.ChildCount() != 0 && Depth(cousin) < depth {
		cousin = cousin.Child(0)
	}
	return cousin
}

func NextSiblingOrCousin(node *sitter.Node) *sitter.Node {
	sibling := node.NextSibling()
	if sibling != nil {
		return sibling
	}
	parent := node.Parent()
	if parent == nil {
		return nil
	}
	uncle := parent.NextSibling()
	if uncle == nil {
		uncle = NextSiblingOrCousin(parent)
		if uncle == nil {
			return nil
		}
	}
	if uncle.ChildCount() == 0 {
		return uncle
	} else {
		return uncle.Child(0)
	}
}

func PrevSiblingOrCousinDepth(node *sitter.Node, depth int) *sitter.Node {
	for node_depth := Depth(node); node_depth > depth; node_depth-- {
		node = node.Parent()
	}
	prev := PrevSiblingOrCousin(node)
	for prev != nil && prev.ChildCount() != 0 && Depth(prev) < depth {
		prev = prev.Child(prev.ChildCount() - 1)
	}
	return prev
}

func PrevSiblingOrCousin(node *sitter.Node) *sitter.Node {
	sibling := node.PrevSibling()
	if sibling != nil {
		return sibling
	}
	parent := node.Parent()
	if parent == nil {
		return nil
	}
	ancle := parent.PrevSibling()
	if ancle == nil {
		ancle = PrevSiblingOrCousin(parent)
		if ancle == nil {
			return nil
		}
	}
	if ancle.ChildCount() == 0 {
		return ancle
	} else {
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
