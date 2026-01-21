package sync

import (
	"encoding/gob"
	"io"
)

// MessageType represents the type of sync message
type MessageType uint8

const (
	// Message types for sync protocol
	MsgSignatureRequest MessageType = iota + 1
	MsgSignatureResponse
	MsgDeltaRequest
	MsgDeltaResponse
	MsgFileTransfer
	MsgAck
	MsgError
)

// Message represents a sync protocol message
type Message struct {
	Type    MessageType
	Payload interface{}
}

// SignatureRequest requests file signatures from a peer
type SignatureRequest struct {
	ContainerID string
	VolumePath  string
}

// SignatureResponse contains file signatures
type SignatureResponse struct {
	Signatures []FileSignature
}

// DeltaRequest requests specific files
type DeltaRequest struct {
	Files []string
}

// FileTransfer contains file data
type FileTransfer struct {
	Meta FileSignature
	Data []byte
}

// AckMessage acknowledges receipt
type AckMessage struct {
	Success bool
	Message string
}

// ErrorMessage reports an error
type ErrorMessage struct {
	Code    int
	Message string
}

// Protocol handles sync message encoding/decoding
type Protocol struct {
	encoder *gob.Encoder
	decoder *gob.Decoder
}

// NewProtocol creates a new protocol handler
func NewProtocol(rw io.ReadWriter) *Protocol {
	return &Protocol{
		encoder: gob.NewEncoder(rw),
		decoder: gob.NewDecoder(rw),
	}
}

// Send sends a message
func (p *Protocol) Send(msg Message) error {
	return p.encoder.Encode(msg)
}

// Receive receives a message
func (p *Protocol) Receive() (Message, error) {
	var msg Message
	err := p.decoder.Decode(&msg)
	return msg, err
}

// SendSignatureRequest sends a signature request
func (p *Protocol) SendSignatureRequest(containerID, volumePath string) error {
	return p.Send(Message{
		Type: MsgSignatureRequest,
		Payload: SignatureRequest{
			ContainerID: containerID,
			VolumePath:  volumePath,
		},
	})
}

// SendSignatures sends file signatures
func (p *Protocol) SendSignatures(signatures []FileSignature) error {
	return p.Send(Message{
		Type: MsgSignatureResponse,
		Payload: SignatureResponse{
			Signatures: signatures,
		},
	})
}

// SendDeltaRequest requests specific files
func (p *Protocol) SendDeltaRequest(files []string) error {
	return p.Send(Message{
		Type: MsgDeltaRequest,
		Payload: DeltaRequest{
			Files: files,
		},
	})
}

// SendFile sends a file
func (p *Protocol) SendFile(meta FileSignature, data []byte) error {
	return p.Send(Message{
		Type: MsgFileTransfer,
		Payload: FileTransfer{
			Meta: meta,
			Data: data,
		},
	})
}

// SendAck sends an acknowledgment
func (p *Protocol) SendAck(success bool, message string) error {
	return p.Send(Message{
		Type: MsgAck,
		Payload: AckMessage{
			Success: success,
			Message: message,
		},
	})
}

// SendError sends an error message
func (p *Protocol) SendError(code int, message string) error {
	return p.Send(Message{
		Type: MsgError,
		Payload: ErrorMessage{
			Code:    code,
			Message: message,
		},
	})
}

func init() {
	// Register types for gob encoding
	gob.Register(SignatureRequest{})
	gob.Register(SignatureResponse{})
	gob.Register(DeltaRequest{})
	gob.Register(FileTransfer{})
	gob.Register(AckMessage{})
	gob.Register(ErrorMessage{})
}
