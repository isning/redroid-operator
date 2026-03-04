package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	redroidv1alpha1 "github.com/isning/redroid-operator/api/v1alpha1"
)

func newTaskCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "task",
		Aliases: []string{"tasks"},
		Short:   "Manage RedroidTask resources",
	}
	c.AddCommand(
		newTaskListCmd(),
		newTaskDescribeCmd(),
		newTaskTriggerCmd(),
		newTaskLogsCmd(),
	)
	return c
}

// ── list ──────────────────────────────────────────────────────────────────────

func newTaskListCmd() *cobra.Command {
	var allNamespaces bool
	c := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "get"},
		Short:   "List RedroidTasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := buildClients()
			if err != nil {
				return err
			}

			ns := resolvedNamespace()
			if allNamespaces {
				ns = ""
			}

			list, err := cl.dynamic.Resource(tasksGVR).Namespace(ns).
				List(context.Background(), metav1.ListOptions{})
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			defer func() { _ = w.Flush() }()

			if allNamespaces {
				_, _ = fmt.Fprintln(w, "NAMESPACE\tNAME\tSCHEDULE\tSUSPEND\tACTIVE JOBS\tLAST SCHEDULE\tAGE")
			} else {
				_, _ = fmt.Fprintln(w, "NAME\tSCHEDULE\tSUSPEND\tACTIVE JOBS\tLAST SCHEDULE\tAGE")
			}

			for _, item := range list.Items {
				task := &redroidv1alpha1.RedroidTask{}
				if err := convertUnstructured(item.Object, task); err != nil {
					continue
				}

				schedule := task.Spec.Schedule
				if schedule == "" {
					schedule = "<one-shot>"
				}

				lastSched := "<never>"
				if task.Status.LastScheduleTime != nil {
					lastSched = humanDuration(time.Since(task.Status.LastScheduleTime.Time))
				}

				age := humanDuration(time.Since(task.CreationTimestamp.Time))

				if allNamespaces {
					_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%v\t%d\t%s\t%s\n",
						task.Namespace, task.Name, schedule, task.Spec.Suspend,
						len(task.Status.ActiveJobs), lastSched, age)
				} else {
					_, _ = fmt.Fprintf(w, "%s\t%s\t%v\t%d\t%s\t%s\n",
						task.Name, schedule, task.Spec.Suspend,
						len(task.Status.ActiveJobs), lastSched, age)
				}
			}
			return nil
		},
	}
	c.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "List across all namespaces")
	return c
}

// ── describe ──────────────────────────────────────────────────────────────────

func newTaskDescribeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "describe <name>",
		Short: "Show detailed information about a RedroidTask",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := buildClients()
			if err != nil {
				return err
			}
			ns := resolvedNamespace()
			u, err := cl.dynamic.Resource(tasksGVR).Namespace(ns).
				Get(context.Background(), args[0], metav1.GetOptions{})
			if err != nil {
				return err
			}
			task := &redroidv1alpha1.RedroidTask{}
			if err := convertUnstructured(u.Object, task); err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			defer func() { _ = w.Flush() }()
			tf := func(format string, a ...interface{}) { _, _ = fmt.Fprintf(w, format, a...) }

			tf("Name:\t%s\n", task.Name)
			tf("Namespace:\t%s\n", task.Namespace)
			tf("Age:\t%s\n", humanDuration(time.Since(task.CreationTimestamp.Time)))
			tf("Schedule:\t%s\n", task.Spec.Schedule)
			tf("Timezone:\t%s\n", task.Spec.Timezone)
			tf("Suspend:\t%v\n", task.Spec.Suspend)
			tf("SuspendInstance:\t%v\n", task.Spec.SuspendInstance)

			for _, inst := range task.Spec.Instances {
				tf("Instance:\t%s\n", inst.Name)
			}
			for _, intg := range task.Spec.Integrations {
				tf("Integration[%s]:\t%s\n", intg.Name, intg.Image)
			}
			tf("ActiveJobs:\t%d\n", len(task.Status.ActiveJobs))
			for _, j := range task.Status.ActiveJobs {
				tf("  - %s\n", j)
			}
			if task.Status.LastScheduleTime != nil {
				tf("LastSchedule:\t%s ago\n",
					humanDuration(time.Since(task.Status.LastScheduleTime.Time)))
			}
			if task.Status.LastSuccessfulTime != nil {
				tf("LastSuccess:\t%s ago\n",
					humanDuration(time.Since(task.Status.LastSuccessfulTime.Time)))
			}
			for _, cond := range task.Status.Conditions {
				tf("Condition[%s]:\t%s — %s\n", cond.Type, cond.Status, cond.Message)
			}
			return nil
		},
	}
}

// ── trigger ───────────────────────────────────────────────────────────────────

func newTaskTriggerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "trigger <name>",
		Short: "Manually trigger a scheduled RedroidTask (creates ad-hoc Jobs)",
		Long: `Create one Job per instance from the CronJob specs of a scheduled RedroidTask.

Only works for tasks with a non-empty spec.schedule. For one-shot tasks simply
create or re-create the RedroidTask resource.

Example:
  kubectl redroid task trigger maa-daily`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := buildClients()
			if err != nil {
				return err
			}
			ns := resolvedNamespace()
			taskName := args[0]

			u, err := cl.dynamic.Resource(tasksGVR).Namespace(ns).
				Get(context.Background(), taskName, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("get task %q: %w", taskName, err)
			}
			task := &redroidv1alpha1.RedroidTask{}
			if err := convertUnstructured(u.Object, task); err != nil {
				return err
			}

			if task.Spec.Schedule == "" {
				return fmt.Errorf("task %q is a one-shot task (no schedule); re-create it to trigger", taskName)
			}

			var triggered []string
			for _, instRef := range task.Spec.Instances {
				cronName := fmt.Sprintf("%s-%s", taskName, instRef.Name)
				cj, err := cl.kube.BatchV1().CronJobs(ns).Get(context.Background(), cronName, metav1.GetOptions{})
				if err != nil {
					return fmt.Errorf("get cronjob %q: %w", cronName, err)
				}

				job := buildJobFromCronJob(cj, fmt.Sprintf("%s-manual-%s", taskName, shortTimestamp()))
				created, err := cl.kube.BatchV1().Jobs(ns).Create(context.Background(), job, metav1.CreateOptions{})
				if err != nil {
					return fmt.Errorf("create job for instance %q: %w", instRef.Name, err)
				}
				triggered = append(triggered, created.Name)
			}

			for _, j := range triggered {
				fmt.Printf("Job %q created\n", j)
			}
			return nil
		},
	}
}

// buildJobFromCronJob creates a new Job from a CronJob's job template.
func buildJobFromCronJob(cj *batchv1.CronJob, name string) *batchv1.Job {
	annotations := map[string]string{
		"redroid.io/triggered-by": "kubectl-redroid",
		"redroid.io/cronjob":      cj.Name,
	}
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   cj.Namespace,
			Labels:      cj.Spec.JobTemplate.Labels,
			Annotations: annotations,
		},
		Spec: cj.Spec.JobTemplate.Spec,
	}
}

// shortTimestamp returns a compact UTC timestamp for generated Job names.
func shortTimestamp() string {
	return time.Now().UTC().Format("20060102-1504")
}

// ── logs ──────────────────────────────────────────────────────────────────────

func newTaskLogsCmd() *cobra.Command {
	var follow bool
	var instance string
	var tailLines int64

	c := &cobra.Command{
		Use:   "logs <name>",
		Short: "Print logs from the most recent Job of a RedroidTask",
		Long: `Fetch the logs of the most recently created Job pod for a RedroidTask.

By default shows logs from all instances. Use --instance to filter to one.

Examples:
  kubectl redroid task logs maa-daily
  kubectl redroid task logs maa-daily --instance maa-0 -f`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := buildClients()
			if err != nil {
				return err
			}
			ns := resolvedNamespace()
			taskName := args[0]

			selector := labels.Set{"redroid.io/task": taskName}
			if instance != "" {
				selector["redroid.io/instance"] = instance
			}

			jobList, err := cl.kube.BatchV1().Jobs(ns).List(context.Background(), metav1.ListOptions{
				LabelSelector: selector.String(),
			})
			if err != nil {
				return fmt.Errorf("list jobs: %w", err)
			}

			if len(jobList.Items) == 0 {
				return fmt.Errorf("no jobs found for task %q (instance filter: %q)", taskName, instance)
			}

			// Sort by creation time descending — pick the newest.
			newest := jobList.Items[0]
			for _, j := range jobList.Items[1:] {
				if j.CreationTimestamp.After(newest.CreationTimestamp.Time) {
					newest = j
				}
			}

			fmt.Fprintf(os.Stderr, "# Logs from Job %q (created %s ago)\n",
				newest.Name, humanDuration(time.Since(newest.CreationTimestamp.Time)))

			// Find the Pod(s) for this job.
			podSelector := labels.Set{"job-name": newest.Name}
			pods, err := cl.kube.CoreV1().Pods(ns).List(context.Background(), metav1.ListOptions{
				LabelSelector: podSelector.String(),
			})
			if err != nil {
				return err
			}
			if len(pods.Items) == 0 {
				return fmt.Errorf("no pods found for job %q", newest.Name)
			}

			for _, pod := range pods.Items {
				fmt.Fprintf(os.Stderr, "# Pod %q\n", pod.Name)

				for _, container := range pod.Spec.InitContainers {
					printContainerLogs(cl, ns, pod.Name, container.Name, follow, tailLines)
				}
				for _, container := range pod.Spec.Containers {
					printContainerLogs(cl, ns, pod.Name, container.Name, follow, tailLines)
				}
			}
			return nil
		},
	}
	c.Flags().BoolVarP(&follow, "follow", "f", false, "Follow log output")
	c.Flags().StringVar(&instance, "instance", "", "Filter to a specific instance name")
	c.Flags().Int64Var(&tailLines, "tail", 0, "Lines of recent log to show (0 = all)")
	return c
}

func printContainerLogs(cl *clients, ns, podName, containerName string, follow bool, tailLines int64) {
	fmt.Fprintf(os.Stderr, "# Container %q\n", containerName)

	opts := &corev1.PodLogOptions{
		Container: containerName,
		Follow:    follow,
	}
	if tailLines > 0 {
		opts.TailLines = &tailLines
	}

	req := cl.kube.CoreV1().Pods(ns).GetLogs(podName, opts)
	stream, err := req.Stream(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "error streaming logs for %q: %v\n", containerName, err)
		return
	}
	defer func() { _ = stream.Close() }()

	if _, err := io.Copy(os.Stdout, stream); err != nil {
		fmt.Fprintf(os.Stderr, "error reading logs: %v\n", err)
	}
}
