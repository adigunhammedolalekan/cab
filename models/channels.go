package models

import (
	"melody"
	"encoding/json"
)

type Channel struct {

	ChannelName string `json:"channel_name"`
	Sessions []*melody.Session

}

func (c *Channel) Send(data interface{}) error {

	msg, err := json.Marshal(data)
	if err != nil {
		return err
	}

	for _, s := range c.Sessions {
		if s != nil {
			err = s.Write(msg)
		}
	}

	return err
}
