//   Copyright Steve Sloka 2021
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package watcher

import (
	"context"
	"io/ioutil"
	log "github.com/sirupsen/logrus"

	"github.com/fsnotify/fsnotify"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

type OperationType int

const (
	Create OperationType = iota
	Remove
	Modify
)

const (
	CM_NAME = "seldon-scheduler-config"
	)

type NotifyMessage struct {
	Operation OperationType
	Contents  []byte
}

func WatchConfigmap(clientset *kubernetes.Clientset, namespace string, notifyCh chan<- NotifyMessage) {
	for {
		watcher, err := clientset.CoreV1().ConfigMaps(namespace).Watch(context.TODO(),
			metav1.SingleObject(metav1.ObjectMeta{Name: CM_NAME, Namespace: namespace}))
		if err != nil {
			log.WithError(err).Fatal("Failed to create configmap watcher")
		}
		updateConfigMap(watcher.ResultChan(), notifyCh)
	}
}

func updateConfigMap(eventChannel <-chan watch.Event, notifyCh chan<- NotifyMessage) {
	for {
		event, open := <-eventChannel
		if open {
			switch event.Type {
			case watch.Added:
				fallthrough
			case watch.Modified:
				// Update our endpoint
				if updatedMap, ok := event.Object.(*corev1.ConfigMap); ok {
					if seldonData, ok := updatedMap.Data["seldon.yaml"]; ok {
						notifyCh <- NotifyMessage{
							Operation: Create,
							Contents:  []byte(seldonData),
						}
					}
				}
			case watch.Deleted:
			default:
				// Do nothing
			}
		} else {
			// If eventChannel is closed, it means the server has closed the connection
			return
		}
	}
}

func WatchFile(directory string, notifyCh chan<- NotifyMessage) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					yamlFile, err := ioutil.ReadFile(event.Name)
					if err != nil {
						return
					}
					notifyCh <- NotifyMessage{
						Operation: Modify,
						Contents:  yamlFile,
					}
				} else if event.Op&fsnotify.Create == fsnotify.Create {
					yamlFile, err := ioutil.ReadFile(event.Name)
					if err != nil {
						return
					}
					notifyCh <- NotifyMessage{
						Operation: Create,
						Contents:  yamlFile,
					}
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					// Do nothing
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(directory)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
