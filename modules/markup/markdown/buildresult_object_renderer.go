package markdown

import (
	"bytes"
	"net/url"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// buildresultsObjectRenderer renders images from br.opensuse.org as <object>.
type buildresultsObjectRenderer struct {
	defaultFuncs map[ast.NodeKind]renderer.NodeRendererFunc
}

// adapter so we can capture funcs into a map
type funcRegisterer struct {
	funcs map[ast.NodeKind]renderer.NodeRendererFunc
}

func (r *funcRegisterer) Register(kind ast.NodeKind, v renderer.NodeRendererFunc) {
	r.funcs[kind] = v
}

func newObjectImageRenderer() *buildresultsObjectRenderer {
	tmp := html.NewRenderer()
	reg := &funcRegisterer{funcs: make(map[ast.NodeKind]renderer.NodeRendererFunc)}
	tmp.RegisterFuncs(reg)
	return &buildresultsObjectRenderer{defaultFuncs: reg.funcs}
}

func (r *buildresultsObjectRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindImage, r.renderImage)
}

func (r *buildresultsObjectRenderer) renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	img := node.(*ast.Image)
	if u, err := url.Parse(string(img.Destination)); err == nil && (u.Host == "br.opensuse.org" || u.Host == "br.suse.de" || u.Host == "gitexplorer.opensuse.org") {
		// collect alt text
		var alt bytes.Buffer
		for c := img.FirstChild(); c != nil; c = c.NextSibling() {
			if t, ok := c.(*ast.Text); ok {
				alt.Write(t.Segment.Value(source))
			}
		}
		// <object data="..."> (let sanitizer control which attrs are kept)
		w.WriteString(`<object data="`)
		w.Write(util.EscapeHTML([]byte(u.String())))
		w.WriteString(`" type="image/svg+xml" aria-label="`)
		w.Write(util.EscapeHTML(alt.Bytes()))
		w.WriteString(`"></object>`)
		return ast.WalkSkipChildren, nil
	}

	// Delegate to stock html image renderer for all other cases
	if f := r.defaultFuncs[ast.KindImage]; f != nil {
		return f(w, source, node, entering)
	}
	return ast.WalkContinue, nil
}
