package versionparams

import (
	"github.com/nickwells/location.mod/location"
	"github.com/nickwells/param.mod/v6/param"
	"github.com/nickwells/param.mod/v6/psetter"
	"golang.org/x/exp/slices"
)

var showVsn = struct {
	parts        []vsnPartName
	shortDisplay bool

	modFilts filter
	bldFilts filter
}{}

type vsnPartName string

const (
	vpGoVsn    vsnPartName = "go-version"
	vpPath     vsnPartName = "path"
	vpMain     vsnPartName = "main"
	vpMods     vsnPartName = "modules"
	vpSettings vsnPartName = "build-settings"
	vpRaw      vsnPartName = "raw"
)

const GroupName = "version-details"

// AddParams will add parameters to the passed param.PSet
func AddParams(ps *param.PSet) error {
	ps.AddGroup(GroupName,
		"parameters relating to program version information."+
			" You can control which parts of the version"+
			" information are shown.")

	const (
		paramNameVersion           = "version"
		paramNameVersionPart       = "version-part"
		paramNameVersionPartShort  = "version-part-short"
		paramNameVersionModuleFltr = "version-module-filter"
		paramNameVersionBuildFltr  = "version-build-filter"
	)

	fullParts := []vsnPartName{
		vpGoVsn,
		vpPath,
		vpMain,
		vpMods,
		vpSettings,
	}

	ps.Add(paramNameVersion, psetter.Nil{},
		"show the complete version details for this program"+
			" in the default format",
		param.PostAction(
			func(_ location.L, p *param.ByName, _ []string) error {
				showVsn.parts = append(showVsn.parts, vpMain)
				return nil
			}),
		param.SeeAlso(paramNameVersionPart),
		param.Attrs(param.CommandLineOnly|param.DontShowInStdUsage),
		param.GroupName(GroupName),
	)

	ps.Add(paramNameVersionPart,
		psetter.EnumList[vsnPartName]{
			Value: &showVsn.parts,
			AllowedVals: psetter.AllowedVals[vsnPartName]{
				vpGoVsn:    "show the Go version used to make the program",
				vpPath:     "show the path of the main package",
				vpMain:     "show the version of the main module",
				vpMods:     "show the module dependencies",
				vpSettings: "show the settings used to make the program",
				vpRaw:      "show the full build information",
			},
			Aliases: psetter.Aliases[vsnPartName]{
				"go":          []vsnPartName{vpGoVsn},
				"go-vsn":      []vsnPartName{vpGoVsn},
				"mods":        []vsnPartName{vpMods},
				"dep":         []vsnPartName{vpMods},
				"build-flags": []vsnPartName{vpSettings},
				"build":       []vsnPartName{vpSettings},
				"settings":    []vsnPartName{vpSettings},
				"all":         fullParts,
				"full":        fullParts,
			},
		},
		"show only the named parts of the version",
		param.AltNames("version-p"),
		param.SeeAlso(paramNameVersionPartShort),
		param.Attrs(param.CommandLineOnly|param.DontShowInStdUsage),
		param.GroupName(GroupName),
	)

	ps.Add(paramNameVersionPartShort,
		psetter.Bool{Value: &showVsn.shortDisplay},
		"show the version parts in simplified form, without headings and"+
			" prompts. This is more useful if you want to use the value"+
			" as you won't need to strip out the other text. Note that"+
			" there is no short form of the '"+string(vpRaw)+"' form."+
			"\n\n"+
			"If this value is set and no version parts have been"+
			" chosen, the "+string(vpMain)+" part will be shown",
		param.AltNames("version-short", "version-s"),
		param.SeeAlso(paramNameVersionPart),
		param.Attrs(param.CommandLineOnly|param.DontShowInStdUsage),
		param.GroupName(GroupName),
	)

	modFilterMap := map[string]bool{}
	ps.Add(paramNameVersionModuleFltr,
		psetter.Map[string]{Value: &modFilterMap},
		"only those module paths matching the given patterns will be"+
			" shown. Note that the patterns are Regular Expressions (RE)"+
			" not Glob patterns. It is an error if the RE will not compile."+
			" If there are no filters given then any value will match."+
			"\n\n"+
			"If the value ends in '=false' then the sense is reversed"+
			" and only module paths not matching the RE will be shown",
		param.AltNames("version-mod-fltr", "version-module", "version-mod"),
		param.SeeAlso(paramNameVersionPart),
		param.Attrs(param.CommandLineOnly|param.DontShowInStdUsage),
		param.GroupName(GroupName),
	)

	bldFilterMap := map[string]bool{}
	ps.Add(paramNameVersionBuildFltr,
		psetter.Map[string]{Value: &bldFilterMap},
		"only those Build Setting Keys matching the given patterns will be"+
			" shown. Note that the patterns are Regular Expressions (RE)"+
			" not Glob patterns. It is an error if the RE will not compile."+
			" If there are no filters given then any value will match."+
			"\n\n"+
			"If the value ends in '=false' then the sense is reversed"+
			" and only Build Setting Keys not matching the RE will be shown",
		param.AltNames("version-bld-fltr", "version-build-key", "version-bk"),
		param.SeeAlso(paramNameVersionPart),
		param.Attrs(param.CommandLineOnly|param.DontShowInStdUsage),
		param.GroupName(GroupName),
	)

	filterErrCount := 0
	ps.AddFinalCheck(func() error {
		var errs []error
		showVsn.modFilts, errs = makeFilterFromMap(modFilterMap)
		filterErrCount += len(errs)
		if len(errs) > 0 {
			ps.AddErr("Bad Version Module Path filters", errs...)
		}
		return nil
	})
	ps.AddFinalCheck(func() error {
		var errs []error
		showVsn.bldFilts, errs = makeFilterFromMap(bldFilterMap)
		filterErrCount += len(errs)
		if len(errs) > 0 {
			ps.AddErr("Bad Version Build Key filters", errs...)
		}
		return nil
	})

	ps.AddFinalCheck(func() error {
		if filterErrCount > 0 {
			return nil
		}

		if showVsn.modFilts.HasFilters() &&
			!slices.Contains(showVsn.parts, vpMods) {
			showVsn.parts = append(showVsn.parts, vpMods)
		}

		if showVsn.bldFilts.HasFilters() &&
			!slices.Contains(showVsn.parts, vpSettings) {
			showVsn.parts = append(showVsn.parts, vpSettings)
		}

		if showVsn.shortDisplay && len(showVsn.parts) == 0 {
			showVsn.parts = append(showVsn.parts, vpMain)
		}

		return showVersion(ps.StdW())
	})

	return nil
}
