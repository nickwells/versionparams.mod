package versionparams

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"github.com/nickwells/col.mod/v3/col"
	"github.com/nickwells/col.mod/v3/col/colfmt"
	"github.com/nickwells/location.mod/location"
	"github.com/nickwells/param.mod/v5/param"
	"github.com/nickwells/param.mod/v5/param/psetter"
)

var (
	vsnPart      []string
	shortDisplay bool
)

const (
	vpGoVersion = "go-version"
	vpPath      = "path"
	vpMain      = "main"
	vpMods      = "modules"
	vpSettings  = "build-settings"
	vpRaw       = "raw"

	noBuildInfo = "Build information not available"

	replIntro   = "   "
	modTypeDep  = "D"
	modTypeRepl = "r"
	modTypeMain = "M"
)

type versionPartShowFunc func(io.Writer, *debug.BuildInfo)

var vsnPartShowFuncMap = map[string]versionPartShowFunc{
	vpGoVersion: showGoVersion,
	vpPath:      showPath,
	vpMain:      showMain,
	vpMods:      showModules,
	vpSettings:  showSettings,
	vpRaw:       showRaw,
}

// depCols returns a set of columns for the modules section of the report
// with the column widths set to the maximum observed value
func depCols(bi *debug.BuildInfo) []*col.Col {
	var (
		colWidthPath    = len(bi.Main.Path)
		colWidthVersion = len(bi.Main.Version)
		colWidthSum     = len(bi.Main.Sum)
	)
	for _, d := range bi.Deps {
		if colWidthPath < len(d.Path) {
			colWidthPath = len(d.Path)
		}
		if colWidthVersion < len(d.Version) {
			colWidthVersion = len(d.Version)
		}
		if colWidthSum < len(d.Sum) {
			colWidthSum = len(d.Sum)
		}
		if d.Replace != nil {
			replPathLen := len(d.Replace.Path) + len(replIntro)
			if colWidthPath < replPathLen {
				colWidthPath = replPathLen
			}
			if colWidthVersion < len(d.Replace.Version) {
				colWidthVersion = len(d.Replace.Version)
			}
			if colWidthSum < len(d.Replace.Sum) {
				colWidthSum = len(d.Replace.Sum)
			}
		}
	}

	return []*col.Col{
		col.New(&colfmt.String{W: colWidthPath}, "Path"),
		col.New(&colfmt.String{W: colWidthVersion}, "Version"),
		col.New(&colfmt.String{W: colWidthSum}, "CheckSum"),
	}
}

// showGoVersion shows the Go Version used to build this executable
func showGoVersion(w io.Writer, bi *debug.BuildInfo) {
	prompt := "Built with Go Version: "
	if shortDisplay {
		prompt = ""
	}

	fmt.Fprintln(w, prompt+bi.GoVersion)
}

// showPath shows the Path of this executable
func showPath(w io.Writer, bi *debug.BuildInfo) {
	prompt := "Path: "
	if shortDisplay {
		prompt = ""
	}

	fmt.Fprintln(w, prompt+bi.Path)
}

// showMain shows the details of the main module of this executable
func showMain(w io.Writer, bi *debug.BuildInfo) {
	prompt := "Module, Version"
	if bi.Main.Sum != "" {
		prompt += ", Checksum"
	}
	prompt += ": "
	if shortDisplay {
		prompt = ""
	}

	fmt.Fprintln(w, prompt+bi.Main.Path, bi.Main.Version, bi.Main.Sum)
}

// showModules shows the module details of this executable
func showModules(w io.Writer, bi *debug.BuildInfo) {
	if !shortDisplay {
		fmt.Fprintln(w, "Modules:")
	}

	cols := depCols(bi)

	hdr := col.NewHeaderOrPanic()
	if shortDisplay {
		col.HdrOptDontPrint(hdr)
	}

	rpt := col.NewReport(hdr, w,
		col.New(&colfmt.String{}, "Type"), cols...)
	_ = rpt.PrintRow(modTypeMain, bi.Main.Path, bi.Main.Version, bi.Main.Sum)

	for _, dep := range bi.Deps {
		modType := modTypeDep
		if dep.Replace != nil {
			modType = modTypeRepl
		}

		_ = rpt.PrintRow(modType, dep.Path, dep.Version, dep.Sum)

		if dep.Replace != nil {
			repl := dep.Replace
			_ = rpt.PrintRow(modTypeDep,
				replIntro+repl.Path, repl.Version, repl.Sum)
		}
	}
}

// showSettings shows the build settings of this executable
func showSettings(w io.Writer, bi *debug.BuildInfo) {
	if !shortDisplay {
		fmt.Fprintln(w, "Build Settings:")
	}

	maxKey := 0
	for _, s := range bi.Settings {
		if len(s.Key) > maxKey {
			maxKey = len(s.Key)
		}
	}

	keyJust := col.Right

	hdr := col.NewHeaderOrPanic()
	if shortDisplay {
		col.HdrOptDontPrint(hdr)
		keyJust = col.Left
	}

	rpt := col.NewReport(hdr, w,
		col.New(&colfmt.String{StrJust: keyJust, W: maxKey}, "Key"),
		col.New(&colfmt.String{}, "Value"))

	for _, s := range bi.Settings {
		_ = rpt.PrintRow(s.Key, s.Value)
	}
}

// showRaw shows the build info in raw form
func showRaw(w io.Writer, bi *debug.BuildInfo) {
	fmt.Fprintln(w, bi)
}

// showVersion shows the version details, if specific parts have been
// requested then just those parts are shown otherwise the full default
// version info is displayed
func showVersion(w io.Writer) error {
	if len(vsnPart) == 0 {
		return nil
	}

	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return errors.New(noBuildInfo)
	}

	shown := make(map[string]bool, len(vsnPart))
	for _, part := range vsnPart {
		if shown[part] {
			continue
		}
		shown[part] = true

		var f versionPartShowFunc
		var ok bool
		if f, ok = vsnPartShowFuncMap[part]; !ok {
			return errors.New("bad version part: " + part)
		}

		f(w, bi)
	}

	os.Exit(0)
	return nil
}

// AddParams will add parameters to the passed param.PSet
func AddParams(ps *param.PSet) error {
	const (
		paramNameVersion          = "version"
		paramNameVersionPart      = "version-part"
		paramNameVersionPartShort = "version-part-short"
	)

	defaultParts := []string{vpGoVersion, vpPath, vpMain, vpMods, vpSettings}

	ps.Add(paramNameVersion, psetter.Nil{},
		"show the complete version details for this program"+
			" in the default format",
		param.PostAction(
			func(_ location.L, p *param.ByName, _ []string) error {
				vsnPart = append(vsnPart, defaultParts...)
				return nil
			}),
		param.SeeAlso(paramNameVersionPart),
		param.Attrs(param.CommandLineOnly|param.DontShowInStdUsage),
	)

	ps.Add(paramNameVersionPart,
		psetter.EnumList{
			Value: &vsnPart,
			AllowedVals: psetter.AllowedVals{
				vpGoVersion: "show the version of Go used to make this program",
				vpPath:      "show the path of the main package",
				vpMain:      "show the version of the main module",
				vpMods:      "show the module dependencies",
				vpSettings:  "show the settings used to make this program",
				vpRaw:       "show the full build information",
			},
			Aliases: psetter.Aliases{
				"go":          []string{vpGoVersion},
				"go-vsn":      []string{vpGoVersion},
				"mods":        []string{vpMods},
				"dep":         []string{vpMods},
				"build-flags": []string{vpSettings},
				"build":       []string{vpSettings},
				"settings":    []string{vpSettings},
				"default":     defaultParts,
			},
		},
		"show only the named parts of the version",
		param.SeeAlso(paramNameVersionPartShort),
		param.Attrs(param.CommandLineOnly|param.DontShowInStdUsage),
	)

	ps.Add(paramNameVersionPartShort,
		psetter.Bool{Value: &shortDisplay},
		"show the version parts in simplified form, without headings and"+
			" prompts. This is more useful if you want to use the value"+
			" as you won't need to strip out the other text. Note that"+
			" there is no short form of the '"+vpRaw+"' form.",
		param.SeeAlso(paramNameVersionPart),
		param.Attrs(param.CommandLineOnly|param.DontShowInStdUsage),
	)

	ps.AddFinalCheck(func() error {
		return showVersion(ps.StdW())
	})

	return nil
}
