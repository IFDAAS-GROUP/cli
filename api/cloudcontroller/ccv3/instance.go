package ccv3

import (
	"encoding/json"
	"strconv"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/types"
)

type Instance struct {
	Index       int
	State       string
	Uptime      int
	CPU         float64
	MemoryUsage types.NullByteSize
	MemoryQuota types.NullByteSize
	DiskUsage   types.NullByteSize
	DiskQuota   types.NullByteSize
}

// UnmarshalJSON helps unmarshal a V3 Cloud Controller Instance response.
func (instance *Instance) UnmarshalJSON(data []byte) error {
	var inputInstance struct {
		State string `json:"state"`
		Usage struct {
			CPU  float64            `json:"cpu"`
			Mem  types.NullByteSize `json:"mem"`
			Disk types.NullByteSize `json:"disk"`
		} `json:"usage"`
		MemQuota  types.NullByteSize `json:"mem_quota"`
		DiskQuota types.NullByteSize `json:"disk_quota"`
		Index     int                `json:"index"`
		Uptime    int                `json:"uptime"`
	}
	if err := json.Unmarshal(data, &inputInstance); err != nil {
		return err
	}

	instance.State = inputInstance.State
	instance.CPU = inputInstance.Usage.CPU
	instance.MemoryUsage = inputInstance.Usage.Mem
	instance.MemoryUsage.IsBytes = true

	instance.DiskUsage = inputInstance.Usage.Disk
	instance.DiskUsage.IsBytes = true

	instance.MemoryQuota = inputInstance.MemQuota
	instance.MemoryQuota.IsBytes = true
	instance.DiskQuota = inputInstance.DiskQuota
	instance.DiskQuota.IsBytes = true
	instance.Index = inputInstance.Index
	instance.Uptime = inputInstance.Uptime

	return nil
}

func (client *Client) DeleteApplicationProcessInstance(appGUID string, processType string, instanceIndex int) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteApplicationProcessInstanceRequest,
		URIParams: map[string]string{
			"app_guid": appGUID,
			"type":     processType,
			"index":    strconv.Itoa(instanceIndex),
		},
	})
	if err != nil {
		return nil, err
	}

	var response cloudcontroller.Response
	err = client.connection.Make(request, &response)

	return response.Warnings, err
}

// GetProcessInstances lists instance stats for a given process.
func (client *Client) GetProcessInstances(processGUID string) ([]Instance, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetProcessInstancesRequest,
		URIParams:   map[string]string{"process_guid": processGUID},
	})
	if err != nil {
		return nil, nil, err
	}

	var fullInstancesList []Instance
	warnings, err := client.paginate(request, Instance{}, func(item interface{}) error {
		if instance, ok := item.(Instance); ok {
			fullInstancesList = append(fullInstancesList, instance)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Instance{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullInstancesList, warnings, err
}
