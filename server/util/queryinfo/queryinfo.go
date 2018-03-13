package queryinfo

import "bitbucket.org/0xor1/task/server/util/time"

func New(query string, args interface{}, startUnixMillis int64) *QueryInfo {
	return &QueryInfo{
		Query:    query,
		Args:     args,
		Duration: time.NowUnixMillis() - startUnixMillis,
	}
}

type QueryInfo struct {
	Query    string      `json:"query"`
	Args     interface{} `json:"args"`
	Duration int64       `json:"duration"`
}
