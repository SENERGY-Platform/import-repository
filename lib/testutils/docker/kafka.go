package docker

import (
	"context"
	"strconv"
	"sync"

	"github.com/SENERGY-Platform/go-service-base/struct-logger/attributes"
	"github.com/SENERGY-Platform/import-repository/lib/log"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func Kafka(ctx context.Context, wg *sync.WaitGroup, zookeeperUrl string) (kafkaUrl string, err error) {
	kafkaport, err := GetFreePort()
	if err != nil {
		return kafkaUrl, err
	}
	provider, err := testcontainers.NewDockerProvider(testcontainers.DefaultNetwork("bridge"))
	if err != nil {
		return kafkaUrl, err
	}
	hostIp, err := provider.GetGatewayIP(ctx)
	if err != nil {
		return kafkaUrl, err
	}
	kafkaUrl = hostIp + ":" + strconv.Itoa(kafkaport)
	log.Logger.Debug("host ip", "value", hostIp)
	log.Logger.Debug("host port", "value", kafkaport)
	log.Logger.Debug("kafka url", "value", kafkaUrl)
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "bitnamilegacy/kafka:3.4.0-debian-11-r21",
			Tmpfs: map[string]string{},
			WaitingFor: wait.ForAll(
				wait.ForLog("INFO Awaiting socket connections on"),
				wait.ForListeningPort("9092/tcp"),
			),
			ExposedPorts:    []string{strconv.Itoa(kafkaport) + ":9092"},
			AlwaysPullImage: true,
			Env: map[string]string{
				"ALLOW_PLAINTEXT_LISTENER":             "yes",
				"KAFKA_LISTENERS":                      "OUTSIDE://:9092",
				"KAFKA_ADVERTISED_LISTENERS":           "OUTSIDE://" + kafkaUrl,
				"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP": "OUTSIDE:PLAINTEXT",
				"KAFKA_INTER_BROKER_LISTENER_NAME":     "OUTSIDE",
				"KAFKA_ZOOKEEPER_CONNECT":              zookeeperUrl,
			},
		},
		Started: true,
	})
	if err != nil {
		return kafkaUrl, err
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		err := c.Terminate(context.Background())
		if err != nil {
			log.Logger.Debug("remove container kafka failed", attributes.ErrorKey, err)
		} else {
			log.Logger.Debug("remove container kafka")
		}
	}()

	containerPort, err := c.MappedPort(ctx, "9092/tcp")
	if err != nil {
		return kafkaUrl, err
	}
	log.Logger.Debug("kafka test container port", "container_port", containerPort.Port(), "host_port", kafkaport)

	return kafkaUrl, err
}

func Zookeeper(ctx context.Context, wg *sync.WaitGroup) (hostPort string, ipAddress string, err error) {
	log.Logger.Info("start zookeeper")
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "wurstmeister/zookeeper:latest",
			Tmpfs: map[string]string{"/opt/zookeeper-3.4.13/data": "rw"},
			WaitingFor: wait.ForAll(
				wait.ForLog("binding to port"),
				wait.ForListeningPort("2181/tcp"),
			),
			ExposedPorts:    []string{"2181/tcp"},
			AlwaysPullImage: true,
		},
		Started: true,
	})
	if err != nil {
		return "", "", err
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		err := c.Terminate(context.Background())
		if err != nil {
			log.Logger.Debug("remove container zookeeper failed", attributes.ErrorKey, err)
		} else {
			log.Logger.Debug("remove container zookeeper")
		}
	}()

	ipAddress, err = c.ContainerIP(ctx)
	if err != nil {
		return "", "", err
	}
	temp, err := c.MappedPort(ctx, "2181/tcp")
	if err != nil {
		return "", "", err
	}
	hostPort = temp.Port()

	return hostPort, ipAddress, err
}
