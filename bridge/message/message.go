package message

import "encoding/json"
import "errors"
import "log"
import "reflect"

type Message struct {
	Evt string
	Data interface{}
}


func NewMessage(b []byte) (*Message, error) {
	var data map[string]interface{}
	var msg = Message{}

	if err:= json.Unmarshal(b, &data); err != nil {
		log.Println(err)
		return &msg, errors.New("Unreadable JSON data");
	} else {
		evt, existe := data["evt"]
		if !existe {
			return &msg, errors.New("No event in parsed JSON")
		} else {
			msg.Evt = evt.(string)
			payload, hayPayload := data["data"]
			if hayPayload {
				switch payload.(type) {
				case string:
					msg.Data = payload.(string)
				case map[string]interface{}:
					msg.Data = payload.(map[string]interface{})
				default:
					log.Println(reflect.ValueOf(payload))
				}
			} else {
				log.Println("No data")
			}

			return &msg, nil
		}
	}
}

func (m *Message) AsMap() map[string]interface{} {
	return map[string]interface{}{"evt": m.Evt, "data": m.Data}
}