package flip

import (
	"math/big"
)

type User string

type UserState struct {
	User       User
	Commitment Secret
	Reveal     Secret
}

func checkReveal(c Secret, r Secret) bool {
	return c.Hash().Eq(r)
}

func checkUser(err *Error, user UserState) {
	if user.Commitment.IsNil() {
		err.addNoCommitment(user.User)
		return
	}
	if user.Reveal.IsNil() {
		err.addNoReveal(user.User)
		return
	}

	if !checkReveal(user.Commitment, user.Reveal) {
		err.addBadCommitment(user.User)
		return
	}

	return
}

func checkUsers(users []UserState) error {
	var err Error
	for _, u := range users {
		checkUser(&err, u)
	}
	if err.IsNil() {
		return nil
	}

	return err
}

func computeSecret(users []UserState) Secret {
	var res Secret
	for _, u := range users {
		res.XOR(u.Commitment)
	}
	return res
}

func Flip(users []UserState) (*PRNG, error) {
	err := checkUsers(users)
	if err != nil {
		return nil, err
	}
	res := computeSecret(users)
	return NewPRNG(res), nil
}

func FlipOneBig(users []UserState, modulus *big.Int) (*big.Int, error) {
	prng, err := Flip(users)
	if err != nil {
		return nil, err
	}
	return prng.Big(modulus), nil
}

func FlipInt(users []UserState, modulus int64) (int64, error) {
	prng, err := Flip(users)
	if err != nil {
		return 0, err
	}
	return prng.Int(modulus), nil
}

func FlipBool(users []UserState) (bool, error) {
	prng, err := Flip(users)
	if err != nil {
		return false, err
	}
	return prng.Bool(), nil
}
