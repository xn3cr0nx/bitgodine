package dgraph

import (
	"context"
	"errors"
	"fmt"

	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

// Clusters represents the set of clusters
type Clusters struct {
	UID     string    `json:"uid,omitempty"`
	Size    uint32    `json:"size,omitempty"`
	Height  int32     `json:"height,omitempty"`
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
func (d *Dgraph) NewClusters() error {
	c := Clusters{
		UID:    "_:cluster",
		Size:   0,
		Height: 0,
		Set: []Cluster{
			{
				UID: "_:init",
			},
		},
	}
	fmt.Println("INSERTING CLUSTER", c)
	err := d.Store(c)
	if err != nil {
		return err
	}
	// uids := resp.GetUids()
	// fmt.Println("CACHED CLUSTER UID AFTER CREATION", uids["cluster"])
	// if err == nil {
	// 	if !d.cache.Set("clusterUID", uids["cluster"], 1) {
	// 		logger.Error("Cache", err, logger.Params{})
	// 	}
	// }

	return nil
}

// GetClusters returnes the set of all clusters stored in dgraph
func (d *Dgraph) GetClusters() (Clusters, error) {
	uid, err := d.GetClusterUID()
	if err != nil {
		return Clusters{}, err
	}

	resp, err := instance.NewReadOnlyTxn().QueryWithVars(context.Background(), `
		query params($u: string) {
			c(func: uid($u)) {
				uid
				size
				height
				parents (orderasc: pos) (first: 1000000000) {
					uid
					pos
					parent
				}
				ranks (orderasc: pos) (first: 1000000000) {
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
		}`, map[string]string{"$u": uid})
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
func (d *Dgraph) GetClusterUID() (string, error) {
	if cached, ok := d.cache.Get("clusterUID"); ok {
		return cached.(string), nil
	}

	resp, err := instance.NewReadOnlyTxn().Query(context.Background(), `{
		var(func: has(set)) {
			H as height
		}
		var() {
			h as max(val(H))
		}
		c(func: has(set)) @filter(eq(height, val(h))) {
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
		if !d.cache.Set("clusterUID", r.C[0].Clusters.UID, 1) {
			logger.Error("Cache", err, logger.Params{})
		}
	}
	return r.C[0].Clusters.UID, nil
}

// GetClusterHeight returns the UID of the cluster
func (d *Dgraph) GetClusterHeight() (int32, error) {
	uid, err := d.GetClusterUID()
	if err != nil {
		return 0, err
	}

	resp, err := instance.NewReadOnlyTxn().QueryWithVars(context.Background(), `
		query params($u: string) {
			c(func: uid($u)) {
				height
			}
		}`, map[string]string{"$u": uid})
	if err != nil {
		return 0, err
	}
	var r ClustersResp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return 0, err
	}
	if len(r.C) == 0 {
		return 0, errors.New("Cluster not found")
	}

	return r.C[0].Clusters.Height, nil
}

// GetSetUID returnes the UID of the specified set of addresses
func (d *Dgraph) GetSetUID(set uint32) (string, error) {
	uid, err := d.GetClusterUID()
	if err != nil {
		return "", err
	}

	resp, err := instance.NewReadOnlyTxn().QueryWithVars(context.Background(), `
		query params($u: string, $d: int) {
			c(func: uid($u)) {
				uid
				set @filter(eq(cluster, $d)) {
					uid
					cluster
					addresses {
						uid
						address
					}
				}
			}
		}`, map[string]string{"$u": uid, "$d": fmt.Sprintf("%d", set)})
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
func (d *Dgraph) UpdateSet(address string, cluster uint32) error {
	setUID, err := d.GetSetUID(cluster)
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
	if err := d.Store(c); err != nil {
		return err
	}
	return nil
}

// NewSet creates a new set in the cluster. A set is composed by
// at least an addres, that is why address is passed as argument
func (d *Dgraph) NewSet(address string, cluster uint32) error {
	uid, err := d.GetClusterUID()
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
	if err := d.Store(set); err != nil {
		return err
	}
	return nil
}

// UpdateSize updates the size of the cluster
func (d *Dgraph) UpdateSize(size uint32) error {
	uid, err := d.GetClusterUID()
	if err != nil {
		return err
	}
	set := Clusters{
		UID:  uid,
		Size: size,
	}
	if err := d.Store(set); err != nil {
		return err
	}
	return nil
}

// UpdateClusterHeight updates the size of the cluster
func (d *Dgraph) UpdateClusterHeight(height int32) error {
	if height == 0 {
		return nil
	}
	uid, err := d.GetClusterUID()
	if err != nil {
		return err
	}
	set := Clusters{
		UID:    uid,
		Height: height,
	}
	if err := d.Store(set); err != nil {
		return err
	}
	return nil
}

// GetParent returns the parent struct at the required position
func (d *Dgraph) GetParent(pos uint32) (Parent, error) {
	uid, err := d.GetClusterUID()
	if err != nil {
		return Parent{}, err
	}

	resp, err := instance.NewReadOnlyTxn().QueryWithVars(context.Background(), `
		query params($u: string, $d: int) {
			c(func: uid($u)) {
				uid
				parents @filter(eq(pos, $d)) {
					uid
					pos
					parent
				}
			}
		}`, map[string]string{"$u": uid, "$d": fmt.Sprintf("%d", pos)})

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
func (d *Dgraph) AddParent(pos, parent uint32) error {
	uid, err := d.GetClusterUID()
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
	if err := d.Store(set); err != nil {
		return err
	}
	return nil
}

// UpdateParent updates the parent tag in parent node based on passed position
func (d *Dgraph) UpdateParent(pos, parent uint32) error {
	p, err := d.GetParent(pos)
	if err != nil {
		return err
	}
	pnt := Parent{
		UID:    p.UID,
		Pos:    pos,
		Parent: parent,
	}
	if err := d.Store(pnt); err != nil {
		return err
	}
	return nil
}

// GetRank returns the parent struct at the required position
func (d *Dgraph) GetRank(pos uint32) (Rank, error) {
	uid, err := d.GetClusterUID()
	if err != nil {
		return Rank{}, err
	}

	resp, err := instance.NewReadOnlyTxn().QueryWithVars(context.Background(), `
		query params($u: string, $d: int) {
			c(func: uid($u)) {
				uid
				ranks @filter(eq(pos, $d)) {
					uid
					pos
					rank
				}
			}
		}`, map[string]string{"$u": uid, "$d": fmt.Sprintf("%d", pos)})
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
func (d *Dgraph) AddRank(pos, rank uint32) error {
	uid, err := d.GetClusterUID()
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
	if err := d.Store(set); err != nil {
		return err
	}
	return nil
}

// UpdateRank updates the parent tag in parent node based on passed position
func (d *Dgraph) UpdateRank(pos, rank uint32) error {
	r, err := d.GetRank(pos)
	if err != nil {
		return err
	}
	rnk := Rank{
		UID:  r.UID,
		Pos:  pos,
		Rank: rank,
	}
	if err := d.Store(rnk); err != nil {
		return err
	}
	return nil
}

// BulkMakeSet take together single update operations and update clusters with a single request
func (d *Dgraph) BulkMakeSet(address string, size, parentPos, rankPos uint32) error {
	uid, err := d.GetClusterUID()
	if err != nil {
		return err
	}
	set := Clusters{
		UID:  uid,
		Size: size + 1,
		Set: []Cluster{
			{
				Cluster: size,
				Addresses: []Address{
					{Address: address},
				},
			},
		},
		Parents: []Parent{
			{
				Pos:    parentPos,
				Parent: size,
			},
		},
		Ranks: []Rank{
			{
				Pos:  rankPos,
				Rank: 0,
			},
		},
	}
	if err := d.Store(set); err != nil {
		return err
	}
	return nil
}
