package cmd

import (
	"fmt"

	"github.com/karozadev/KzNginxGenerator/internal/server"
	"github.com/spf13/cobra"
)

var uiPort int

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Lance l'interface web locale de KzNginxGenerator",
	Long: `Démarre un serveur HTTP local servant l'interface web de
KzNginxGenerator : construction visuelle de la configuration
(upstreams, locations, SSL, ...) avec génération et aperçu en temps réel.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		addr := fmt.Sprintf(":%d", uiPort)
		srv := server.New(server.Options{Addr: addr, Version: Version, Revision: GitCommit})
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "KzNginxGenerator UI disponible sur http://localhost%s\n", addr); err != nil {
			return err
		}
		return srv.ListenAndServe()
	},
}

func init() {
	uiCmd.Flags().IntVar(&uiPort, "port", 8080, "port d'écoute du serveur web")
	rootCmd.AddCommand(uiCmd)
}
