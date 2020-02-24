package apple

type DeviceCheck interface {
	IsValidDeviceToken(string) (bool, error)
}
