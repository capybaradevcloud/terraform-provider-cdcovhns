package api

const (
	NSExternal string = "external"
	NSHosted   string = "hosted"
)

type NameServerType struct {
	NameServerType string `json:"nameServerType"`
}

type NameServerUpdateRequest struct {
	NameServers []*NameServerCreatePayload `json:"nameServers"`
}

type NameServerCreatePayload struct {
	Host string `json:"host,omitempty"`
	IP   string `json:"ip,omitempty"`
}

type TaskReposnse struct {
	Status string `json:"status"`
}

type NameServerOvhResponse struct {
	Host     *string `json:"host,omitempty"`
	Id       int     `json:"id"`
	IP       *string `json:"ip,omitempty"`
	IsUsed   bool    `json:"isUsed"`
	ToDelete bool    `json:"toDelete"`
}

func (n *NameServerOvhResponse) GetHost() string {
	if n.Host == nil {
		return ""
	}
	return *n.Host
}

func (n *NameServerOvhResponse) GetIP() string {
	if n.IP == nil {
		return ""
	}
	return *n.IP
}

type NameServerTask struct {
	ID          int64  `json:"id"`
	ServiceName string `json:"domain"`
}
