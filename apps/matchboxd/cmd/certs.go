package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"crackerboxd/pkg/pki"

	"github.com/spf13/cobra"
)

var (
	certsCertDir    string
	certsHostnames  []string
	certsForce      bool
)

var certsCmd = &cobra.Command{
	Use:   "certs",
	Short: "Manage mTLS certificates",
	Long:  `Manage the mutual TLS certificates used for daemon-manager communication.`,
}

var certsGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a full PKI bundle (CA, server, client certificates)",
	Long: `Generates a complete PKI bundle for the Crackerbox cluster:
  - CA certificate and key
  - Server certificate for crackerboxd
  - Client certificate for crackerbox-manager

The certificates are written to the specified directory (default: /etc/crackerbox/pki).
This command is safe to run multiple times — it will skip generation if certs already exist.

Examples:
  crackerboxd certs generate
  crackerboxd certs generate --cert-dir /opt/crackerbox/pki
  crackerboxd certs generate --hostnames "myhost.local,192.168.1.100"
  crackerboxd certs generate --force`,

	RunE: func(cmd *cobra.Command, args []string) error {
		if pki.CertsExist(certsCertDir) && !certsForce {
			fmt.Printf("✅ PKI certificates already exist at %s\n", certsCertDir)
			fmt.Println("   Use --force to regenerate")
			return nil
		}

		fmt.Printf("🔐 Generating PKI bundle at %s...\n", certsCertDir)
		fmt.Printf("   Server SANs: %v\n", certsHostnames)

		if err := pki.GenerateFullPKI(certsCertDir, certsHostnames...); err != nil {
			return fmt.Errorf("generate PKI: %w", err)
		}

		fmt.Println()
		fmt.Println("   ✅ CA certificate:      " + filepath.Join(certsCertDir, pki.CACertFile))
		fmt.Println("   ✅ CA key:              " + filepath.Join(certsCertDir, pki.CAKeyFile))
		fmt.Println("   ✅ Server certificate:  " + filepath.Join(certsCertDir, pki.ServerCertFile))
		fmt.Println("   ✅ Server key:          " + filepath.Join(certsCertDir, pki.ServerKeyFile))
		fmt.Println("   ✅ Client certificate:  " + filepath.Join(certsCertDir, pki.ClientCertFile))
		fmt.Println("   ✅ Client key:          " + filepath.Join(certsCertDir, pki.ClientKeyFile))
		fmt.Println()
		fmt.Println("   Copy the client cert + key + CA cert to the manager machine.")
		return nil
	},
}

var certsInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Display information about the current certificates",
	Long:  `Shows the paths and existence status of all PKI files.`,
	Run: func(cmd *cobra.Command, args []string) {
		files := []struct {
			name string
			path string
		}{
			{"CA Certificate", filepath.Join(certsCertDir, pki.CACertFile)},
			{"CA Key", filepath.Join(certsCertDir, pki.CAKeyFile)},
			{"Server Certificate", filepath.Join(certsCertDir, pki.ServerCertFile)},
			{"Server Key", filepath.Join(certsCertDir, pki.ServerKeyFile)},
			{"Client Certificate", filepath.Join(certsCertDir, pki.ClientCertFile)},
			{"Client Key", filepath.Join(certsCertDir, pki.ClientKeyFile)},
		}

		fmt.Printf("🔐 PKI Certificate Directory: %s\n\n", certsCertDir)

		allExist := true
		for _, f := range files {
			status := "✅"
			if _, err := os.Stat(f.path); os.IsNotExist(err) {
				status = "❌"
				allExist = false
			}
			fmt.Printf("   %s %-25s %s\n", status, f.name, f.path)
		}

		fmt.Println()
		if allExist {
			fmt.Println("   All certificates are present.")
		} else {
			fmt.Println("   Some certificates are missing. Run 'crackerboxd certs generate' to create them.")
		}
	},
}

func init() {
	certsCmd.PersistentFlags().StringVar(&certsCertDir, "cert-dir", pki.DefaultCertDir, "Certificate directory")

	certsGenerateCmd.Flags().StringSliceVar(&certsHostnames, "hostnames", []string{"localhost", "127.0.0.1"}, "Server certificate SANs (hostnames and IPs)")
	certsGenerateCmd.Flags().BoolVar(&certsForce, "force", false, "Force regeneration of existing certificates")

	certsCmd.AddCommand(certsGenerateCmd)
	certsCmd.AddCommand(certsInfoCmd)
	rootCmd.AddCommand(certsCmd)
}
