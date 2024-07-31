package models

import (
	genprotos "github.com/ruziba3vich/users/genprotos/users_submodule/protos"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	POST   Method = "POST"
	UPDATE Method = "UPDATE"
	DELETE Method = "DELETE"
)

type (
	User struct {
		Id       primitive.ObjectID `bson:"_id"`
		Username string             `bson:"username"`
		Email    string             `bson:"email"`
		Password string             `bson:"password"`
		Profile  Profile            `bson:"profile"`
		Deleted  bool               `bson:"deleted"`
		// Method   Method
	}

	Method string

	Profile struct {
		Name    string `bson:"name"`
		Address string `bson:"address"`
	}

	GetByFieldRequest struct {
		Field string
		Value string
	}
)

func (u *User) Update(obj *genprotos.UpdateUserReuqest) {
	if len(obj.User.Email) > 0 {
		u.Email = obj.User.Email
	}
	if len(obj.User.Username) > 0 {
		u.Username = obj.User.Username
	}
	if obj.User.Profile != nil {
		u.Profile.FromProto(obj.User.Profile)
	}
}

func (p Profile) ToProtoProfile() *genprotos.Profile {
	return &genprotos.Profile{
		Name:    p.Name,
		Address: p.Address,
	}
}

func (p *Profile) FromProto(obj *genprotos.Profile) {
	p.Name = obj.Name
	p.Address = obj.Address
}

func (u *User) ToProtoUser() *genprotos.User {
	return &genprotos.User{
		UserId:   u.Id.Hex(),
		Username: u.Username,
		Email:    u.Email,
		Password: u.Password,
		Profile:  u.Profile.ToProtoProfile(),
		Deleted:  u.Deleted,
	}
}

func (u *User) FromProto(data *genprotos.CreateUserReuest) {
	u.Email = data.Email
	u.Password = data.Password
	u.Username = data.Username
	u.Deleted = data.Deleted
	u.Profile.FromProto(data.Profile)
}

func (u *User) ToCreateUserRequest() *genprotos.CreateUserReuest {
	return &genprotos.CreateUserReuest{
		Username: u.Username,
		Email:    u.Email,
		Password: u.Password,
		Profile:  u.Profile.ToProtoProfile(),
		Deleted:  u.Deleted,
	}
}
