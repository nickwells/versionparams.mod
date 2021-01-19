package versionparams

import (
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

	return nil
}
