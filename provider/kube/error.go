package kube

type ErrNotFound string

func (e ErrNotFound) Error() string {
	return "not found: " + string(e)
}
