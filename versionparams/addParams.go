package versionparams

import (
	"github.com/nickwells/location.mod/location"
	"github.com/nickwells/param.mod/v6/param"
	"github.com/nickwells/param.mod/v6/psetter"
	"golang.org/x/exp/slices"
)

var vsn = struct {
	parts        []vsnPartName
	shortDisplay bool
	showChecksum bool

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

var (
	modFilterMap = map[string]bool{}
	bldFilterMap = map[string]bool{}
)

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
		paramNameVersionShowCkSum  = "version-show-checksum"
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
				vsn.parts = append(vsn.parts, vpMain)
				return nil
			}),
		param.SeeAlso(paramNameVersionPart),
		param.Attrs(param.CommandLineOnly|param.DontShowInStdUsage),
		param.GroupName(GroupName),
	)

	ps.Add(paramNameVersionPart,
		psetter.EnumList[vsnPartName]{
			Value: &vsn.parts,
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
		psetter.Bool{Value: &vsn.shortDisplay},
		"show the version parts in simplified form, without headings and"+
			" prompts. This is more useful if you want to use the value"+
			" as you won't need to strip out the other text. Note that"+
			" there is no short form of the '"+string(vpRaw)+"' form."+
			"\n\n"+
			"If this value is set and no version parts have been"+
			" chosen, the '"+string(vpMain)+"' part will be shown",
		param.AltNames("version-short", "version-s"),
		param.SeeAlso(paramNameVersionPart),
		param.Attrs(param.CommandLineOnly|param.DontShowInStdUsage),
		param.GroupName(GroupName),
	)

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

	ps.Add(paramNameVersionShowCkSum,
		psetter.Bool{Value: &vsn.showChecksum},
		"show module checksums."+
			" This only changes the appearance of the"+
			" '"+string(vpMain)+"' and '"+string(vpMods)+"' parts"+
			" as these are the only parts of the version information"+
			" that have associated checksums."+
			"\n\n"+
			"If this value is set and no version parts have been"+
			" chosen, the '"+string(vpMain)+"' part will be shown",
		param.AltNames("version-checksum", "version-cksum"),
		param.SeeAlso(paramNameVersionPart),
		param.Attrs(param.CommandLineOnly|param.DontShowInStdUsage),
		param.GroupName(GroupName),
	)

	addFinalChecks(ps)
	return nil
}

// addFinalChecks adds the final check functions
func addFinalChecks(ps *param.PSet) {
	filterErrCount := 0
	ps.AddFinalCheck(func() error {
		var errs []error
		vsn.modFilts, errs = makeFilterFromMap(modFilterMap)
		filterErrCount += len(errs)
		if len(errs) > 0 {
			ps.AddErr("Bad Version Module Path filters", errs...)
		}
		return nil
	})
	ps.AddFinalCheck(func() error {
		var errs []error
		vsn.bldFilts, errs = makeFilterFromMap(bldFilterMap)
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

		if vsn.modFilts.HasFilters() &&
			!slices.Contains(vsn.parts, vpMods) {
			vsn.parts = append(vsn.parts, vpMods)
		}

		if vsn.bldFilts.HasFilters() &&
			!slices.Contains(vsn.parts, vpSettings) {
			vsn.parts = append(vsn.parts, vpSettings)
		}

		if vsn.shortDisplay && len(vsn.parts) == 0 {
			vsn.parts = append(vsn.parts, vpMain)
		}

		if vsn.showChecksum && len(vsn.parts) == 0 {
			vsn.parts = append(vsn.parts, vpMain)
		}

		return showVersion(ps.StdW())
	})
}
