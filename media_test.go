package onvif

import (
	"fmt"
	"log"
	"testing"
)

func TestGetProfiles(t *testing.T) {
	log.Println("Test GetProfiles")

	res, err := testDevice.GetProfiles()
	if err != nil {
		t.Error(err)
	}

	js := prettyJSON(&res)
	fmt.Println(js)
}

func TestGetStreamURI(t *testing.T) {
	log.Println("Test GetStreamURI")

	res, err := testDevice.GetStreamURI("profile_1_h264", "UDP")
	if err != nil {
		t.Error(err)
	}

	js := prettyJSON(&res)
	fmt.Println(js)
}
