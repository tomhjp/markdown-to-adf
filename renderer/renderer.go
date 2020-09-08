package renderer

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extAst "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
)

var _ renderer.Renderer = &ADFRenderer{}

// ADFRenderer implements goldmark.Renderer
type ADFRenderer struct {
	document     *Node          // Root node
	blockContext blockNodeStack // Track where we are in the structure of the document
}

type Node struct {
	Type       Type        `json:"type"`
	Version    int         `json:"version,omitempty"`
	Attributes *Attributes `json:"attrs,omitempty"`
	Content    []*Node     `json:"content,omitempty"`
	Marks      []Mark      `json:"marks,omitempty"`
	Text       string      `json:"text,omitempty"`
}

func (n *Node) AddContent(c *Node) {
	n.Content = append(n.Content, c)
}

type Attributes struct {
	Width  float32 `json:"width,omitempty"`
	Layout Layout  `json:"layout,omitempty"`
	Level  int     `json:"level,omitempty"`
}

// Type represents the type of a node
type Type string

// Node types
const (
	TypeNone        = "none"
	TypeBlockquote  = "blockquote"
	TypeBulletList  = "bulletList"
	TypeCodeBlock   = "codeBlock"
	TypeHeading     = "heading"
	TypeMediaGroup  = "mediaGroup"
	TypeMediaSingle = "mediaSingle"
	TypeOrderedList = "orderedList"
	TypePanel       = "panel"
	TypeParagraph   = "paragraph"
	TypeRule        = "rule"
	TypeTable       = "table"
	TypeListItem    = "listItem"
	TypeMedia       = "media"
	TypeTableCell   = "table_cell"
	TypeTableHeader = "table_header"
	TypeTableRow    = "table_row"
	TypeEmoji       = "emoji"
	TypeHardBreak   = "hardBreak"
	TypeInlineCard  = "inlineCard"
	TypeMention     = "mention"
	TypeText        = "text"
)

type Layout string

// Enum values for Layout in Attributes struct
const (
	LayoutWrapLeft   = "wrap-left"
	LayoutCenter     = "center"
	LayoutWrapRight  = "wrap-right"
	LayoutWide       = "wide"
	LayoutFullWidth  = "full-width"
	LayoutAlignStart = "align-start"
	LayoutAlignEnd   = "align-end"
)

type blockNodeStack struct {
	data []*Node
}

func (s *blockNodeStack) Push(node *Node) {
	fmt.Printf("Adding node %+v\n", *node)
	// Update the actual document
	s.data[len(s.data)-1].AddContent(node)
	// Update the context stack
	s.data = append(s.data, node)
}

// Intentionally unsafe because we should never peek an empty stack
func (s *blockNodeStack) Peek() *Node {
	return s.data[len(s.data)-1]
}

// Intentionally unsafe because we should never pop an empty stack
func (s *blockNodeStack) Pop() *Node {
	last := len(s.data) - 1
	node := s.data[last]
	s.data = s.data[:last]
	return node
}

// Mark represents a text formatting directive
type Mark string

// Enum values for Mark text formatting
const (
	MarkCode      Mark = "code"
	MarkEm        Mark = "em"
	MarkLink      Mark = "link"
	MarkStrike    Mark = "strike"
	MarkStrong    Mark = "strong"
	MarkSubsup    Mark = "subsup"
	MarkTextcolor Mark = "textColor"
	MarkUnderline Mark = "underline"
)

func NewRenderer() *ADFRenderer {
	root := Node{
		Version: 1,
		Type:    "doc",
	}
	return &ADFRenderer{
		document: &root,
		blockContext: blockNodeStack{
			data: []*Node{&root},
		},
	}
}

func Render(w io.Writer, source []byte) error {
	gm := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM, // GitHub flavoured markdown.
		),
		goldmark.WithParserOptions(
			parser.WithAttribute(), // Enables # headers {#custom-ids}.
		),
		goldmark.WithRenderer(NewRenderer()),
	)

	if err := gm.Convert(source, w); err != nil {
		return err
	}

	return nil
}

func astToADFType(n ast.Node) Type {
	switch n.(type) {
	case *ast.Document:
	case *ast.Paragraph,
		*ast.TextBlock:
		return TypeParagraph
	case *ast.Heading:
		return TypeHeading
	case *ast.Text,
		*ast.String,
		*extAst.Strikethrough,
		*ast.Emphasis:
		return TypeText
	case *ast.CodeSpan,
		*ast.CodeBlock,
		*ast.FencedCodeBlock:
		return TypeCodeBlock
	case *ast.ThematicBreak:
	case *ast.Blockquote:
		return TypeBlockquote
	case *ast.List:
		if n.(*ast.List).IsOrdered() {
			return TypeOrderedList
		}
		return TypeBulletList
	case *ast.ListItem:
		return TypeListItem
	case *ast.Link:
	case *ast.Image:
		return TypeMedia
	case *ast.HTMLBlock:
	case *ast.RawHTML:
	case *extAst.Table:
		return TypeTable
	case *extAst.TableHeader:
		return TypeTableHeader
	case *extAst.TableRow:
		return TypeTableRow
	case *extAst.TableCell:
		return TypeTableCell
	}

	return TypeNone
}

func (r *ADFRenderer) renderSingle(w io.Writer, source []byte, n ast.Node, entering bool) ast.WalkStatus {
	if !entering {
		if astToADFType(n) == r.blockContext.Peek().Type {
			r.blockContext.Pop()
		}
		return ast.WalkContinue
	}

	adfNode := &Node{Type: astToADFType(n)}

	switch n.(type) {
	case *ast.Document:
		fmt.Printf("Node: %s, entering: %v, value: %v, children: %d\n", "*ast.Document", entering, string(n.Text(source)), n.ChildCount())
	case *ast.TextBlock:
		r.blockContext.Push(adfNode)
		fmt.Printf("Node: %s, entering: %v, value: %v, children: %d\n", "*ast.TextBlock", entering, string(n.Text(source)), n.ChildCount())
		//r.paragraph(tnode, entering)
	case *ast.Paragraph:
		r.blockContext.Push(adfNode)
		fmt.Printf("Node: %s, entering: %v, value: %v, children: %d\n", "*ast.Paragraph", entering, string(n.Text(source)), n.ChildCount())
		//r.paragraph(tnode, entering)
	case *ast.Heading:
		adfNode.Attributes = &Attributes{
			Level: n.(*ast.Heading).Level,
		}
		r.blockContext.Push(adfNode)
		fmt.Printf("Node: %s, entering: %v, value: %v, children: %d\n", "*ast.Heading", entering, string(n.Text(source)), n.ChildCount())
		// if entering {
		// 	children := r.renderChildren(source, n)
		// 	r.header(tnode, children)
		// }
		// return ast.WalkSkipChildren
	case *ast.Text:
		fmt.Printf("Node: %s, entering: %v, value: %v, children: %d\n", "*ast.Text", entering, string(n.Text(source)), n.ChildCount())
		adfNode.Text = string(n.Text(source))
		if len(adfNode.Text) == 0 {
			// TODO: Uh what's happening here? Not sure why goldmark is splitting up paragraph text in this way.
			adfNode.Text = " "
		}
		r.blockContext.Peek().AddContent(adfNode)
		// r.normalText(tnode, source, entering)
	case *ast.String:
		fmt.Printf("Node: %s, entering: %v, value: %v, children: %d\n", "*ast.String", entering, string(n.Text(source)), n.ChildCount())
		// r.string(tnode, source, entering)
	case *ast.CodeSpan:
		fmt.Printf("Node: %s, entering: %v, value: %v, children: %d\n", "*ast.CodeSpan", entering, string(n.Text(source)), n.ChildCount())
		// if entering {
		// 	r.codeSpan(tnode, source)
		// }
		// return ast.WalkSkipChildren
	case *extAst.Strikethrough:
		fmt.Printf("Node: %s, entering: %v, value: %v, children: %d\n", "*extAst.Strikethrough", entering, string(n.Text(source)), n.ChildCount())
		// if entering {
		// 	children := r.renderChildren(source, n)
		// 	r.strikeThrough(children)
		// }
		// return ast.WalkSkipChildren
	case *ast.Emphasis:
		fmt.Printf("Node: %s, entering: %v, value: %v, children: %d\n", "*ast.Emphasis", entering, string(n.Text(source)), n.ChildCount())
		// if entering {
		// 	children := r.renderChildren(source, n)
		// 	r.emphasis(tnode, children)
		// }
		// return ast.WalkSkipChildren
	case *ast.ThematicBreak:
		fmt.Printf("Node: %s, entering: %v, value: %v, children: %d\n", "*ast.ThematicBreak", entering, string(n.Text(source)), n.ChildCount())
		// if entering {
		// 	r.hrule()
		// }
	case *ast.Blockquote:
		fmt.Printf("Node: %s, entering: %v, value: %v, children: %d\n", "*ast.Blockquote", entering, string(n.Text(source)), n.ChildCount())
		// r.blockQuote(entering)
	case *ast.List:
		r.blockContext.Push(adfNode)
		fmt.Printf("Node: %s, entering: %v, value: %v, children: %d\n", "*ast.List", entering, string(n.Text(source)), n.ChildCount())
		// r.list(tnode, entering)
	case *ast.ListItem:
		r.blockContext.Push(adfNode)
		fmt.Printf("Node: %s, entering: %v, value: %v, children: %d\n", "*ast.ListItem", entering, string(n.Text(source)), n.ChildCount())
		// r.item(tnode, entering, source)
	case *ast.Link:
		fmt.Printf("Node: %s, entering: %v, value: %v, children: %d\n", "*ast.Link", entering, string(n.Text(source)), n.ChildCount())
		// if entering {
		// 	children := r.renderChildren(source, n)
		// 	r.link(tnode.Destination, tnode.Title, children)
		// }
		// return ast.WalkSkipChildren
	case *ast.Image:
		fmt.Printf("Node: %s, entering: %v, value: %v, children: %d\n", "*ast.Image", entering, string(n.Text(source)), n.ChildCount())
		// if entering {
		// 	children := r.renderChildren(source, n)
		// 	r.image(tnode.Destination, tnode.Title, children)
		// }
		// return ast.WalkSkipChildren
	case *ast.CodeBlock:
		fmt.Printf("Node: %s, entering: %v, value: %v, children: %d\n", "*ast.CodeBlock", entering, string(n.Text(source)), n.ChildCount())
		// if entering {
		// 	r.blockCode(tnode, source)
		// }
	case *ast.FencedCodeBlock:
		fmt.Printf("Node: %s, entering: %v, value: %v, children: %d\n", "*ast.FencedCodeBlock", entering, string(n.Text(source)), n.ChildCount())
		// if entering {
		// 	r.blockCode(tnode, source)
		// }
	case *ast.HTMLBlock:
		fmt.Printf("Node: %s, entering: %v, value: %v, children: %d\n", "*ast.HTMLBlock", entering, string(n.Text(source)), n.ChildCount())
		// if entering {
		// 	r.blockHtml(tnode, source)
		// }
	case *ast.RawHTML:
		fmt.Printf("Node: %s, entering: %v, value: %v, children: %d\n", "*ast.RawHTML", entering, string(n.Text(source)), n.ChildCount())
		// if entering {
		// 	r.rawHtml(tnode, source)
		// }
		// return ast.WalkSkipChildren
	case *extAst.Table:
		fmt.Printf("Node: %s, entering: %v, value: %v, children: %d\n", "*extAst.Table", entering, string(n.Text(source)), n.ChildCount())
		// r.table(tnode, entering)
	case *extAst.TableHeader:
		fmt.Printf("Node: %s, entering: %v, value: %v, children: %d\n", "*extAst.TableHeader", entering, string(n.Text(source)), n.ChildCount())
		// if entering {
		// 	r.tableIsHeader = true
		// }
	case *extAst.TableRow:
		fmt.Printf("Node: %s, entering: %v, value: %v, children: %d\n", "*extAst.TableRow", entering, string(n.Text(source)), n.ChildCount())
		// if entering {
		// 	r.tableIsHeader = false
		// }
	case *extAst.TableCell:
		fmt.Printf("Node: %s, entering: %v, value: %v, children: %d\n", "*extAst.TableCell", entering, string(n.Text(source)), n.ChildCount())
		// if entering {
		// 	children := r.renderChildren(source, n)
		// 	if r.tableIsHeader {
		// 		r.tableHeaderCell(children, tnode.Alignment)
		// 	} else {
		// 		r.tableCell(children)
		// 	}
		// }
		// return ast.WalkSkipChildren
	default:
		panic("unknown type" + n.Kind().String())
	}

	// if !entering {
	// 	r.buf.WriteTo(writer)
	// 	r.buf.Reset()
	// 	r.buf = bytes.NewBuffer(nil)
	// }

	return ast.WalkContinue
}

// Render implements goldmark.Renderer interface.
func (r *ADFRenderer) Render(w io.Writer, source []byte, n ast.Node) error {
	for current := n.FirstChild(); current != nil; current = current.NextSibling() {
		ast.Walk(current, func(current ast.Node, entering bool) (ast.WalkStatus, error) {
			return r.renderSingle(w, source, current, entering), nil
		})
	}

	b, err := json.MarshalIndent(r.document, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	w.Write(b)

	return nil
}

func (*ADFRenderer) AddOptions(...renderer.Option) {
	//panic("No options for ADF renderer")
}
