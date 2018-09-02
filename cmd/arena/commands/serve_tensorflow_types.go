package commands

// Common configuration for loading a model being served.
type TensorflowModelConfig struct {
	// Name of the model.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Base path to the model, excluding the version directory.
	// E.g> for a model at /foo/bar/my_model/123, where 123 is the version, the
	// base path is /foo/bar/my_model.
	//
	// (This can be changed once a model is in serving, *if* the underlying data
	// remains the same. Otherwise there are no guarantees about whether the old
	// or new data will be used for model versions currently loaded.)
	BasePath string `protobuf:"bytes,2,opt,name=base_path,json=basePath,proto3" json:"base_path,omitempty"`
	// Type of model.
	ModelType TensorflowModelType `protobuf:"varint,3,opt,name=model_type,json=modelType,proto3,enum=tensorflow.serving.ModelType" json:"model_type,omitempty"`
	// Type of model (e.g. "tensorflow").
	//
	// (This cannot be changed once a model is in serving.)
	ModelPlatform string `protobuf:"bytes,4,opt,name=model_platform,json=modelPlatform,proto3" json:"model_platform,omitempty"`
	// Version policy for the model indicating which version(s) of the model to
	// load and make available for serving simultaneously.
	// The default option is to serve only the latest version of the model.
	//
	// (This can be changed once a model is in serving.)
	ModelVersionPolicy *TensorflowFileSystemStoragePathSourceConfig_ServableVersionPolicy `protobuf:"bytes,7,opt,name=model_version_policy,json=modelVersionPolicy,proto3" json:"model_version_policy,omitempty"`
	// Configures logging requests and responses, to the model.
	//
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

// A policy that dictates which version(s) of a servable should be served.
type TensorflowFileSystemStoragePathSourceConfig_ServableVersionPolicy struct {
	// Types that are valid to be assigned:
	//	*FileSystemStoragePathSourceConfig_ServableVersionPolicy_Latest_
	//	*FileSystemStoragePathSourceConfig_ServableVersionPolicy_All_
	//	*FileSystemStoragePathSourceConfig_ServableVersionPolicy_Specific_
	Latest   *TensorflowFileSystemStoragePathSourceConfig_ServableVersionPolicy_Latest   `protobuf:"bytes,100,opt,name=latest,proto3,oneof" json:"latest,omitempty"`
	All      *TensorflowFileSystemStoragePathSourceConfig_ServableVersionPolicy_All      `protobuf:"bytes,101,opt,name=all,proto3,oneof" json:"all,omitempty"`
	Specific *TensorflowFileSystemStoragePathSourceConfig_ServableVersionPolicy_Specific `protobuf:"bytes,102,opt,name=specific,proto3,oneof" json:"specific,omitempty"`
}

type TensorflowFileSystemStoragePathSourceConfig_ServableVersionPolicy_Latest struct {
	// Number of latest versions to serve. (The default is 1.)
	NumVersions uint32 `protobuf:"varint,1,opt,name=num_versions,json=numVersions,proto3" json:"num_versions,omitempty"`
}

// Serve all versions found on disk.
type TensorflowFileSystemStoragePathSourceConfig_ServableVersionPolicy_All struct {
}

// Serve a specific version (or set of versions).
type TensorflowFileSystemStoragePathSourceConfig_ServableVersionPolicy_Specific struct {
	// The version numbers to serve.
	Versions []int64 `protobuf:"varint,1,rep,packed,name=versions,proto3" json:"versions,omitempty"`
}

// The type of model.
type TensorflowModelType int32

const (
	ModelType_MODEL_TYPE_UNSPECIFIED TensorflowModelType = 0 // Deprecated: Do not use.
	ModelType_TENSORFLOW             TensorflowModelType = 1 // Deprecated: Do not use.
	ModelType_OTHER                  TensorflowModelType = 2 // Deprecated: Do not use.
)

var ModelType_name = map[int32]string{
	0: "MODEL_TYPE_UNSPECIFIED",
	1: "TENSORFLOW",
	2: "OTHER",
}
var ModelType_value = map[string]int32{
	"MODEL_TYPE_UNSPECIFIED": 0,
	"TENSORFLOW":             1,
	"OTHER":                  2,
}
