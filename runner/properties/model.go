package properties

import (
	"fmt"
	"io"
	"strings"
)

type Model struct {
	nodes []node
}

func (p *Model) Properties() []*Property {
	properties := make([]*Property, 0)
	for _, n := range p.nodes {
		if p, ok := n.(*Property); ok {
			properties = append(properties, p)
		}
	}
	return properties
}

func (p *Model) Get(key string) (string, bool) {
	for _, n := range p.nodes {
		if p, ok := n.(*Property); ok && p.Key == key {
			return p.Value, true
		}
	}
	return "", false
}

func (p *Model) Set(key, value string) {
	for _, n := range p.nodes {
		if p, ok := n.(*Property); ok && p.Key == key {
			p.Value = value
			return
		}
	}
	p.nodes = append(p.nodes, &Property{Key: key, Value: value})
}

func (p *Model) Append(key, value string) {
	for _, n := range p.nodes {
		if p, ok := n.(*Property); ok && p.Key == key {
			p.Value += value
			return
		}
	}
	p.nodes = append(p.nodes, &Property{Key: key, Value: value})
}

func (p *Model) Write(writer io.Writer) error {
	for _, n := range p.nodes {
		switch n := n.(type) {
		case *comment:
			_, err := fmt.Fprintf(writer, "%s\n", n.text)
			if err != nil {
				return err
			}
		case *Property:
			_, err := fmt.Fprintf(writer, "%s=%s\n", n.Key, n.Value)
			if err != nil {
				return err
			}
		case *whitespace:
			_, err := fmt.Fprintln(writer)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unexpected node type %T", n)
		}
	}
	return nil
}

func (p *Model) String() string {
	var builder strings.Builder
	_ = p.Write(&builder)
	return builder.String()
}

type node interface {
	node()
}

type comment struct {
	text string
}

func (comment) node() {}

type Property struct {
	Key   string
	Value string
}

func (Property) node() {}

type whitespace struct {
}

func (whitespace) node() {}

func Load(data []byte) (*Model, error) {
	text := string(data)
	lines := strings.Split(text, "\n")

	nodes := make([]node, 0)

	i := 0
	for i < len(lines) {
		line := strings.ReplaceAll(lines[i], "\r", "")

		if len(line) == 0 || strings.Count(line, " ")+strings.Count(line, "\t") == len(line) {
			nodes = append(nodes, &whitespace{})
			i++
			continue
		}

		if line[0] == '#' {
			nodes = append(nodes, &comment{text: line})
			i++
			continue
		}

		if !strings.Contains(line, "=") {
			return nil, fmt.Errorf("expected key=value pair in line %d", i+1)
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("expected Key=Value pair in line %d", i+1)
		}

		val := parts[1]

		// Multiline Value
		for strings.HasSuffix(val, "\\") {
			val = val[:len(val)-1]
			if i >= len(lines)-1 {
				return nil, fmt.Errorf("unexpected end of file in multiline Value in line %d", i+1)
			}
			i++
			val += strings.ReplaceAll(lines[i], "\r", "")
		}

		nodes = append(nodes, &Property{Key: parts[0], Value: val})
		i++
	}

	return &Model{
		nodes: nodes,
	}, nil
}
