package hexerr

// hexerr is a Hexagonal architecture error wrapper

import (
	"bytes"
	"fmt"
	"runtime"

	"github.com/go-errors/errors"
)

const (
	InterfaceLayer   Layer = 1
	ApplicationLayer Layer = 2
	DomainLayer      Layer = 3
	Infrastructure   Layer = 4

	EntityNotFound            Code = 1
	Forbidden                 Code = 2
	InvalidArguments          Code = 3
	Internal                  Code = 4
	InvalidRequest            Code = 5
	UniqueConstraintViolation Code = 6
)

type Layer int
type Code int

type HexagonalError struct {
	Layer        Layer
	Code         Code
	Message      string
	ThrownAtLine string
	stack        []uintptr
	frames       []errors.StackFrame
}

func NewInterfaceLayerError(code Code, message string) error {
	_, file, line, _ := runtime.Caller(1)

	stack := make([]uintptr, errors.MaxStackDepth)
	length := runtime.Callers(2, stack[:])

	return &HexagonalError{
		Layer:        InterfaceLayer,
		Code:         code,
		Message:      message,
		ThrownAtLine: fmt.Sprintf("%v:%v", file, line),
		stack:        stack[:length],
	}
}

func NewApplicationLayerError(code Code, message string) error {
	_, file, line, _ := runtime.Caller(1)

	stack := make([]uintptr, errors.MaxStackDepth)
	length := runtime.Callers(2, stack[:])

	return &HexagonalError{
		Layer:        ApplicationLayer,
		Code:         code,
		Message:      message,
		ThrownAtLine: fmt.Sprintf("%v:%v", file, line),
		stack:        stack[:length],
	}
}

func NewInfrastructureLayerError(code Code, message string) error {
	_, file, line, _ := runtime.Caller(1)

	stack := make([]uintptr, errors.MaxStackDepth)
	length := runtime.Callers(2, stack[:])

	return &HexagonalError{
		Layer:        Infrastructure,
		Code:         code,
		Message:      message,
		ThrownAtLine: fmt.Sprintf("%v:%v", file, line),
		stack:        stack[:length],
	}
}

func NewDomainLayerError(code Code, message string) error {
	_, file, line, _ := runtime.Caller(1)

	stack := make([]uintptr, errors.MaxStackDepth)
	length := runtime.Callers(2, stack[:])

	return &HexagonalError{
		Layer:        DomainLayer,
		Code:         code,
		Message:      message,
		ThrownAtLine: fmt.Sprintf("%v:%v", file, line),
		stack:        stack[:length],
	}
}

func (h *HexagonalError) Error() string {
	return h.Message
}

func (h *HexagonalError) Details() string {
	return fmt.Sprintf(
		"error: %v, code: %v, layer: %v, at: %v",
		h.Message,
		h.Code,
		h.Layer,
		h.ThrownAtLine,
	)
}

func (h *HexagonalError) StackTrace() string {
	return fmt.Sprintf("error: %v\n%v", h.Message, string(h.stackBytes()))
}

func (h *HexagonalError) stackBytes() []byte {
	buf := bytes.Buffer{}

	for _, frame := range h.stackFrames() {
		buf.WriteString(frame.String())
	}

	return buf.Bytes()
}

func (h *HexagonalError) stackFrames() []errors.StackFrame {
	if h.frames == nil {
		h.frames = make([]errors.StackFrame, len(h.stack))

		for i, pc := range h.stack {
			h.frames[i] = errors.NewStackFrame(pc)
		}
	}

	return h.frames
}
