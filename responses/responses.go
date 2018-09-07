package responses

type StatusResponse struct {
	Name       string `json:"name"`
	QueueName  string `json:"queue_name"`
	Status     string `json:"status"`
	StatusTime string `json:"status_time"`
}

type RereadResponse struct {
	Changes []string `json:"changes"`
	Removes []string `json:"removes"`
}
