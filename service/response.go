package service

import (
	"os"
	"path/filepath"

	"github.com/TheLeeeo/grpc-hole/cli/vars"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/spf13/viper"
)

const (
	responsesDirName = "responses"
)

func SaveResponseFile(serviceDir string, method string, data []byte) error {
	responseDir := filepath.Join(serviceDir, responsesDirName)

	if err := createDir(responseDir); err != nil {
		return err
	}

	path := filepath.Join(responseDir, method+".json")

	if err := os.WriteFile(path, data, os.ModePerm); err != nil {
		return err
	}

	return nil
}

func createDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.Mkdir(path, os.ModePerm)
	}

	return nil
}

func LoadResponse(serviceName, methodName string, mDesc *desc.MessageDescriptor) (*dynamic.Message, error) {
	baseDir := viper.GetString(vars.SerivceDirKey)
	responseDir := filepath.Join(baseDir, serviceName, responsesDirName)
	path := filepath.Join(responseDir, methodName+".json")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	msg := dynamic.NewMessage(mDesc)
	if err := msg.UnmarshalJSON(data); err != nil {
		return nil, err
	}

	return msg, nil
}