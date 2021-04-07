package pipeline

type Op interface {
	Execute(environment *Environment) error
}
