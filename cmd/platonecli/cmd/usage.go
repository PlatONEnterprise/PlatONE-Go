// This is referenced from ./platone/usage.go with tiny modification (line: 198)
package cmd

import (
	"io"
	"sort"
	"strings"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"gopkg.in/urfave/cli.v1"
)

// AppHelpTemplate is the test template for the default, global app help topic.
var AppHelpTemplate = `NAME:
   {{.App.Name}} - {{.App.Usage}}

USAGE:
   {{.App.HelpName}} [options]{{if .App.Commands}} command [command options]{{end}} {{if .App.ArgsUsage}}{{.App.ArgsUsage}}{{else}}[arguments...]{{end}}
   {{if .App.Version}}
VERSION:
   {{.App.Version}}
   {{end}}{{if len .App.Authors}}
AUTHOR(S):
   {{range .App.Authors}}{{ . }}{{end}}
   {{end}}{{if .App.Commands}}
COMMANDS:
   {{range .App.Commands}}{{join .Names ", "}}{{ "\t" }}{{.Usage}}
   {{end}}{{end}}{{if .FlagGroups}}
{{range .FlagGroups}}{{.Name}} OPTIONS:
  {{range .Flags}}{{.}}
  {{end}}
{{end}}{{end}}{{if .App.Copyright }}
COPYRIGHT:
   {{.App.Copyright}}
   {{end}}
`

// flagGroup is a collection of flags belonging to a single topic.
type flagGroup struct {
	Name  string
	Flags []cli.Flag
}

// AppHelpFlagGroups is the application flags, grouped by functionality.
var AppHelpFlagGroups = []flagGroup{
	{
		Name: "GLOBAL",
		Flags: []cli.Flag{
			UrlFlags,
			AccountFlags,
			GasFlags,
			GasPriceFlags,
			LocalFlags,
			KeyfileFlags,
			SyncFlags,
			DefaultFlags,
		},
	},
	{
		Name:  "COMMON",
		Flags: []cli.Flag{},
	},
	{
		Name: "ACCOUNT",
		Flags: []cli.Flag{
			UserRemarkFlags,
			TelFlags,
			EmailFlags,
			OrganizationFlags,
			AddressFlags,
			UserIDFlags,
			UserRoleFlag,
			RolesFlag,
		},
	},
	{
		Name: "ADMIN",
		Flags: []cli.Flag{
			NodeDescFlags,
			NodeDelayNumFlags,
			NodeTypeFlags,
			NodeP2pPortFlags,
			NodeRpcPortFlags,
			NodePublicKeyFlags,
			NameFlags,

			AdminApproveFlags,
			AdminDeleteFlags,
		},
	},
	{
		Name: "CONTRACT",
		Flags: []cli.Flag{
			ContractParamFlag,
			ContractIDFlag,
			ContractAbiFilePathFlag,
			ContractVmFlags,
			TransferValueFlag,
			ShowContractMethodsFlag,
		},
	},
	{
		Name: "CNS",
		Flags: []cli.Flag{
			CnsVersionFlags,
		},
	},
	{
		Name: "FIREWALL",
		Flags: []cli.Flag{
			FilePathFlags,
			FwActionFlags,
		},
	},
	{
		Name: "SYSTEM_CONFIG",
		Flags: []cli.Flag{
			BlockGasLimitFlags,
			TxGasLimitFlags,
			IsTxUseGasFlags,
			IsApproveDeployedContractFlags,
			IsCheckContractDeployPermissionFlags,
			IsProduceEmptyBlockFlags,
			GasContractNameFlags,
		},
	},
	{
		Name: "MISC",
		// list the flags that are not categorized
		// ShowAllFlags,
		// FwClearAllFlags,
		// NodeStatusFlags,
		// UserStatusFlag,
		// PageNumFlags,
		// PageSizeFlags,
	},
}

// byCategory sorts an array of flagGroup by Name in the order
// defined in AppHelpFlagGroups.
type byCategory []flagGroup

func (a byCategory) Len() int      { return len(a) }
func (a byCategory) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byCategory) Less(i, j int) bool {
	iCat, jCat := a[i].Name, a[j].Name
	iIdx, jIdx := len(AppHelpFlagGroups), len(AppHelpFlagGroups) // ensure non categorized flags come last

	for i, group := range AppHelpFlagGroups {
		if iCat == group.Name {
			iIdx = i
		}
		if jCat == group.Name {
			jIdx = i
		}
	}

	return iIdx < jIdx
}

func flagCategory(flag cli.Flag) string {
	for _, category := range AppHelpFlagGroups {
		for _, flg := range category.Flags {
			if flg.GetName() == flag.GetName() {
				return category.Name
			}
		}
	}
	return "MISC"
}

func init() {
	// Override the default app help template
	cli.AppHelpTemplate = AppHelpTemplate

	// Define a one shot struct to pass to the usage template
	type helpData struct {
		App        interface{}
		FlagGroups []flagGroup
	}

	// Override the default app help printer, but only for the global app help
	originalHelpPrinter := cli.HelpPrinter
	cli.HelpPrinter = func(w io.Writer, tmpl string, data interface{}) {
		if tmpl == AppHelpTemplate {
			// Iterate over all the flags and add any uncategorized ones
			categorized := make(map[string]struct{})
			for _, group := range AppHelpFlagGroups {
				for _, flag := range group.Flags {
					categorized[flag.String()] = struct{}{}
				}
			}
			uncategorized := []cli.Flag{}
			for _, flag := range data.(*cli.App).Flags {
				if _, ok := categorized[flag.String()]; !ok {
					if strings.HasPrefix(flag.GetName(), "dashboard") {
						continue
					}
					uncategorized = append(uncategorized, flag)
				}
			}
			if len(uncategorized) > 0 {
				// Append all ungategorized options to the misc group
				miscs := len(AppHelpFlagGroups[len(AppHelpFlagGroups)-1].Flags)
				AppHelpFlagGroups[len(AppHelpFlagGroups)-1].Flags = append(AppHelpFlagGroups[len(AppHelpFlagGroups)-1].Flags, uncategorized...)

				// Make sure they are removed afterwards
				defer func() {
					AppHelpFlagGroups[len(AppHelpFlagGroups)-1].Flags = AppHelpFlagGroups[len(AppHelpFlagGroups)-1].Flags[:miscs]
				}()
			}

			// new added
			newHelpFlagGroups := append([]flagGroup{}, AppHelpFlagGroups[0], AppHelpFlagGroups[len(AppHelpFlagGroups)-1])

			// Render out custom usage screen
			originalHelpPrinter(w, tmpl, helpData{data, newHelpFlagGroups})
		} else if tmpl == utils.CommandHelpTemplate {
			// Iterate over all command specific flags and categorize them
			categorized := make(map[string][]cli.Flag)
			for _, flag := range data.(cli.Command).Flags {
				if _, ok := categorized[flag.String()]; !ok {
					categorized[flagCategory(flag)] = append(categorized[flagCategory(flag)], flag)
				}
			}

			// sort to get a stable ordering
			sorted := make([]flagGroup, 0, len(categorized))
			for cat, flgs := range categorized {
				sorted = append(sorted, flagGroup{cat, flgs})
			}
			sort.Sort(byCategory(sorted))

			// add sorted array to data and render with default printer
			originalHelpPrinter(w, tmpl, map[string]interface{}{
				"cmd":              data,
				"categorizedFlags": sorted,
			})
		} else {
			originalHelpPrinter(w, tmpl, data)
		}
	}
}
