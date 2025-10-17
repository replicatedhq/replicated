// Package lint2 provides Helm chart linting functionality that integrates with
// the replicated CLI tool resolver infrastructure.
//
// This package enables automatic downloading and execution of helm lint commands
// on Helm charts specified in .replicated configuration files. It handles:
//
//   - Chart path expansion (including glob patterns)
//   - Chart validation (ensuring valid Helm chart structure)
//   - Helm binary resolution via tool-resolver (automatic download/caching)
//   - Helm lint output parsing into structured results
//
// # Usage
//
// The typical workflow is:
//
//  1. Load configuration using tools.ConfigParser
//  2. Extract and validate chart paths
//  3. Resolve helm binary (downloads if not cached)
//  4. Execute helm lint on each chart
//  5. Parse and display results
//
// Example:
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
package lint2
