package properties

import (
	"testing"
)

func TestLoadComment(t *testing.T) {
	data :=
		`
# This is a comment
`
	model, err := Load([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if len(model.nodes) != 3 {
		t.Error("expected 3 nodes, got", len(model.nodes))
	}
	if _, ok := model.nodes[1].(*comment); !ok {
		t.Error("expected comment node")
	}
	comment := model.nodes[1].(*comment)
	if comment.text != "# This is a comment" {
		t.Error("expected '# This is a comment', got", comment.text)
	}
}

func TestLoadProperty(t *testing.T) {
	data :=
		`
hello=world
abc=def
`
	model, err := Load([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if len(model.nodes) != 4 {
		t.Error("expected 4 nodes, got", len(model.nodes))
	}
	expectProperty(t, model.nodes[1], "hello", "world")
	expectProperty(t, model.nodes[2], "abc", "def")
}

func TestLoad(t *testing.T) {
	data :=
		`
hello=world \
this is a test \
of multiline properties
test=abc\
def
# this is a comment
another=Property\
hello
`
	model, err := Load([]byte(data))
	if err != nil {
		t.Error(err)
	}

	if len(model.nodes) != 6 {
		t.Error("expected 6 nodes, got", len(model.nodes))
	}

	expectProperty(t, model.nodes[1], "hello", "world this is a test of multiline properties")
	expectProperty(t, model.nodes[2], "test", "abcdef")
	expectProperty(t, model.nodes[4], "another", "Propertyhello")
}

func TestModel_Write(t *testing.T) {
	data :=
		`
# Comment
hello=world
abc=def
text=hello \
world
`

	model, err := Load([]byte(data))
	if err != nil {
		t.Error(err)
	}

	expected := "\n# Comment\nhello=world\nabc=def\ntext=hello world\n\n"
	if model.String() != expected {
		t.Error("expected", expected, "got", model.String())
	}
}

func expectProperty(t *testing.T, n node, key, value string) {
	if _, ok := n.(*Property); !ok {
		t.Error("expected Property node")
	}
	prop := n.(*Property)
	if prop.Key != key {
		t.Error("expected", key, "got", prop.Key)
	}
	if prop.Value != value {
		t.Error("expected", value, "got", prop.Value)
	}
}

func TestModel_Set(t *testing.T) {
	data :=
		`
hello=world
`
	model, err := Load([]byte(data))
	if err != nil {
		t.Error(err)
	}
	model.Set("hello", "world2")
	if val, _ := model.Get("hello"); val != "world2" {
		t.Error("expected 'world2', got", model.nodes[1].(Property).Value)
	}
}
