package gofaster

import "strings"

type treeNode struct {
	name       string
	childrens  []*treeNode
	routerName string
}

// put path: /user/get/:id
func (t *treeNode) Put(path string) {
	tmp := t
	spilt := strings.Split(path, "/")
	for index, name := range spilt {
		if index == 0 {
			continue
		}
		children := tmp.childrens
		isMatch := false
		for _, node := range children {
			if node.name == name {
				tmp = node
				isMatch = true
				break
			}
		}
		if !isMatch {
			node := &treeNode{
				name:      name,
				childrens: make([]*treeNode, 0),
			}
			tmp.childrens = append(tmp.childrens, node)
			tmp = node
		}
	}
}

// get path:/usr/get/1
func (t *treeNode) Get(name string) *treeNode {
	tmp := t
	routerName := ""
	split := strings.Split(name, "/")
	for index, name := range split {
		if index == 0 {
			continue
		}
		isMatch := false
		children := tmp.childrens
		for _, node := range children {
			if node.name == name ||
				node.name == "*" ||
				strings.Contains(node.name, ":") {
				tmp = node
				routerName += "/" + node.name
				node.routerName = routerName
				isMatch = true
				if index == len(split)-1 {
					return node
				}
				break
			}
		}
		if !isMatch {
			for _, node := range children {
				routerName += "/" + node.name
				node.routerName = routerName
				if node.name == "**" {
					return node
				}
			}
		}
	}

	return nil
}
