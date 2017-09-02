package v2action

import "io"

type UI interface {
	GetIn() io.Reader
	GetOut() io.Writer
	GetErr() io.Writer
}
