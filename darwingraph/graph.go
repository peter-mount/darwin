package darwingraph

import (
	"encoding/xml"
	"errors"
	"flag"
	"github.com/peter-mount/golib/kernel"
	"log"
	"os"
	"sync"
)

// DarwinGraph is a Kernel service which maintains a TiplocGraph and delegate methods to it
type DarwinGraph struct {
	mutex            sync.Mutex   // Mutex to protect graph
	graph            *TiplocGraph // The graph representing the UK rail network
	importFileName   *string      // -import filename to create an initial model
	stationsFileName *string      // -kbstation filename to import data from the NRE Knowledge Base
	xmlFileName      *string      // -xml filename to load/save the model
}

func (d *DarwinGraph) Name() string {
	return "DarwinGraph"
}

func (d *DarwinGraph) Init(k *kernel.Kernel) error {
	d.importFileName = flag.String("import", "", "Import tiploc data")
	d.xmlFileName = flag.String("xml", "", "xml filename for the graph")
	d.stationsFileName = flag.String("kbstation", "", "xml to import KB data into the graph")
	return nil
}

func (d *DarwinGraph) PostInit() error {
	if *d.xmlFileName == "" {
		return errors.New("-xml is required")
	}
	return nil
}

func (d *DarwinGraph) Start() error {
	d.graph = NewTiplocGraph()

	// Import the model on start
	if *d.importFileName != "" {
		err := d.importFile()
		if err != nil {
			return err
		}
	} else {
		// If not importing the model, load the graph
		err := d.LoadGraph()
		if err != nil {
			return err
		}
	}

	if *d.stationsFileName != "" {
		// Import information from the NRE KB feed
		err := d.importKBStations()
		if err != nil {
			return err
		}
	}
	// Once started save the current graph (if enabled)
	return d.SaveGraph()
}

func (d *DarwinGraph) LoadGraph() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	log.Printf("Restoring graph from %s", *d.xmlFileName)
	f, err := os.Open(*d.xmlFileName)
	if err != nil {
		return err
	}
	defer f.Close()

	e := xml.NewDecoder(f)
	err = e.Decode(d.graph)
	if err != nil {
		return err
	}

	log.Printf("Loaded graph from %s Nodes %d\n",
		*d.xmlFileName,
		len(d.graph.ids),
	)

	return nil
}

func (d *DarwinGraph) SaveGraph() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	log.Printf("Persisting graph to %s", *d.xmlFileName)

	f, err := os.Create(*d.xmlFileName)
	if err != nil {
		return err
	}
	defer f.Close()

	e := xml.NewEncoder(f)
	e.Indent("", "  ")
	err = e.Encode(d.graph)
	if err != nil {
		return err
	}

	log.Printf("Persisted graph to %s", *d.xmlFileName)
	return nil
}

// GetTiplocNode returns an existing TiplocNode or nil if it doesn't exist
func (d *DarwinGraph) GetTiplocNode(tiploc string) *TiplocNode {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.graph.GetNode(tiploc)
}

// ComputeIfAbsent returns an existing TiplocNode or creates one using the supplied function
func (d *DarwinGraph) ComputeIfAbsent(tiploc string, f func() *TiplocNode) *TiplocNode {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.graph.ComputeIfAbsent(tiploc, f)
}
