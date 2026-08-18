package cmd

import (
	"fmt"
	"os"

	"github.com/karozadev/KzNginxGenerator/internal/nginx"
	"github.com/spf13/cobra"
)

var (
	genDomain    string
	genProxy     string
	genOut       string
	genSSLCert   string
	genSSLKey    string
	genWebSocket bool
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Génère une configuration Nginx en ligne de commande",
	Long: `Génère une configuration Nginx de reverse proxy pour un domaine et un
backend donnés, écrite sur stdout ou dans le fichier indiqué par --out.

Exemple :
  kznginx generate --domain example.com --proxy http://localhost:8000 \
    --out /etc/nginx/sites-available/karoza-app.conf`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if genDomain == "" {
			return fmt.Errorf("--domain est requis")
		}
		if genProxy == "" {
			return fmt.Errorf("--proxy est requis")
		}

		cfg := nginx.NewSimpleReverseProxyConfig(nginx.SimpleReverseProxyOptions{
			Domain:                genDomain,
			ProxyPass:             genProxy,
			WebSocket:             genWebSocket,
			SSLCertificatePath:    genSSLCert,
			SSLCertificateKeyPath: genSSLKey,
		})

		output, err := nginx.Render(cfg)
		if err != nil {
			return fmt.Errorf("échec de la génération: %w", err)
		}

		if genOut == "" {
			_, err := fmt.Fprint(cmd.OutOrStdout(), output)
			return err
		}

		if err := os.WriteFile(genOut, []byte(output), 0o644); err != nil {
			return fmt.Errorf("écriture de %s: %w", genOut, err)
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Configuration écrite dans %s\n", genOut)
		return err
	},
}

func init() {
	generateCmd.Flags().StringVar(&genDomain, "domain", "", "nom de domaine du virtual host (requis)")
	generateCmd.Flags().StringVar(&genProxy, "proxy", "", "URL du backend à reverse-proxifier (requis)")
	generateCmd.Flags().StringVar(&genOut, "out", "", "fichier de sortie (stdout par défaut)")
	generateCmd.Flags().StringVar(&genSSLCert, "ssl-cert", "", "chemin du certificat SSL (active HTTPS)")
	generateCmd.Flags().StringVar(&genSSLKey, "ssl-key", "", "chemin de la clé privée SSL (active HTTPS)")
	generateCmd.Flags().BoolVar(&genWebSocket, "websocket", false, "active le support WebSocket sur la location racine")
	rootCmd.AddCommand(generateCmd)
}
