package models

import (
	genprotos "github.com/ruziba3vich/devices/genprotos/devices_submodule"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	Device struct {
		Id       primitive.ObjectID `bson:"_id" json:"id"`
		Name     string             `bson:"name" json:"name"`
		Type     string             `bson:"type" json:"type"`
		Status   string             `bson:"status" json:"status"`
		Location string             `bson:"location" json:"location"`
		Deleted  bool               `bson:"deleted" json:"deleted"`
	}

	CreateDeviceRequest struct {
		Name     string `json:"name"`
		Type     string `json:"type"`
		Status   string `json:"status"`
		Location string `json:"location"`
	}

	UpdateDeviceRequest struct {
		DeviceId string `json:"device_id"`
		Name     string `json:"name"`
		Type     string `json:"type"`
		Status   string `json:"status"`
		Location string `json:"location"`
	}

	DeleteDeviceRequest struct {
		DeviceId string `json:"device_id"`
	}
)

func (d *Device) ToProtoDevice() *genprotos.Device {
	return &genprotos.Device{
		Id:       d.Id.Hex(),
		Name:     d.Name,
		Type:     d.Type,
		Status:   d.Status,
		Location: d.Location,
	}
}

func (d *Device) FromProto(data *genprotos.Device) {
	d.Id, _ = primitive.ObjectIDFromHex(data.Id)
	d.Name = data.Name
	d.Type = data.Type
	d.Status = data.Status
	d.Location = data.Location
}

func (d *Device) ToCreateDeviceRequest() *genprotos.CreateDeviceRequest {
	return &genprotos.CreateDeviceRequest{
		Device: &genprotos.Device{
			Name:     d.Name,
			Type:     d.Type,
			Status:   d.Status,
			Location: d.Location,
		},
	}
}

func (d *Device) ToUpdateDeviceRequest() *genprotos.UpdateDeviceRequest {
	return &genprotos.UpdateDeviceRequest{
		Device: &genprotos.Device{
			Name:     d.Name,
			Type:     d.Type,
			Status:   d.Status,
			Location: d.Location,
		},
	}
}

func (d *Device) FromCreateDeviceRequest(data *genprotos.CreateDeviceRequest) {
	d.Name = data.Device.Name
	d.Type = data.Device.Type
	d.Status = data.Device.Status
	d.Location = data.Device.Location
}

func (d *Device) FromUpdateDeviceRequest(data *genprotos.UpdateDeviceRequest) {
	d.Id, _ = primitive.ObjectIDFromHex(data.Device.Id)
	d.Name = data.Device.Name
	d.Type = data.Device.Type
	d.Status = data.Device.Status
	d.Location = data.Device.Location
}
