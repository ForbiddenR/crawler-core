package entity

// IPCMessage defines a message emitted by a spider process for Crawlab IPC.
type IPCMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
	IPC     bool        `json:"ipc"`
}
