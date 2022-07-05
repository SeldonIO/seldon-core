package triton

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	pb "github.com/seldonio/seldon-core/scheduler/pkg/agent/repository/triton/config"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

func TestFindModelVersionFolder(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name            string
		folders         []string
		artifactVersion *uint32
		found           bool
		expectedFolder  string
	}

	getArtifactVersion := func(version uint32) *uint32 {
		return &version
	}
	tests := []test{
		{
			name:            "Simple",
			folders:         []string{"1", "2", "3"},
			artifactVersion: getArtifactVersion(1),
			found:           true,
			expectedFolder:  "1",
		},
		{
			name:            "MidVersion",
			folders:         []string{"1", "2", "3"},
			artifactVersion: getArtifactVersion(2),
			found:           true,
			expectedFolder:  "2",
		},
		{
			name:           "HighestVersion",
			folders:        []string{"1", "2", "3"},
			found:          true,
			expectedFolder: "3",
		},
		{
			name:    "NoVersionFolders",
			folders: []string{"x"},
			found:   false,
		},
		{
			name:            "NoVersionFoldersArtifactVersion",
			folders:         []string{"x"},
			artifactVersion: getArtifactVersion(2),
			found:           false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rclonePath := t.TempDir()
			modelFolder := filepath.Join(rclonePath, "12341234") // pretend hash
			for _, folder := range test.folders {
				versionFolder := filepath.Join(modelFolder, folder)
				err := os.MkdirAll(versionFolder, fs.ModePerm)
				g.Expect(err).To(BeNil())
			}
			logger := log.New()
			triton := TritonRepositoryHandler{logger: logger}
			foundPath, err := triton.FindModelVersionFolder("foo", test.artifactVersion, modelFolder)
			if !test.found {
				g.Expect(err).ToNot(BeNil())
				g.Expect(foundPath).To(Equal(""))
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(filepath.Base(foundPath)).To(Equal(test.expectedFolder))
			}
		})
	}
}

func TestUpdateModelRepository(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name       string
		config     *pb.ModelConfig
		repoConfig *pb.ModelConfig
		modelName  string
	}

	tests := []test{
		{
			name: "Simple",
			config: &pb.ModelConfig{
				Name:         "densenet_onnx",
				Platform:     "onnxruntime_onnx",
				MaxBatchSize: 0,
				Input: []*pb.ModelInput{
					{
						Name:     "data_0",
						DataType: pb.DataType_TYPE_FP32,
						Format:   pb.ModelInput_FORMAT_NCHW,
						Dims:     []int64{3, 224, 224},
						Reshape:  &pb.ModelTensorReshape{Shape: []int64{1, 3, 224, 224}},
					},
				},
				Output: []*pb.ModelOutput{
					{
						Name:          "fc6_1",
						DataType:      pb.DataType_TYPE_FP32,
						Dims:          []int64{1000},
						Reshape:       &pb.ModelTensorReshape{Shape: []int64{1, 1000, 1, 1}},
						LabelFilename: "densenet_labels.txt",
					},
				},
			},
			modelName: "foo",
		},
		{
			name: "Simple",
			config: &pb.ModelConfig{
				Name:     "densenet_onnx",
				Platform: "onnxruntime_onnx",
			},
			modelName: "foo",
			repoConfig: &pb.ModelConfig{
				Name:     "foo",
				Platform: "onnxruntime_onnx",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rclonePath := t.TempDir()
			repoPath := t.TempDir()
			// Create rclone config.pbtxt
			configPathRclone := filepath.Join(rclonePath, TritonConfigFile)
			data, err := prototext.Marshal(test.config)
			g.Expect(err).To(BeNil())
			err = os.WriteFile(configPathRclone, data, fs.ModePerm)
			g.Expect(err).To(BeNil())
			repoPathConfig := filepath.Join(repoPath, TritonConfigFile)
			// Create repo config.pbtxt
			if test.repoConfig != nil {
				data, err := prototext.Marshal(test.repoConfig)
				g.Expect(err).To(BeNil())
				err = os.WriteFile(repoPathConfig, data, fs.ModePerm)
				g.Expect(err).To(BeNil())
			}
			versionPath := filepath.Join(rclonePath, "1")
			logger := log.New()
			triton := TritonRepositoryHandler{logger: logger}
			err = triton.UpdateModelRepository(test.modelName, versionPath, repoPath)
			g.Expect(err).To(BeNil())
			_, err = os.Stat(repoPathConfig)
			g.Expect(err).To(BeNil())
			config, err := triton.loadConfigFromFile(repoPathConfig)
			g.Expect(err).To(BeNil())
			g.Expect(config.Name).To(Equal(test.modelName))
		})
	}
}

func TestLoadFromBytes(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name     string
		data     []byte
		expected *pb.ModelConfig
		error    bool
	}

	tests := []test{
		{
			name: "onnx",
			data: []byte(`name: "densenet_onnx"
platform: "onnxruntime_onnx"
max_batch_size : 0
input [
  {
    name: "data_0"
    data_type: TYPE_FP32
    format: FORMAT_NCHW
    dims: [ 3, 224, 224 ]
    reshape { shape: [ 1, 3, 224, 224 ] }
  }
]
output [
  {
    name: "fc6_1"
    data_type: TYPE_FP32
    dims: [ 1000 ]
    reshape { shape: [ 1, 1000, 1, 1 ] }
    label_filename: "densenet_labels.txt"
  }
]`),
			expected: &pb.ModelConfig{
				Name:         "densenet_onnx",
				Platform:     "onnxruntime_onnx",
				MaxBatchSize: 0,
				Input: []*pb.ModelInput{
					{
						Name:     "data_0",
						DataType: pb.DataType_TYPE_FP32,
						Format:   pb.ModelInput_FORMAT_NCHW,
						Dims:     []int64{3, 224, 224},
						Reshape:  &pb.ModelTensorReshape{Shape: []int64{1, 3, 224, 224}},
					},
				},
				Output: []*pb.ModelOutput{
					{
						Name:          "fc6_1",
						DataType:      pb.DataType_TYPE_FP32,
						Dims:          []int64{1000},
						Reshape:       &pb.ModelTensorReshape{Shape: []int64{1, 1000, 1, 1}},
						LabelFilename: "densenet_labels.txt",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := &TritonRepositoryHandler{}
			c, err := s.loadConfigFromBytes(test.data)
			if !test.error {
				g.Expect(err).To(BeNil())
				g.Expect(proto.Equal(c, test.expected)).To(BeTrue())
			} else {
				g.Expect(err).ToNot(BeNil())
			}
		})
	}
}
