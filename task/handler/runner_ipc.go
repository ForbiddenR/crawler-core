package handler

import (
	"encoding/json"
	"fmt"

	"github.com/apex/log"
	"github.com/crawlab-team/crawlab-core/constants"
	"github.com/crawlab-team/crawlab-core/entity"
	grpc "github.com/crawlab-team/crawlab-grpc"
)

func (r *Runner) handleIPCLine(line string) bool {
	var ipcMsg entity.IPCMessage
	if err := json.Unmarshal([]byte(line), &ipcMsg); err != nil || !ipcMsg.IPC {
		return false
	}

	switch ipcMsg.Type {
	case "", constants.IPCMessageData:
		if err := r.writeIPCData(ipcMsg.Payload); err != nil {
			log.Errorf("handle ipc data: %v", err)
		}
	case constants.IPCMessageLog:
		r.writeIPCLogs(ipcMsg.Payload, line)
	default:
		log.Warnf("unsupported ipc message type: %s", ipcMsg.Type)
	}

	return true
}

func (r *Runner) writeIPCData(payload interface{}) error {
	records, err := normalizeIPCRecords(payload)
	if err != nil {
		return err
	}
	if len(records) == 0 {
		return nil
	}

	for i := range records {
		records[i].SetTaskId(r.tid)
	}

	data, err := json.Marshal(&entity.StreamMessageTaskData{
		TaskId:  r.tid,
		Records: records,
	})
	if err != nil {
		return err
	}

	if r.sub == nil {
		return fmt.Errorf("task stream is not initialized")
	}

	msg := &grpc.StreamMessage{
		Code: grpc.StreamMessageCode_INSERT_DATA,
		Data: data,
	}
	return r.sub.Send(msg)
}

func normalizeIPCRecords(payload interface{}) ([]entity.Result, error) {
	switch p := payload.(type) {
	case map[string]interface{}:
		return []entity.Result{entity.Result(p)}, nil
	case []interface{}:
		records := make([]entity.Result, 0, len(p))
		for i, item := range p {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid ipc record at index %d: %T", i, item)
			}
			records = append(records, entity.Result(itemMap))
		}
		return records, nil
	case []map[string]interface{}:
		records := make([]entity.Result, 0, len(p))
		for _, item := range p {
			records = append(records, entity.Result(item))
		}
		return records, nil
	case []entity.Result:
		return p, nil
	case nil:
		return nil, fmt.Errorf("empty ipc payload")
	default:
		return nil, fmt.Errorf("unsupported ipc payload type: %T", payload)
	}
}

func (r *Runner) writeIPCLogs(payload interface{}, fallback string) {
	switch p := payload.(type) {
	case string:
		r.writeLogLines([]string{p})
	case []interface{}:
		lines := make([]string, 0, len(p))
		for _, item := range p {
			lines = append(lines, fmt.Sprint(item))
		}
		r.writeLogLines(lines)
	default:
		r.writeLogLines([]string{fallback})
	}
}
