/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package triton

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"
	pb "github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/repository/triton/config"
)

func TestFindModelVersionFolder(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name            string
		folders         []string
		artifactVersion *uint32
		found           bool
		expectedFolder  string
		error           bool
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
			error:           false,
		},
		{
			name:            "MidVersion",
			folders:         []string{"1", "2", "3"},
			artifactVersion: getArtifactVersion(2),
			found:           true,
			expectedFolder:  "2",
			error:           false,
		},
		{
			name:           "HighestVersion",
			folders:        []string{"1", "2", "3"},
			found:          true,
			expectedFolder: "3",
			error:          false,
		},
		{ // We allow users to put their artifacts in the top level folder so no error here even though this is not a
			// proper Triton structure which assumes numbered version folders.
			name:           "NoVersionFolders",
			folders:        []string{},
			found:          false,
			expectedFolder: "hash",
			error:          false,
		},
		{
			name:            "NoVersionFoldersArtifactVersion",
			folders:         []string{"x"},
			artifactVersion: getArtifactVersion(2),
			found:           false,
			error:           true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rclonePath := t.TempDir()
			modelFolder := filepath.Join(rclonePath, "hash") // pretend hash
			err := os.MkdirAll(modelFolder, fs.ModePerm)
			g.Expect(err).To(BeNil())
			for _, folder := range test.folders {
				versionFolder := filepath.Join(modelFolder, folder)
				err := os.MkdirAll(versionFolder, fs.ModePerm)
				g.Expect(err).To(BeNil())
			}
			logger := log.New()
			triton := TritonRepositoryHandler{logger: logger}
			foundPath, found, err := triton.FindModelVersionFolder("foo", test.artifactVersion, modelFolder)
			if test.error {
				g.Expect(err).ToNot(BeNil())
				g.Expect(found).To(BeFalse())
			} else {
				if !test.found {
					g.Expect(found).To(BeFalse())
					g.Expect(foundPath).To(Equal(modelFolder))
				} else {
					g.Expect(err).To(BeNil())
					g.Expect(filepath.Base(foundPath)).To(Equal(test.expectedFolder))
				}
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
			err = triton.UpdateModelRepository(test.modelName, versionPath, true, repoPath)
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

func TestCopyNonConfigFilesToModelRepo(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                 string
		rcloneFiles          []string
		rcloneFolders        []string
		repoFoldersToCreate  []string
		expectedRepoPaths    []string
		notExpectedRepoPaths []string
	}

	tests := []test{
		{
			name:              "files",
			rcloneFiles:       []string{"t"},
			expectedRepoPaths: []string{"t"},
		},
		{
			name:                 "files and folders",
			rcloneFiles:          []string{"t"},
			rcloneFolders:        []string{"folder"},
			expectedRepoPaths:    []string{"t"},
			notExpectedRepoPaths: []string{"folder"},
		},
		{
			name:                 "files and folders with config.pbtxt and existing folders in repo",
			rcloneFiles:          []string{"t", "config.pbtxt"},
			rcloneFolders:        []string{"folder"},
			repoFoldersToCreate:  []string{"1"},
			expectedRepoPaths:    []string{"t", "1"},
			notExpectedRepoPaths: []string{"folder", "config.pbtxt"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rclonePath := t.TempDir()
			repoPath := t.TempDir()
			for _, filename := range test.rcloneFiles {
				path := filepath.Join(rclonePath, filename)
				err := os.WriteFile(path, []byte{}, fs.ModePerm)
				g.Expect(err).To(BeNil())
			}
			for _, filename := range test.rcloneFolders {
				path := filepath.Join(rclonePath, filename)
				err := os.MkdirAll(path, os.ModePerm)
				g.Expect(err).To(BeNil())
			}
			for _, filename := range test.repoFoldersToCreate {
				path := filepath.Join(repoPath, filename)
				err := os.MkdirAll(path, os.ModePerm)
				g.Expect(err).To(BeNil())
			}
			err := copyNonConfigFilesToModelRepo(rclonePath, repoPath)
			g.Expect(err).To(BeNil())
			for _, filename := range test.expectedRepoPaths {
				path := filepath.Join(repoPath, filename)
				_, err := os.Stat(path)
				g.Expect(err).To(BeNil())
			}
			for _, filename := range test.notExpectedRepoPaths {
				path := filepath.Join(repoPath, filename)
				_, err := os.Stat(path)
				g.Expect(err).ToNot(BeNil())
			}
		})
	}
}

func TestGetModelConfig(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                  string
		modelConfig           *pb.ModelConfig
		expectedInstanceCount uint32
	}

	tests := []test{
		{
			name: "onnxruntime backend without a count defaults to 2",
			modelConfig: &pb.ModelConfig{
				Backend:       "onnxruntime",
				InstanceGroup: []*pb.ModelInstanceGroup{{Kind: pb.ModelInstanceGroup_KIND_CPU}},
			},
			expectedInstanceCount: 2,
		}, {
			name: "tensorflow backend without a count defaults to 2",
			modelConfig: &pb.ModelConfig{
				Backend:       "tensorflow",
				InstanceGroup: []*pb.ModelInstanceGroup{{Kind: pb.ModelInstanceGroup_KIND_CPU}},
			},
			expectedInstanceCount: 2,
		}, {
			name: "other backend without a count defaults to 2",
			modelConfig: &pb.ModelConfig{
				Backend:       "other",
				InstanceGroup: []*pb.ModelInstanceGroup{{Kind: pb.ModelInstanceGroup_KIND_CPU}},
			},
			expectedInstanceCount: 1,
		}, {
			name: "not KIND_CPU defaults to 1",
			modelConfig: &pb.ModelConfig{
				Backend:       "tensorflow",
				InstanceGroup: []*pb.ModelInstanceGroup{{Kind: pb.ModelInstanceGroup_KIND_GPU}},
			},
			expectedInstanceCount: 1,
		}, {
			name: "onnxruntime backend with count",
			modelConfig: &pb.ModelConfig{
				Backend:       "onnxruntime",
				InstanceGroup: []*pb.ModelInstanceGroup{{Count: 8, Kind: pb.ModelInstanceGroup_KIND_CPU}},
			},
			expectedInstanceCount: 8,
		}, {
			name: "tensorflow backend with count",
			modelConfig: &pb.ModelConfig{
				Backend:       "tensorflow",
				InstanceGroup: []*pb.ModelInstanceGroup{{Count: 5, Kind: pb.ModelInstanceGroup_KIND_CPU}},
			},
			expectedInstanceCount: 5,
		}, {
			name: "no instance group defaults to 1",
			modelConfig: &pb.ModelConfig{
				Backend: "tensorflow",
			},
			expectedInstanceCount: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rclonePath := t.TempDir()
			// Create rclone config.pbtxt
			configPathRclone := filepath.Join(rclonePath, TritonConfigFile)
			data, err := prototext.Marshal(test.modelConfig)
			g.Expect(err).To(BeNil())
			err = os.WriteFile(configPathRclone, data, fs.ModePerm)
			g.Expect(err).To(BeNil())

			logger := log.New()
			triton := TritonRepositoryHandler{logger: logger}
			modelConfig, err := triton.GetModelConfig(rclonePath)

			g.Expect(err).To(BeNil())
			g.Expect(modelConfig.InstanceCount).To(Equal(test.expectedInstanceCount))
			g.Expect(modelConfig.Resource).To(Equal(agent.ModelConfig_MEMORY))
		})
	}
}
