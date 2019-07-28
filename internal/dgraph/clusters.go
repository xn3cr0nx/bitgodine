package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/allegro/bigcache"
	"github.com/xn3cr0nx/bitgodine_code/internal/cache"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// Clusters represents the set of clusters
type Clusters struct {
	UID     string    `json:"uid,omitempty"`
	Size    uint32    `json:"size,omitempty"`
	Parents []Parent  `json:"parents,omitempty"`
	Ranks   []Rank    `json:"ranks,omitempty"`
	Set     []Cluster `json:"set,omitempty"`
}

// Cluster set of addresses related to the same entity
type Cluster struct {
	UID       string    `json:"uid,omitempty"`
	Addresses []Address `json:"addresses,omitempty"`
	Cluster   uint32    `json:"cluster,omitempty"`
}

// Address node
type Address struct {
	UID     string `json:"uid,omitempty"`
	Address string `json:"address"`
}

// Parent persist the parent tag and its position
type Parent struct {
	UID    string `json:"uid,omitempty"`
	Pos    uint32 `json:"pos"`
	Parent uint32 `json:"parent"`
}

// Rank persist rank and its position
type Rank struct {
	UID  string `json:"uid,omitempty"`
	Pos  uint32 `json:"pos"`
	Rank uint32 `json:"rank"`
}

// ClustersResp basic structure to unmarshall cluster query
type ClustersResp struct {
	C []struct{ Clusters }
}

// NewClusters stores the basic struct to manage the cluster sets
func NewClusters() error {
	c := Clusters{
		Size: 0,
		Set: []Cluster{
			{
				UID: "_:init",
			},
		},
	}
	if err := Store(c); err != nil {
		return err
	}
	return nil
}

// GetClusters returnes the set of all clusters stored in dgraph
func GetClusters() (Clusters, error) {
	resp, err := instance.NewTxn().Query(context.Background(), `{
		c(func: has(set)) {
			uid
			size
			parents (orderasc: pos) (first: 100000) {
				uid
				pos
				parent
			}
			ranks (orderasc: pos) (first: 100000) {
				uid
				pos
				rank
			}
			set {
				uid
				cluster
				addresses {
					uid
					address
				}
			}
		}
	}`)
	if err != nil {
		return Clusters{}, err
	}
	var r ClustersResp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return Clusters{}, err
	}
	if len(r.C) == 0 {
		return Clusters{}, errors.New("Cluster not found")
	}

	logger.Debug("Dgraph Cluster", "Retrieving Clusters", logger.Params{"size_parents": len(r.C[0].Clusters.Parents), "size_ranks": len(r.C[0].Clusters.Ranks)})
	return r.C[0].Clusters, nil
}

// GetClusterUID returns the UID of the cluster
func GetClusterUID() (string, error) {
	c, err := cache.Instance(bigcache.Config{})
	if err != nil {
		return "", err
	}
	cached, err := c.Get("clusterUID")
	if len(cached) != 0 {
		return string(cached), nil
	}

	resp, err := instance.NewTxn().Query(context.Background(), `{
		c(func: has(set)) {
			uid
		}
	}`)
	if err != nil {
		return "", err
	}
	var r ClustersResp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return "", err
	}
	if len(r.C) == 0 {
		return "", errors.New("Cluster not found")
	}

	if err == nil {
		if err := c.Set("clusterUID", []byte(r.C[0].Clusters.UID)); err != nil {
			logger.Error("Cache", err, logger.Params{})
		}
	}
	return r.C[0].Clusters.UID, nil
}

// GetSetUID returnes the UID of the specified set of addresses
func GetSetUID(set uint32) (string, error) {
	resp, err := instance.NewTxn().Query(context.Background(), fmt.Sprintf(`{
		c(func: has(set)) {
			uid
			set @filter(eq(cluster, %d)) {
				uid
				cluster
				addresses {
					uid
					address
				}
			}
		}
	}`, set))
	if err != nil {
		return "", err
	}
	var r ClustersResp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return "", err
	}
	if len(r.C) == 0 {
		return "", errors.New("Cluster not found")
	}
	if len(r.C[0].Clusters.Set) == 0 {
		return "", errors.New("Set not found")
	}
	return r.C[0].Clusters.Set[0].UID, nil
}

// UpdateSet adds an address to a cluster
func UpdateSet(address string, cluster uint32) error {
	setUID, err := GetSetUID(cluster)
	if err != nil {
		return err
	}
	c := Cluster{
		UID:     setUID,
		Cluster: cluster,
		Addresses: []Address{
			{Address: address},
		},
	}
	logger.Debug("Dgraph Clusters", "Updating set", logger.Params{"set": c})
	if err := Store(c); err != nil {
		return err
	}
	return nil
}

// NewSet creates a new set in the cluster. A set is composed by
// at least an addres, that is why address is passed as argument
func NewSet(address string, cluster uint32) error {
	uid, err := GetClusterUID()
	if err != nil {
		return err
	}
	c := []Cluster{
		{
			Cluster: cluster,
			Addresses: []Address{
				{Address: address},
			},
		},
	}
	set := Clusters{
		UID: uid,
		Set: c,
	}
	logger.Debug("Dgraph Clusters", "Creating New Set", logger.Params{"set": set})
	if err := Store(set); err != nil {
		return err
	}
	return nil
}

// UpdateSize updates the size of the cluster
func UpdateSize(size uint32) error {
	uid, err := GetClusterUID()
	if err != nil {
		return err
	}
	set := Clusters{
		UID:  uid,
		Size: size,
	}
	if err := Store(set); err != nil {
		return err
	}
	return nil
}

// GetParent returns the parent struct at the required position
func GetParent(pos uint32) (Parent, error) {
	resp, err := instance.NewTxn().Query(context.Background(), fmt.Sprintf(`{
		c(func: has(set)) {
    	uid
    	parents @filter(eq(pos, %d)) {
    	  uid
    	  pos
    	  parent
			}
		}
  }`, pos))
	if err != nil {
		return Parent{}, err
	}
	var r ClustersResp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return Parent{}, err
	}
	if len(r.C) == 0 {
		return Parent{}, errors.New("Cluster not found")
	}
	if len(r.C[0].Parents) == 0 {
		return Parent{}, errors.New("Parent not found")
	}
	if len(r.C[0].Parents) > 1 {
		return Parent{}, errors.New("More than a parent found, something is wrong")
	}
	return r.C[0].Parents[0], nil
}

// AddParent appends a rank to the cluster
func AddParent(pos, parent uint32) error {
	uid, err := GetClusterUID()
	if err != nil {
		return err
	}
	p := []Parent{
		{
			Pos:    pos,
			Parent: parent,
		},
	}
	set := Clusters{
		UID:     uid,
		Parents: p,
	}
	logger.Debug("Dgraph Clusters", "Adding parent", logger.Params{"set": set, "pos": pos, "parent": parent, "p": p})
	if err := Store(set); err != nil {
		return err
	}
	return nil
}

// UpdateParent updates the parent tag in parent node based on passed position
func UpdateParent(pos, parent uint32) error {
	p, err := GetParent(pos)
	if err != nil {
		return err
	}
	logger.Debug("Dgraph Clusters", "Updating parent", logger.Params{"pos": pos, "parent": parent, "uid": p.UID, "prev_parent": p.Parent, "prev_pos": p.Pos})
	pnt := Parent{
		UID:    p.UID,
		Pos:    pos,
		Parent: parent,
	}
	if err := Store(pnt); err != nil {
		return err
	}
	return nil
}

// GetRank returns the parent struct at the required position
func GetRank(pos uint32) (Rank, error) {
	resp, err := instance.NewTxn().Query(context.Background(), fmt.Sprintf(`{
		c(func: has(set)) {
    	uid
    	ranks @filter(eq(pos, %d)) {
    	  uid
    	  pos
    	  rank
			}
		}
  }`, pos))
	if err != nil {
		return Rank{}, err
	}
	var r ClustersResp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return Rank{}, err
	}
	if len(r.C) == 0 {
		return Rank{}, errors.New("Cluster not found")
	}
	if len(r.C[0].Ranks) == 0 {
		return Rank{}, errors.New("Rank not found")
	}
	if len(r.C[0].Ranks) > 1 {
		return Rank{}, errors.New("More than a parent found, something is wrong")
	}
	return r.C[0].Ranks[0], nil
}

// AddRank appends a rank to the cluster
func AddRank(pos, rank uint32) error {
	uid, err := GetClusterUID()
	if err != nil {
		return err
	}
	r := []Rank{
		{
			Pos:  pos,
			Rank: rank,
		},
	}
	set := Clusters{
		UID:   uid,
		Ranks: r,
	}
	logger.Debug("Dgraph Clusters", "Adding Rank", logger.Params{"set": set, "pos": pos, "rank": rank, "r": r})
	if err := Store(set); err != nil {
		return err
	}
	return nil
}

// UpdateRank updates the parent tag in parent node based on passed position
func UpdateRank(pos, rank uint32) error {
	r, err := GetRank(pos)
	if err != nil {
		return err
	}
	logger.Debug("Dgraph Clusters", "Updating Rank", logger.Params{"pos": pos, "rank": rank, "uid": r.UID, "prev_rank": r.Rank, "prev_pos": r.Pos})
	rnk := Rank{
		UID:  r.UID,
		Pos:  pos,
		Rank: rank,
	}
	if err := Store(rnk); err != nil {
		return err
	}
	return nil
}
