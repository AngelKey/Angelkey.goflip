package flip

import (
	"context"
	"encoding/base64"
	"github.com/keybase/go-codec/codec"
	"math/big"
	"strings"
)

type GameMessageWrappedEncoded struct {
	Header UserDevice
	Body   string // base64-encoded GameMessaageBody that comes in over chat
}

type GameMessageWrapped struct {
	Header UserDevice
	Msg    GameMessage
}

type Chatter interface {
	ReadChat(context.Context) (*GameMessageWrappedEncoded, error)
	SendChat(context.Context, string) error
	ReportHook(context.Context, GameMessageWrapped)
	ResultHook(context.Context) (GameMetadata, *Result, error)
	CLogf(ctx context.Context, fmt string, args ...interface{})
}

type Dealer struct {
	chatter Chatter
	games   map[GameKey]Game
}

type IntResult struct {
	b   *bool
	i   *int64
	big *big.Int
}

type Permutation []int

type GameMetadata struct {
	GameID    GameID
	Initiator UserDevice
}

func (g GameMessageWrapped) GameMetadata() GameMetadata {
	return GameMetadata{GameID: g.Msg.GameID, Initiator: g.Header}
}

func (g GameMetadata) ToKey() GameKey {
	return GameKey(strings.Join([]string{g.Initiator.User.String(), g.Initiator.Device.String(), g.GameID.String()}, ","))
}

type GameKey string

type Result struct {
	P Permutation
	I []IntResult
}

type Game struct {
}

func codecHandle() *codec.MsgpackHandle {
	var mh codec.MsgpackHandle
	mh.WriteExt = true
	return &mh
}

func msgpackDecode(dst interface{}, src []byte) error {
	h := codecHandle()
	return codec.NewDecoderBytes(src, h).Decode(dst)
}

func (e *GameMessageWrappedEncoded) Decode() (*GameMessageWrapped, error) {
	raw, err := base64.StdEncoding.DecodeString(e.Body)
	if err != nil {
		return nil, err
	}
	var msg GameMessage
	err = msgpackDecode(&msg, raw)
	if err != nil {
		return nil, err
	}
	ret := GameMessageWrapped{Header: e.Header, Msg: msg}
	return &ret, nil
}

func (d *Dealer) handleMessageStart(c context.Context, msg *GameMessageWrapped) error {
	return nil
}

func (d *Dealer) handleMessage(c context.Context, emsg *GameMessageWrappedEncoded) error {
	msg, err := emsg.Decode()
	if err != nil {
		return err
	}
	s, err := msg.Msg.Body.S()
	if err != nil {
		return err
	}
	switch s {
	case Stage_START:
		return d.handleMessageStart(c, msg)
	}
	return nil
}

func NewDealer(c Chatter) *Dealer {
	return &Dealer{chatter: c}
}

func (d *Dealer) Run(ctx context.Context) error {
	for {
		msg, err := d.chatter.ReadChat(ctx)
		if err != nil {
			return err
		}
		err = d.handleMessage(ctx, msg)
		if err != nil {
			d.chatter.CLogf(ctx, "Error reading message: %s", err.Error())
			return err
		}
	}
	return nil
}
