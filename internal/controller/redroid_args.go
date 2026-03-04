package controller

import (
	"fmt"

	redroidv1alpha1 "github.com/isning/redroid-operator/api/v1alpha1"
)

// buildRedroidArgs constructs the androidboot.* argument list for a Redroid container
// based on the RedroidInstanceSpec. The returned slice is suitable for use as the
// container command arguments.
//
// When spec.BaseMode is true, androidboot.use_redroid_overlayfs is forced to 0
// so the container writes directly to the mounted /data volume without overlayfs.
func buildRedroidArgs(spec redroidv1alpha1.RedroidInstanceSpec) []string {
	args := []string{
		fmt.Sprintf("androidboot.redroid_gpu_mode=%s", spec.GPUMode),
		"androidboot.use_memfd=1",
	}

	// In base mode disable overlayfs so writes go directly to /data.
	if spec.BaseMode {
		args = append(args, "androidboot.use_redroid_overlayfs=0")
	} else {
		args = append(args, "androidboot.use_redroid_overlayfs=1")
	}

	// GPU node (DRM device path).
	if spec.GPUNode != "" {
		args = append(args, "androidboot.redroid_gpu_node="+spec.GPUNode)
	}

	// Screen resolution and display settings.
	if s := spec.Screen; s != nil {
		if s.Width != nil {
			args = append(args, fmt.Sprintf("androidboot.redroid_width=%d", *s.Width))
		}
		if s.Height != nil {
			args = append(args, fmt.Sprintf("androidboot.redroid_height=%d", *s.Height))
		}
		if s.DPI != nil {
			args = append(args, fmt.Sprintf("androidboot.redroid_dpi=%d", *s.DPI))
		}
		if s.FPS != nil {
			args = append(args, fmt.Sprintf("androidboot.redroid_fps=%d", *s.FPS))
		}
	}

	// Network: DNS and proxy settings.
	if n := spec.Network; n != nil {
		if len(n.DNS) > 0 {
			args = append(args, fmt.Sprintf("androidboot.redroid_net_ndns=%d", len(n.DNS)))
			for i, dns := range n.DNS {
				args = append(args, fmt.Sprintf("androidboot.redroid_net_dns%d=%s", i+1, dns))
			}
		}
		if p := n.Proxy; p != nil {
			if p.Type != "" {
				args = append(args, "androidboot.redroid_net_proxy_type="+p.Type)
			}
			if p.Host != "" {
				args = append(args, "androidboot.redroid_net_proxy_host="+p.Host)
			}
			if p.Port != nil {
				args = append(args, fmt.Sprintf("androidboot.redroid_net_proxy_port=%d", *p.Port))
			}
			if p.ExcludeList != "" {
				args = append(args, "androidboot.redroid_net_proxy_exclude_list="+p.ExcludeList)
			}
			if p.PAC != "" {
				args = append(args, "androidboot.redroid_net_proxy_pac="+p.PAC)
			}
		}
	}

	// User-supplied extra arguments. Supports $(VAR_NAME) env var substitution
	// from ExtraEnv (resolved by the kubelet, same as container command args).
	args = append(args, spec.ExtraArgs...)

	return args
}
