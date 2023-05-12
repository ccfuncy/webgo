package gee

import "strings"

type node struct {
	pattern  string  //请求路由 /path/user/name
	part     string  //请求路径的一部分 user
	children []*node //子节点
	isWiled  bool    //是否精确匹配
}

//返回当前层的匹配的第一个节点
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWiled {
			return child
		}
	}
	return nil
}

//返回当前层所有匹配的节点
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWiled {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

func (n *node) insert(pattern string, parts []string, height int) {
	if len(parts) == height {
		n.pattern = pattern
		return
	}
	part := parts[height]
	child := n.matchChild(part)
	if child == nil {
		child = &node{
			part:    part,
			isWiled: part[0] == ':' || part[0] == '*',
		}
		n.children = append(n.children, child)
	}
	child.insert(pattern, parts, height+1)
}

func (n *node) search(parts []string, height int) *node {
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.pattern == "" {
			return nil
		}
		return n
	}
	part := parts[height]
	children := n.matchChildren(part)
	for _, child := range children {
		n := child.search(parts, height+1)
		if n != nil {
			return n
		}
	}
	return nil
}
