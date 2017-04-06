package edgerouter

type Device interface {
	DeviceID() string
	DeviceType() string
}
