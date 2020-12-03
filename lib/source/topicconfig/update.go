/*
 * Copyright 2020 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */


package topicconfig

import (
	"errors"
	"github.com/Shopify/sarama"
	"github.com/wvanbergen/kazoo-go"
	"io/ioutil"
	"log"
)

func EnsureWithZk(zkUrl string, topic string, config map[string]string) (err error) {
	controller, err := getKafkaController(zkUrl)
	if err != nil {
		log.Println("ERROR: unable to find controller", err)
		return err
	}
	if controller == "" {
		log.Println("ERROR: unable to find controller")
		return errors.New("unable to find controller")
	}
	return EnsureWithBroker(controller, topic, config)
}

func EnsureWithBroker(broker string, topic string, config map[string]string) (err error) {
	sconfig := sarama.NewConfig()
	sconfig.Version = sarama.V2_4_0_0
	admin, err := sarama.NewClusterAdmin([]string{broker}, sconfig)
	if err != nil {
		return err
	}

	temp := map[string]*string{}
	for key, value := range config {
		tempValue := value
		temp[key] = &tempValue
	}

	err = set(admin, topic, temp)
	if err != nil {
		log.Println("WARNING: ", err)
		err = create(admin, topic, temp)
	}

	return err
}

func set(admin sarama.ClusterAdmin, topic string, config map[string]*string) (err error) {
	return admin.AlterConfig(sarama.TopicResource, topic, config, false)
}

func create(admin sarama.ClusterAdmin, topic string, config map[string]*string) (err error) {
	return admin.CreateTopic(topic, &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
		ConfigEntries:     config,
	}, false)
}

func getKafkaController(zkUrl string) (controller string, err error) {
	zookeeper := kazoo.NewConfig()
	zookeeper.Logger = log.New(ioutil.Discard, "", 0)
	zk, chroot := kazoo.ParseConnectionString(zkUrl)
	zookeeper.Chroot = chroot
	kz, err := kazoo.NewKazoo(zk, zookeeper)
	if err != nil {
		return controller, err
	}
	controllerId, err := kz.Controller()
	if err != nil {
		return controller, err
	}
	brokers, err := kz.Brokers()
	kz.Close()
	if err != nil {
		return controller, err
	}
	return brokers[controllerId], err
}
