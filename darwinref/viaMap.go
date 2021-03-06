package darwinref

import (
	"encoding/json"
	bolt "github.com/etcd-io/bbolt"
	"sync"
)

type ViaMap struct {
	mutex sync.Mutex
	via   map[string]map[string][]*Via
}

func (vm *ViaMap) View(f func() error) error {
	vm.mutex.Lock()
	defer vm.mutex.Unlock()
	return f()
}

func (d *DarwinReference) NewViaMap() *ViaMap {
	vm := &ViaMap{}
	vm.via = make(map[string]map[string][]*Via)
	go vm.View(func() error {
		return d.View(func(tx *bolt.Tx) error {
			return tx.Bucket([]byte("DarwinVia")).
				ForEach(func(k, v []byte) error {
					via := &Via{}
					err := json.Unmarshal(v, via)
					if err != nil {
						return err
					}
					vm.addVia(via)
					return nil
				})
		})
	})
	return vm
}

func (vm *ViaMap) addVia(v *Via) {
	m1, ok := vm.via[v.At]
	if !ok {
		m1 = make(map[string][]*Via)
		vm.via[v.At] = m1
	}

	m1[v.Dest] = append(m1[v.Dest], v)
}

func (r *DarwinReference) ResolveVia(at string, dest string, loc []string) *Via {
	var via *Via
	r.viaMap.View(func() error {
		if m1, ok := r.viaMap.via[at]; ok {
			if m2, ok := m1[dest]; ok {
				for _, v := range m2 {
					for i1, l1 := range loc {
						// Run rest of array as 2nd loc against l1
						for _, l2 := range loc[i1+1:] {
							if v.Loc1 == l1 && v.Loc2 == l2 {
								via = v
								return nil
							}
						}
						// Only 1 entry & it matches then use it
						if v.Loc1 == l1 && v.Loc2 == "" {
							via = v
							return nil
						}
					}
				}
			}
		}
		return nil
	})
	return via
}
