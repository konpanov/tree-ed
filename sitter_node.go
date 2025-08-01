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

func NextSiblingOrCousinDepth(node *sitter.Node, depth int) *sitter.Node {
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
	ancle := parent.NextSibling()
	if ancle == nil {
		ancle = NextSiblingOrCousin(parent)
		if ancle == nil {
			return nil
		}
	}
	if ancle.ChildCount() == 0 {
		return ancle
	} else {
		return ancle.Child(0)
	}
}

func PrevSiblingOrCousinDepth(node *sitter.Node, depth int) *sitter.Node {
	prev := PrevSiblingOrCousin(node)
	for prev != nil && prev.ChildCount() != 0 && Depth(prev) < depth {
		prev = prev.Child(prev.ChildCount() - 1)
	}
	return prev
}

func NextAncle(node *sitter.Node) *sitter.Node {
	if node == nil {
		return nil
	}
	parent := node.Parent()
	if parent == nil {
		return nil
	}
	ancle := parent.NextSibling()
	if ancle == nil {
		ancle = NextAncle(parent)
	}
	return ancle
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
