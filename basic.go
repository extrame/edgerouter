package edgerouter

func panicWrapping(f func()) {
	defer func() {
		recover()
	}()
	f()
}
