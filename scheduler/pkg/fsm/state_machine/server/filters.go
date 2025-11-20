package server

type Filter string

const (
	FilterReplicas      Filter = "filter_replicas"
	FilterDeletedServer Filter = "filter_deleted_server"
)

func (ss *Snapshot) Filter(filter Filter) bool {
	switch filter {
	case FilterReplicas:
		return len(ss.Replicas) > 0
	case FilterDeletedServer:
		return ss.ExpectedReplicas != 0
	}

	return false
}
