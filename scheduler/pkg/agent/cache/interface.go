package cache

type CacheManager interface {
	// evict the least priority node and return key, this should start a transaction (lock on id)
	StartEvict() (string, int64, error)
	// end the transaction, if false is provided the id should be put back in the queue
	EndEvict(string, int64, bool) error
	// add a new node with specific id and priority/value
	Add(id string, value int64) error
	// add a new node with specific id and default priority/value
	AddDefault(id string) error
	// update value for given id, which would reflect in order
	Update(id string, value int64) error
	// default bump value for given id, which would reflect in order
	UpdateDefault(id string) error
	// check if value exists
	Exists(id string) bool
	// get value/priority of given id
	Get(id string) (int64, error)
	// delete item with id from cache
	Delete(id string) error
	// get a list of all keys / values
	GetItems() ([]string, []int64)
}
