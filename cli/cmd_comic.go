package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) comicCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "comic <hid-or-slug>",
		Short: "Fetch full metadata for a single comic by hid or slug",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a.progressf("fetching comic %q...", args[0])
			comic, err := a.client.GetComic(cmd.Context(), args[0])
			if err != nil {
				return mapFetchErr(err)
			}
			return a.render(comic)
		},
	}
}
