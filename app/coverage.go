package app

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2022 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"bufio"
	"bytes"
	"fmt"
	"go/build"
	"html/template"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/essentialkaos/ek/v12/color"
	"github.com/essentialkaos/ek/v12/fsutil"

	"golang.org/x/tools/cover"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// TEMPLATE is HTML template
const TEMPLATE = `<!DOCTYPE html="en">
<html>
  <head>
  	<title>Coverage</title>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8">
    <style>
      body {
        background: #222;
        color: #777;
      }

      body, pre, #legend span {
        font-family: 'JetBrains Mono', 'Fira Code', Consolas, Menlo, monospace;
        font-size: 15px;
        font-variant-ligatures: none;
      }

      #topbar {
        background: #111;
        border-bottom: 1px solid #444;
        font-weight: bold;
        height: 42px;
        position: fixed;
        top: 0; left: 0; right: 0;
      }

      #content {
        margin-top: 50px;
        margin-left: 4px;
      }

      #nav, #legend {
        float: left;
        margin-left: 10px;
      }

      #legend {
        margin-top: 12px;
      }

      #nav {
        margin-top: 10px;
      }

      #legend span {
        margin: 0 5px;
      }

{{colors}}
    </style>
  </head>
  <body>
    <div id="topbar">
      <div id="nav">
        <select id="files">
        {{range $i, $f := .Files}}
        <option value="file{{$i}}">{{$f.Name}} ({{printf "%.1f" $f.Coverage}}%)</option>
        {{end}}
        </select>
      </div>
      <div id="legend">
        <span>not tracked</span>
      {{if .IsSet}}
        <span class="cov0">not covered</span>
        <span class="cov8">covered</span>
      {{else}}
        <span class="cov0">no coverage</span>
        <span class="cov1">low coverage</span>
        <span class="cov2">●</span>
        <span class="cov3">●</span>
        <span class="cov4">●</span>
        <span class="cov5">●</span>
        <span class="cov6">●</span>
        <span class="cov7">●</span>
        <span class="cov8">●</span>
        <span class="cov9">●</span>
        <span class="cov10">high coverage</span>
      {{end}}
      </div>
    </div>
    <div id="content">
    {{range $i, $f := .Files}}
    <pre class="file" id="file{{$i}}" {{if $i}}style="display: none"{{end}}>{{$f.Data}}</pre>
    {{end}}
    </div>
  </body>
  <script>
  (function() {
    var files = document.getElementById('files');
    var visible = document.getElementById('file0');
    files.addEventListener('change', onChange, false);
    function onChange() {
      visible.style.display = 'none';
      visible = document.getElementById(files.value);
      visible.style.display = 'block';
      window.scrollTo(0, 0);
    }
  })();
  </script>
</html>
`

// TAB_SIZE is number of spaces for tab symbols
const TAB_SIZE = 4

// ////////////////////////////////////////////////////////////////////////////////// //

type CoverData struct {
	Files []*FileCover
	IsSet bool
}

type FileCover struct {
	Name     string
	Data     template.HTML
	Coverage float64
}

// ////////////////////////////////////////////////////////////////////////////////// //

var (
	covNoneColor = color.Hex(0xDD3D27)
	covMinColor  = color.Hex(0x989997)
	covMaxColor  = color.Hex(0x77D300)
)

// ////////////////////////////////////////////////////////////////////////////////// //

func convertProfile(profileFile, outputFile string) error {
	profiles, err := cover.ParseProfiles(profileFile)

	if err != nil {
		return err
	}

	coverData := &CoverData{}

	for _, profile := range profiles {
		if profile.Mode == "set" {
			coverData.IsSet = true
		}

		fileCover, err := generateFileCover(profile)

		if err != nil {
			return err
		}

		coverData.Files = append(coverData.Files, fileCover)
	}

	return writeCoverReport(coverData, outputFile)
}

// writeCoverReport writes coverage report into file
func writeCoverReport(data *CoverData, outputFile string) error {
	fd, err := os.OpenFile(outputFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)

	if err != nil {
		return fmt.Errorf("Can't write output to %s: %w", outputFile, err)
	}

	defer fd.Close()

	w := bufio.NewWriter(fd)
	tmpl := template.Must(template.New("html").Funcs(template.FuncMap{
		"colors": getCoverageColors,
	}).Parse(TEMPLATE))

	err = tmpl.Execute(w, data)

	if err != nil {
		return fmt.Errorf("Can't create report: %w", err)
	}

	return w.Flush()
}

// generateFileCover generates coverage report for file
func generateFileCover(profile *cover.Profile) (*FileCover, error) {
	file := profile.FileName
	data, err := getSourceData(file)

	if err != nil {
		return nil, err
	}

	buf, err := generateHTML(profile, data)

	if err != nil {
		return nil, err
	}

	return &FileCover{
		Name:     file,
		Data:     template.HTML(buf.String()),
		Coverage: calculateCoverage(profile),
	}, nil
}

// generateHTML generates HTML code for given source data
func generateHTML(profile *cover.Profile, data []byte) (bytes.Buffer, error) {
	var buf bytes.Buffer

	boundaries := profile.Boundaries(data)

	for i := range data {
		for len(boundaries) > 0 && boundaries[0].Offset == i {
			boundary := boundaries[0]

			if boundary.Start {
				n := 0

				if boundary.Count > 0 {
					n = int(math.Floor(boundary.Norm*9)) + 1
				}

				fmt.Fprintf(&buf, `<span class="cov%v" title="%v">`, n, boundary.Count)
			} else {
				buf.WriteString(`</span>`)
			}

			boundaries = boundaries[1:]
		}

		switch data[i] {
		case '>':
			buf.WriteString("&gt;")
		case '<':
			buf.WriteString("&lt;")
		case '&':
			buf.WriteString("&amp;")
		case '\t':
			buf.WriteString(strings.Repeat(" ", TAB_SIZE))
		default:
			buf.WriteByte(data[i])
		}
	}

	return buf, nil
}

// calculateCoverage calculates coverage as a percentage
func calculateCoverage(profile *cover.Profile) float64 {
	var total, covered float64

	for _, block := range profile.Blocks {
		total += float64(block.NumStmt)

		if block.Count > 0 {
			covered += float64(block.NumStmt)
		}
	}

	if total == 0 {
		return 0
	}

	return covered / total * 100
}

// getSourceData reads source data from given file
func getSourceData(file string) ([]byte, error) {
	dir, file := filepath.Split(file)
	pkg, err := build.Import(dir, ".", build.FindOnly)

	if err != nil {
		return nil, fmt.Errorf("Can't find file %q: %w", file, err)
	}

	srcFile := filepath.Join(pkg.Dir, file)
	err = fsutil.ValidatePerms("FRS", srcFile)

	if err != nil {
		return nil, err
	}

	return ioutil.ReadFile(srcFile)
}

// getCoverageColors generates CSS with colors for different coverage levels
func getCoverageColors() template.CSS {
	var buf bytes.Buffer

	minColor := covMinColor.ToRGB().ToHSV()
	maxColor := covMaxColor.ToRGB().ToHSV()
	vDiff := (maxColor.V - minColor.V) / 10.0

	fmt.Fprintf(&buf, "      .cov0 { color: %s; }\n", covNoneColor.ToWeb(false))

	for i := 0; i < 10; i++ {
		c := color.HSV{
			H: maxColor.H,
			S: (float64(i) / 10.0) + 0.1,
			V: minColor.V + (vDiff * float64(i)),
		}

		fmt.Fprintf(&buf, "      .cov%d { color: %s; }\n", i+1, c.ToRGB().ToHex().ToWeb(false))
	}

	return template.CSS(buf.String())
}
