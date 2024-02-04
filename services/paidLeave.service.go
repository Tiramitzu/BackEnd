package services

import (
	db "attendit/backend/models/db"
	"errors"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetActiveRequest(userId primitive.ObjectID) (*db.PaidLeave, error) {
	paidLeave := &db.PaidLeave{}
	err := mgm.Coll(paidLeave).First(bson.M{"userId": userId, "accepted": false}, paidLeave)

	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			return nil, nil
		}

		return nil, err
	}

	return paidLeave, nil
}

func GetPaidLeaves() ([]*db.PaidLeave, error) {
	var paidLeaves []*db.PaidLeave
	var users []*db.User

	err := mgm.Coll(&db.PaidLeave{}).SimpleFind(&paidLeaves, bson.M{})
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			return nil, nil
		}
		return nil, errors.New("Gagal mendapatkan data cuti")
	}

	err = mgm.Coll(&db.User{}).SimpleFind(&users, bson.M{})
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			return nil, nil
		}

		return nil, errors.New("Gagal mendapatkan data user")
	}

	for _, paidLeave := range paidLeaves {
		for _, user := range users {
			if user.ID == paidLeave.UserId {
				paidLeave.User = user
			}
		}
	}

	return paidLeaves, nil
}

func CreatePaidLeave(userId primitive.ObjectID, reason string, startDate string, days int) (*db.PaidLeave, error) {
	paidLeave := db.NewPaidLeave(userId, 0, primitive.NilObjectID, reason, startDate, days)
	err := mgm.Coll(paidLeave).Create(paidLeave)

	if err != nil {
		return nil, errors.New("Gagal membuat permintaan cuti")
	}

	return paidLeave, nil
}

func UpdatePaidLeaveStatus(paidLeaveId primitive.ObjectID, status int) (*db.PaidLeave, error) {
	paidLeave := &db.PaidLeave{}
	err := mgm.Coll(paidLeave).First(bson.M{"_id": paidLeaveId}, paidLeave)

	if err != nil {
		return nil, errors.New("Gagal mendapatkan data cuti")
	}

	paidLeave.Status = status
	err = mgm.Coll(paidLeave).Update(paidLeave)

	if err != nil {
		return nil, errors.New("Gagal mengupdate data cuti")
	}

	return paidLeave, nil
}
