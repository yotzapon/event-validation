package validate

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/labstack/echo"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type Request struct {
	ValidationType validationType `json:"validationType"`
	Values         struct {
		Name   string                 `json:"name"`
		Schema map[string]interface{} `json:"schema"`
	} `json:"values"`
}

type eventYamlModel struct {
	Channels map[string]interface{}
}

type publishModel struct {
	Message struct {
		Examples []yamlExample `json:"examples"`
	} `json:"message"`
}

type yamlExample struct {
	Payload map[string]interface{} `json:"payload"`
}

type validationType string

var (
	errInvalidName   = errors.New("invalid name")
	errInvalidSchema = errors.New("invalid schema type")

	topic         validationType = "topic"
	Schema        validationType = "schema"
	supportedType                = []validationType{topic, Schema}
)

func (s *service) Validate(c echo.Context) error {
	req := new(Request)
	if err := c.Bind(req); err != nil {
		return c.JSON(400, ErrorResp{&Resp{
			Code: 400, Message: "invalid json",
		}})
	}

	if !isValidType(req.ValidationType) {
		return c.JSON(400, ErrorResp{&Resp{
			Code: 400, Message: "unknown validationType",
		}})
	}

	directoryPath := s.eventsFileConfig.Dir
	yamlFiles, err := listYAMLFilesInDirectory(directoryPath)
	if err != nil {
		fmt.Printf("error listing YAML files: %v\n", err)
		return err
	}

	if len(yamlFiles) == 0 {
		fmt.Println("no .yaml files found in the directory.")
		return errors.New("file not found")
	}

	switch req.ValidationType {
	case topic:
		err = s.validateEventName(yamlFiles, req.Values.Name)
		if err != nil {
			return c.JSON(400, ErrorResp{&Resp{
				Code: 400, Message: err.Error(),
			}})
		}
		return c.JSON(200, Resp{200, "Ok"})

	case Schema:

		err = s.validateEventName(yamlFiles, req.Values.Name)
		if err != nil {
			return c.JSON(400, ErrorResp{&Resp{
				Code: 400, Message: err.Error(),
			}})
		}

		err := s.validateSchema(req.Values.Name, req.Values.Schema)
		if err != nil {
			return c.JSON(400, ErrorResp{&Resp{
				Code: 400, Message: err.Error(),
			}})
		}
		return c.JSON(200, Resp{200, "Ok"})

	default:
		return c.JSON(400, ErrorResp{&Resp{
			Code: 400, Message: "validationType is not supported",
		}})
	}
}

func isValidType(t validationType) bool {
	for _, vt := range supportedType {
		if t == vt {
			return true
		}
	}
	return false
}

func listYAMLFilesInDirectory(dirPath string) ([]string, error) {
	var yamlFiles []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println(err)
			return err
		}

		// Check if it's a regular file and has a ".yaml" extension
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".yaml") {
			yamlFiles = append(yamlFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return yamlFiles, nil
}

func (s *service) validateEventName(yamlFiles []string, eventName string) error {
	for _, file := range yamlFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			log.Printf("error reading the YAML file: %v", err.Error())
			return err
		}

		// Unmarshal the YAML data into the eventConfig struct
		var eventData eventYamlModel
		if err = yaml.Unmarshal(data, &eventData); err != nil {
			log.Printf("error unmarshaling the YAML data: %v", err)
			return err
		}

		if isValidEventName(&eventData, eventName) {
			s.eventData = &eventData
			return nil
		}

		s.eventData = &eventData
	}
	return errInvalidName
}

func (s *service) validateSchema(evtName string, reqSchema map[string]interface{}) error {
	exampleMap, err := s.getExampleYaml(evtName)
	if err != nil {
		return err
	}

	err = s.inspectJSON(exampleMap, reqSchema)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) inspectJSON(exampleMap, reqSchema map[string]interface{}) error {
	if len(exampleMap) != len(reqSchema) {
		return fmt.Errorf("%v: %v", errInvalidSchema, "Some fields are added or missing")
	}

	for ek, ev := range exampleMap {
		// check yaml field is existing in json
		jsonFieldVal, ok := reqSchema[ek]
		if !ok {
			return fmt.Errorf("%v: %v", errInvalidSchema, ek)
		}

		// check type of yaml field and json field is the same
		if reflect.TypeOf(ev) != reflect.TypeOf(jsonFieldVal) {
			return fmt.Errorf("%v: %v", errInvalidSchema, ek)
		}

		// case: yaml field is nest object
		yamlNestObj, ok := ev.(map[string]interface{})
		if ok {
			err := s.inspectJSON(yamlNestObj, jsonFieldVal.(map[string]interface{}))
			if err != nil {
				return err
			}
		}

		// case: yaml field is array object
		yamlArr, ok := ev.([]interface{})
		if ok {
			yamlArrMap, ok := yamlArr[0].(map[string]interface{})
			if !ok {
				return fmt.Errorf("%v: %v", errInvalidSchema, ek)
			}

			jsonArr, ok := jsonFieldVal.([]interface{})
			if !ok {
				return fmt.Errorf("%v: %v", errInvalidSchema, ek)
			}

			for _, jv := range jsonArr {
				jsonArrMap, ok := jv.(map[string]interface{})
				if !ok {
					return fmt.Errorf("%v: %v", errInvalidSchema, ek)
				}
				err := s.inspectJSON(yamlArrMap, jsonArrMap)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (s *service) getExampleYaml(name string) (map[string]interface{}, error) {
	yamlSchema, ok := s.eventData.Channels[name].(map[string]interface{})
	if !ok {
		msg := fmt.Sprintf("The '%v' field is not a valid map.", name)
		return nil, errors.New(msg)
	}

	yamlByte, err := json.Marshal(yamlSchema["publish"])
	if err != nil {
		return nil, err
	}

	yamlPublishMap := new(publishModel)
	err = json.Unmarshal(yamlByte, yamlPublishMap)
	if err != nil {
		return nil, err
	}

	return yamlPublishMap.Message.Examples[0].Payload, nil

}
func isValidEventName(eventCfg *eventYamlModel, eventName string) bool {
	val := eventCfg.Channels[eventName]
	if val == nil {
		return false
	}
	return true
}
