// Package cmd implements the kubectl-redroid cobra command tree.
package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

// Version is set at build time via -ldflags.
var Version = "dev"

// globalFlags holds flags shared by all subcommands.
type globalFlags struct {
	Kubeconfig string
	Context    string
	Namespace  string
}

var globals globalFlags

var rootCmd = &cobra.Command{
	Use:   "kubectl-redroid",
	Short: "Manage Redroid Android instances and tasks on Kubernetes",
	Long: `kubectl-redroid is a kubectl plugin for the redroid-operator.

Commands:
  instance list/describe/port-forward/adb/shell/logs/suspend/resume
  task     list/describe/trigger/logs`,
	Version:       Version,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&globals.Kubeconfig, "kubeconfig", "",
		"Path to kubeconfig file (defaults to $KUBECONFIG then ~/.kube/config)")
	rootCmd.PersistentFlags().StringVar(&globals.Context, "context", "",
		"Kubernetes context to use (defaults to current context)")
	rootCmd.PersistentFlags().StringVarP(&globals.Namespace, "namespace", "n", "",
		"Namespace to operate in (defaults to current context namespace)")

	rootCmd.AddCommand(newInstanceCmd())
	rootCmd.AddCommand(newTaskCmd())

	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

// loadingRules returns ClientConfigLoadingRules from global flags.
func loadingRules() *clientcmd.ClientConfigLoadingRules {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	if globals.Kubeconfig != "" {
		rules.ExplicitPath = globals.Kubeconfig
	}
	return rules
}

// configOverrides returns ConfigOverrides from global flags.
func configOverrides() *clientcmd.ConfigOverrides {
	o := &clientcmd.ConfigOverrides{}
	if globals.Context != "" {
		o.CurrentContext = globals.Context
	}
	return o
}

// resolvedNamespace returns the effective namespace: flag -> context default -> "default".
func resolvedNamespace() string {
	if globals.Namespace != "" {
		return globals.Namespace
	}
	ns, _, err := buildClientConfig().Namespace()
	if err != nil || ns == "" {
		return "default"
	}
	return ns
}

func buildClientConfig() clientcmd.ClientConfig {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules(),
		configOverrides(),
	)
}
