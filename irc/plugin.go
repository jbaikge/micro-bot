package irc

import "context"

type Plugin interface {
	Run(context.Context)
}
