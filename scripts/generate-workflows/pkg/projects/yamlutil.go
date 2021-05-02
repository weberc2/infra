package projects

import "gopkg.in/yaml.v3"

func scalar(s string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Value: s}
}

func list(v ...*yaml.Node) *yaml.Node {
	return &yaml.Node{Kind: yaml.SequenceNode, Content: v, Style: yaml.FlowStyle}
}

type field struct {
	key   string
	value *yaml.Node
}

func mapping(fields ...field) *yaml.Node {
	nodes := make([]*yaml.Node, 2*len(fields))
	for i, field := range fields {
		nodes[2*i] = scalar(field.key)
		nodes[2*i+1] = field.value
	}
	return &yaml.Node{Kind: yaml.MappingNode, Content: nodes}
}
