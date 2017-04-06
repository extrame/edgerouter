package edgerouter

type Transport interface {
	Connect(string) error
	Send(*BytesMessage) error
	SetController(Controller)
}

type Server interface {
	Run()
}
