// Package config loads typed application configuration from environment
// variables, then runs validation hooks.
//
// Struct field tags use the github.com/kusold/glib/env tag format, including
// env, envDefault, required, and notEmpty. File-backed configuration is
// intentionally out of scope.
package config
