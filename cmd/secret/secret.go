package secret

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/zarigata/budgie/internal/cmdutil"
	"github.com/zarigata/budgie/internal/secrets"
)

var secretCmd = &cobra.Command{
	Use:   "secret",
	Short: "Manage secrets",
	Long:  `Manage encrypted secrets for use in containers.`,
}

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a secret from standard input",
	Long: `Create a new encrypted secret. The secret value is read from standard input.

Examples:
  echo "mysecretvalue" | budgie secret create my-secret
  budgie secret create db-password < password.txt`,
	Args: cobra.ExactArgs(1),
	RunE: createSecret,
}

var lsCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List secrets",
	RunE:    listSecrets,
}

var rmCmd = &cobra.Command{
	Use:     "rm <name>",
	Aliases: []string{"remove"},
	Short:   "Remove a secret",
	Args:    cobra.ExactArgs(1),
	RunE:    removeSecret,
}

var inspectCmd = &cobra.Command{
	Use:   "inspect <name>",
	Short: "Display secret metadata (not the value)",
	Args:  cobra.ExactArgs(1),
	RunE:  inspectSecret,
}

func createSecret(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Validate secret name
	if strings.ContainsAny(name, " \t\n") {
		return fmt.Errorf("secret name cannot contain whitespace")
	}

	dataDir := cmdutil.GetDataDir()
	sm, err := secrets.NewSecretManager(dataDir)
	if err != nil {
		return fmt.Errorf("failed to initialize secret manager: %w", err)
	}

	// Read secret value from stdin
	reader := bufio.NewReader(os.Stdin)
	data, err := reader.ReadBytes('\n')
	if err != nil && err.Error() != "EOF" {
		// Try reading without newline delimiter
		data, err = os.ReadFile("/dev/stdin")
		if err != nil {
			return fmt.Errorf("failed to read secret from stdin: %w", err)
		}
	}

	// Trim trailing newline
	data = []byte(strings.TrimSuffix(string(data), "\n"))

	if len(data) == 0 {
		return fmt.Errorf("secret value cannot be empty")
	}

	secret, err := sm.CreateSecret(name, data)
	if err != nil {
		return err
	}

	fmt.Println(secret.ID)
	return nil
}

func listSecrets(cmd *cobra.Command, args []string) error {
	dataDir := cmdutil.GetDataDir()
	sm, err := secrets.NewSecretManager(dataDir)
	if err != nil {
		return fmt.Errorf("failed to initialize secret manager: %w", err)
	}

	secretList := sm.ListSecrets()

	if len(secretList) == 0 {
		fmt.Println("No secrets found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tCREATED\tUPDATED")

	for _, s := range secretList {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			s.ID,
			s.Name,
			formatTime(s.CreatedAt),
			formatTime(s.UpdatedAt))
	}

	return w.Flush()
}

func removeSecret(cmd *cobra.Command, args []string) error {
	name := args[0]

	dataDir := cmdutil.GetDataDir()
	sm, err := secrets.NewSecretManager(dataDir)
	if err != nil {
		return fmt.Errorf("failed to initialize secret manager: %w", err)
	}

	if err := sm.RemoveSecret(name); err != nil {
		return err
	}

	fmt.Println(name)
	return nil
}

func inspectSecret(cmd *cobra.Command, args []string) error {
	name := args[0]

	dataDir := cmdutil.GetDataDir()
	sm, err := secrets.NewSecretManager(dataDir)
	if err != nil {
		return fmt.Errorf("failed to initialize secret manager: %w", err)
	}

	secretList := sm.ListSecrets()
	var found *secrets.SecretInfo
	for _, s := range secretList {
		if s.Name == name {
			found = s
			break
		}
	}

	if found == nil {
		return fmt.Errorf("secret not found: %s", name)
	}

	fmt.Printf("ID: %s\n", found.ID)
	fmt.Printf("Name: %s\n", found.Name)
	fmt.Printf("Created: %s\n", found.CreatedAt.Format(time.RFC3339))
	fmt.Printf("Updated: %s\n", found.UpdatedAt.Format(time.RFC3339))

	return nil
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "N/A"
	}
	duration := time.Since(t)
	if duration < time.Hour {
		return fmt.Sprintf("%d minutes ago", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%d hours ago", int(duration.Hours()))
	}
	return fmt.Sprintf("%d days ago", int(duration.Hours()/24))
}

func GetSecretCmd() *cobra.Command {
	return secretCmd
}

func init() {
	secretCmd.AddCommand(createCmd)
	secretCmd.AddCommand(lsCmd)
	secretCmd.AddCommand(rmCmd)
	secretCmd.AddCommand(inspectCmd)
}
