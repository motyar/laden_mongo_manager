package mongo

import (
	"encoding/json"
	"log"
	"sync"

	. "github.com/ory/ladon"
	"github.com/pkg/errors"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type rdbSchema struct {
	ID          string          `json:"id"`
	Description string          `json:"description"`
	Subjects    []string        `json:"subjects"`
	Effect      string          `json:"effect"`
	Resources   []string        `json:"resources"`
	Actions     []string        `json:"actions"`
	Conditions  json.RawMessage `json:"conditions"`
}

func rdbToPolicy(s *rdbSchema) (*DefaultPolicy, error) {
	if s == nil {
		return nil, nil
	}

	ret := &DefaultPolicy{
		ID:          s.ID,
		Description: s.Description,
		Subjects:    s.Subjects,
		Effect:      s.Effect,
		Resources:   s.Resources,
		Actions:     s.Actions,
		Conditions:  Conditions{},
	}

	return ret, nil

}

func rdbFromPolicy(p Policy) (*rdbSchema, error) {
	cs, err := p.GetConditions().MarshalJSON()
	if err != nil {
		return nil, err
	}
	return &rdbSchema{
		ID:          p.GetID(),
		Description: p.GetDescription(),
		Subjects:    p.GetSubjects(),
		Effect:      p.GetEffect(),
		Resources:   p.GetResources(),
		Actions:     p.GetActions(),
		Conditions:  cs,
	}, err
}

// NewMongoManager initializes a new MongoManager for given session.
func NewMongoManager(collection *mgo.Collection) *MongoManager {
	// ensure we can query fast on the policy id
	// TODO: we might want to unforce this less frequent
	collection.EnsureIndex(mgo.Index{
		Key:        []string{"id"},
		Unique:     true,
		DropDups:   true,
		Background: true,
	})

	return &MongoManager{
		Collection: collection,
	}
}

// MongoManager is a mongodb implementation of Manager to store policies persistently.
type MongoManager struct {
	Collection *mgo.Collection
	sync.RWMutex
	Policies map[string]Policy
}

// GetAll returns all policies.
func (m *MongoManager) GetAll(limit, offset int64) (Policies, error) {
	q := m.Collection.Find(nil).Skip(int(offset)).Limit(int(limit))

	iter := q.Iter()
	if iter.Err() != nil {
		return nil, iter.Err()
	}

	var ps Policies
	m.Policies = map[string]Policy{}
	var s rdbSchema

	for iter.Next(&s) {

		// Convert to Plicies
		p, err := rdbToPolicy(&s)
		if err != nil {
			return nil, err
		}
		ps = append(ps, p)
		m.Policies[s.ID] = p
	}
	return ps, nil
}

// ColdStart loads all policies from mongodb into memory.
func (m *MongoManager) ColdStart() error {

	policies, err := m.GetAll(10, 0)
	if err != nil {
		return errors.WithStack(err)
	}

	m.Lock()
	defer m.Unlock()

	for k, v := range policies {
		log.Printf("key=%v, value=%v", k, v)
	}

	return nil
}

// Create inserts a new policy.
func (m *MongoManager) Create(policy Policy) error {
	if err := m.publishCreate(policy); err != nil {
		return err
	}

	return nil
}

// Update updatess a policy.
func (m *MongoManager) Update(policy Policy) error {
	if err := m.publishUpdate(policy); err != nil {
		return err
	}

	return nil
}

// Get retrieves a policy.
func (m *MongoManager) Get(id string) (Policy, error) {
	m.RLock()
	defer m.RUnlock()

	p, ok := m.Policies[id]
	if !ok {
		return nil, errors.New("Not found")
	}

	return p, nil
}

// Delete removes a policy.
func (m *MongoManager) Delete(id string) error {
	if err := m.publishDelete(id); err != nil {
		return err
	}

	return nil
}

func (m *MongoManager) FindRequestCandidates(r *Request) (Policies, error) {
	m.RLock()
	defer m.RUnlock()
	ps := make(Policies, len(m.Policies))
	var count int
	for _, p := range m.Policies {
		ps[count] = p
		count++
	}
	return ps, nil
}

func (m *MongoManager) publishCreate(policy Policy) error {
	p, err := rdbFromPolicy(policy)
	if err != nil {
		return err
	}
	if err := m.Collection.Insert(p); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (m *MongoManager) publishUpdate(policy Policy) error {
	p, err := rdbFromPolicy(policy)
	if err != nil {
		return err
	}
	if err := m.Collection.Update(bson.M{"id": p.ID}, p); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (m *MongoManager) publishDelete(id string) error {
	if err := m.Collection.Remove(bson.M{"id": id}); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

/*
// Watch is used to watch for changes on mongodb
// and updates manager's policy accordingly.
func (m *MongoManager) Watch(ctx context.Context) {
	go retry(time.Second*15, time.Minute, func() error {

		policies, err := m.Table.Changes().Run(m.Session)
		if err != nil {
			return errors.WithStack(err)
		}

		defer policies.Close()

		var update = make(map[string]*rdbSchema)
		for policies.Next(&update) {
			logrus.Debug("Received update from RethinkDB Cluster in policy manager.")
			newVal, err := rdbToPolicy(update["new_val"])
			if err != nil {
				logrus.Error(err)
				continue
			}

			oldVal, err := rdbToPolicy(update["old_val"])
			if err != nil {
				logrus.Error(err)
				continue
			}

			m.Lock()

			if newVal == nil && oldVal != nil {
				delete(m.Policies, oldVal.GetID())
			} else if newVal != nil && oldVal != nil {
				delete(m.Policies, oldVal.GetID())
				m.Policies[newVal.GetID()] = newVal
			} else {
				m.Policies[newVal.GetID()] = newVal
			}

			m.Unlock()
		}

		if policies.Err() != nil {
			logrus.Error(errors.Wrap(policies.Err(), ""))
		}
		return nil
	})
}

func retry(maxWait time.Duration, failAfter time.Duration, f func() error) (err error) {
	var lastStart time.Time
	err = errors.New("Did not connect.")
	loopWait := time.Millisecond * 500
	retryStart := time.Now()
	for retryStart.Add(failAfter).After(time.Now()) {
		lastStart = time.Now()
		if err = f(); err == nil {
			return nil
		}

		if lastStart.Add(maxWait * 2).Before(time.Now()) {
			retryStart = time.Now()
		}

		logrus.Infof("Retrying in %f seconds...", loopWait.Seconds())
		time.Sleep(loopWait)
		loopWait = loopWait * time.Duration(int64(2))
		if loopWait > maxWait {
			loopWait = maxWait
		}
	}
	return err
}
*/
