// Package lint2 provides linting functionality for Replicated resources that integrates with
// the replicated CLI tool resolver infrastructure.
//
// This package enables automatic downloading and execution of linting commands for:
//   - Helm charts (via helm lint)
//   - Preflight specs (via preflight lint from troubleshoot.sh)
//   - Support Bundle specs (via support-bundle lint from troubleshoot.sh)
//
// # Features
//
// Common functionality across all linters:
//   - Resource path expansion (including glob patterns)
//   - Resource validation (ensuring valid structure)
//   - Binary resolution via tool-resolver (automatic download/caching)
//   - Output parsing into structured results
//   - Support for custom tool versions
//
// Glob pattern support (powered by doublestar library):
//   - Basic patterns: * (any chars), ? (one char), [abc] (char class)
//   - Recursive matching: ** (matches zero or more directories)
//   - Brace expansion: {alt1,alt2} (matches alternatives)
//   - Pattern validation: Early syntax checking during config parse
//
// Helm-specific:
//   - Chart directory validation (Chart.yaml presence)
//   - Multi-chart linting with summary results
//
// Troubleshoot (Preflight/Support Bundle) specific:
//   - Multi-document YAML parsing with yaml.NewDecoder
//   - JSON output parsing with generic type-safe implementation
//   - Auto-discovery of Support Bundles from manifest files
//
// # Usage
//
// The typical workflow for any linter is:
//
//  1. Load configuration using tools.ConfigParser
//  2. Extract and validate resource paths
//  3. Resolve tool binary (downloads if not cached)
//  4. Execute lint command on each resource
//  5. Parse and display results
//
// # Example - Helm Charts
//
//	parser := tools.NewConfigParser()
//	config, err := parser.FindAndParseConfig(".")
//	if err != nil {
//	    return err
//	}
//
//	chartPaths, err := lint2.GetChartPathsFromConfig(config)
//	if err != nil {
//	    return err
//	}
//
//	for _, chartPath := range chartPaths {
//	    result, err := lint2.LintChart(ctx, chartPath, helmVersion)
//	    if err != nil {
//	        return err
//	    }
//	    // Process result...
//	}
//
// # Example - Preflight Specs
//
//	preflightPaths, err := lint2.GetPreflightPathsFromConfig(config)
//	if err != nil {
//	    return err
//	}
//
//	for _, specPath := range preflightPaths {
//	    result, err := lint2.LintPreflight(ctx, specPath, preflightVersion)
//	    if err != nil {
//	        return err
//	    }
//	    // Process result...
//	}
//
// # Example - Support Bundle Specs
//
//	// Support bundles are auto-discovered from manifest files
//	sbPaths, err := lint2.DiscoverSupportBundlesFromManifests(config.Manifests)
//	if err != nil {
//	    return err
//	}
//
//	for _, specPath := range sbPaths {
//	    result, err := lint2.LintSupportBundle(ctx, specPath, sbVersion)
//	    if err != nil {
//	        return err
//	    }
//	    // Process result...
//	}
package lint2
