package receiver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

const (
	propertyToken      = "token"
	propertyTime       = "time"
	propertyDistinctId = "distinct_id"
	propertyInsertId   = "$insert_id"
)

type KeysAndValues map[string]interface{}

type Event struct {
	Event      string     `json:"event"`
	Properties Properties `json:"properties"`
}

type RawEvent struct {
	Event      string          `json:"event"`
	Properties json.RawMessage `json:"properties"`
}

func (e *Event) UnmarshalJSON(b []byte) error {
	ej := &RawEvent{}
	err := json.Unmarshal(b, ej)
	if err != nil {
		fmt.Println("unable to unmarshall into EventJson")
		return err
	}

	if strings.TrimSpace(ej.Event) == "" {
		return &json.UnmarshalTypeError{
			Value: "missing event name",
			Type:  reflect.TypeOf(""),
			Field: "event",
		}
	}

	if len(ej.Properties) == 0 {
		return &json.UnmarshalTypeError{
			Value: "empty properties",
			Type:  reflect.TypeOf(map[string]string{}),
			Field: "properties",
		}
	}

	props := &Properties{}
	err = json.Unmarshal(ej.Properties, props)
	if err != nil {
		return err
	}

	e.Event = ej.Event
	e.Properties = *props

	return nil
}

type Properties struct {
	Token      string `json:"token"`
	Time       int    `json:"time"`
	DistinctId string `json:"distinct_id"`
	InsertId   string `json:"$insert_id"`
	Extra      KeysAndValues
}

func (p *Properties) UnmarshalJSON(b []byte) error {
	requiredProperties := map[string]struct{}{
		propertyToken:      {},
		propertyTime:       {},
		propertyDistinctId: {},
		propertyInsertId:   {},
	}

	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()

	values := map[string]interface{}{}
	err := d.Decode(&values)
	if err != nil {
		return err
	}

	extra := map[string]interface{}{}

	for k, v := range values {
		switch k {
		case propertyToken:
			asString, ok := v.(string)
			if !ok {
				return &json.UnmarshalTypeError{
					Value: "token",
					Type:  reflect.TypeOf(""),
				}
			}
			p.Token = asString
			delete(requiredProperties, propertyToken)
		case propertyTime:
			asNumber, ok := v.(json.Number)
			if !ok {
				return &json.UnmarshalTypeError{
					Value: "time",
					Type:  reflect.TypeOf(1),
				}
			}

			asInt, err := asNumber.Int64()
			if err != nil {
				return &json.UnmarshalTypeError{
					Value: "time",
					Type:  reflect.TypeOf(1),
				}
			}

			p.Time = int(asInt)
			delete(requiredProperties, propertyTime)
		case propertyDistinctId:
			asString, ok := v.(string)
			if !ok {
				return &json.UnmarshalTypeError{
					Value: "distinct_id",
					Type:  reflect.TypeOf(""),
				}
			}
			p.DistinctId = asString
			delete(requiredProperties, propertyDistinctId)
		case propertyInsertId:
			asString, ok := v.(string)
			if !ok {
				return &json.UnmarshalTypeError{
					Value: "$insert_id",
					Type:  reflect.TypeOf(""),
				}
			}
			p.InsertId = asString
			delete(requiredProperties, propertyInsertId)
		default:
			extra[k] = v
		}
	}

	if len(requiredProperties) > 0 {
		numSeen := 0

		errBuilder := strings.Builder{}
		errBuilder.WriteString("required fields not provided:")

		for k := range requiredProperties {
			numSeen++
			errBuilder.WriteString(" ")
			errBuilder.WriteString(k)
			if numSeen < len(requiredProperties) {
				errBuilder.WriteString(",")
			}
		}
		return fmt.Errorf(errBuilder.String())
	}

	if len(extra) > 0 {
		p.Extra = extra
	}
	return nil
}

func (p *Properties) MarshalJSON() ([]byte, error) {
	fields := KeysAndValues{}

	for k, v := range p.Extra {
		fields[k] = v
	}
	fields[propertyDistinctId] = p.DistinctId
	fields[propertyInsertId] = p.InsertId
	fields[propertyToken] = p.Token
	fields[propertyTime] = p.Time

	return json.Marshal(fields)
}

type Status struct {
	Status uint   `json:"status"`
	Error  string `json:"error"`
}
