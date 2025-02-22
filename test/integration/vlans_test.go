package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/linode/linodego"
)

func TestVLANs_List(t *testing.T) {
	instancePrefix := "linodego-testing-" + randString(12, lowerBytes, upperBytes)
	vlanName := "linodego-really-cool-vlan-" + randString(12, lowerBytes, upperBytes)

	client, fixturesTeardown := createTestClient(t, "fixtures/TestVLANs_List")
	defer fixturesTeardown()

	var instances []*linodego.Instance
	for i := 0; i < 2; i++ {
		instance, instanceTeardown, err := createVLANInstance(t, client, fmt.Sprintf("%s-%d", instancePrefix, i), vlanName)
		if err != nil {
			t.Fatal(err)
		}
		defer instanceTeardown()

		instances = append(instances, instance)
	}

	for _, instance := range instances {
		if _, err := client.WaitForInstanceStatus(context.Background(), instance.ID, linodego.InstanceRunning, 240); err != nil {
			t.Error(err)
		}
	}

	vlans, err := client.ListVLANs(context.TODO(), &linodego.ListOptions{
		Filter: fmt.Sprintf("{\"label\": \"%s\"}", vlanName),
	})
	if err != nil {
		t.Error(err)
	}

	if len(vlans) < 1 {
		t.Error("expected list of vlans but got none")
	}

	for _, instance := range instances {
		if !vlanHasLinodeID(&vlans[0], instance.ID) {
			t.Errorf("instance %d not found in vlan", instance.ID)
		}
	}
}

func createVLANInstance(t *testing.T, client *linodego.Client, instanceName, vlanName string) (*linodego.Instance, func(), error) {
	t.Helper()

	trueBool := true

	instance, err := createInstance(t, client, func(opts *linodego.InstanceCreateOptions) {
		opts.Interfaces = []linodego.InstanceConfigInterface{
			{
				Label:   vlanName,
				Purpose: linodego.InterfacePurposeVLAN,
			},
		}

		opts.Booted = &trueBool
		opts.Label = instanceName
		opts.Region = "us-southeast"
	})
	if err != nil {
		return nil, nil, err
	}

	teardown := func() {
		if terr := client.DeleteInstance(context.Background(), instance.ID); terr != nil {
			t.Errorf("Error deleting test Instance: %s", terr)
		}
	}

	return instance, teardown, err
}

func vlanHasLinodeID(vlan *linodego.VLAN, linodeID int) bool {
	for _, id := range vlan.Linodes {
		if id == linodeID {
			return true
		}
	}

	return false
}
