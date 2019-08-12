package types

// this file is used to define serving type

type ServingType string

// three serving types.
const (
	// tensorflow
	ServingTF ServingType = "TENSORFLOW"
	// tensorrt
	ServingTRT ServingType = "TENSORRT"
	// custom
	ServingCustom ServingType = "CUSTOM"
)

var SERVING_CHARTS = map[string]string{
	"tensorflow-serving-0.2.0":        "Tensorflow",
	"tensorrt-inference-server-0.0.1": "TensorRT",
}

func KeyMapServingType(servingKey string) ServingType {
	switch servingKey {
	case "tf", "tf-serving", "tensorflow-serving":
		return ServingTF
	case "trt", "trt-serving", "tensorrt-serving":
		return ServingTRT
	case "custom", "custom-serving":
		return ServingCustom
	default:
		return ServingType("")
	}
}
