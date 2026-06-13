package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) searchCmd() *cobra.Command {
	var (
		sort    string
		country string
		typ     string
	)
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search for manga and comics by keyword",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			n := a.effectiveLimit(20)
			a.progressf("searching for %q...", args[0])
			comics, err := a.client.Search(cmd.Context(), args[0], n, sort, country, typ)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(comics, len(comics))
		},
	}
	cmd.Flags().StringVar(&sort, "sort", "view", "sort order: view, new, trending, follow, rating")
	cmd.Flags().StringVar(&country, "country", "", "filter by country code (jp, kr, cn, us, ...)")
	cmd.Flags().StringVar(&typ, "type", "", "filter by type: comic, novel, oneshot")
	return cmd
}
