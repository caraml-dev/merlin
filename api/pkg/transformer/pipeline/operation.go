package pipeline

import "context"

type Op interface {
	Execute(context context.Context, environment *Environment) error
}
