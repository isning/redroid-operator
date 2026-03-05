package main

import (
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/cache"
)

// parseWatchNamespaces parses a comma-separated list of namespace names into a
// cache.Config map suitable for controller-runtime's cache.Options.DefaultNamespaces.
//
// Rules:
//   - An empty input string returns a nil map (caller should use cluster-wide mode).
//   - Each entry is trimmed of whitespace; blank entries are dropped.
//   - If all entries are blank the map is nil (cluster-wide fallback).
func parseWatchNamespaces(raw string) map[string]cache.Config {
	if raw == "" {
		return nil
	}
	nsMap := map[string]cache.Config{}
	for _, ns := range strings.Split(raw, ",") {
		if ns = strings.TrimSpace(ns); ns != "" {
			nsMap[ns] = cache.Config{}
		}
	}
	if len(nsMap) == 0 {
		return nil
	}
	return nsMap
}
