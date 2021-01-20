package versionparams

import (
	"fmt"

	"github.com/nickwells/location.mod/location"
	"github.com/nickwells/param.mod/v5/param"
	"github.com/nickwells/param.mod/v5/param/paction"
	"github.com/nickwells/param.mod/v5/param/psetter"
	"github.com/nickwells/version.mod/version"
)

// AddParams will add parameters to the passed param.PSet
func AddParams(ps *param.PSet) error {
	ps.Add("version", psetter.Nil{},
		"show the version details for this program",
		param.PostAction(paction.ReportAndExit(version.All()+"\n")),
		param.Attrs(param.CommandLineOnly|param.DontShowInStdUsage),
	)

	var vsnPart string
	const (
		vpTag       = "tag"
		vpCommit    = "commit"
		vpAuthor    = "author"
		vpDate      = "date"
		vpBuildDate = "build-date"
		vpBuiltBy   = "built-by"
	)
	ps.Add("version-part",
		psetter.Enum{
			Value: &vsnPart,
			AllowedVals: psetter.AllowedVals{
				vpTag:       "show the version tag",
				vpCommit:    "show the latest commit ID",
				vpAuthor:    "show the author of the latest commit",
				vpDate:      "show the date of the latest commit",
				vpBuildDate: "show the build date",
				vpBuiltBy:   "show the name of the user who built the program",
			},
			AllowInvalidInitialValue: true,
		},
		"show the named part of the version",
		param.PostAction(
			func(_ location.L, p *param.ByName, _ []string) error {
				switch vsnPart {
				case vpTag:
					fmt.Fprintln(p.StdWriter(), version.Tag())
				case vpCommit:
					fmt.Fprintln(p.StdWriter(), version.Commit())
				case vpAuthor:
					fmt.Fprintln(p.StdWriter(), version.Author())
				case vpDate:
					fmt.Fprintln(p.StdWriter(), version.Date())
				case vpBuildDate:
					fmt.Fprintln(p.StdWriter(), version.BuildDate())
				case vpBuiltBy:
					fmt.Fprintln(p.StdWriter(), version.BuildUser())
				default:
					fmt.Fprintln(p.StdWriter(), "bad version part: "+vsnPart)
				}
				return nil
			}),
		param.Attrs(param.CommandLineOnly|param.DontShowInStdUsage),
	)

	return nil
}
