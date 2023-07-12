package api

import (
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const CHECK_STATUS_WAIT_TIME time.Duration = 15 * time.Second

func (c APIClient) DeleteNameServers(serviceName string) error {
	return c.SetNameServerType(serviceName, NSHosted)
}

func (c APIClient) UpdateNameServers(serviceName string, data *NameServerUpdateRequest) (NameServerTask, error) {
	endpoint := fmt.Sprintf("/domain/%s/nameServers/update", serviceName)
	response := NameServerTask{}

	tflog.Debug(c.ctx, fmt.Sprintf("[CDC_OVH] UpdateNameServers ENDPOINT: %s", endpoint))
	err := c.Client.Post(
		endpoint,
		data,
		&response,
	)
	tflog.Debug(c.ctx, fmt.Sprintf("[CDC_OVH] UpdateNameServers RESP: %v", response))
	tflog.Debug(c.ctx, fmt.Sprintf("[CDC_OVH] UpdateNameServers ERR: %v", err))

	return response, err
}

func (c APIClient) GetNameServersType(serviceName string) (NameServerType, error) {
	endpoint := fmt.Sprintf("/domain/%s", serviceName)
	response := NameServerType{}

	tflog.Debug(c.ctx, fmt.Sprintf("[CDC_OVH] GetNameServersType ENDPOINT: %s", endpoint))
	err := c.Client.Get(
		endpoint,
		&response,
	)
	tflog.Debug(c.ctx, fmt.Sprintf("[CDC_OVH] GetNameServersType RESP: %v", response))
	tflog.Debug(c.ctx, fmt.Sprintf("[CDC_OVH] GetNameServersType ERR: %v", err))

	return response, err
}

func (c APIClient) SetNameServerType(serviceName string, nsType string) error {
	var nsTypeObject NameServerType

	if nsType == NSExternal {
		nsTypeObject.NameServerType = NSExternal
	} else if nsType == NSHosted {
		nsTypeObject.NameServerType = NSHosted
	} else {
		return fmt.Errorf("wrong name server type. Use hosted or external")
	}

	endpoint := fmt.Sprintf("/domain/%s", serviceName)
	tflog.Debug(c.ctx, fmt.Sprintf("[CDC_OVH] SetNameServerType ENDPOINT: %s", endpoint))
	err := c.Client.Put(
		endpoint,
		nsTypeObject,
		nil,
	)
	tflog.Debug(c.ctx, fmt.Sprintf("[CDC_OVH] SetNameServerType ERR: %v", err))
	return err
}

func (c APIClient) CheckOVHTask(err chan<- error, domain string, id int64) {
	endpoint := fmt.Sprintf("/domain/%s/task/%d", domain, id)

	tflog.Debug(c.ctx, fmt.Sprintf("[CDC_OVH] CheckOVHTask ENDPOINT: %s", endpoint))

	apiErr := error(nil)
	for {
		time.Sleep(CHECK_STATUS_WAIT_TIME)

		response := TaskReposnse{}
		err := c.Client.Get(
			endpoint,
			&response,
		)

		tflog.Debug(c.ctx, "[CDC_OVH] CheckOVHTask after get call")

		if err != nil {
			tflog.Debug(c.ctx, fmt.Sprintf("[CDC_OVH]CheckTask API CALL ERROR: %v", err))
			apiErr = err
			break
		}

		if response.Status == "doing" || response.Status == "todo" {
			tflog.Debug(c.ctx, fmt.Sprintf("[CDC_OVH] CheckOVHTask status is %s, waiting", response.Status))
			continue
		}

		if response.Status == "done" {
			tflog.Debug(c.ctx, fmt.Sprintf("[CDC_OVH] CheckOVHTask status is %s, breaking loop", response.Status))
			break
		}

		if response.Status == "error" || response.Status == "cancelled" {
			tflog.Debug(c.ctx, fmt.Sprintf("[CDC_OVH] CheckOVHTask status is %s, error occurred, breaking", response.Status))
			apiErr = fmt.Errorf("task status %s. check OVH Panel", response.Status)
			break
		}

		tflog.Debug(c.ctx, fmt.Sprintf("[CDC_OVH] CheckOVHTask wrong task status: %s, breaking loop to prevent next api calls", response.Status))

		apiErr = fmt.Errorf("unhandled task status: %s, report this to provider developer", response.Status)
		break

	}
	err <- apiErr
}

func (c APIClient) GetNameServersFromAPI(serviceName string) (map[string]NameServerOvhResponse, error) {
	var ids []uint64
	nameServers := make(map[string]NameServerOvhResponse)

	endpoint := fmt.Sprintf("/domain/%s/nameServer", serviceName)
	tflog.Debug(c.ctx, fmt.Sprintf("[CDC_OVH] GetNameServersFromAPI ENDPOINT: %v", endpoint))
	err := c.Client.Get(
		endpoint,
		&ids,
	)

	tflog.Debug(c.ctx, fmt.Sprintf("[CDC_OVH] GetNameServersFromAPI RESP: %v", ids))

	if err != nil {
		tflog.Debug(c.ctx, fmt.Sprintf("[CDC_OVH] GetNameServersFromAPI ERR: %v", err))
		return nil, err
	}

	for key, id := range ids {
		// Get NS data
		nsResponse := NameServerOvhResponse{}
		nsDataEndpoint := fmt.Sprintf("/domain/%s/nameServer/%v", serviceName, id)
		tflog.Debug(c.ctx, fmt.Sprintf("[CDC_OVH] GetNameServersFromAPI loop ENDPOINT: %v", nsDataEndpoint))

		err := c.Client.Get(
			nsDataEndpoint,
			&nsResponse,
		)

		tflog.Debug(c.ctx, fmt.Sprintf("[CDC_OVH] GetNameServersFromAPI RESP: %v", nsResponse))

		if err != nil {
			tflog.Debug(c.ctx, fmt.Sprintf("[CDC_OVH] GetNameServersFromAPI ERR: %v", err))
			return nil, err
		}

		nameServers["ns"+strconv.Itoa(key+1)] = nsResponse
	}

	return nameServers, nil
}

func (c APIClient) CheckCurrentTaskState(serviceName string) error {
	var ids []uint64

	checkDoingTasksEndpoint := fmt.Sprintf("/domain/%s/task?status=%s", serviceName, "doing")
	tflog.Debug(c.ctx, fmt.Sprintf("[CDC_OVH] CheckCurrentTaskState doing ENDPOINT: %v", checkDoingTasksEndpoint))

	doingErr := c.Client.Get(
		checkDoingTasksEndpoint,
		&ids,
	)

	if doingErr != nil {
		tflog.Debug(c.ctx, fmt.Sprintf("[CDC] CheckCurrentTaskState doing ERR: %v", doingErr))
		return doingErr
	}
	tflog.Debug(c.ctx, fmt.Sprintf("[CDC_OVH] CheckCurrentTaskState doing RESP: %v", ids))

	if len(ids) > 0 {
		return fmt.Errorf("some tasks are already in doing state for domain %s", serviceName)
	}

	checkTodoTasksEndpoint := fmt.Sprintf("/domain/%s/task?status=%s", serviceName, "todo")
	tflog.Debug(c.ctx, fmt.Sprintf("[CDC_OVH] CheckCurrentTaskState todo ENDPOINT: %v", checkTodoTasksEndpoint))
	todoErr := c.Client.Get(
		checkTodoTasksEndpoint,
		&ids,
	)
	if todoErr != nil {
		tflog.Debug(c.ctx, fmt.Sprintf("[CDC] CheckCurrentTaskState todo ERR: %v", todoErr))
		return todoErr
	}
	tflog.Debug(c.ctx, fmt.Sprintf("[CDC_OVH] CheckCurrentTaskState todo RESP: %v", ids))

	if len(ids) > 0 {
		return fmt.Errorf("some tasks are already in todo state for domain %s", serviceName)
	}

	return nil
}
