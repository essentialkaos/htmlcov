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

	"github.com/essentialkaos/ek/v12/color"
	"github.com/essentialkaos/ek/v12/fsutil"
	"github.com/essentialkaos/ek/v12/mathutil"
	"github.com/essentialkaos/ek/v12/path"
	"github.com/essentialkaos/ek/v12/strutil"

	"golang.org/x/tools/cover"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// TEMPLATE is HTML template
const TEMPLATE = `<!DOCTYPE html="en">
<html>
  <head>
    <title>Coverage report</title>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
    <link rel="icon" href="data:image/svg+xml,<svg xmlns=%22http://www.w3.org/2000/svg%22 viewBox=%220 0 100 100%22><text y=%22.9em%22 font-size=%2290%22>📃</text></svg>" />
    <style>
      body {
        background: #191919;
        color: #777;
      }

      body, pre, #files, #legend span {
        font-family: 'JetBrains Mono', 'Fira Code', Consolas, Menlo, monospace;
        font-size: 15px;
        font-variant-ligatures: none;
      }

      .point {
        background: none !important;
      }

      #topbar {
        background: #111;
        border-bottom: 1px solid #444;
        font-weight: bold;
        height: 44px;
        position: fixed;
        top: 0; left: 0; right: 0;
      }

      #files {
        background-color: #191919;
        border-radius: 4px;
        border: 1px solid #444;
        color: #ccc;
        padding: 2px;
        font-size: 12px;
      }

      #numbers {
        color: #333;
        cursor: default;
        float: left;
        margin-right: 8px;
        overflow-y: hidden;
        text-align: right;
        user-select: none;
        width: 38px;
      }

      #source {
        margin-top: 50px;
        margin-left: 4px;
        tab-size: 4;
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
        margin: 0 4px;
        background: #222;
        border-radius: 8px;
        cursor: help;
        font-size: 14px;
        padding: 2px 8px;
      }

      @media (max-width: 860px) {
        body, pre, #legend span {
          font-size: 14px;
        }

        #source {
          tab-size: 2;
        }
      }

      @media (max-width: 680px) {
        body, pre, #legend span {
          font-size: 12px;
        }
      }

{{colors}}
    </style>
  </head>
  <body>
    <div id="topbar">
      <div id="nav">
        <select id="files">
        {{- range $i, $f := .Files}}
        <option value="file{{$i}}">{{$f.Name}} ({{printf "%.1f" $f.Coverage}}%)</option>
        {{- end}}
        </select>
      </div>
      <div id="legend">
        <span title="Code which can't be tested">not tracked</span>
      {{- if .IsSet}}
        <span class="cov0" title="Code without test coverage">not covered</span>
        <span class="cov8" title="Code covered by tests">covered</span>
      {{- else}}
        <span class="cov0" title="Code without test coverage">no coverage</span>
        <span class="cov1" title="Code with low coverage (counter = 1)">low coverage</span>
        <span class="cov2 point" title="Code with coverage counter = 2">●</span>
        <span class="cov3 point" title="Code with coverage counter = 3">●</span>
        <span class="cov4 point" title="Code with coverage counter = 4">●</span>
        <span class="cov5 point" title="Code with coverage counter = 5">●</span>
        <span class="cov6 point" title="Code with coverage counter = 6">●</span>
        <span class="cov7 point" title="Code with coverage counter = 7">●</span>
        <span class="cov8 point" title="Code with coverage counter = 8">●</span>
        <span class="cov9 point" title="Code with coverage counter = 9">●</span>
        <span class="cov10" title="Code with high coverage (counter = 10)">high coverage</span>
      {{- end}}
      </div>
    </div>
      <div id="content">
      <div id="numbers">{{range .Lines}}{{.}}<br/>{{end}}</div>
      <div id="source">
      {{range $i, $f := .Files}}
      <pre class="file" id="file{{$i}}" data-name="{{$f.Name}}" data-lines="{{$f.Lines}}" data-cover="{{printf "%.1f" $f.Coverage}}%"{{if $i}} style="display: none"{{end}}>{{$f.Data}}</pre>
      {{end}}
      </div>
     </div>
  </body>
  <script>
    var files
    var current

    function main() {
      files = document.getElementById('files');
      current = document.getElementById('file0');

      if (window.location.hash != "") {
        var selectedFile = window.location.hash.substr(1);
        swapFiles('file' + selectedFile);
        files.selectedIndex = parseInt(selectedFile);
      } else {
        updateViewport();
        updateInfo(current.id);
      }

      files.addEventListener('change', onChange, false);
    }

    function onChange() {
      swapFiles(files.value);
    }

    function swapFiles(file) {
      current.style.display = 'none';
      current = document.getElementById(file);
      current.style.display = 'block';

      updateViewport();
      updateInfo(file);

      window.scrollTo(0, 0);
    }

    function updateInfo(file) {
      window.location.hash = file.substr(4);
      document.title = current.getAttribute('data-cover') + ' • ' + current.getAttribute('data-name');
    }

    function updateViewport() {
      document.getElementById('numbers').style.maxHeight = current.offsetHeight;
    }

    main();
  </script>
</html>
`

// ////////////////////////////////////////////////////////////////////////////////// //

type CoverData struct {
	Files []*FileCover
	IsSet bool
}

type FileCover struct {
	Name     string
	Data     template.HTML
	Lines    int
	Coverage float64
}

// ////////////////////////////////////////////////////////////////////////////////// //

var (
	covNoneColor = color.Hex(0xDD3D27)
	covMinColor  = color.Hex(0x989997)
	covMaxColor  = color.Hex(0x77D300)
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Lines returns slice with line numbers
func (c *CoverData) Lines() []int {
	var maxLines int

	for _, f := range c.Files {
		maxLines = mathutil.Max(maxLines, f.Lines)
	}

	result := make([]int, maxLines)

	for i := 0; i < maxLines; i++ {
		result[i] = i + 1
	}

	return result
}

// ////////////////////////////////////////////////////////////////////////////////// //

// convertProfile converts profiles into HTML and writes it to the given file
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

	uniquifyFileNames(coverData)

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

	buf, lines, err := generateHTML(profile, data)

	if err != nil {
		return nil, err
	}

	return &FileCover{
		Name:     file,
		Data:     template.HTML(buf.String()),
		Lines:    lines + 1,
		Coverage: calculateCoverage(profile),
	}, nil
}

// generateHTML generates HTML code for given source data
func generateHTML(profile *cover.Profile, data []byte) (bytes.Buffer, int, error) {
	var buf bytes.Buffer
	var lines int

	boundaries := profile.Boundaries(data)

	for i := range data {
		for len(boundaries) > 0 && boundaries[0].Offset == i {
			boundary := boundaries[0]

			if boundary.Start {
				n := 0

				if boundary.Count > 0 {
					n = int(math.Floor(boundary.Norm*9)) + 1
				}

				if profile.Mode == "set" {
					fmt.Fprintf(&buf, `<span class="cov%v">`, n)
				} else {
					fmt.Fprintf(&buf, `<span class="cov%v" title="Count: %v">`, n, boundary.Count)
				}
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
		case '\n':
			buf.WriteRune('\n')
			lines++
		default:
			buf.WriteByte(data[i])
		}
	}

	return buf, lines, nil
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

// uniquifyFileNames removes the repeated part from file names
func uniquifyFileNames(data *CoverData) {
	if len(data.Files) == 1 {
		data.Files[0].Name = path.Base(data.Files[0].Name)
		return
	}

	var samePart string

MAIN:
	for i := 0; i < 1024; i++ {
		var cr byte

		for j, f := range data.Files {
			if j == 0 && len(f.Name) > i {
				cr = f.Name[i]
				continue
			}

			if cr != f.Name[i] {
				break MAIN
			}
		}

		samePart += string(cr)
	}

	if samePart != "" {
		for _, f := range data.Files {
			f.Name = strutil.Exclude(f.Name, samePart)
		}
	}
}
