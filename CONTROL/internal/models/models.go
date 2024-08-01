package models

type (
	TYPE string
)

const (
	TURNDEVICEONQUEUE  TYPE = "turn_device_on_queue"
	TURNDEVICEOFFQUEUE TYPE = "turn_device_off_queue"
	ADDUSERQUEUE       TYPE = "add_user_queue"
	REMOVEUSERQUEUE    TYPE = "remove_user_queue"
)
