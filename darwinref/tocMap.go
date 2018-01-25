package darwinref

import (
  bolt "github.com/coreos/bbolt"
  "bytes"
  "encoding/json"
//  "encoding/xml"
  "github.com/peter-mount/golib/rest"
  "sort"
  "strings"
)

type TocMap struct {
  m map[string]*Toc
}

func NewTocMap() *TocMap {
  return &TocMap{ m: make( map[string]*Toc ) }
}

// AddTiploc adds a Tiploc to the response
func (r *TocMap) Add( t *Toc ) {
  if _, ok := r.m[ t.Toc ]; !ok {
    r.m[ t.Toc ] = t
  }
}

// AddTiplocs adds an array of Tiploc's to the response
func (r *TocMap) AddAll( t []*Toc ) {
  for _, e := range t {
    r.Add( e )
  }
}

func (r *TocMap) AddToc( dr *DarwinReference, tx *bolt.Tx, t string ) {
  if _, ok := r.m[ t ]; !ok {
    if loc, exists := dr.GetToc( tx, t ); exists {
      r.m[ t ] = loc
    }
  }
}

func (r *TocMap) AddTocs( dr *DarwinReference, tx *bolt.Tx, ts []string ) {
  bucket := tx.Bucket( []byte("DarwinToc") )
  for _, t := range ts {
    if _, ok := r.m[ t ]; !ok {
      if loc, exists := dr.GetTocBucket( bucket, t ); exists {
        r.m[ t ] = loc
      }
    }
  }
}

func (r *TocMap) Get( n string ) ( *Toc, bool ) {
  t, e := r.m[ n ]
  return t, e
}

// Self sets the Self field to match this request
func (r *TocMap) Self( rs *rest.Rest ) {
  for _, v := range r.m {
    v.Self = rs.Self( rs.Context() + "/ref/toc/" + v.Toc )
  }
}

func (t *TocMap) MarshalJSON() ( []byte, error ) {
  // Tiploc sorted by NLC
  var vals []*Toc
  for _, v := range t.m {
    vals = append( vals, v )
  }

  sort.SliceStable( vals, func( i, j int ) bool {
    return strings.Compare( vals[i].Toc, vals[j].Toc ) < 0
  })

  b := &bytes.Buffer{}
  b.WriteByte( '{' )

  for i, v := range vals {
    if i > 0 {
      b.WriteByte( ',' )
    }
    b.WriteByte( '"' )
    b.WriteString( v.Toc )
    b.WriteByte( '"' )
    b.WriteByte( ':' )

    if eb, err := json.Marshal( v ); err != nil {
      return nil, err
    } else {
      b.Write( eb )
    }
  }

  b.WriteByte( '}' )
  return b.Bytes(), nil
}