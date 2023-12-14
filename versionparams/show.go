package versionparams

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"github.com/nickwells/col.mod/v4/col"
	"github.com/nickwells/col.mod/v4/colfmt"
)

const replIntro = "   "

// showPrompt prints the prompt if the shortDisplay flag is not set
func showPrompt(w io.Writer, prompt string) {
	if vsn.shortDisplay {
		return
	}
	fmt.Fprint(w, prompt)
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

	cols := []*col.Col{
		col.New(&colfmt.String{W: uint(colWidthPath)}, "Path"),
		col.New(&colfmt.String{W: uint(colWidthVersion)}, "Version"),
	}
	if vsn.showChecksum {
		cols = append(cols,
			col.New(&colfmt.String{W: uint(colWidthSum)}, "CheckSum"))
	}

	return cols
}

// showGoVersion shows the Go Version used to build this executable
func showGoVersion(w io.Writer, bi *debug.BuildInfo) {
	showPrompt(w, "Built with Go Version: ")

	fmt.Fprintln(w, bi.GoVersion)
}

// showPath shows the Path of this executable
func showPath(w io.Writer, bi *debug.BuildInfo) {
	showPrompt(w, "Path: ")

	fmt.Fprintln(w, bi.Path)
}

// showMain shows the details of the main module of this executable
func showMain(w io.Writer, bi *debug.BuildInfo) {
	prompt := "Version"

	sum := bi.Main.Sum
	if !vsn.showChecksum {
		sum = ""
	}
	if sum != "" {
		prompt += ", Checksum"
		sum = ", " + sum
	}
	prompt += ": "
	showPrompt(w, prompt)

	fmt.Fprintln(w, bi.Main.Version+sum)
}

// printModRow prints the report row for the supplied module taking into
// account the showChecksum value
func printModRow(rpt *col.Report, mType, path string, m *debug.Module) {
	if vsn.showChecksum {
		_ = rpt.PrintRow(mType, path, m.Version, m.Sum)
	} else {
		_ = rpt.PrintRow(mType, path, m.Version)
	}
}

// showModules shows the module details of this executable
func showModules(w io.Writer, bi *debug.BuildInfo) {
	const (
		modTypeDep  = "D"
		modTypeRepl = "r"
		modTypeMain = "M"
	)

	showPrompt(w, "Modules:\n")

	hdr := col.NewHeaderOrPanic()
	if vsn.shortDisplay {
		col.HdrOptDontPrint(hdr)
	}

	rpt := col.NewReportOrPanic(hdr, w,
		col.New(&colfmt.String{}, "Type"), depCols(bi)...)

	if vsn.modFilts.Passes(bi.Main.Path) {
		printModRow(rpt, modTypeMain, bi.Main.Path, &bi.Main)
	}

	for _, m := range bi.Deps {
		if !vsn.modFilts.Passes(m.Path) {
			continue
		}

		modType := modTypeDep
		if m.Replace != nil {
			modType = modTypeRepl
		}

		printModRow(rpt, modType, m.Path, m)

		if m.Replace != nil {
			m := m.Replace
			path := replIntro + m.Path
			printModRow(rpt, modType, path, m)
		}
	}
}

// showSettings shows the build settings of this executable
func showSettings(w io.Writer, bi *debug.BuildInfo) {
	showPrompt(w, "Build Settings:\n")

	maxKey := 0
	for _, s := range bi.Settings {
		if !vsn.bldFilts.Passes(s.Key) {
			continue
		}
		if len(s.Key) > maxKey {
			maxKey = len(s.Key)
		}
	}

	keyJust := col.Right

	hdr := col.NewHeaderOrPanic()
	if vsn.shortDisplay {
		col.HdrOptDontPrint(hdr)
		keyJust = col.Left
	}

	rpt := col.NewReportOrPanic(hdr, w,
		col.New(&colfmt.String{StrJust: keyJust, W: uint(maxKey)}, "Key"),
		col.New(&colfmt.String{}, "Value"))

	for _, s := range bi.Settings {
		if !vsn.bldFilts.Passes(s.Key) {
			continue
		}
		_ = rpt.PrintRow(s.Key, s.Value)
	}
}

// showRaw shows the build info in raw form
func showRaw(w io.Writer, bi *debug.BuildInfo) {
	fmt.Fprintln(w, bi)
}

// showVersion shows the version details, if specific parts have been
// requested then just those parts are shown otherwise the full default
// version info is displayed. Note that unless there is a problem retrieving
// the BuildInfo or an unknown part is requested then this will exit with
// status 0.
func showVersion(w io.Writer) error {
	type versionPartShowFunc func(io.Writer, *debug.BuildInfo)

	vsnPartShowFuncMap := map[vsnPartName]versionPartShowFunc{
		vpGoVsn:    showGoVersion,
		vpPath:     showPath,
		vpMain:     showMain,
		vpMods:     showModules,
		vpSettings: showSettings,
		vpRaw:      showRaw,
	}

	if len(vsn.parts) == 0 {
		return nil
	}

	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return errors.New("Build information not available")
	}

	shown := make(map[vsnPartName]bool, len(vsn.parts))
	for _, part := range vsn.parts {
		if shown[part] {
			continue
		}
		shown[part] = true

		var f versionPartShowFunc
		var ok bool
		if f, ok = vsnPartShowFuncMap[part]; !ok {
			return errors.New("bad version part: " + string(part))
		}

		f(w, bi)
	}

	os.Exit(0)
	return nil
}
