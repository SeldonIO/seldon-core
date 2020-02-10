//+build ignore

// TensorFlow Serving gRPC interface generator.
//
// This script works around a bunch of issues (as of 2019-08-25) between Go's
// protobuf compiler plugin, Go modules, and definitions of TensorFlow and
// TensorFlow Serving proto files. It assumes that protoc and protoc-gen-go are
// on your PATH.
//
//	git clone -b r1.15 https://github.com/tensorflow/tensorflow.git
//	git clone -b r1.14 https://github.com/tensorflow/serving.git
//	go run protoc.go
//	go mod edit -replace=github.com/tensorflow/tensorflow/tensorflow/go/core=./proto/tensorflow/core
//	cd proto/tensorflow/core && go mod init github.com/tensorflow/tensorflow/tensorflow/go/core && cd -
//	go build ./proto/tensorflow/serving
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

const protoDir = "proto/"

var opts = []string{"-Iserving", "-Itensorflow"}
var cmds = []ProtocCmd{{
	PkgDir: protoDir + "tensorflow/core/example",
	Inputs: []string{"tensorflow/tensorflow/core/example/*.proto"},
}, {
	PkgDir: protoDir + "tensorflow/core/framework",
	Inputs: []string{"tensorflow/tensorflow/core/framework/*.proto"},
}, {
	PkgDir: protoDir + "tensorflow/core/lib/core",
	Inputs: []string{"tensorflow/tensorflow/core/lib/core/*.proto"},
}, {
	GoOpts: "import_path=protobuf",
	PkgDir: protoDir + "tensorflow/core/protobuf",
	Inputs: []string{
		"tensorflow/tensorflow/core/protobuf/*.proto",
		"tensorflow/tensorflow/stream_executor/*.proto",
	},
}, {
	GoOpts: "plugins=grpc,import_path=serving",
	PkgDir: protoDir + "tensorflow/serving",
	Inputs: []string{
		"serving/tensorflow_serving/apis/*.proto",
		"serving/tensorflow_serving/config/*.proto",
		"serving/tensorflow_serving/core/*.proto",
		"serving/tensorflow_serving/sources/storage_path/*.proto",
		"serving/tensorflow_serving/util/*.proto",
	},
}}

func main() {
	for _, cmd := range cmds {
		fmt.Fprintln(os.Stderr, "==>", cmd.PkgDir)
		if err := cmd.Run(); err != nil {
			if e, ok := err.(*exec.ExitError); ok {
				os.Exit(e.ExitCode())
			}
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

// ProtocCmd executes protoc to generate sources for a single Go package.
type ProtocCmd struct {
	GoOpts string   // --go_out options
	PkgDir string   // Final output directory
	Inputs []string // Input files or glob patterns
}

func (pc *ProtocCmd) Run() error {
	// Use a temporary protoc output directory
	root := filepath.Dir(pc.PkgDir)
	os.MkdirAll(root, 0777)
	tmp, err := ioutil.TempDir(root, filepath.Base(pc.PkgDir)+".")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	// Run protoc
	cmd := exec.Command("protoc", opts...)
	cmd.Args = append(cmd.Args, "--go_out="+pc.GoOpts+":"+tmp)
	for _, in := range pc.Inputs {
		files, err := filepath.Glob(in)
		if err != nil {
			return err
		}
		cmd.Args = append(cmd.Args, files...)
	}
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	err = cmd.Run()

	// Move generated files to PkgDir
	os.RemoveAll(pc.PkgDir)
	if err := os.MkdirAll(pc.PkgDir, 0777); err != nil {
		return err
	}
	walkErr := filepath.Walk(tmp, func(path string, fi os.FileInfo, err error) error {
		if err == nil && fi.Mode().IsRegular() {
			err = os.Rename(path, filepath.Join(pc.PkgDir, fi.Name()))
		}
		return err
	})
	if err == nil {
		err = walkErr
	}

	//for _, in := range pc.Inputs {
	//	files, err := filepath.Glob(in)
	//	if err != nil {
	//		return err
	//	}
	//	for _, path := range files {
	//		err = os.Rename(path, filepath.Join(pc.PkgDir, filepath.Base(path)))
	//	}
	//}

	return err
}
