package libs

// 用户文档结构
type User struct {
	UserID    string `json:"userid"`
	Password  string `json:"password"`
	Latitude  string `json:"latitude,omitempty"`  // 纬度
	Longitude string `json:"longitude,omitempty"` // 经度
	IsVIP     bool   `json:"isvip,omitempty"`     // 是否是VIP用户
	GroupID   int    `json:"groupid,omitempty"`
}
