package models

import (
	usersprotos "github.com/ruziba3vich/smart-house/genprotos/submodules/users_submodule/protos"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	User struct {
		Id       primitive.ObjectID `bson:"_id" json:"id"`
		Username string             `bson:"username" json:"username"`
		Email    string             `bson:"email" json:"email"`
		Password string             `bson:"password" json:"password"`
		Profile  Profile            `bson:"profile" json:"profile"`
		Deleted  bool               `bson:"deleted" json:"deleted"`
		// Method   Method
	}

	Profile struct {
		Name    string `bson:"name" json:"name"`
		Address string `bson:"address" json:"address"`
	}

	GetByFieldRequest struct {
		Field string
		Value string
	}

	DeleteUserRequest struct {
		UserId string `json:"user_id"`
	}
)

func (u *User) Update(obj *usersprotos.UpdateUserReuqest) {
	if len(obj.User.Email) > 0 {
		u.Email = obj.User.Email
	}
	if len(obj.User.Username) > 0 {
		u.Username = obj.User.Username
	}
	if len(obj.User.Profile.Name) > 0 {
		u.Profile.Name = obj.User.Profile.Name
	}
	if len(obj.User.Profile.Address) > 0 {
		u.Profile.Address = obj.User.Profile.Address
	}
}

func (p Profile) ToProtoProfile() *usersprotos.Profile {
	return &usersprotos.Profile{
		Name:    p.Name,
		Address: p.Address,
	}
}

func (p *Profile) FromProto(obj *usersprotos.Profile) {
	p.Name = obj.Name
	p.Address = obj.Address
}

func (u *User) ToProtoUser() *usersprotos.User {
	return &usersprotos.User{
		UserId:   u.Id.Hex(),
		Username: u.Username,
		Email:    u.Email,
		Password: u.Password,
		Profile:  u.Profile.ToProtoProfile(),
		Deleted:  u.Deleted,
	}
}

func (u *User) FromProto(data *usersprotos.CreateUserReuest) {
	u.Email = data.Email
	u.Password = data.Password
	u.Username = data.Username
	u.Deleted = data.Deleted
	u.Profile.FromProto(data.Profile)
}

func (u *User) ToCreateUserRequest() *usersprotos.CreateUserReuest {
	return &usersprotos.CreateUserReuest{
		Username: u.Username,
		Email:    u.Email,
		Password: u.Password,
		Profile:  u.Profile.ToProtoProfile(),
		Deleted:  u.Deleted,
	}
}
