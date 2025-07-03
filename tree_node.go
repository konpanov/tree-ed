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
