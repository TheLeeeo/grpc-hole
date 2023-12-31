package methodhandler

import (
	"encoding/json"
	"os"

	"github.com/TheLeeeo/grpc-hole/fieldselector"
	"github.com/TheLeeeo/grpc-hole/service"
	"github.com/TheLeeeo/grpc-hole/templateparse"
	"github.com/hashicorp/go-hclog"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"google.golang.org/grpc"
)

// A static handler returns a response based on a template file.
type staticHandler struct {
	method *desc.MethodDescriptor
	lg     hclog.Logger
}

func NewStaticHandler(method *desc.MethodDescriptor, logger hclog.Logger) Handler {
	return &staticHandler{
		method: method,
		lg:     logger,
	}
}

func (h *staticHandler) Name() string {
	return h.method.GetName()
}

func (h *staticHandler) Handle(stream grpc.ServerStream) error {
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
		out = CreatePopulatedMessage(outType, 0)
	} else {
		var inputMap map[string]any
		if err := json.Unmarshal(inputJSON, &inputMap); err != nil {
			return err
		}

		var templateMap map[string]any
		if err := json.Unmarshal(respTemplate, &templateMap); err != nil {
			return err
		}

		outMap, parseErrs := templateparse.ParseTemplate(fieldselector.Root, inputMap, templateMap)
		// Parsing a template proceeds even if there are errors.
		// We log any the errors and continue.
		for _, err := range parseErrs {
			h.lg.Warn("Encountered error parsing template", "error", err, "location", err.Location())
		}

		outJSON, err := json.Marshal(outMap)
		if err != nil {
			h.lg.Error("Failed to marshal json", "Method", h.method.GetName(), "Error", err)
			return err
		}

		out = dynamic.NewMessage(outType)
		if err := out.UnmarshalJSON(outJSON); err != nil {
			h.lg.Error("Failed to unmarshal json", "Method", h.method.GetName(), "Error", err)
			return err
		}
	}

	return stream.SendMsg(out)
}
