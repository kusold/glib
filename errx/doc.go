// Package errx provides coded application errors with safe public messages.
//
// Coded errors keep transport-neutral application codes separate from internal
// causes. The public message is safe to return to callers, while the wrapped
// cause remains available to standard errors.Is and errors.As checks.
package errx
