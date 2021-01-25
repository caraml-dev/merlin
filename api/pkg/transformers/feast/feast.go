package feast

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"unsafe"

	"github.com/buger/jsonparser"
	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/pkg/errors"

	"github.com/gojek/merlin/pkg/transformers/feast/spec"
)

// Options for the Feast transformer.
type Options struct {
	ServingAddress string `envconfig:"FEAST_SERVING_ADDRESS" required:"true"`
	ServingPort    int    `envconfig:"FEAST_SERVING_PORT" required:"true"`
}

// Feast wraps Feast serving client to retrieve features.
type Feast struct {
	options *Options

	client *feast.GrpcClient

	requestParsing []spec.RequestParsing
}

// New initializes a new Feast client.
func New(o *Options, requestParsingConfig []spec.RequestParsing) *Feast {
	cli, err := feast.NewGrpcClient(o.ServingAddress, o.ServingPort)
	if err != nil {
		panic(err)
	}
	return &Feast{
		options:        o,
		client:         cli,
		requestParsing: requestParsingConfig,
	}
}

type FeastFeature struct {
	Columns []string  `json:"columns"`
	Data    []float64 `json:"data"`
}

func (f *Feast) TransformHandler(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	log.Println("TransformHandler")

	ctx := r.Context()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	feastFeatures := make(map[string]FeastFeature, len(f.requestParsing))

	for _, config := range f.requestParsing {
		// TODO: validate the config
		log.Println(config)

		entities := []feast.Row{}
		entityIDs := make(map[string]int, len(config.Entities))

		for _, entity := range config.Entities {
			val, err := getValueFromJSONPayload(body, *entity)
			if err != nil {
				log.Printf("JSON Path %s not found in request payload", entity.JsonPath)
			}
			entities = append(entities, feast.Row{
				entity.Name: feast.StrVal(val),
			})

			entityIDs[entity.Name] = 1
		}

		features := []string{}
		for _, feature := range config.Features {
			features = append(features, feature.Name)
		}

		feastRequest := feast.OnlineFeaturesRequest{
			Project:  config.Project,
			Entities: entities,
			Features: features,
		}

		feastResponse, err := f.client.GetOnlineFeatures(ctx, &feastRequest)
		if err != nil {
			return nil, err
		}

		log.Println(feastResponse.Rows())

		data := []float64{}
		for i, feastRow := range feastResponse.Rows() {
			log.Println(i, feastRow)
			for featureID, featureValue := range feastRow {
				log.Println("ID:", featureID, "Value:", featureValue)
				if _, ok := entityIDs[featureID]; ok {
					continue
				}

				data = append(data, featureValue.GetDoubleVal())
			}
		}

		feastFeatures[config.Entities[0].Name] = FeastFeature{
			Columns: features,
			Data:    data,
		}
	}

	feastFeatureJSON, err := json.Marshal(feastFeatures)
	if err != nil {
		return nil, err
	}

	out, err := jsonparser.Set(body, feastFeatureJSON, "feast_features")
	if err != nil {
		return nil, err
	}

	return out, err
}

// TODO: return feastTypes.Value instead of string
func getValueFromJSONPayload(body []byte, entity spec.Entity) (string, error) {
	// Retrieve value using JSON path
	value, typez, _, _ := jsonparser.Get(body, strings.Split(entity.JsonPath, ".")...)

	switch typez {
	case jsonparser.String, jsonparser.Number, jsonparser.Boolean:
		// See: https://github.com/buger/jsonparser/blob/master/bytes_unsafe.go#L31
		return *(*string)(unsafe.Pointer(&value)), nil
	case jsonparser.Null:
		return "", nil
	case jsonparser.NotExist:
		return "", errors.Errorf("Field %s not found in the request payload: Key path not found", entity.JsonPath)
	default:
		return "", errors.Errorf(
			"Field %s can not be parsed, unsupported type: %s", entity.JsonPath, typez.String())
	}
}
