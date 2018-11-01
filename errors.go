package commander

type applicationError struct {
	error
}

func isApplicationError(err error) bool {
	_, ok := err.(applicationError)
	return ok
}
