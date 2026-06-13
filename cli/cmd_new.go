package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) newCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "new",
		Short: "List the newest manga and comics added to the catalogue",
		RunE: func(cmd *cobra.Command, _ []string) error {
			n := a.effectiveLimit(20)
			a.progressf("fetching newest comics...")
			comics, err := a.client.New(cmd.Context(), n)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(comics, len(comics))
		},
	}
}
