package entity

type EventData struct {
	Event string
	Data  any
}

func (d *EventData) GetEvent() string {
	return d.Event
}

func (d *EventData) GetData() any {
	return d.Data
}
