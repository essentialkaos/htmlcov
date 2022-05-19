package app

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2022 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"os"
	"strings"

	"github.com/essentialkaos/ek/v12/fmtc"
	"github.com/essentialkaos/ek/v12/fsutil"
	"github.com/essentialkaos/ek/v12/options"
	"github.com/essentialkaos/ek/v12/usage"
	"github.com/essentialkaos/ek/v12/usage/completion/bash"
	"github.com/essentialkaos/ek/v12/usage/completion/fish"
	"github.com/essentialkaos/ek/v12/usage/completion/zsh"
	"github.com/essentialkaos/ek/v12/usage/man"
	"github.com/essentialkaos/ek/v12/usage/update"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Basic utility info
const (
	APP  = "htmlcov"
	VER  = "0.0.1"
	DESC = "Utility for converting coverage profiles into HTML pages"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Options
const (
	OPT_OUTPUT   = "o:output"
	OPT_REMOVE   = "r:remove"
	OPT_NO_COLOR = "nc:no-color"
	OPT_HELP     = "h:help"
	OPT_VER      = "v:version"

	OPT_COMPLETION   = "completion"
	OPT_GENERATE_MAN = "generate-man"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// optMap contains information about all supported options
var optMap = options.Map{
	OPT_OUTPUT:   {Value: "coverage.html"},
	OPT_REMOVE:   {Type: options.BOOL},
	OPT_NO_COLOR: {Type: options.BOOL},
	OPT_HELP:     {Type: options.BOOL, Alias: "u:usage"},
	OPT_VER:      {Type: options.BOOL, Alias: "ver"},

	OPT_COMPLETION:   {},
	OPT_GENERATE_MAN: {Type: options.BOOL},
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Init is main function
func Init() {
	args, errs := options.Parse(optMap)

	if len(errs) != 0 {
		for _, err := range errs {
			printError(err.Error())
		}

		os.Exit(1)
	}

	preConfigureUI()

	if options.Has(OPT_COMPLETION) {
		os.Exit(genCompletion())
	}

	if options.Has(OPT_GENERATE_MAN) {
		os.Exit(genMan())
	}

	configureUI()

	if options.GetB(OPT_VER) {
		os.Exit(showAbout())
	}

	if options.GetB(OPT_HELP) || len(args) == 0 {
		os.Exit(showUsage())
	}

	process(args)
}

// preConfigureUI preconfigures UI based on information about user terminal
func preConfigureUI() {
	term := os.Getenv("TERM")

	fmtc.DisableColors = true

	if term != "" {
		switch {
		case strings.Contains(term, "xterm"),
			strings.Contains(term, "color"),
			term == "screen":
			fmtc.DisableColors = false
		}
	}

	if !fsutil.IsCharacterDevice("/dev/stdout") && os.Getenv("FAKETTY") == "" {
		fmtc.DisableColors = true
	}

	if os.Getenv("NO_COLOR") != "" {
		fmtc.DisableColors = true
	}
}

// configureUI configures user interface
func configureUI() {
	if options.GetB(OPT_NO_COLOR) {
		fmtc.DisableColors = true
	}
}

// process starts processing
func process(args options.Arguments) {
	covFile := args.Get(0).Clean().String()
	err := fsutil.ValidatePerms("FRS", covFile)

	if err != nil {
		printErrorAndExit(err.Error())
	}

	output := options.GetS(OPT_OUTPUT)
	err = convertProfile(covFile, output)

	if err != nil {
		printErrorAndExit(err.Error())
	}

	if options.GetB(OPT_REMOVE) {
		os.Remove(covFile)
	}

	fmtc.Printf("{g}Report successfully saved as {g*}%s{!}\n", output)
}

// printError prints error message to console
func printError(f string, a ...interface{}) {
	if len(a) == 0 {
		fmtc.Fprintln(os.Stderr, "{r}"+f+"{!}")
	} else {
		fmtc.Fprintf(os.Stderr, "{r}"+f+"{!}\n", a...)
	}
}

// printError prints warning message to console
func printWarn(f string, a ...interface{}) {
	if len(a) == 0 {
		fmtc.Fprintln(os.Stderr, "{y}"+f+"{!}")
	} else {
		fmtc.Fprintf(os.Stderr, "{y}"+f+"{!}\n", a...)
	}
}

// printErrorAndExit print error mesage and exit with exit code 1
func printErrorAndExit(f string, a ...interface{}) {
	printError(f, a...)
	os.Exit(1)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// showUsage prints usage info
func showUsage() int {
	genUsage().Render()
	return 0
}

// showAbout prints info about version
func showAbout() int {
	genAbout().Render()
	return 0
}

// genCompletion generates completion for different shells
func genCompletion() int {
	info := genUsage()

	switch options.GetS(OPT_COMPLETION) {
	case "bash":
		fmt.Printf(bash.Generate(info, "htmlcov"))
	case "fish":
		fmt.Printf(fish.Generate(info, "htmlcov"))
	case "zsh":
		fmt.Printf(zsh.Generate(info, optMap, "htmlcov"))
	default:
		return 1
	}

	return 0
}

// genMan generates man page
func genMan() int {
	fmt.Println(
		man.Generate(
			genUsage(),
			genAbout(),
		),
	)

	return 0
}

// genUsage generates usage info
func genUsage() *usage.Info {
	info := usage.NewInfo("", "coverage-file")

	info.AddOption(OPT_OUTPUT, "Output file {s-}(default: coverage.html){!}", "file")
	info.AddOption(OPT_REMOVE, "Delete input file after successfull generation")
	info.AddOption(OPT_NO_COLOR, "Disable colors in output")
	info.AddOption(OPT_HELP, "Show this help message")
	info.AddOption(OPT_VER, "Show version")

	return info
}

// genAbout generates info about version
func genAbout() *usage.About {
	return &usage.About{
		App:           APP,
		Version:       VER,
		Desc:          DESC,
		Year:          2009,
		Owner:         "ESSENTIAL KAOS",
		License:       "Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>",
		UpdateChecker: usage.UpdateChecker{"essentialkaos/htmlcov", update.GitHubChecker},
	}
}

// ////////////////////////////////////////////////////////////////////////////////// //
