package main

import (
	"context"
	"fmt"
	"io"
	// "os"
	"strings"
)

type Node func(r Renderer)

type Attribute struct {
	Key   string
	Value string
}

type Attributes []Attribute

type Renderer interface {
	Element(tag string, attrs Attributes, children ...Node)
	Text(content string)
}

type html struct {
	Context context.Context
	Writer  io.Writer
}

func HtmlWriter(ctx context.Context, w io.Writer, rootNode Node) {
	rootNode(&html{
		Context: ctx,
		Writer:  w,
	})
}

func HtmlString(ctx context.Context, rootNode Node) string {
	var w strings.Builder

	rootNode(&html{
		Context: ctx,
		Writer:  &w,
	})

	return w.String()
}

func (r *html) Element(tag string, attrs Attributes, children ...Node) {
	fmt.Fprintf(r.Writer, "<%s", tag)
	for _, a := range attrs {
		fmt.Fprintf(r.Writer, " %s", a.Key)
		// TODO: check for props with no value (e.g. disabled)
		fmt.Fprintf(r.Writer, `="`)
		fmt.Fprintf(r.Writer, "%s", a.Value)
		fmt.Fprintf(r.Writer, `"`)
	}
	fmt.Fprintf(r.Writer, ">")

	for _, child := range children {
		if child != nil {
			child(r)
		}
	}

	fmt.Fprintf(r.Writer, "</%s>", tag)
}

func (r *html) Text(content string) {
	io.WriteString(r.Writer, content)
}

func h(tag string, attrs Attributes, children ...Node) Node {
	return func(r Renderer) {
		r.Element(tag, attrs, children...)
	}
}

func Div(attrs Attributes, children ...Node) Node {
	return h("div", attrs, children...)
}

func Text(content string) Node {
	return func(r Renderer) {
		r.Text(content)
	}
}

func Fragment(children ...Node) Node {
	return func(r Renderer) {
		for _, child := range children {
			if child != nil {
				child(r)
			}
		}
	}
}

func If(cond bool, then Node, otherwise Node) Node {
	return func(r Renderer) {
		if cond {
			if then != nil {
				then(r)
			}
		} else {
			if otherwise != nil {
				otherwise(r)
			}
		}
	}
}

// struct HtmlAttributes {
// 	id: string;
// }
//
// struct DivAttributes : HtmlAttributes {
// }

// extern fn div(props: DivAttributes) Element

// app.gloss
//
//	fn App() Element {
//		let time = date.now()
//
//		return <div id="app">
//			{if time > 0 {
//				Current time: {time}
//			}}
//		</div>
//	}
//
// app.go

type Time interface {
	Now() int
}

type Runtime interface {
	Time() Time
}

type State struct {
	global Runtime
}

func NewState(runtime Runtime) *State {
	return &State{global: runtime}
}

func App(s *State) Node {
	time := s.global.Time().Now()

	return Div(Attributes{
		Attribute{Key: "id", Value: "app"},
	},
		If(time > 0, Fragment(
			Text("Time is: "),
			Text(fmt.Sprintf("%d", time)),
		), nil),
	)
}

// Implementation of the runtime
type time struct{}

func (t time) Now() int { return -1 }

type MyRuntime struct{}

func (g MyRuntime) Time() Time {
	return time{}
}

func main() {
	runtime := MyRuntime{}
	state := NewState(runtime)
	node := App(state)
	// HtmlWriter(context.Background(), os.Stdout, node)
	fmt.Println(HtmlString(context.Background(), node))
}
