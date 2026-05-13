package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show cluster status",
	Long:  `Displays the current status of the Crackerbox cluster including registered nodes and running VMs.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔥 Crackerbox Cluster Status")
		fmt.Println("════════════════════════════════════")
		fmt.Printf("  Version:    %s\n", Version)
		fmt.Printf("  Nodes:      %d\n", 0)
		fmt.Printf("  VMs:        %d\n", 0)
		fmt.Printf("  Status:     ● Not Connected\n")
		fmt.Println("════════════════════════════════════")
		fmt.Println()
		fmt.Println("  Hint: Start the manager with 'crackerbox-manager serve'")
	},
}

var nodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "List registered daemon nodes",
	Long:  `Lists all crackerboxd nodes registered with this cluster manager.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("No nodes registered yet.")
		fmt.Println()
		fmt.Println("To register a node, start crackerboxd on the target machine")
		fmt.Println("and point it to this manager's address.")
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(nodesCmd)
}
