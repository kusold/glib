// Package app coordinates application component lifecycles.
//
// Components run until their context is canceled or they return. When the
// parent context is canceled, a component exits, or a component fails, Run
// requests graceful shutdown for every component and waits for the running
// components to return.
package app
