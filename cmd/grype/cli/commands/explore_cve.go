package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/anchore/clio"
	"github.com/anchore/grype/cmd/grype/cli/options"
	"github.com/anchore/grype/grype"
	"github.com/anchore/grype/grype/db"
	"github.com/anchore/grype/grype/store"
	"github.com/anchore/grype/grype/vulnerability"
	"github.com/anchore/grype/internal/bus"
	"github.com/anchore/grype/internal/log"
	"github.com/hashicorp/go-multierror"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type cveExploreOptions struct {
	Output string           `yaml:"output" json:"output" mapstructure:"output"`
	cveID  string           `yaml:"cveID" json:"cveID" mapstructure:"cveID"`
	DB     options.Database `yaml:"db" json:"db" mapstructure:"db"`
}

var _ clio.FlagAdder = (*cveExploreOptions)(nil)

func (c *cveExploreOptions) AddFlags(flags clio.FlagSet) {
	flags.StringVarP(&c.Output, "output", "o", "format to display results (available=[table, json])")
}

func ExploreCVE(app clio.Application) *cobra.Command {
	opts := &cveExploreOptions{
		Output: "table",
		DB:     options.DefaultDatabase(app.ID()),
	}

	return app.SetupCommand(&cobra.Command{
		Use:   "cve [flags] cve_id",
		Short: "explore a vulnerability and display information",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) (err error) {
			var cveID string
			cveID = args[0]
			return runExploreCVE(opts, cveID)
		},
	}, opts)
}

func runExploreCVE(opts *cveExploreOptions, cveID string) (errs error) {
	var str *store.Store
	var status *db.Status
	var dbCloser *db.Closer

	err := parallel(
		func() (err error) {
			log.Debug("loading DB")
			str, status, dbCloser, err = grype.LoadVulnerabilityDB(opts.DB.ToCuratorConfig(), opts.DB.AutoUpdate)
			return validateDBLoad(err, status)
		},
	)

	if err != nil {
		return err
	}

	if dbCloser != nil {
		defer dbCloser.Close()
	}

	vulnerabilities, err := str.Get(cveID, "")
	if err != nil {
		return err
	}

	sb := &strings.Builder{}
	if len(vulnerabilities) == 0 {
		sb.WriteString("CVE doesn't exist in the DB\n")
	} else {
		err := present(opts.Output, vulnerabilities, sb)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	bus.Report(sb.String())

	return errs
}

func present(outputFormat string, vulnerabilities []vulnerability.Vulnerability, output io.Writer) error {
	if vulnerabilities == nil {
		return nil
	}

	switch outputFormat {
	case "table":
		rows := [][]string{}
		for _, v := range vulnerabilities {
			rows = append(rows, []string{v.ID, v.PackageName, v.Namespace, v.Constraint.String()})
		}

		table := tablewriter.NewWriter(output)
		columns := []string{"ID", "Package Name", "Namespace", "Version Constraint"}

		table.SetHeader(columns)
		table.SetAutoWrapText(false)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAlignment(tablewriter.ALIGN_LEFT)

		table.SetHeaderLine(false)
		table.SetBorder(false)
		table.SetAutoFormatHeaders(true)
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetRowSeparator("")
		table.SetTablePadding("  ")
		table.SetNoWhiteSpace(true)

		table.AppendBulk(rows)
		table.Render()
	case "json":
		enc := json.NewEncoder(output)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", " ")
		if err := enc.Encode(vulnerabilities); err != nil {
			return fmt.Errorf("failed to encode diff information: %+v", err)
		}
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
	return nil
}