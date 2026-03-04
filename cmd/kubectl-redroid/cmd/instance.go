package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	redroidv1alpha1 "github.com/isning/redroid-operator/api/v1alpha1"
)

func newInstanceCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "instance",
		Aliases: []string{"inst", "instances"},
		Short:   "Manage RedroidInstance resources",
	}
	c.AddCommand(
		newInstanceListCmd(),
		newInstanceGetCmd(),
		newInstancePortForwardCmd(),
		newInstanceADBCmd(),
		newInstanceShellCmd(),
		newInstanceLogsCmd(),
		newInstanceSuspendCmd(),
		newInstanceResumeCmd(),
	)
	return c
}

// ── list ─────────────────────────────────────────────────────────────────────

func newInstanceListCmd() *cobra.Command {
	var allNamespaces bool
	c := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "get"},
		Short:   "List RedroidInstances",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := buildClients()
			if err != nil {
				return err
			}

			ns := resolvedNamespace()
			if allNamespaces {
				ns = ""
			}

			list, err := cl.dynamic.Resource(instancesGVR).Namespace(ns).
				List(context.Background(), metav1.ListOptions{})
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			defer func() { _ = w.Flush() }()

			if allNamespaces {
				_, _ = fmt.Fprintln(w, "NAMESPACE\tNAME\tINDEX\tPHASE\tADB ADDRESS\tSUSPEND\tSUSPENDED\tAGE")
			} else {
				_, _ = fmt.Fprintln(w, "NAME\tINDEX\tPHASE\tADB ADDRESS\tSUSPEND\tSUSPENDED\tAGE")
			}

			for _, item := range list.Items {
				inst := &redroidv1alpha1.RedroidInstance{}
				if err := convertUnstructured(item.Object, inst); err != nil {
					continue
				}

				suspended := ""
				if inst.Status.Suspended != nil {
					suspended = inst.Status.Suspended.Actor
					if suspended == "" {
						suspended = "yes"
					}
				}

				age := humanDuration(time.Since(inst.CreationTimestamp.Time))
				adb := inst.Status.ADBAddress
				if adb == "" {
					adb = "<none>"
				}

				if allNamespaces {
					_, _ = fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%v\t%s\t%s\n",
						inst.Namespace, inst.Name, inst.Spec.Index,
						inst.Status.Phase, adb, inst.Spec.Suspend, suspended, age)
				} else {
					_, _ = fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%v\t%s\t%s\n",
						inst.Name, inst.Spec.Index,
						inst.Status.Phase, adb, inst.Spec.Suspend, suspended, age)
				}
			}
			return nil
		},
	}
	c.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "List across all namespaces")
	return c
}

// ── get ──────────────────────────────────────────────────────────────────────

func newInstanceGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "describe <name>",
		Short: "Show detailed information about a RedroidInstance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := buildClients()
			if err != nil {
				return err
			}
			ns := resolvedNamespace()
			u, err := cl.dynamic.Resource(instancesGVR).Namespace(ns).
				Get(context.Background(), args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}
			inst := &redroidv1alpha1.RedroidInstance{}
			if err := convertUnstructured(u.Object, inst); err != nil {
				return err
			}

			printInstanceDetail(inst)
			return nil
		},
	}
}

func printInstanceDetail(inst *redroidv1alpha1.RedroidInstance) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer func() { _ = w.Flush() }()
	tf := func(format string, a ...interface{}) { _, _ = fmt.Fprintf(w, format, a...) }

	tf("Name:\t%s\n", inst.Name)
	tf("Namespace:\t%s\n", inst.Namespace)
	tf("Age:\t%s\n", humanDuration(time.Since(inst.CreationTimestamp.Time)))
	tf("Index:\t%d\n", inst.Spec.Index)
	tf("Image:\t%s\n", inst.Spec.Image)
	tf("BaseMode:\t%v\n", inst.Spec.BaseMode)
	tf("GPUMode:\t%s\n", inst.Spec.GPUMode)
	tf("Spec.Suspend:\t%v\n", inst.Spec.Suspend)
	tf("Phase:\t%s\n", inst.Status.Phase)
	tf("PodName:\t%s\n", inst.Status.PodName)
	tf("ADBAddress:\t%s\n", inst.Status.ADBAddress)

	if inst.Status.Suspended != nil {
		s := inst.Status.Suspended
		tf("Suspended.Reason:\t%s\n", s.Reason)
		tf("Suspended.Actor:\t%s\n", s.Actor)
		if s.Until != nil {
			tf("Suspended.Until:\t%s\n", s.Until.Format(time.RFC3339))
		}
	}

	for _, cond := range inst.Status.Conditions {
		tf("Condition[%s]:\t%s — %s\n", cond.Type, cond.Status, cond.Message)
	}
}

// ── port-forward ─────────────────────────────────────────────────────────────

func newInstancePortForwardCmd() *cobra.Command {
	var localPort int
	c := &cobra.Command{
		Use:   "port-forward <name>",
		Short: "Forward the ADB port of a RedroidInstance to localhost",
		Long: `Forward the instance's ADB TCP port to a local port.

Examples:
  # Forward to a random local port
  kubectl redroid instance port-forward my-instance

  # Forward to a specific local port
  kubectl redroid instance port-forward my-instance --local-port 15555`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := buildClients()
			if err != nil {
				return err
			}
			inst, err := getInstance(cl, resolvedNamespace(), args[0])
			if err != nil {
				return err
			}

			svcName, podPort, err := instanceServiceAndPort(inst)
			if err != nil {
				return err
			}

			if localPort == 0 {
				localPort = int(podPort)
			}

			fmt.Printf("Forwarding %s/%s ADB port %d → localhost:%d\n",
				inst.Namespace, inst.Name, podPort, localPort)

			// Blocking call — returns on Ctrl-C.
			_, err = startPortForward(portForwardOptions{
				REST:        cl.cfg,
				Namespace:   inst.Namespace,
				ServiceName: svcName,
				LocalPort:   localPort,
				PodPort:     int(podPort),
				Out:         os.Stdout,
				ErrOut:      os.Stderr,
			})
			return err
		},
	}
	c.Flags().IntVar(&localPort, "local-port", 0, "Local TCP port to forward to (0 = same as pod port)")
	return c
}

// ── adb ───────────────────────────────────────────────────────────────────────

func newInstanceADBCmd() *cobra.Command {
	var localPort int
	c := &cobra.Command{
		Use:   "adb <name> -- [adb-args...]",
		Short: "Run an adb command against a RedroidInstance",
		Long: `Port-forward the ADB port and execute an adb command.

Requires the 'adb' binary to be present in $PATH.

Examples:
  kubectl redroid instance adb my-instance -- devices
  kubectl redroid instance adb my-instance -- install /path/to/app.apk
  kubectl redroid instance adb my-instance -- logcat -d`,
		Args:               cobra.MinimumNArgs(1),
		DisableFlagParsing: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWithADB(args[0], localPort, func(host string, port int) error {
				adbArgs := args[1:]
				return runADB(host, port, adbArgs...)
			})
		},
	}
	c.Flags().IntVar(&localPort, "local-port", 0, "Local TCP port (0 = use pod ADB port)")
	return c
}

// ── shell ─────────────────────────────────────────────────────────────────────

func newInstanceShellCmd() *cobra.Command {
	var localPort int
	c := &cobra.Command{
		Use:   "shell <name>",
		Short: "Open an interactive ADB shell in a RedroidInstance",
		Long: `Port-forward the ADB port then opens an interactive adb shell.

Requires the 'adb' binary to be present in $PATH.

Example:
  kubectl redroid instance shell my-instance`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWithADB(args[0], localPort, func(host string, port int) error {
				return runADB(host, port, "shell")
			})
		},
	}
	c.Flags().IntVar(&localPort, "local-port", 0, "Local TCP port (0 = use pod ADB port)")
	return c
}

// runWithADB sets up a background port-forward and calls fn(host, port).
func runWithADB(instanceName string, localPort int, fn func(host string, port int) error) error {
	cl, err := buildClients()
	if err != nil {
		return err
	}

	if _, err := exec.LookPath("adb"); err != nil {
		return fmt.Errorf("'adb' not found in $PATH; install Android SDK platform-tools first")
	}

	inst, err := getInstance(cl, resolvedNamespace(), instanceName)
	if err != nil {
		return err
	}

	svcName, podPort, err := instanceServiceAndPort(inst)
	if err != nil {
		return err
	}

	if localPort == 0 {
		localPort = int(podPort)
	}

	readyCh := make(chan struct{})
	stopCh := make(chan struct{})

	result, err := startPortForward(portForwardOptions{
		REST:        cl.cfg,
		Namespace:   inst.Namespace,
		ServiceName: svcName,
		LocalPort:   localPort,
		PodPort:     int(podPort),
		Out:         io.Discard,
		ErrOut:      os.Stderr,
		StopCh:      stopCh,
		ReadyCh:     readyCh,
	})
	if err != nil {
		return fmt.Errorf("port-forward: %w", err)
	}
	defer close(result.StopCh)

	// Connect adb to the forwarded port.
	if out, err := exec.Command("adb", "-H", "localhost", "-P",
		fmt.Sprint(localPort), "connect", fmt.Sprintf("localhost:%d", localPort)).
		CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "adb connect: %s\n", out)
	}

	return fn("localhost", localPort)
}

// runADB runs an adb command targeting host:port.
func runADB(host string, port int, adbArgs ...string) error {
	args := append([]string{"-H", host, "-P", fmt.Sprint(port)}, adbArgs...)
	c := exec.Command("adb", args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

// ── logs ──────────────────────────────────────────────────────────────────────

func newInstanceLogsCmd() *cobra.Command {
	var follow bool
	var previous bool
	var tailLines int64
	var container string

	c := &cobra.Command{
		Use:   "logs <name>",
		Short: "Print logs from the Pod of a RedroidInstance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := buildClients()
			if err != nil {
				return err
			}
			inst, err := getInstance(cl, resolvedNamespace(), args[0])
			if err != nil {
				return err
			}

			// Find the pod via label selector so we don't depend on status.PodName.
			podList, err := cl.kube.CoreV1().Pods(inst.Namespace).List(context.Background(),
				metav1.ListOptions{LabelSelector: "redroid.io/instance=" + inst.Name})
			if err != nil {
				return fmt.Errorf("list pods for instance %q: %w", inst.Name, err)
			}
			if len(podList.Items) == 0 {
				return fmt.Errorf("instance %q has no running Pod (phase: %s)", args[0], inst.Status.Phase)
			}
			podName := podList.Items[0].Name

			opts := &corev1.PodLogOptions{
				Container: container,
				Follow:    follow,
				Previous:  previous,
			}
			if tailLines > 0 {
				opts.TailLines = &tailLines
			}

			req := cl.kube.CoreV1().Pods(inst.Namespace).GetLogs(podName, opts)
			stream, err := req.Stream(context.Background())
			if err != nil {
				return fmt.Errorf("get logs: %w", err)
			}
			defer func() { _ = stream.Close() }()

			_, err = io.Copy(os.Stdout, stream)
			return err
		},
	}
	c.Flags().BoolVarP(&follow, "follow", "f", false, "Follow log output")
	c.Flags().BoolVar(&previous, "previous", false, "Print previous terminated container logs")
	c.Flags().Int64Var(&tailLines, "tail", 0, "Lines of recent log to show (0 = all)")
	c.Flags().StringVarP(&container, "container", "c", "", "Container name (defaults to main redroid container)")
	return c
}

// ── suspend ───────────────────────────────────────────────────────────────────

func newInstanceSuspendCmd() *cobra.Command {
	var reason string
	var actor string
	var duration time.Duration

	c := &cobra.Command{
		Use:   "suspend <name>",
		Short: "Temporarily suspend a RedroidInstance (sets status.suspended)",
		Long: `Set status.suspended on a RedroidInstance to stop its Pod without modifying spec.suspend.

Unlike spec.suspend this field is stored in status and is therefore NOT reconciled by
GitOps tools such as Flux, preventing config drift.

Examples:
  kubectl redroid instance suspend my-inst --reason "nightly maintenance" --actor "manual"
  kubectl redroid instance suspend my-inst --duration 2h --actor "manual"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := buildClients()
			if err != nil {
				return err
			}
			ns := resolvedNamespace()

			suspended := map[string]interface{}{
				"reason": reason,
				"actor":  actor,
			}
			if duration > 0 {
				until := time.Now().Add(duration).UTC().Format(time.RFC3339)
				suspended["until"] = until
			}

			patch := map[string]interface{}{
				"status": map[string]interface{}{
					"suspended": suspended,
				},
			}
			data, _ := json.Marshal(patch)

			_, err = cl.dynamic.Resource(instancesGVR).Namespace(ns).
				Patch(context.Background(), args[0], types.MergePatchType, data,
					metav1.PatchOptions{}, "status")
			if err != nil {
				return err
			}
			fmt.Printf("Instance %q suspended (actor=%s, reason=%s)\n", args[0], actor, reason)
			return nil
		},
	}
	c.Flags().StringVar(&reason, "reason", "", "Human-readable reason for the suspend")
	c.Flags().StringVar(&actor, "actor", "manual", "Actor setting the suspend (stored for audit)")
	c.Flags().DurationVar(&duration, "duration", 0, "Auto-expire after this duration (e.g. 2h, 30m)")
	return c
}

// ── resume ────────────────────────────────────────────────────────────────────

func newInstanceResumeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "resume <name>",
		Short: "Clear the temporary suspend on a RedroidInstance (clears status.suspended)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := buildClients()
			if err != nil {
				return err
			}
			ns := resolvedNamespace()

			patch := []byte(`{"status":{"suspended":null}}`)
			_, err = cl.dynamic.Resource(instancesGVR).Namespace(ns).
				Patch(context.Background(), args[0], types.MergePatchType, patch,
					metav1.PatchOptions{}, "status")
			if err != nil {
				return err
			}
			fmt.Printf("Instance %q resumed (status.suspended cleared)\n", args[0])
			return nil
		},
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

// getInstance fetches and decodes a single RedroidInstance.
func getInstance(cl *clients, ns, name string) (*redroidv1alpha1.RedroidInstance, error) {
	u, err := cl.dynamic.Resource(instancesGVR).Namespace(ns).
		Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get instance %q: %w", name, err)
	}
	inst := &redroidv1alpha1.RedroidInstance{}
	return inst, convertUnstructured(u.Object, inst)
}

// instanceServiceAndPort derives the stable Service name from the instance name
// (no status.PodName dependency) and returns the configured ADB port.
// The operator ensures a ClusterIP Service named redroid-instance-<name> always
// exists for every RedroidInstance, even when the Pod is restarting.
func instanceServiceAndPort(inst *redroidv1alpha1.RedroidInstance) (svcName string, adbPort int32, err error) {
	if inst.Status.Phase == "" {
		return "", 0, fmt.Errorf("instance %q status unknown", inst.Name)
	}
	adbPort = defaultADBPortValue
	if inst.Spec.ADBPort != nil {
		adbPort = *inst.Spec.ADBPort
	}
	return fmt.Sprintf("redroid-instance-%s", inst.Name), adbPort, nil
}

const defaultADBPortValue = int32(5555)

// convertUnstructured converts an unstructured map to a typed object via JSON round-trip.
func convertUnstructured(obj map[string]interface{}, out interface{}) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

// humanDuration formats a duration like kubectl age columns.
func humanDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	days := int(d.Hours() / 24)
	if days < 365 {
		return fmt.Sprintf("%dd", days)
	}
	return fmt.Sprintf("%dy", days/365)
}
