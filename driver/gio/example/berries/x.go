package main

type err string

func (err err) Error() string { return string(err) }

const ErrMustNotBeNil = err("must not be nil")

func try[T any](result T, err error) T {
	noerr(err)
	return result
}

func must[T any](result T) T {
	if any(result) == nil {
		panic(ErrMustNotBeNil)
	}
	return result
}

func noerr(err error) {
	if err != nil {
		panic(err)
	}
}

func catch(errs ...*error) {
	if err, ok := recover().(error); ok {
		if len(errs) > 0 && errs[0] != nil {
			*errs[0] = err
		} else {
			println(err.Error())
		}
	}
}
