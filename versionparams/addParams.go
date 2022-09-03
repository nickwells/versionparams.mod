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

var vsnPart []string

const (
	vpGoVersion = "go-version"
	vpPath      = "path"
	vpMain      = "main"
	vpMods      = "modules"
	vpSettings  = "build-settings"

	noBuildInfo = "Build information not available"

	replIntro   = "   "
	modTypeDep  = "D"
	modTypeRepl = "r"
	modTypeMain = "M"
)

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
func showGoVersion(bi *debug.BuildInfo, w io.Writer) error {
	if bi == nil {
		return errors.New("Could not show the Go Version - no build info")
	}
	fmt.Fprintln(w, "Built with Go Version: ", bi.GoVersion)
	return nil
}

// showPath shows the Path of this executable
func showPath(bi *debug.BuildInfo, w io.Writer) error {
	if bi == nil {
		return errors.New("Could not show the Path - no build info")
	}
	fmt.Fprintln(w, "Path: ", bi.Path)
	return nil
}

// showMain shows the details of the main module of this executable
func showMain(bi *debug.BuildInfo, w io.Writer) error {
	if bi == nil {
		return errors.New("Could not show the modules - no build info")
	}

	fmt.Fprintln(w, bi.Main.Path, bi.Main.Version, bi.Main.Sum)

	return nil
}

// showModules shows the module details of this executable
func showModules(bi *debug.BuildInfo, w io.Writer) error {
	if bi == nil {
		return errors.New("Could not show the modules - no build info")
	}

	fmt.Fprintln(w, "Modules:")

	cols := depCols(bi)
	rpt := col.NewReport(nil, w,
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

	return nil
}

// showSettings shows the environment settings of this executable
func showSettings(bi *debug.BuildInfo, w io.Writer) error {
	if bi == nil {
		return errors.New("Could not show the build settings - no build info")
	}
	fmt.Fprintln(w, "Build Settings:")
	maxKey := 0
	for _, s := range bi.Settings {
		if len(s.Key) > maxKey {
			maxKey = len(s.Key)
		}
	}
	rpt := col.NewReport(nil, w,
		col.New(&colfmt.String{StrJust: col.Right, W: maxKey},
			"Key"),
		col.New(&colfmt.String{}, "Value"))
	fmt.Fprintln(w, "Build Settings:")
	for _, s := range bi.Settings {
		_ = rpt.PrintRow(s.Key, s.Value)
	}
	return nil
}

// showVersionPart shows the selected part of the version details
func showVersionPart(w io.Writer) error {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return errors.New(noBuildInfo)
	}

	var err error
	partShown := false
	for _, part := range vsnPart {
		switch part {
		case vpGoVersion:
			err = showGoVersion(bi, w)
		case vpPath:
			err = showPath(bi, w)
		case vpMain:
			err = showMain(bi, w)
		case vpMods:
			err = showModules(bi, w)
		case vpSettings:
			err = showSettings(bi, w)
		default:
			badVsnPart := "bad version part: " + part
			err = errors.New(badVsnPart)
			fmt.Fprintln(w, badVsnPart)
		}
		partShown = true
	}
	if partShown {
		os.Exit(0)
	}
	return err
}

// AddParams will add parameters to the passed param.PSet
func AddParams(ps *param.PSet) error {
	ps.Add("version", psetter.Nil{},
		"show the version details for this program",
		param.PostAction(
			func(_ location.L, p *param.ByName, _ []string) error {
				if bi, ok := debug.ReadBuildInfo(); !ok {
					fmt.Fprintln(p.StdWriter(), noBuildInfo)
				} else {
					fmt.Fprintln(p.StdWriter(), bi.String())
				}
				os.Exit(0)
				return nil
			}),
		param.Attrs(param.CommandLineOnly|param.DontShowInStdUsage),
	)

	ps.Add("version-part",
		psetter.EnumList{
			Value: &vsnPart,
			AllowedVals: psetter.AllowedVals{
				vpGoVersion: "show the version of Go that produced this binary",
				vpPath:      "show the path of the main package",
				vpMain:      "show the version of the main module",
				vpMods:      "show the module dependencies",
				vpSettings:  "show other information about the build",
			},
		},
		"show the named part of the version",
		param.Attrs(param.CommandLineOnly|param.DontShowInStdUsage),
	)

	ps.AddFinalCheck(func() error {
		return showVersionPart(ps.StdW())
	})

	return nil
}
