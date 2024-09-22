package token

import (
	"aidanwoods.dev/go-paseto"
	"errors"
	"golang.org/x/crypto/chacha20poly1305"
	"strconv"
	"time"
)

type PasetoMaker struct {
	paseto       paseto.V4SymmetricKey
	symmetricKey []byte
}

func NewPasetoMaker(symmetricKey string) (Maker, error) {

	if len(symmetricKey) != chacha20poly1305.KeySize {
		return nil, errors.New("symmetric key too short")
	}

	maker := &PasetoMaker{
		paseto:       paseto.NewV4SymmetricKey(),
		symmetricKey: []byte(symmetricKey),
	}

	return maker, nil
}

func (maker PasetoMaker) CreateToken(userId int64, duration time.Duration) (string, *Payload, error) {
	payload, err := NewPayload(userId, duration)
	if err != nil {
		return "", payload, err
	}

	token := paseto.NewToken()
	token.SetExpiration(payload.ExpireAt)
	token.SetIssuedAt(payload.IssuedAt)
	token.SetSubject(strconv.FormatInt(payload.UserID, 10))
	token.SetNotBefore(time.Now())
	_ = token.Set("payload", payload)

	return token.V4Encrypt(maker.paseto, maker.symmetricKey), payload, nil
}

func (maker PasetoMaker) VerifyToken(token string) (*Payload, error) {
	parser := paseto.NewParser()
	parsedToken, err := parser.ParseV4Local(maker.paseto, token, maker.symmetricKey)

	if err != nil {
		return nil, errors.Join(ErrInvalidToken, err)
	}

	payload := &Payload{}

	err = parsedToken.Get("payload", payload)

	if err != nil {
		return nil, ErrInvalidToken
	}

	return payload, nil
}
