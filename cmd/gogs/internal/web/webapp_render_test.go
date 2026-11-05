package web

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gogs.io/gogs/internal/context"
)

func TestRenderIndex_injection(t *testing.T) {
	customDir := t.TempDir()
	injectDir := filepath.Join(customDir, "templates", "inject")
	require.NoError(t, os.MkdirAll(injectDir, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(injectDir, "head.tmpl"), []byte(`<meta name="head-marker">`), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(injectDir, "footer.tmpl"), []byte(`<script>footerMarker()</script>`), 0600))
	t.Setenv("GOGS_CUSTOM", customDir)

	inject, err := readInjectContent()
	require.NoError(t, err)

	shell := `<html><head>{{.WebContext}}</head><body><div id="root"></div></body></html>`
	got, err := renderIndex([]byte(shell), context.WebContext{Lang: "en-US"}, inject)
	require.NoError(t, err)

	out := string(got)
	assert.Contains(t, out, `<meta name="head-marker">`)
	assert.Contains(t, out, `<script>footerMarker()</script></body>`)
	// The head injection lands inside <head>, before </head>.
	assert.Less(t, strings.Index(out, `<meta name="head-marker">`), strings.Index(out, `</head>`))
}

func TestRenderIndex_assetPaths(t *testing.T) {
	shell := `<html><head>{{.WebContext}}<link rel="stylesheet" href="./assets/index-abc.css"></head>` +
		`<body><script type="module" src="./assets/index-abc.js"></script></body></html>`

	tests := []struct {
		name   string
		subURL string
		want   string
	}{
		{name: "root mount", subURL: "", want: "/assets/index-abc.js"},
		{name: "subpath mount", subURL: "/gogs", want: "/gogs/assets/index-abc.js"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := renderIndex([]byte(shell), context.WebContext{Lang: "en-US", SubURL: test.subURL}, injectContent{})
			require.NoError(t, err)
			assert.Contains(t, string(got), `src="`+test.want+`"`)
			assert.NotContains(t, string(got), `"./assets/`)
		})
	}
}
