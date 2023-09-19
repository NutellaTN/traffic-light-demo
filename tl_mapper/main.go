package main

import (
	"encoding/json"
	"fmt"
	"os"
	"syscall"
	"time"
	"github.com/yosssi/gmq/mqtt"
	"github.com/yosssi/gmq/mqtt/client"
	logger "github.com/d2r2/go-logger"
  	"github.com/stianeikeland/go-rpio"
  	"os/signal"
)

var lg = logger.NewPackageLogger("main",
	logger.DebugLevel,
	// logger.InfoLevel,
)

//DeviceStateUpdate is the structure used in updating the device state
type DeviceStateUpdate struct {
	State string `json:"state,omitempty"`
}

//BaseMessage the base struct of event message
type BaseMessage struct {
	EventID   string `json:"event_id"`
	Timestamp int64  `json:"timestamp"`
}

//TwinValue the struct of twin value
type TwinValue struct {
	Value    *string        `json:"value, omitempty"`
	Metadata *ValueMetadata `json:"metadata,omitempty"`
}

//ValueMetadata the meta of value
type ValueMetadata struct {
	Timestamp int64 `json:"timestamp, omitempty"`
}

//TypeMetadata the meta of value type
type TypeMetadata struct {
	Type string `json:"type,omitempty"`
}

//TwinVersion twin version
type TwinVersion struct {
	CloudVersion int64 `json:"cloud"`
	EdgeVersion  int64 `json:"edge"`
}

//MsgTwin the struct of device twin
type MsgTwin struct {
	Expected        *TwinValue    `json:"expected,omitempty"`
	Actual          *TwinValue    `json:"actual,omitempty"`
	Optional        *bool         `json:"optional,omitempty"`
	Metadata        *TypeMetadata `json:"metadata,omitempty"`
	ExpectedVersion *TwinVersion  `json:"expected_version,omitempty"`
	ActualVersion   *TwinVersion  `json:"actual_version,omitempty"`
}

//DeviceTwinUpdate the struct of device twin update
type DeviceTwinUpdate struct {
	BaseMessage
	Twin map[string]*MsgTwin `json:"twin"`
}

func main() {
	
	defer logger.FinalizeLogger()
	
	if err := rpio.Open(); err != nil {
    		fmt.Println(err)
    		os.Exit(1)
  	}
	
	// Get the pin for each of the lights
  	redPin := rpio.Pin(9)
  	yellowPin := rpio.Pin(10)
  	greenPin := rpio.Pin(11)
	
	redPin.Output()
  	yellowPin.Output()
  	greenPin.Output()
	
	// Clean up on ctrl-c and turn lights out
  	c := make(chan os.Signal, 1)
  	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
  	go func() {
    		<-c
    		redPin.Low()
    		yellowPin.Low()
    		greenPin.Low()
    		lg.Info("exited")
    		os.Exit(0)
  	}()
	defer rpio.Close()
	
	// Turn lights off to start.
	redPin.Low()
  	yellowPin.Low()
  	greenPin.Low()

	// connect to Mqtt broker
	cli := connectToMqtt()

	
	
	for {
		
		// Red
    		redPin.High()
    		publishToMqtt(cli, "red", "ON")
    		time.Sleep(time.Second * 4)

    		// Red and yellow
    		yellowPin.High()
    		publishToMqtt(cli, "yellow", "ON")
    		time.Sleep(time.Second * 2)

    		// Green
    		redPin.Low()
    		yellowPin.Low()
    		greenPin.High()
    		publishToMqtt(cli, "red", "OFF")
    		publishToMqtt(cli, "yellow", "OFF")
    		publishToMqtt(cli, "green", "ON")
    		time.Sleep(time.Second * 6)

    		// Yellow
    		greenPin.Low()
    		yellowPin.High()
    		publishToMqtt(cli, "green", "OFF")
    		publishToMqtt(cli, "yellow", "ON")
    		time.Sleep(time.Second * 3)

    		// Yellow off
    		yellowPin.Low()
    		publishToMqtt(cli, "yellow", "OFF")
		
	}
}

func connectToMqtt() *client.Client {
	cli := client.New(&client.Options{
		// Define the processing of the error handler.
		ErrorHandler: func(err error) {
			fmt.Println(err)
		},
	})
	defer cli.Terminate()

	// Connect to the MQTT Server.
	err := cli.Connect(&client.ConnectOptions{
		Network:  "tcp",
		Address:  "127.0.0.1:1883",
		ClientID: []byte("receive-client"),
	})
	if err != nil {
		panic(err)
	}
	return cli
}

func publishToMqtt(cli *client.Client, property string, propertyValue string) {
	deviceTwinUpdate := "$hw/events/device/" + "traffic-light-instance-01" + "/twin/update"

	updateMessage := createActualUpdateMessage(propertyValue, property)
	twinUpdateBody, _ := json.Marshal(updateMessage)

	cli.Publish(&client.PublishOptions{
		TopicName: []byte(deviceTwinUpdate),
		QoS:       mqtt.QoS0,
		Message:   twinUpdateBody,
	})
}

//createActualUpdateMessage function is used to create the device twin update message
func createActualUpdateMessage(actualValue string, property string) DeviceTwinUpdate {
	var deviceTwinUpdateMessage DeviceTwinUpdate
	actualMap := map[string]*MsgTwin{property: {Actual: &TwinValue{Value: &actualValue}, Metadata: &TypeMetadata{Type: "Updated"}}}
	deviceTwinUpdateMessage.Twin = actualMap
	return deviceTwinUpdateMessage
}
