package commands

import "context"

type Command[T any] interface {
	Run(context.Context, T) error
}
