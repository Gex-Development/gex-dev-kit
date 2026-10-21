package app

func Explain(err error) Problem {
	if err == nil {
		return Problem{}
	}
	return Problem{Title: "Erro", Cause: err.Error()}
}
