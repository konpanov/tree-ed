package main

import (
	"log"

	sitter "github.com/smacker/go-tree-sitter"
)

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
		return ancle.Child(0)
	}
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

func Depth(node *sitter.Node) int {
	depth := 0
	for node.Parent() != nil {
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
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if int(child.StartByte()) <= index && int(child.EndByte()) > index {
			return NodeLeaf(child, index)
		}

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

func PrevCousinDepth(node *sitter.Node, depth int) *sitter.Node {
	cousin := PrevCousin(node)
	for cousin != nil && cousin.ChildCount() != 0 && Depth(cousin) < depth {
		cousin = cousin.Child(int(cousin.ChildCount()) - 1)
	}
	return cousin
}
