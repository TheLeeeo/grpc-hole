package methodhandler

import (
	"encoding/json"
	"os"

	"github.com/TheLeeeo/grpc-hole/service"
	"github.com/hashicorp/go-hclog"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"google.golang.org/grpc"
)

type MethodHandler interface {
	Name() string
	Handle(stream grpc.ServerStream) error
}

type methodHandler struct {
	method *desc.MethodDescriptor
	lg     hclog.Logger
}

func New(method *desc.MethodDescriptor, logger hclog.Logger) MethodHandler {
	return &methodHandler{
		method: method,
		lg:     logger,
	}
}

func (h *methodHandler) Name() string {
	return h.method.GetName()
}

func (h *methodHandler) Handle(stream grpc.ServerStream) error {
	inputMsg := dynamic.NewMessage(h.method.GetInputType())
	if err := stream.RecvMsg(inputMsg); err != nil {
		return err
	}

	inputJSON, _ := inputMsg.MarshalJSON()
	h.lg.Info("Received request", "Method", h.method.GetName(), "Input", string(inputJSON))

	outType := h.method.GetOutputType()
	var out *dynamic.Message

	respTemplate, err := service.LoadResponse(h.method.GetService().GetName(), h.method.GetName())
	if err != nil {
		// If the error is something else than "file not found", return it.
		if !os.IsNotExist(err) {
			return err
		}
		out = CreatePopulatedMessage(outType)
	}

	if respTemplate != nil {
		var inputMap map[string]any
		if err := json.Unmarshal(inputJSON, &inputMap); err != nil {
			return err
		}

		outJson, err := service.ParseTemplate(inputMap, respTemplate)
		if err != nil {
			h.lg.Error("Failed to parse template", "Method", h.method.GetName(), "Error", err)
			return err
		}

		out = dynamic.NewMessage(outType)
		if err := out.UnmarshalJSON(outJson); err != nil {
			h.lg.Error("Failed to unmarshal json", "Method", h.method.GetName(), "Error", err)
			return err
		}
	}

	return stream.SendMsg(out)
}