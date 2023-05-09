package gofaster

import "testing"

func TestTreeNode(t *testing.T) {
	root := &treeNode{
		name:      "/",
		childrens: make([]*treeNode, 0),
	}
	root.Put("/user/add/:id")
	root.Put("/user/create/hello")
	root.Put("/user/create/aaa")
	root.Put("/order/get/*/test")

	println(root.Get("/user/add/1").name)
	println(root.Get("/user/create/hello").name)
	println(root.Get("/user/create/aaa").name)
	println(root.Get("/order/get/aaa/test").name)

}
