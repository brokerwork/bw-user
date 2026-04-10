package model

type DashboardConfig struct {
	TenantId string   `json:"tenantId" bson:"tenantId"`
	UserId   string   `json:"userId" bson:"userId"`
	Type     []string `json:"type" bson:"type"`
	Time     int64    `json:"time" bson:"time"`
}
