package utils

func PanicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}
