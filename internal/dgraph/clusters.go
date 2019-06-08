package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// Clusters represents the set of clusters
type Clusters struct {
	UID     string    `json:"uid,omitempty"`
	Size    int       `json:"size"`
	Parents []Parent  `json:"parents,omitempty"`
	Ranks   []Rank    `json:"ranks,omitempty"`
	Set     []Cluster `json:"set,omitempty"`
}

// Cluster set of addresses related to the same entity
type Cluster struct {
	UID       string    `json:"uid,omitempty"`
	Addresses []Address `json:"addresses,omitempty"`
	Cluster   int       `json:"cluster"`
}

// Address node
type Address struct {
	UID     string `json:"uid,omitempty"`
	Address string `json:"address"`
}

// Parent persist the parent tag and its position
type Parent struct {
	UID    string `json:"uid,omitempty"`
	Pos    int    `json:"pos"`
	Parent int    `json:"parent"`
}

// Rank persist rank and its position
type Rank struct {
	UID  string `json:"uid,omitempty"`
	Pos  int    `json:"pos"`
	Rank int    `json:"rank"`
}

// ClustersResp basic structure to unmarshall cluster query
type ClustersResp struct {
	C []struct{ Clusters }
}

// GetClusters returnes the set of all clusters stored in dgraph
func GetClusters() (Clusters, error) {
	resp, err := instance.NewTxn().Query(context.Background(), `{
		c(func: has(set)) {
			uid
			size
			parents {
				uid
				pos
				parent
			}
			ranks {
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
	return r.C[0].Clusters, nil
}

// GetClusterUID returns the UID of the cluster
func GetClusterUID() (string, error) {
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
	return r.C[0].Clusters.UID, nil
}

// GetSetUID returnes the UID of the specified set of addresses
func GetSetUID(set int) (string, error) {
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
func UpdateSet(address string, cluster int) error {
	uid, err := GetClusterUID()
	if err != nil {
		return err
	}
	clusterUID, err := GetSetUID(cluster)
	if err != nil {
		return err
	}
	c := []Cluster{
		{
			UID: clusterUID,
			Addresses: []Address{
				{Address: address},
			},
		},
	}
	set := Clusters{
		UID: uid,
		Set: c,
	}
	if err := Store(set); err != nil {
		return err
	}
	return nil
}

// NewSet creates a new set in the cluster. A set is composed by
// at least an addres, that is why address is passed as argument
func NewSet(address string, cluster int) error {
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
	if err := Store(set); err != nil {
		return err
	}
	return nil
}

// UpdateSize updates the size of the cluster
func UpdateSize(size int) error {
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
func GetParent(pos int) (Parent, error) {
	resp, err := instance.NewTxn().Query(context.Background(), fmt.Sprintf(`
		c(func: has(set)) {
    	uid
    	parents @filter(eq(pos, %d)) {
    	  uid
    	  pos
    	  parent
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
func AddParent(pos, parent int) error {
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
	if err := Store(set); err != nil {
		return err
	}
	return nil
}

// UpdateParent updates the parent tag in parent node based on passed position
func UpdateParent(pos, parent int) error {
	cuid, err := GetClusterUID()
	if err != nil {
		return err
	}
	p, err := GetParent(pos)
	if err != nil {
		return err
	}
	c := Clusters{
		UID: cuid,
		Parents: []Parent{
			{
				UID:    p.UID,
				Parent: parent,
			},
		},
	}
	if err := Store(c); err != nil {
		return err
	}
	return nil
}

// GetRank returns the parent struct at the required position
func GetRank(pos int) (Rank, error) {
	resp, err := instance.NewTxn().Query(context.Background(), fmt.Sprintf(`
		c(func: has(set)) {
    	uid
    	ranks @filter(eq(pos, %d)) {
    	  uid
    	  pos
    	  rank
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
	if len(r.C[0].Parents) == 0 {
		return Rank{}, errors.New("Parent not found")
	}
	if len(r.C[0].Parents) > 1 {
		return Rank{}, errors.New("More than a parent found, something is wrong")
	}
	return r.C[0].Ranks[0], nil
}

// AddRank appends a rank to the cluster
func AddRank(pos, rank int) error {
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
	if err := Store(set); err != nil {
		return err
	}
	return nil
}

// UpdateRank updates the parent tag in parent node based on passed position
func UpdateRank(pos, rank int) error {
	cuid, err := GetClusterUID()
	if err != nil {
		return err
	}
	r, err := GetRank(pos)
	if err != nil {
		return err
	}
	c := Clusters{
		UID: cuid,
		Ranks: []Rank{
			{
				UID:  r.UID,
				Rank: rank,
			},
		},
	}
	if err := Store(c); err != nil {
		return err
	}
	return nil
}
