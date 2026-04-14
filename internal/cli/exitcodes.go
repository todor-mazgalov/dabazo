// Package cli exit code constants matching the dabazo specification.
package cli

const (
	// ExitSuccess indicates the command completed successfully.
	ExitSuccess = 0
	// ExitGeneric indicates a generic failure.
	ExitGeneric = 1
	// ExitUsage indicates a usage error (bad flags, missing required args).
	ExitUsage = 2
	// ExitNotFound indicates the instance was not found in the registry.
	ExitNotFound = 3
	// ExitAlreadyExists indicates the instance name is already taken.
	ExitAlreadyExists = 4
	// ExitPkgManager indicates a package manager operation failed.
	ExitPkgManager = 5
	// ExitDBOperation indicates a database operation failed.
	ExitDBOperation = 6
	// ExitAborted indicates the user aborted a confirmation prompt.
	ExitAborted = 7
)
